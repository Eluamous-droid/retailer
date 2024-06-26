package main

import (
	"os"
	"sort"
	"strings"
)

func getFilesForMonitoring(files []os.DirEntry, maxLength int) []os.DirEntry {
	files = removeIneligibleFiles(files)
	files = sortFilesByModTime(files)

	sliceSize := getSmallestInt(len(files), maxLength)
	filesSlice := files[len(files)-sliceSize:]

	return filesSlice
}

func sortFilesByModTime(files []os.DirEntry) []os.DirEntry {
	sort.Slice(files, func(i, j int) bool {
		fileI, err := files[i].Info()
		if err != nil {
			println("Unable to read file %s , while sorting", fileI.Name())
			os.Exit(1)
		}
		fileJ, err := files[j].Info()
		if err != nil {
			println("Unable to read file %s , while sorting", fileJ.Name())
			os.Exit(1)
		}
		return fileI.ModTime().Before(fileJ.ModTime())

	})
	return files
}

func removeIneligibleFiles(s []os.DirEntry) []os.DirEntry {
	var eligibleFiles []os.DirEntry
	for _, file := range s {
		fi, err := file.Info()
		if err != nil {
			println("File is unreadable: ", file.Name())
		}
		if isEligibleFile(fi) {
			eligibleFiles = append(eligibleFiles, file)
		}
	}
	return eligibleFiles
}

func isEligibleFile(fi os.FileInfo) bool {
	if fi.IsDir() {
		return false
	}

	fn := fi.Name()
	for _, s := range excludedFiles {
		if strings.Contains(fn, s) {
			return false
		}
	}

	return true
}

func getSmallestInt(a int, b int) int {
	if a > b {
		return b
	}
	return a
}
