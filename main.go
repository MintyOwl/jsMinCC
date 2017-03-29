package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/MintyOwl/closureCompiler"

	es "github.com/MintyOwl/elementScraper"
)

func osArgs(exec string) string {
	i := strings.LastIndex(exec, "/")
	if i > 0 {
		return exec[i+1:]
	}
	i = strings.LastIndex(exec, `\`)
	if i > 0 {
		return exec[i+1:]
	}
	return exec
}

func usage() {
	fmt.Println("\n  USAGE ")
	fmt.Printf("  Example 1 %v -ua=\"Custom User Agent Goes Here\" ./mySourceFile.js ", osArgs(os.Args[0]))
	fmt.Printf("\n  Example 2 %v ./mySourceFile.js ../minifiedDestinationFile.js ", osArgs(os.Args[0]))
	fmt.Printf("\n  Example 3 %v ./mySourceFile.js  ", osArgs(os.Args[0]))
	fmt.Println("\n  Destination file is optional. When not provided, it creates file with the source name ending with cc.js extension ")
	fmt.Printf("\n  Example 4 %v -html ./myHTMLSourceFile.html  ", osArgs(os.Args[0]))
	fmt.Println("\n  If you have HTML file containing inline scripts then use '-html' option as shown above ")
	fmt.Println("\n  When using HTML file, '-html' flag is a must, otherwise its considered a js file ")
	fmt.Println("  Destination file is optional as well in this case \n")
	fmt.Println("  IMPORTANT ")
	fmt.Println("  flags must be provided in the beginning before other arguments \n")
}

var cce *closureCompiler.CCEval

func handleFileErr(err error, path string) {

	if err != nil {
		absPath, e := filepath.Abs(path)
		if e == nil {
			fmt.Printf("Could not open file %v. Error is %v", absPath, err)
			os.Exit(2)
		}
		os.Exit(1)
	}
	return
}

func handlFprintErr(err error) {
	if err != nil {
		fmt.Println("Error while Fprintf ", err)
		os.Exit(2)
	}
	return
}

func ccScrapeImpl(raw string) string {
	cce := closureCompiler.NewCCEval(raw, "")
	s, err := cce.Run()
	if err != nil {
		fmt.Println("Error while minifying js code via ClosureCompiler >> ", err)
		return raw
	}
	return s
}

func addendum(path string) {
	absPath, err := filepath.Abs(path)
	if err == nil {
		fmt.Printf("Output available at %v", absPath)
		os.Exit(0)
	}
	os.Exit(1)
}

func run(ua *string, html *bool, opts []string) {
	if args := len(opts); args > 0 {
		path, err := filepath.Abs(opts[0])
		f, err := os.Open(path)
		handleFileErr(err, path)
		b, err := ioutil.ReadAll(f)
		if err != nil {
			fmt.Printf("Could not read file %v. Ended with an error %v", path, err)
			os.Exit(1)
		}
		var o string

		if *html == false {
			cce = closureCompiler.NewCCEval(string(b), *ua)
			o, err = cce.Run()
		}

		if err == nil {
			if args == 2 {
				fl := opts[1]
				f, err := os.OpenFile(fl, os.O_CREATE, 0644)
				handleFileErr(err, fl)
				if *html == false {
					_, err = fmt.Fprintf(f, o)
					handlFprintErr(err)
					addendum(fl)
				}

				elmScrpr := es.NewElementScraper(string(b), `<script>`).ElementFunc(ccScrapeImpl)

				_, err = fmt.Fprintf(f, elmScrpr.Run())
				handlFprintErr(err)
				addendum(fl)
			}
			if *html == false {
				newPath := path + "cc.js"
				f, err := os.OpenFile(newPath, os.O_CREATE, 0644)
				handleFileErr(err, newPath)
				_, err = fmt.Fprintf(f, o)
				handlFprintErr(err)
				addendum(newPath)
			}
			newPath := path + "cc.html"
			f, err := os.OpenFile(newPath, os.O_CREATE, 0644)
			handleFileErr(err, newPath)

			elmScrpr := es.NewElementScraper(string(b), `<script>`).ElementFunc(ccScrapeImpl)

			_, err = fmt.Fprintf(f, elmScrpr.Run())
			handlFprintErr(err)
			addendum(newPath)

		}
		fmt.Println(err)
		os.Exit(1)
	}
	usage()
}

func main() {
	ua := flag.String("ua", "", "User Agent for Closure Compiler")
	html := flag.Bool("html", false, "Enter path to your HTML file containing inline scripts")
	flag.Parse()
	opts := flag.Args()
	run(ua, html, opts)
}
