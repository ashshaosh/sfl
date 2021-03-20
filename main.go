package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
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
	dir               http.Dir
	er                error
	programName       string
	tmplt             *template.Template
	wlcm, upld, cntnt template.HTML

	args        = make([]string, 15, 15)
	currentPath = ""
	port        = ":8080"
	ssize       = false
	modt        = false
	full        = false
	optQuiet    = false
)

//Page struct to describe page template

type Page struct {
	Title         string
	Welcome       template.HTML
	UploadForm    template.HTML
	ContentFolder template.HTML
}

//Article struct for article inside page
type Welcome struct {
	Title, Text string
}
type Upload struct {
}

// Fily struct to hold single file
type Fily struct {
	filepath, filename string
	size               int64
	moddate            time.Time
	isdir              bool
}

func init() { // get template to know html files needed
	tmplt = template.Must(template.ParseFiles("./templates/index.html", "./templates/welcome.html", "./templates/form_upload.html"))
}

func main() {
	// app params
	cPath, err := os.Getwd()
	currentPath = cPath
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
			Usage:       "Show the Sizes of files",
			EnvVar:      "",
			Hidden:      false,
			Destination: new(bool),
		},
		cli.BoolFlag{
			Name:        "modified, m",
			Usage:       "Show the last Modification time",
			EnvVar:      "",
			Hidden:      false,
			Destination: new(bool),
		},
		cli.BoolFlag{
			Name:        "full, f",
			Usage:       "Print Full path to files",
			EnvVar:      "",
			Hidden:      false,
			Destination: new(bool),
		},
		cli.BoolFlag{
			Name:        "quiet, q",
			Usage:       "ElideQ printing of non-critical error messages.",
			EnvVar:      "",
			Hidden:      false,
			Destination: new(bool),
		},
		cli.BoolFlag{
			Name:        "upload, u",
			Usage:       "Enable Upload to folder.",
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
		optQuiet = c.GlobalBool("upload")
		//modt = c.GlobalString("modt")
		fmt.Printf("Path: %s, Port %s\n\n", dir, port)
		return nil
	}
	app.Run(os.Args)

	//diveIntoFolder(currentPath)
	diveDirTree(currentPath)
	//scanFolder(currentPath)

	if err != nil {
		log.Println(err)
	}

	intro := &Welcome{
		Title: `Welcome lol`,
		Text:  `Welcome Welcome Welcome Welcome Welcome Welcome Welcome Welcome Welcome Welcome Welcome Welcome Welcome Welcome Welcome `,
	}

	form := &Upload{}

	var b bytes.Buffer
	//
	tmplt.ExecuteTemplate(&b, "welcome.html", intro)    // parse var with template for article into buffer
	wlcm = template.HTML(b.String())                    // fill var with string from buffer
	b.Reset()                                           // clear buffer
	tmplt.ExecuteTemplate(&b, "form_upload.html", form) // parse into buffer
	upld = template.HTML(b.String())                    // fill var ...
	b.Reset()

	http.HandleFunc("/", fileForm)
	http.ListenAndServe(port, nil)
}

func diveDirTree(path string) {
	fmt.Println("\nfilepath.Walk------------------")
	args = nil
	er = filepath.Walk(path,
		func(path string, info os.FileInfo, errr error) error {
			if er != nil {
				return er
			}
			if info.IsDir() == true {
				args = append(args, "+")
			}
			if full != false {
				args = append(args, path)
			} else {
				args = append(args, info.Name())
			}
			if ssize != false && info.IsDir() == false {
				args = append(args, strconv.FormatInt(info.Size()/1000, 10)+"Kb")
			}
			// if modt != false {
			// 	args = append(args, info.ModTime().Local().String())
			// }

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

func scanFolder(path string) {
	err := godirwalk.Walk(path, &godirwalk.Options{
		Callback: func(osPathname string, de *godirwalk.Dirent) error {
			fmt.Printf("%s %s\n", de.ModeType(), osPathname)
			return nil
		},
		ErrorCallback: func(osPathname string, err error) godirwalk.ErrorAction {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
			return godirwalk.SkipNode // TODO: hold error
		},
		Unsorted: true, // set true for faster yet non-deterministic enumeration (see godoc)
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
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

func fileForm(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		p := &Page{ // make var to hold whole page content
			Title:         "Lol upload", // assign page itle
			Welcome:       wlcm,         // post var to parameter
			UploadForm:    upld,
			ContentFolder: wlcm,
		}
		tmplt.ExecuteTemplate(w, "index.html", p)
		// t, _ := template.ParseFiles("templates/form_upload.html")
		// t.Execute(w, nil)
		fmt.Printf("Upload to: %s\n", currentPath)
	} else {
		mr, err := r.MultipartReader()
		values := make(map[string][]string)
		if err != nil {
			panic("Failed to read multipart message: ")
		}

		//length := r.ContentLength
		maxValueBytes := int64(10 << 20)
		//fmt.Printf("%f", float64(maxValueBytes))
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			name := part.FormName()
			if name == "" {
				continue
			}
			var b bytes.Buffer
			filename := part.FileName()

			if filename == "" {
				n, err := io.CopyN(&b, part, maxValueBytes)
				if err != nil && err != io.EOF {
					fmt.Fprint(w, "Error processing form")
					return
				}
				maxValueBytes -= n
				if maxValueBytes == 0 {
					fmt.Fprint(w, "multipart message too large")
					return
				}
				values[name] = append(values[name], b.String())
				continue
			}

			now := time.Now().Format("(Jan _2 15-04-05)-")
			fileName := currentPath + "/" + now + filename
			dst, err := os.Create(fileName)
			defer dst.Close()
			if err != nil {
				return
			}
			for {
				buffer := make([]byte, 999999)
				cBytes, err := part.Read(buffer)
				if err == io.EOF {
					break
				}
				dst.Write(buffer[0:cBytes])
			}
		}

		fmt.Println("Upload done")
		fmt.Println(values)
		fmt.Fprint(w, "Upload complete")
	}
}

//wlcm, upld, cntnt
func displayPage(rw http.ResponseWriter, r *http.Request) {

	p := &Page{ // make var to hold whole page content
		Title:         "Lol upload", // assign page itle
		Welcome:       wlcm,         // post var to parameter
		UploadForm:    upld,
		ContentFolder: wlcm,
	}
	tmplt.ExecuteTemplate(rw, "index.html", p) // parse var with index.html template
}
