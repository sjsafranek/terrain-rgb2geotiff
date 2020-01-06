package main

import (
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

// createTempFile
func createTempFile(directory, filename string, content []byte) (*os.File, error) {
	// create temp file
	file, err := ioutil.TempFile(directory, filename)
	if nil != err {
		// cleanup bad file
		defer os.Remove(file.Name())
		return nil, err
	}

	// write file contents
	if _, err := file.Write(content); err != nil {
		return nil, err
	}

	// close file
	if err := file.Close(); err != nil {
		return nil, err
	}

	return file, nil
}

// executeScript executes shell command with supplied arguments.
func executeScript(script string, args ...string) (string, error) {
	cmd := exec.Command(script, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if nil != err {
		return "", errors.New(stderr.String())
	}
	return stdout.String(), nil
}

// createAndExecuteScript
func createAndExecuteScript(directory, filename, content string) error {

	log.Println(content)

	file, err := createTempFile(directory, filename, []byte(content))
	if nil != err {
		return err
	}

	results, err := executeScript("/bin/sh", file.Name())

	log.Println(err)
	log.Println(results)

	return err
}
