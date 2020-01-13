package main

import (
	"bufio"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/radovskyb/watcher"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "soturon_maker",
		Short: "soturon_maker",
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	file    = ""
	absPath = ""
)

func Execute() error {
	return rootCmd.Execute()
}

func main() {
	run := &cobra.Command{
		Use:   "run",
		Short: "Run",
		RunE:  runCmd,
	}
	rootCmd.AddCommand(run)
	rootCmd.PersistentFlags().StringVar(&file, "file", "", "markdown file")
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func runCmd(c *cobra.Command, args []string) (err error) {
	w := watcher.New()
	go func() {
		for {
			select {
			case event := <-w.Event:
				fmt.Println(event)
				convert()
			case err := <-w.Error:
				log.Fatalln(err)
			case <-w.Closed:
				return
			}
		}
	}()

	// Get FileAbsPath and set watch list
	absPath, err = filepath.Abs(file)
	if err != nil {
		panic(err)
	}
	fmt.Println(absPath)
	if err := w.Add(absPath); err != nil {
		panic(err)
	}

	go func() {
		w.Wait()
		w.TriggerEvent(watcher.Create, nil)
		w.TriggerEvent(watcher.Remove, nil)
	}()

	// Start the watching process - it'll check for changes every 100ms.
	if err := w.Start(time.Millisecond * 100); err != nil {
		log.Fatalln(err)
	}

	return nil
}

func convert() {
	convertMdToLatex()
	insertTemplate()
	convertLatexToPdf()
}

func convertMdToLatex() {
	err := exec.Command("pandoc", "-r", "markdown-auto_identifiers", absPath, "-o", "tmp.tex").Run()
	if err != nil {
		panic(err)
	}
}

func insertTemplate() {
	var writer *bufio.Writer
	tpl := template.Must(template.ParseFiles("theis_template.tpl"))
	bytes, err := ioutil.ReadFile("tmp.tex")
	if err != nil {
		panic(err)
	}
	// output file
	outputFile, err := os.OpenFile(
		strings.Replace(
			filepath.Base(file), "md", "tex", 1,
		),
		os.O_WRONLY|os.O_CREATE,
		0600,
	)
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()
	writer = bufio.NewWriter(outputFile)
	err = tpl.Execute(writer, string(bytes))
	if err != nil {
		panic(err)
	}
	writer.Flush()
}

func convertLatexToPdf() {
	filename := strings.Replace(
		filepath.Base(file), "md", "tex", 1,
	)
	err := exec.Command(
		"scp", filename, "uaizu:~/",
	).Run()
	if err != nil {
		panic(err)
	}

	err = exec.Command(
		"ssh",
		"uaizu",
		"/usr/local/texlive/bin/platex",
		filename,
	).Run()
	if err != nil {
		panic(err)
	}

	err = exec.Command(
		"ssh",
		"uaizu",
		"/usr/local/texlive/bin/dvipdfmx",
		strings.Replace(filename, "tex", "dvi", 1),
	).Run()
	if err != nil {
		panic(err)
	}

	err = exec.Command(
		"scp",
		"uaizu:~/"+strings.Replace(filename, "tex", "pdf", 1),
		".",
	).Run()
	if err != nil {
		panic(err)
	}
	fmt.Println("create ", strings.Replace(filename, "tex", "pdf", 1))
}
