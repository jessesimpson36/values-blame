package main

import (
	"fmt"
	chart "helm.sh/helm/v4/pkg/chart/v2"
	util "helm.sh/helm/v4/pkg/chart/v2/util"
	"os"
	"reflect"
	"strings"
)

type ValuesYAMLLine struct {
	FileName string `yaml:"fileName"`
	Option   string `yaml:"option"`
	Line     string `yaml:"line"`
}

type ValuesYAMLNode struct {
	Value *ValuesYAMLLine `yaml:"value"`
	Next  *ValuesYAMLNode `yaml:"next,omitempty"`
}

// traverse all options in the values yaml and determine if the new file sets them during coalescing
func traverseValues(valuesCoalesced map[string]interface{}, valuesApplied map[string]interface{}, fileNameApplied string, optionPrefix string, options *map[string]ValuesYAMLLine, node *ValuesYAMLNode) {
	for key, value := range valuesCoalesced {
		fullKey := optionPrefix + key
		valuesLine := ValuesYAMLLine{
			FileName: "blah",
			Option:   fullKey,
			Line:     "blah",
		}
		if node.Next == nil {
			node.Next = &ValuesYAMLNode{
				Value: &valuesLine,
				Next:  nil,
			}
		} else {
			if node.Next.Value.Option != fullKey {
				currentNext := node.Next
				node.Next = &ValuesYAMLNode{
					Value: &valuesLine,
					Next:  currentNext,
				}
			}
		}
		if subMap, ok := value.(map[string]interface{}); ok {
			appliedValue, ok := valuesApplied[key]
			if !ok {
				if node.Next != nil {
					node.Next = node.Next.Next
				}
				continue
			}
			if appliedSubMap, ok := appliedValue.(map[string]interface{}); ok {
				numberOfPeriods := strings.Count(fullKey, ".")
				valuesLine.FileName = fileNameApplied
				valuesLine.Line = fmt.Sprintf("%s%s", strings.Repeat("  ", numberOfPeriods), key)
				prev := node
				for node.Next != nil {
					if node.Next.Value.Option != fullKey {
						prev = node
						node = node.Next
					} else {
						break
					}
				}
				node = prev
				node.Next.Value = &valuesLine
				(*options)[optionPrefix + key] = valuesLine
				traverseValues(subMap, appliedSubMap, fileNameApplied, optionPrefix + key + ".", options, node.Next)
				for node.Next != nil {
					if strings.HasPrefix(node.Next.Value.Option, fullKey + ".") {
						node = node.Next
					} else {
						break
					}
				}
			}
		} else {
			appliedValue, ok := valuesApplied[key]
			if !ok {
				if node.Next != nil {
					node.Next = node.Next.Next
				}
				continue
			}
			if strings.HasPrefix(reflect.TypeOf(appliedValue).String(), "[]") || appliedValue == value {
				numberOfPeriods := strings.Count(fullKey, ".")
				valuesLine.FileName = fileNameApplied
				valuesLine.Line = fmt.Sprintf("%s%s: %v", strings.Repeat("  ", numberOfPeriods), key, value)
				node.Next.Value = &valuesLine
				node = node.Next
				(*options)[optionPrefix + key] = valuesLine
			}
		}
	}
}

func numberOfCharsInBiggestFileName(options map[string]ValuesYAMLLine) int {
	max := 0
	for _, option := range options {
		if len(option.FileName) > max {
			max = len(option.FileName)
		}
	}
	return max
}


func main() {

	// process values files via -f filename -f filename2
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go -f values.yaml -f values-2.yaml")
		return
	}

	valuesFilesList := []string{}
	for i := 1; i < len(os.Args); i++ {
		if strings.HasPrefix(os.Args[i], "-f") {
			continue
		}
		if strings.HasPrefix(os.Args[i], "--values") {
			continue
		}
		valuesFilesList = append(valuesFilesList, os.Args[i])
	}


	fmt.Printf("Processing values file: %s\n", valuesFilesList[0])
	v1, err := util.ReadValuesFile(valuesFilesList[0])
	if err != nil {
		fmt.Printf("Error reading values file %s: %v\n", valuesFilesList[0], err)
		return
	}

	head := &ValuesYAMLNode{}
	options := make(map[string]ValuesYAMLLine)
	valuesCoalesced := v1

	c := &chart.Chart{
		Metadata: &chart.Metadata{
			Name:    "moby",
			Version: "1.2.3",
		},
		Templates: []*chart.File{},
		Values: v1,
	}
	node := head
	traverseValues(valuesCoalesced, v1, valuesFilesList[0], "", &options, node)

	for _, valuesFile := range valuesFilesList[1:] {
		fmt.Printf("Processing values file: %s\n", valuesFile)
		v1, err := util.ReadValuesFile(valuesFile)
		if err != nil {
			fmt.Printf("Error reading values file %s: %v\n", valuesFile, err)
			return
		}
		valuesCoalesced, err = util.CoalesceValues(c, v1)
		if err != nil {
			fmt.Printf("Error coalescing values from %s: %v\n", valuesFile, err)
			return
		}
		c = &chart.Chart{
			Metadata: &chart.Metadata{
				Name:    "moby",
				Version: "1.2.3",
			},
			Templates: []*chart.File{},
			Values: valuesCoalesced,
		}
		node := head
		traverseValues(valuesCoalesced, v1, valuesFile, "", &options, node)
	}



	maxFileNameLength := numberOfCharsInBiggestFileName(options)



	current := head
	for current != nil {
		// Skip the head node
		if current == head {
			current = current.Next
			continue
		}
		fmt.Printf("%-*s :  %s\n", maxFileNameLength, options[current.Value.Option].FileName, options[current.Value.Option].Line)
		current = current.Next
	}

	strValues, err := valuesCoalesced.YAML()
	fmt.Printf("Coalesced Values:\n%+v\n", strValues)

}
