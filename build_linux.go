package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	GO_ARCH_386   GoArch = "386"
	GO_ARCH_AMD64 GoArch = "amd64"

	REPOSITORY = "virifi/gobuild"
)

type GoArch string

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func toSlashAbsPath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("getting absolute path for out directory failed : %v", err)
	}
	if runtime.GOOS != "windows" {
		return absPath, nil
	}
	// C:\Go\hoge -> C:/Go/hoge -> /C/Go/hoge
	s := strings.SplitN(filepath.ToSlash(absPath), ":", 2)
	if len(s) != 2 {
		return "", fmt.Errorf("Unsupported absolute path format '%v'", absPath)
	}
	return "/" + s[0] + s[1], nil
}

func buildGo(outDir string, goArch GoArch, commitHash string) error {
	if goArch != GO_ARCH_386 && goArch != GO_ARCH_AMD64 {
		return fmt.Errorf("unsupported GoArch '%v'", goArch)
	}
	absPath, err := toSlashAbsPath("out")
	if err != nil {
		return fmt.Errorf("toSlashAbsPath failed : %v", err)
	}
	fmt.Println("absPath =" + absPath)
	tag := REPOSITORY
	if err := runCommand("docker", "build", "-t", tag, "linux"); err != nil {
		return err
	}
	if err := runCommand("docker", "run", "-v", absPath+":/out", "-t", tag, "/build.sh", commitHash); err != nil {
		return err
	}
	return nil
}

func main() {
	commitHash := "91191e7b7bc8c0e1a6d49c7a9b3adeb1ab39a423"
	//commitHash := "714291f2d80bab1599a866f266a4fc6546e61632"
	//commitHash := "8017ace496f5a21bcd55377e250e325f8ba11d45"
	if err := buildGo("out", GO_ARCH_AMD64, commitHash); err != nil {
		fmt.Errorf("buildGo Failed : %v\n", err)
		os.Exit(1)
	}
}
