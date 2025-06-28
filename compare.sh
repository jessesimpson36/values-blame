#!/bin/bash

## print with all args
./values-blame $@ -n > test.txt
sed -i 's/\(\[.*\]\)/"\1"/' test.txt
yq 'sortKeys(..)' test.txt > test.yaml

./values-blame $@ -c > test2.txt
yq 'sortKeys(..)' test2.txt > test2.yaml

nvim -d test.yaml test2.yaml

#rm test.txt test.yaml test2.txt test2.yaml
