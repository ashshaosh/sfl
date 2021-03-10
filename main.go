package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/karrick/godirwalk"
	"gopkg.in/urfave/cli.v1"
)

var (
	args        = make([]string, 15, 15)
	currentPath = ""
	port        = ":8080"
	dir         http.Dir
	ssize       = false
	modt        = false
	full        = false
	er          error
	optQuiet    = false
	programName string
)

// Fily struct to hold single file
type Fily struct {
	filepath, filename string
	size               int64
	moddate            time.Time
	isdir              bool
}

func main() {
	// app params
	currentPath, err := os.Getwd()
	dir = http.Dir(currentPath)
	//file params

	//prefiks := http.StripPrefix("/static/", http.FileServer(dir))

	if err != nil {
		er = err
		log.Println(err)
	}

	//http.Handle("/static/", prefiks)

	app := cli.NewApp()
	app.Name = "sfl"
	app.Usage = "Serve any folder to local network"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "directory, d",
			Usage:       "Path to folder to serve to",
			EnvVar:      "",
			Hidden:      false,
			Value:       "./",
			Destination: new(string),
		},
		cli.StringFlag{
			Name:        "port, p",
			Usage:       "Port to open",
			EnvVar:      "",
			Hidden:      false,
			Value:       ":8080",
			Destination: new(string),
		},
		cli.BoolFlag{
			Name:        "size, s",
			Usage:       "Show the sizes of files",
			EnvVar:      "",
			Hidden:      false,
			Destination: new(bool),
		},
		cli.BoolFlag{
			Name:        "modified, m",
			Usage:       "Show the last modification time",
			EnvVar:      "",
			Hidden:      false,
			Destination: new(bool),
		},
		cli.BoolFlag{
			Name:        "full, f",
			Usage:       "Print full path",
			EnvVar:      "",
			Hidden:      false,
			Destination: new(bool),
		},
		cli.BoolFlag{
			Name:        "quiet, q",
			Usage:       "Elide printing of non-critical error messages.",
			EnvVar:      "",
			Hidden:      false,
			Destination: new(bool),
		},
	}
	app.Action = func(c *cli.Context) error {
		currentPath = c.GlobalString("directory")
		dir = http.Dir(currentPath)
		port = c.GlobalString("port")
		ssize = c.GlobalBool("size")
		full = c.GlobalBool("full")
		optQuiet = c.GlobalBool("quiet")
		//modt = c.GlobalString("modt")
		fmt.Printf("Path: %s, Port %s\n\n", dir, port)
		return nil
	}
	app.Run(os.Args)

	//diveIntoFolder(currentPath)
	//diveDirTree(currentPath)
	scanFolder(currentPath)

	if err != nil {
		log.Println(err)
	}

	http.ListenAndServe(port, http.FileServer(dir))
}

func diveDirTree(path string) {
	fmt.Println("\nfilepath.Walk------------------")
	args = nil
	er = filepath.Walk(path,
		func(path string, info os.FileInfo, errr error) error {
			if er != nil {
				return er
			}
			if info.IsDir() {
				args = append(args, "+")
			}
			if full != false {
				args = append(args, path)
			} else {
				args = append(args, info.Name())
			}
			if ssize != false && !info.IsDir() {
				args = append(args, strconv.FormatInt(info.Size()/1000, 10)+"Kb")
			}
			if modt != false {
				args = append(args, info.ModTime().Local().String())
			}

			args = append(args, "\n")
			return nil
		})

	printList(args)
	args = nil
}

func diveIntoFolder(path string) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		if full != false { // how to print path
			args = append(args, path+"/"+f.Name())
		} else {
			args = append(args, f.Name())
		}

		if ssize != false && f.IsDir() != true { // when to print sizes
			fl := float64(f.Size()) / 1000000
			sz := fmt.Sprintf("%.3f", fl)
			args = append(args, " "+sz+"Mb ")
		}
		if modt != false {
			args = append(args, " "+f.ModTime().Local().String()+" ")
		}

		if f.IsDir() != false {
			args = append([]string{"\n|><| "}, args...)
			args = append(args, "\n")
			diveIntoFolder(path + "/" + f.Name())
			args = append(args, "\n")
		} else {
			args = append(args, "\n")
		}
		printList(args)
		args = nil
	}
}

func getFInfo(f os.FileInfo) {
	if f.IsDir() == true {
		args = append(args, "\n|=| ")
		diveIntoFolder(currentPath + "/" + f.Name())
	}
}

func printList(args []string) {
	s := strings.Join(args, "")
	fmt.Printf(s)
}

func scanFolder(path string) {
	err := godirwalk.Walk(path, &godirwalk.Options{
		Callback: func(osPathname string, de *godirwalk.Dirent) error {
			fmt.Printf("%s %s\n", de.ModeType(), osPathname)
			return nil
		},
		ErrorCallback: func(osPathname string, err error) godirwalk.ErrorAction {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
			// For the purposes of this example, a simple SkipNode will suffice,
			// although in reality perhaps additional logic might be called for.
			return godirwalk.SkipNode
		},
		Unsorted: true, // set true for faster yet non-deterministic enumeration (see godoc)
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func stderr(f string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, programName+": "+fmt.Sprintf(f, args...)+"\n")
}

func fatal(f string, args ...interface{}) {
	stderr(f, args...)
	os.Exit(1)
}

func warning(f string, args ...interface{}) {
	if !optQuiet {
		stderr(f, args...)
	}
}
