package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"time"
)

var (
	indexHtmlString    string
	indexHtmlBytes     []byte
	commandLineHistory []string
)

type templateInput struct {
	StandardOutput     string
	StandardError      string
	CommandLineHistory []string
	ExitError          error
}

type webPBTerminal struct{}

func getCommand(req *http.Request) string {
	return req.FormValue("commandlinecommandname")
}

func (w webPBTerminal) ServeHTTP(responseWriter http.ResponseWriter, req *http.Request) {
	req.ParseForm()

	latestCommand := getCommand(req)
	commandLineHistory = append(commandLineHistory, latestCommand)
	command := exec.Command("sh", "-c", latestCommand)
	stdout, err := command.StdoutPipe()
	if err != nil {
		fmt.Println("stdoutpipe", err)
		panic(err)
	}
	stderr, err := command.StderrPipe()
	if err != nil {
		fmt.Println("stderrpipe", err)
		panic(err)
	}
	if err = command.Start(); err != nil {
		fmt.Println("start", err)
		panic(err)
	}
	allstdout, err := ioutil.ReadAll(stdout)
	if err != nil {
		fmt.Println("readallstdout", err)
	}
	allstderr, err := ioutil.ReadAll(stderr)
	if err != nil {
		fmt.Println("readallstderr", err)
	}
	if err = command.Wait(); err != nil {
		fmt.Println("wait", err)
	}
	fmt.Println(string(allstdout), string(allstderr))

	ourTemplateInput := templateInput{}
	ourTemplateInput.StandardOutput = string(allstdout)
	ourTemplateInput.StandardError = string(allstderr)
	ourTemplateInput.CommandLineHistory = commandLineHistory
	tmpl, err := template.New("index").Parse(indexHtmlString)
	if err != nil {
		fmt.Println("template.New", err)
	}
	tmpl.Execute(responseWriter, ourTemplateInput)
	//template.New("index")
	//	io.WriteString(responseWriter)
}

func init() {
	var err error
	indexHtmlBytes, err = ioutil.ReadFile("index.html")
	indexHtmlString = string(indexHtmlBytes)
	if err != nil {
		panic(err)
	}

}

func main() {
	myHandler := webPBTerminal{}

	s := &http.Server{
		Addr:           "localhost:9600",
		Handler:        myHandler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 1,
	}
	log.Fatal(s.ListenAndServe())
}
