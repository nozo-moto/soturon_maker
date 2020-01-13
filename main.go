package main

import (
	"fmt"
	"log"
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
	err := exec.Command("pandoc", absPath, "-o", strings.Replace(filepath.Base(file), "md", "tex", 1)).Run()
	if err != nil {
		panic(err)
	}
}
