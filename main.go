package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func ShellProcess() error {

	wd, _ := os.Getwd()

	fmt.Print(wd + "> ")

	tempfile, err := os.CreateTemp("", "")

	if err != nil {
		return err
	}

	defer os.Remove(tempfile.Name())
	defer tempfile.Close()

	// fmt.Println("Created temp file: ", tempfile.Name())
	reader := bufio.NewReader(os.Stdin)
	commandRaw, err := reader.ReadString('\n')

	if err != nil {
		return err
	}

	command := strings.TrimRight(commandRaw, "\n")

	// fmt.Println("command", command)

	program := strings.Split(command, " ")

	// fmt.Println("argv", program)

	// fmt.Println("program", program[0])

	path, err := CommandExists(program[0])

	// fmt.Println("path", path)

	if err != nil {
		return err
	}

	proc, err := StartNewProcess(path, program, tempfile.Fd())

	if err != nil {
		return err
	}

	_, err = proc.Wait()

	if err != nil {
		return err
	}

	// println("output:")

	// Seek the pointer to the beginning
	tempfile.Seek(0, 0)
	s := bufio.NewScanner(tempfile)
	for s.Scan() {
		fmt.Println(s.Text())
		// println("exit code:", state.ExitCode())
	}
	if err = s.Err(); err != nil {
		log.Fatal("error reading temp file", err)
	}

	return nil

}

func main() {
	fmt.Println("welcome to shell")

	for {
		err := ShellProcess()

		if err != nil {
			fmt.Println("error: ", err)
		}
	}
}

func CommandExists(cmd string) (string, error) {
	return exec.LookPath(cmd)
}

func StartNewProcess(path string, argv []string, stdout uintptr) (*os.Process, error) {

	wd, err := os.Getwd()

	if err != nil {
		return nil, err
	}

	pid, err := syscall.ForkExec(path, argv, &syscall.ProcAttr{
		Dir:   wd,
		Env:   os.Environ(),
		Files: []uintptr{0, stdout, stdout},
	})

	if err != nil {

		if err != nil {
			return nil, err
		}
	}

	proc, err := os.FindProcess(pid)

	if err != nil {
		return nil, err
	}

	return proc, nil

}

func StartNewProcessDeprecated() {
	fmt.Println(os.Getpid())
	// This will not work as fork was built at a time where processes had single thread of execution.
	// Go lang uses mutiple threads per process so it really doesn't work well with this.
	// We can call fork and exec a new process to fix it and there is already a helper for that
	childId, returnCode, _ := syscall.Syscall(syscall.SYS_FORK, 0, 0, 0)
	if returnCode == 0 {
		fmt.Println("hellow from child process", childId)
	} else if returnCode > 0 {
		fmt.Println("hellow from parent process", os.Getpid())
	} else {
		fmt.Println("fork failed", os.Getpid())
	}
}
