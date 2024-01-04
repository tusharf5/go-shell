package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

type Command struct {
	stdin   *os.File
	stdout  *os.File
	program string
	args    []string
}

type CommandSet struct {
	finalOut *os.File
	commands []Command
}

var wdRaw string

func parseCommands(input string) CommandSet {

	var commands []Command
	allCommands := strings.Split(input, " | ")
	var finalOut *os.File
	for _, command := range allCommands {
		stdin, _ := newTempFile()
		stdout, _ := newTempFile()
		finalOut = stdout
		args := strings.Split(command, " ")

		commands = append(commands, Command{
			stdin:   stdin,
			stdout:  stdout,
			args:    args,
			program: args[0],
		})

	}

	return CommandSet{
		finalOut: finalOut,
		commands: commands,
	}
}

func promptPrefix() string {
	homedir, _ := os.UserHomeDir()
	wd := strings.Replace(wdRaw, homedir, "~", 1)
	return wd + "> "
}

func newTempFile() (*os.File, error) {
	tempfile, err := os.CreateTemp("", "")

	if err != nil {
		return nil, err
	}

	return tempfile, nil
}

func readPrompt(stdin *os.File) (string, error) {
	reader := bufio.NewReader(stdin)

	command, err := reader.ReadString('\n')

	if err != nil {
		return "", err
	}

	command = strings.Trim(strings.TrimSuffix(command, "\n"), " ")

	return command, nil
}

func readTempFile(file *os.File) (string, error) {
	var output string = ""
	file.Seek(0, 0)
	s := bufio.NewScanner(file)

	for s.Scan() {
		output = output + s.Text() + "\n"
	}

	if err := s.Err(); err != nil {
		return "", err
	}

	return strings.TrimSuffix(output, "\n"), nil
}

func interpolateInput(str string) string {
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

func executeProgram(wd, path string, argv []string, stdin uintptr, stdout uintptr) (*os.Process, error) {

	pid, err := syscall.ForkExec(path, argv, &syscall.ProcAttr{
		Dir:   wd,
		Env:   os.Environ(),
		Files: []uintptr{stdin, stdout, stdout},
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

func handleShellCommand(cmd Command) (exit bool, err error) {
	homedir, _ := os.UserHomeDir()

	if cmd.program == "cd" {

		if len(cmd.args) == 1 {
			wdRaw = homedir
			return true, nil
		}

		if len(cmd.args) == 2 {

			if cmd.args[1] == "." {
				return true, nil
			} else if cmd.args[1] == ".." {
				lastIndex := strings.LastIndex(wdRaw, "/")
				wdRaw = wdRaw[:lastIndex]
				return true, nil
			} else {

				entries, err := os.ReadDir(wdRaw)

				if err != nil {
					return true, nil
				}

				for _, e := range entries {
					if e.IsDir() && cmd.args[1] == e.Name() {
						wdRaw = wdRaw + "/" + e.Name()
						return true, nil
					}
				}

				return true, errors.New("directoy not found")

			}

		}

		if len(cmd.args) > 2 {
			return true, errors.New("invalid cd arguments")
		}

	}

	return false, nil
}

func runCommand(cmd Command, stdin uintptr, stdout uintptr) (*os.ProcessState, error) {

	exit, err := handleShellCommand(cmd)

	if err != nil || exit {
		return nil, nil
	}

	binartPath, err := exec.LookPath(cmd.program)

	if err != nil {
		return nil, err
	}

	proc, err := executeProgram(wdRaw, binartPath, cmd.args, stdin, stdout)

	if err != nil {
		return nil, err
	}

	state, err := proc.Wait()

	if err != nil {
		return nil, err
	}

	return state, nil

}

func newSession() error {

	fmt.Print(promptPrefix())

	input, _ := readPrompt(os.Stdin)

	input = interpolateInput(input)

	commandset := parseCommands(input)

	defer os.Remove(commandset.finalOut.Name())
	defer commandset.finalOut.Close()

	for index, command := range commandset.commands {

		defer os.Remove(command.stdin.Name())
		defer os.Remove(command.stdout.Name())
		defer command.stdin.Close()
		defer command.stdout.Close()

		if index == 0 {
			_, err := runCommand(command, command.stdin.Fd(), command.stdout.Fd())
			if err != nil {
				return err
			}

		} else {
			prevComm := commandset.commands[index-1]
			prevComm.stdout.Seek(0, 0)
			_, err := runCommand(command, prevComm.stdout.Fd(), command.stdout.Fd())

			if err != nil {
				return err
			}

		}

	}

	output, _ := readTempFile(commandset.finalOut)

	fmt.Println(output)

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
		err := newSession()

		if err != nil {
			fmt.Println("error: ", err)
		}
	}
}
