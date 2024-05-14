package cmd

import (
	"os"
	"os/exec"
	"sort"

	"github.com/fsnotify/fsnotify"
)

func MonitorDir(path string, maxTails int){
	files, err := os.ReadDir(path)
	if err != nil{
		println("Unable to read directory, exiting.")
		os.Exit(1)
	}

	queue := make(chan *os.Process, maxTails)
	counter := 0
	filesForMonitoring := getNewestExistingFiles(files, maxTails)

	for _,f := range filesForMonitoring{
		if !f.IsDir() {
			finfo,err := f.Info()
			if err != nil {
				
				println("Unable to read file %s , skipping", f.Name())
				continue
			}
			queue <- tailFile(finfo.Name())
			counter++
		}
	}

	defer killAllTails(queue)
	monitorDirectory(path, queue, counter, maxTails)

}

func monitorDirectory(path string, tails chan *os.Process, counter int, maxTails int){

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()
	err = watcher.Add(path)
	if err != nil {
		println("Unable to add watcher")
		os.Exit(1)
	}
	// We should never leave this function unless the program ends
	for {
		select{
		case err, ok := <-watcher.Errors:
			if !ok { // Channel was closed (i.e. Watcher.Close() was called).
				return
			}
			println("ERROR: %s", err)

		case event,ok := <-watcher.Events:
			if !ok{
				println("event was not ok") 
				return
			}

			if event.Has(fsnotify.Create){
				counter = newFileCreated(event, counter, maxTails, tails)
			}
		}	
	}
}

func newFileCreated(event fsnotify.Event, counter int, maxTails int, tails chan *os.Process) int{

	f,err := os.Open(event.Name)
	defer f.Close()
	if err != nil {
		println("Newly created file doesnt exist anymore")
		return counter
	}
	fi,_ := f.Stat()

	if fi.IsDir(){
		return counter
	}
	if counter == maxTails{
		process := <- tails
		process.Kill()
		counter--
	}
	tails <- tailFile(event.Name)
	counter++
	return counter
}

func getNewestExistingFiles(files []os.DirEntry, maxLength int) []os.DirEntry{
	files = removeDirs(files)
	files = sortFilesByModTime(files)

	sliceSize := getSmallestInt(len(files), maxLength)
	filesSlice := files[len(files)-sliceSize:]

	return filesSlice
}

func sortFilesByModTime(files []os.DirEntry) []os.DirEntry{
	sort.Slice(files, func(i,j int) bool{
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

func tailFile(filePath string) *os.Process {

	app := "tail"
	args := "-f" 

	cmd := exec.Command(app, args, filePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()
	return cmd.Process
}
func killAllTails(queue chan *os.Process){

	for p := range queue{
		p.Kill()
	}
}

func removeDirs(s []os.DirEntry) []os.DirEntry{
	var dirless []os.DirEntry
	for _,file := range s{
		if !file.IsDir() {
			dirless = append(dirless, file)
		}
	}
	return dirless
}

func getSmallestInt(a int, b int) int {
	if a > b{ return b}
	return a
}