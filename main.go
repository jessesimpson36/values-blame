package main

import (
	"fmt"
	chart "helm.sh/helm/v4/pkg/chart/v2"
	util "helm.sh/helm/v4/pkg/chart/v2/util"
	"os"
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

func numberOfCharsInBiggestFileName(options map[string]ValuesYAMLLine) int {
	max := 0
	for _, option := range options {
		if len(option.FileName) > max {
			max = len(option.FileName)
		}
	}
	return max
}

func isOptionWithinAppliedValues(option string, valuesApplied map[string]interface{}) bool {
	optionArray := strings.Split(option, ".")
	current := valuesApplied
	for i, valueKey := range optionArray {
		currentValue, ok := current[valueKey]
		if !ok {
			return false
		}
		if current, ok = currentValue.(map[string]interface{}); !ok {
			// if current is a primitive, and is the last option, then the option is within the applied values
			return i >= len(optionArray)-1
		}
	}
	return true
}

func getListOfKeys(values map[string]interface{}) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
		if subMap, ok := values[key].(map[string]interface{}); ok {
			subKeys := getListOfKeys(subMap)
			for _, subKey := range subKeys {
				keys = append(keys, key+"."+subKey)
			}
		}
	}
	return keys
}

func getValueAtKey(values map[string]interface{}, key string) interface{} {
	keys := strings.Split(key, ".")
	current := values
	for _, k := range keys {
		if value, ok := current[k]; ok {
			if subMap, ok := value.(map[string]interface{}); ok {
				current = subMap
			} else {
				return value
			}
		} else {
			return nil
		}
	}
	return ""
}

func traverseValues(valuesCoalesced map[string]interface{}, valuesApplied map[string]interface{}, fileNameApplied string, options *map[string]ValuesYAMLLine) {

	keys := getListOfKeys(valuesCoalesced)
	for _, key := range keys {
		if isOptionWithinAppliedValues(key, valuesApplied) {
			lastKey := key
			lastDotIndex := strings.LastIndex(key, ".")
			if lastDotIndex != -1 {
				lastKey = key[lastDotIndex+1:] // Get the last part of the key after the last dot
			}
			line := ValuesYAMLLine{
				FileName: fileNameApplied,
				Option:   key,
				Line:     fmt.Sprintf("%s: %v", lastKey, getValueAtKey(valuesCoalesced, key)),
			}
			(*options)[key] = line
		}
	}
}

func traverseCoalesceAndPrintFile(valuesCoalesced map[string]interface{}, options *map[string]ValuesYAMLLine, prefix string, maxFileNameLength int, indentDepth int, dontPrintFileNames bool) {
	for key, value := range valuesCoalesced {
		if subMap, ok := value.(map[string]interface{}); ok {
			checkedPrefix := ""
			if prefix != "" {
				checkedPrefix = prefix + "."
			}
			line := (*options)[checkedPrefix+key]

			indent := strings.Repeat("  ", indentDepth)
			if dontPrintFileNames {
				fmt.Printf("%s%s\n", indent, line.Line)
			} else {
				fmt.Printf("%-*s %s%s\n", maxFileNameLength, line.FileName, indent, line.Line)
			}

			traverseCoalesceAndPrintFile(subMap, options, checkedPrefix+key, maxFileNameLength, indentDepth+1, dontPrintFileNames)
		} else {
			checkedPrefix := ""
			if prefix != "" {
				checkedPrefix = prefix + "."
			}
			line := (*options)[checkedPrefix+key]
			indent := strings.Repeat("  ", indentDepth)
			if dontPrintFileNames {
				fmt.Printf("%s%s\n", indent, line.Line)
			} else {
				fmt.Printf("%-*s %s%s\n", maxFileNameLength, line.FileName, indent, line.Line)
			}
		}
	}

}

func main() {

	onlyPrintCoalesced := false
	dontPrintFileNames := false

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
		if strings.HasPrefix(os.Args[i], "-c") {
			onlyPrintCoalesced = true
			continue
		}
		if strings.HasPrefix(os.Args[i], "-n") {
			dontPrintFileNames = true
			continue
		}

		valuesFilesList = append(valuesFilesList, os.Args[i])
	}


	v1, err := util.ReadValuesFile(valuesFilesList[0])
	if err != nil {
		fmt.Printf("Error reading values file %s: %v\n", valuesFilesList[0], err)
		return
	}

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
	traverseValues(valuesCoalesced, v1, valuesFilesList[0], &options)

	for _, valuesFile := range valuesFilesList[1:] {
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
		traverseValues(valuesCoalesced, v1, valuesFile, &options)
	}



	maxFileNameLength := numberOfCharsInBiggestFileName(options)
	if !onlyPrintCoalesced {
		traverseCoalesceAndPrintFile(valuesCoalesced, &options, "", maxFileNameLength, 0, dontPrintFileNames)
	}

	strValues, err := valuesCoalesced.YAML()
	if onlyPrintCoalesced {
		fmt.Printf("Coalesced Values:\n%+v\n", strValues)
	}

}
