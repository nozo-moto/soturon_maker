package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/joho/godotenv"
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
	output  = ""
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
	rootCmd.PersistentFlags().StringVar(&output, "output", "", "output file")
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

	// https://github.com/jgm/pandoc/issues/1023
	// https://github.com/jgm/pandoc/issues/1023#issuecomment-83501964
	// result.gsub!('\begin{longtable}', '\begin{center}\begin{supertabular}')
	// result.gsub!('\endhead', '')
	// result.gsub!('\end{longtable}', '\end{supertabular}\end{center}')

	data, err := ioutil.ReadFile("tmp.tex")
	if err != nil {
		panic(err)
	}

	result := strings.Replace(
		strings.Replace(
			strings.Replace(
				strings.Replace(
					strings.Replace(
						strings.Replace(
							strings.Replace(
								strings.Replace(
									strings.Replace(
										strings.Replace(
											strings.Replace(
												string(data),
												"\\begin{longtable}",
												"\\begin{center}\\begin{supertabular}",
												-1,
											),
											"\\endhead",
											"",
											-1,
										),
										"\\end{longtable}",
										"\\end{supertabular}\\end{center}",
										-1,
									),
									"0.27\\columnwidth",
									"0.50\\columnwidth",
									-1,
								),
								"0.36\\columnwidth",
								"0.36\\columnwidth",
								-1,
							),
							"\\end{supertabular}\\end{center}",
							"\\end{supertabular}\\end{center}\n\\end{table*}",
							-1,
						),
						"\\begin{center}\\begin{supertabular}[]{@{}llll@{}}",
						"\\begin{table*}[!h]\n\\begin{center}\\begin{supertabular}[]{@{}llll@{}}",
						-1,
					),
					"0.06",
					"0.24",
					-1,
				),
				"\\tightlist",
				"",
				-1,
			),
			"\\includegraphics",
			"\\includegraphics[clip,keepaspectratio, width = 8.5cm]",
			-1,
		),
		"\\begin{figure}",
		"\\begin{figure}[h]",
		-1,
	)
	err = ioutil.WriteFile("tmp.tex", []byte(result), 0666)
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
	type Theis struct {
		Author     string
		StudentId  string
		Supervisor string
		Title      string
		Body       string
	}
	err = godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}
	theis := Theis{
		Author:     os.Getenv("AUTHOR"),
		StudentId:  os.Getenv("STUDENTID"),
		Supervisor: os.Getenv("SUPERVISOR"),
		Title:      os.Getenv("TITLE"),
		Body:       string(bytes),
	}
	err = tpl.Execute(writer, theis)
	if err != nil {
		panic(err)
	}
	writer.Flush()
}

func convertLatexToPdf() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	host := os.Getenv("HOST")
	filename := strings.Replace(
		filepath.Base(file), "md", "tex", 1,
	)
	err = exec.Command(
		"scp",
		"-r",
		filename,
		"bibfile.bib",
		"image",
		host+":~/",
	).Run()
	if err != nil {
		panic(err)
	}
	err = exec.Command(
		"ssh",
		host,
		"/usr/local/texlive/bin/platex",
		filename,
	).Run()
	if err != nil {
		panic(err)
	}
	err = exec.Command(
		"ssh",
		host,
		"/usr/local/texlive/bin/pbibtex",
		filename,
	).Run()
	if err != nil {
		panic(err)
	}

	err = exec.Command(
		"ssh",
		host,
		"/usr/local/texlive/bin/dvipdfmx",
		strings.Replace(filename, "tex", "dvi", 1),
	).Run()
	if err != nil {
		panic(err)
	}

	if output == "" {
		output = strings.Replace(filename, "tex", "pdf", 1)
	}
	err = exec.Command(
		"scp",
		host+":~/"+strings.Replace(filename, "tex", "pdf", 1),
		"./"+output,
	).Run()
	if err != nil {
		panic(err)
	}
	fmt.Println("create ", output)
}
