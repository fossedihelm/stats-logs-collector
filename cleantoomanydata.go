package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

func main() {
	// user your path
	csvFile, err := os.Open("/home/fossedihelm/memstat.csv")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Successfully Opened CSV file")
	defer csvFile.Close()
	reader := csv.NewReader(csvFile)
	//reader.Comma = '\t'
	csvLines, err := reader.ReadAll()
	if err != nil {
		fmt.Println(err)
	}
	// user your path
	csvPath := filepath.Join("/home/fossedihelm/", "cleaned.csv")
	f, err := os.Create(csvPath)
	if err != nil {
		fmt.Println("Error")
		os.Exit(1)
	}
	defer csvFile.Close()
	writer := csv.NewWriter(f)
	defer writer.Flush()

	filemap := make(map[string]string, 7)
	for index, line := range csvLines {
		if index < 65882 {
			if filemap[line[0]] == "" {
				filemap[line[0]] = line[1]
				fmt.Println(filemap)
				writer.Write(line)
				continue
			}
			curTs, _ := strconv.ParseInt(line[1], 10, 64)
			prevTs, _ := strconv.ParseInt(filemap[line[0]], 10, 64)

			if prevTs+300 < curTs {
				filemap[line[0]] = line[1]
				writer.Write(line)
				continue
			}
		} else {
			writer.Write(line)
		}
	}
}
