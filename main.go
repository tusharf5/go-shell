package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

var wdRaw string

func ShellProcess() error {

	homedir, _ := os.UserHomeDir()

	wd := strings.Replace(wdRaw, homedir, "~", 1)

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

	commandRaw = strings.Trim(strings.TrimSuffix(commandRaw, "\n"), " ")

	command := Interpolate(commandRaw)

	// fmt.Println("command", command)

	args := strings.Split(command, " ")

	program := args[0]

	if program == "cd" {

		if len(args) == 1 {
			wdRaw = homedir
			return nil
		}

		if len(args) == 2 {

			if args[1] == "." {
				return nil
			} else if args[1] == ".." {
				lastIndex := strings.LastIndex(wdRaw, "/")
				wdRaw = wdRaw[:lastIndex]
				return nil
			} else {

				entries, err := os.ReadDir(wdRaw)

				if err != nil {
					return nil
				}

				for _, e := range entries {
					if e.IsDir() && args[1] == e.Name() {
						wdRaw = wdRaw + "/" + e.Name()
						return nil
					}
				}

				return errors.New("directoy not found")

			}

		}

		if len(args) > 2 {
			return errors.New("invalid cd arguments")
		}

	}

	// fmt.Println("argv", program)

	// fmt.Println("program", program[0])

	path, err := CommandExists(program)

	// fmt.Println("path", path)

	if err != nil {
		return err
	}

	// interpolatedArgs = strings.Replacer

	proc, err := StartNewProcess(wdRaw, path, args, tempfile.Fd())

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

	wdRaw, _ = os.Getwd()

	init := exec.Cmd{
		Path:   "/bin/sh",
		Args:   []string{"/Users/tusharf5/.zshrc"},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	err := init.Start()

	if err != nil {
		fmt.Println("initialization error: ", err)
	}

	fmt.Println("welcome to shell")

	for {
		err := ShellProcess()

		if err != nil {
			fmt.Println("error: ", err)
		}
	}
}

func Interpolate(str string) string {
	var envVarsList [][]string

	for _, x := range os.Environ() {
		val := strings.Split(x, "=")
		envVarsList = append(envVarsList, []string{"$" + val[0], val[1]})
		envVarsList = append(envVarsList, []string{"${" + val[0] + "}", val[1]})
	}

	for _, pair := range envVarsList {
		str = strings.ReplaceAll(str, pair[0], pair[1])
	}

	return str
}

func CommandExists(cmd string) (string, error) {
	return exec.LookPath(cmd)
}

func StartNewProcess(wd, path string, argv []string, stdout uintptr) (*os.Process, error) {

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
