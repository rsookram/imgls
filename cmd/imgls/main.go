package main

import (
	"flag"
	"fmt"
	"image"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

type imageMeta struct {
	format   string
	width    int
	height   int
	fileSize int64
	fileName string
}

func main() {
	flag.Parse()

	fileNames := flag.Args()
	if len(fileNames) == 0 {
		// If no filenames are provided, use the current directory
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Couldn't get current directory. "+err.Error())
			os.Exit(1)
		}

		fileNames = []string{cwd}
	}

	metas := getImageMetas(collectFileNames(fileNames))
	printOutput(metas)
}

// Checks if input files array contain directories and adds it's contents to
// the file list if so. Otherwise just adds a file to the list.
func collectFileNames(inputFileNames []string) []string {
	fileNames := make([]string, 0)

	for _, inputFileName := range inputFileNames {
		info, err := os.Stat(inputFileName)
		if err != nil {
			continue
		}

		if !info.IsDir() {
			fileNames = append(fileNames, inputFileName)
			continue
		}

		files, err := ioutil.ReadDir(inputFileName)
		if err != nil {
			continue
		}

		dirPath := filepath.Clean(inputFileName)
		for _, fileInDir := range files {
			if !fileInDir.IsDir() {
				fileNames = append(fileNames, dirPath+string(os.PathSeparator)+fileInDir.Name())
			}
		}
	}

	return fileNames
}

func getImageMetas(filenames []string) []imageMeta {
	metas := make([]imageMeta, 0)

	for _, fileName := range filenames {
		config, format, fileSize, err := decode(fileName)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to decode "+fileName+". "+err.Error())
			continue
		}

		meta := imageMeta{
			format:   format,
			width:    config.Width,
			height:   config.Height,
			fileSize: fileSize,
			fileName: fileName,
		}

		metas = append(metas, meta)
	}

	return metas
}

func decode(fileName string) (*image.Config, string, int64, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, "", -1, err
	}

	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, "", -1, err
	}

	fileSize := fileInfo.Size()

	config, format, err := image.DecodeConfig(file)
	return &config, format, fileSize, err
}

func printOutput(metas []imageMeta) {
	columns := []string{"NUM", "FORMAT", "WIDTH", "HEIGHT", "SIZE", "FILENAME"}
	fmt.Println(strings.Join(columns, "\t"))

	for i, meta := range metas {
		fmt.Printf(
			"%d\t%s\t%d\t%d\t%s\t%s\n",
			i+1, meta.format, meta.width, meta.height, humanReadableFileSize(meta.fileSize), meta.fileName,
		)
	}
}

func humanReadableFileSize(fileSize int64) string {
	switch {
	case fileSize < 1e3:
		return strconv.FormatInt(fileSize, 10)
	case fileSize < 1e6:
		return strconv.FormatInt(fileSize/1e3, 10) + "K"
	case fileSize < 1e9:
		return strconv.FormatInt(fileSize/1e6, 10) + "M"
	default:
		return strconv.FormatInt(fileSize/1e9, 10) + "G"
	}
}
