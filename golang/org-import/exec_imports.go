package main

import (
	"bytes"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"

	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

/* global variable declaration */
var sess *session.Session
var sessionerr error

const (
	scpFilename         string = "imports_policies.tf"
	rootFilename        string = "imports_roots.tf"
	ouFilename          string = "imports_organizations.tf"
	acFilename          string = "imports_accounts.tf"
	attachmentsFilename string = "imports_attachments.tf"
)

// initiliaser for global variables
func init() {
	sess, sessionerr = session.NewSession(&aws.Config{
		Region: aws.String("ap-southeast-2")},
	)
	errCheck("AWS Session Error: ", sessionerr)
}

func main() {
	fmt.Println("<<<<<<<<<<<<<<<SCP>>>>>>>>>>>>>>>>>")
	importScps()
	fmt.Println("<<<<<<<<<<<<<<<ORG>>>>>>>>>>>>>>>>>")
	rootId := importRoot()
	fmt.Println("<<<<<<<<<<<<<<<OUS>>>>>>>>>>>>>>>>>")
	importOUs(rootId)
	fmt.Println("<<<<<<<<<<<<<<<ACC>>>>>>>>>>>>>>>>>")
	importAccounts()
	CmdExec("terraform", "fmt")
}

// CmdExec Execute bash command
// ex: CmdExec("aws", "s3", "ls")
func CmdExec(args ...string) (string, string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	baseCmd := args[0]
	cmdArgs := args[1:]

	fmt.Println("Shell Execution : %v", args)

	cmd := exec.Command(baseCmd, cmdArgs...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Shell Execution Error Encountered %v\n", err)
	}
	if stdout.String() != "" {
		fmt.Println("---------------STDOUT-----------------")
		fmt.Print(stdout.String())
	}
	if stderr.String() != "" {
		fmt.Println("---------------STDERR-----------------")
		fmt.Print(stderr.String())
	}
	return stdout.String(), stderr.String(), err
}

func ioErrCheck(msg string, e error) {
	if e != nil {
		fmt.Println(msg, e)
	}
}

func errCheck(msg string, e error) {
	if e != nil {
		log.Fatal(msg, e)
	}
}

// DumpTextResult from string
func DumpTextResult(str string) []string {
	scanner := bufio.NewScanner(strings.NewReader(str))
	// Set the split function for the scanning operation.
	scanner.Split(bufio.ScanWords)
	// Count the result set.
	count := 0
	for scanner.Scan() {
		count++
	}
	fmt.Printf("Result Set Count = %d\n", count)
	scanner = bufio.NewScanner(strings.NewReader(str))
	result := make([]string, count)
	i := 0
	for scanner.Scan() {
		if scanner.Text() != "" {
			result[i] = scanner.Text()
			i++
		}
	}
	return result
}

// AppendToFile append a line to file
func AppendToFile(filename string, strText string) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	defer f.Close()
	ioErrCheck("AppendToFile Error", err)
	w := bufio.NewWriter(f)
	n, err := w.WriteString(fmt.Sprintf("[%s]\n", strText))
	ioErrCheck("AppendToFile Error", err)
	fmt.Printf("wrote %d bytes\n", n)
	w.Flush()
}
