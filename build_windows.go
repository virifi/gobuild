package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	GO_ARCH_386   GoArch = "386"
	GO_ARCH_AMD64 GoArch = "amd64"

	GO_REPOSITORY = "https://github.com/golang/go.git"
)

type GoArch string

type BuildEnv struct {
	BuildDir         string // This path used as embedded GOROOT of go tool
	BootstrapGo386   string
	BootstrapGoAmd64 string
	TdmGcc32Path     string
	TdmGcc64Path     string
}

func buildGo(env BuildEnv, goArch GoArch, commitHash, outZipPath string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(wd)

	outZipPath, err = filepath.Abs(outZipPath)
	if err != nil {
		return err
	}

	if goArch != GO_ARCH_386 && goArch != GO_ARCH_AMD64 {
		return fmt.Errorf("goArch '%v' is not supported.", goArch)
	}
	if _, err := os.Stat(env.BootstrapGo386); err != nil {
		return fmt.Errorf("BootstrapGo386 '%v' does not exist.", env.BootstrapGo386)
	}
	if _, err := os.Stat(env.BootstrapGoAmd64); err != nil {
		return fmt.Errorf("BootstrapGoAmd64 '%v' does not exit.", env.BootstrapGoAmd64)
	}
	if _, err := os.Stat(env.TdmGcc32Path); err != nil {
		return fmt.Errorf("TdmGcc32Path '%v' does not exit.", env.TdmGcc32Path)
	}
	if _, err := os.Stat(env.TdmGcc64Path); err != nil {
		return fmt.Errorf("TdmGcc64Path '%v' does not exit.", env.TdmGcc64Path)
	}
	if err := checkout(env.BuildDir, commitHash); err != nil {
		return fmt.Errorf("Checkout failed '%v'.", err)
	}

	envMap := getEnvAsMap()
	var bootstrapGoPath, tdmGccPath, arch string
	if goArch == GO_ARCH_386 {
		bootstrapGoPath = env.BootstrapGo386
		tdmGccPath = env.TdmGcc32Path
		arch = "386"
	} else {
		bootstrapGoPath = env.BootstrapGoAmd64
		tdmGccPath = env.TdmGcc64Path
		arch = "amd64"
	}

	envMap = prependPath(envMap, filepath.Join(tdmGccPath, "bin"))
	envMap["GOROOT_BOOTSTRAP"] = bootstrapGoPath
	envMap["GOARCH"] = arch
	if err := os.Chdir(filepath.Join(env.BuildDir, "src")); err != nil {
		return err
	}
	if err := runCommandWithEnv(envMap, "all.bat"); err != nil {
		return err
	}
	return ZipDir(outZipPath, env.BuildDir, "go", true)
}

func ZipDir(destZip, srcDir, rootDirName string, ignoreDotGit bool) error {
	os.MkdirAll(filepath.Dir(destZip), 0755) // Ignore error

	f, err := os.Create(destZip)
	if err != nil {
		return err
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		if ignoreDotGit && strings.HasPrefix(relPath, ".git") {
			return nil
		}

		pathInZip := filepath.ToSlash(filepath.Join(rootDirName, relPath))
		efi, err := os.Lstat(path)
		if err != nil {
			return err
		}
		header, err := zip.FileInfoHeader(efi)
		if err != nil {
			return err
		}
		header.Name = pathInZip
		header.Method = zip.Deflate
		ew, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}
		// if current file is symlink, write link destination as data
		if efi.Mode()&os.ModeSymlink == os.ModeSymlink {
			realPath, err := os.Readlink(path)
			if err != nil {
				return err
			}
			if _, err := ew.Write([]byte(realPath)); err != nil {
				return err
			}
			return nil
		}

		sf, err := os.Open(path)
		if err != nil {
			return err
		}
		if _, err := io.Copy(ew, sf); err != nil {
			sf.Close()
			return err
		}
		sf.Close()

		return nil
	})

	if err != nil {
		return err
	}

	return zw.Flush()
}

func prependPath(envMap map[string]string, path string) map[string]string {
	origPath, ok := envMap["Path"]
	if ok {
		envMap["Path"] = path + ";" + origPath
		return envMap
	}
	origPath, ok = envMap["PATH"]
	if ok {
		envMap["PATH"] = path + ";" + origPath
		return envMap
	}
	panic("PATH environment variable not found")
}

func checkout(buildDir, commitHash string) error {
	if _, err := os.Stat(buildDir); err != nil {
		if err := runCommand("git", "clone", GO_REPOSITORY, buildDir); err != nil {
			return err
		}
	}
	if err := os.Chdir(buildDir); err != nil {
		return err
	}
	if err := runCommand("git", "checkout", commitHash); err != nil {
		return err
	}
	return nil
}

func runCommand(name string, args ...string) error {
	return runCommandWithEnv(nil, name, args...)
}

func runCommandWithEnv(envMap map[string]string, name string, args ...string) error {
	var env []string
	for k, v := range envMap {
		env = append(env, k+"="+v)
	}
	cmd := exec.Command(name, args...)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func getEnvAsMap() map[string]string {
	data := os.Environ()
	items := make(map[string]string)
	for _, val := range data {
		splits := strings.SplitN(val, "=", 2)
		key := splits[0]
		value := splits[1]
		items[key] = value
	}
	return items
}

func main() {
	env := BuildEnv{
		BuildDir:         `C:\Go`,
		BootstrapGo386:   `C:\apps\go1.4.2_386`,
		BootstrapGoAmd64: `C:\apps\go1.4.2_amd64`,
		TdmGcc32Path:     `C:\TDM-GCC-32`,
		TdmGcc64Path:     `C:\TDM-GCC-64`,
	}
	commitHash := "8017ace496f5a21bcd55377e250e325f8ba11d45"

	outZip386 := filepath.Join("out", fmt.Sprintf("go.windows-386.%v.zip", commitHash))
	outZipAmd64 := filepath.Join("out", fmt.Sprintf("go.windows-amd64.%v.zip", commitHash))

	if err := buildGo(env, GO_ARCH_386, commitHash, outZip386); err != nil {
		fmt.Fprintf(os.Stderr, "Building GOARCH=386 failed : %v\n", err)
		os.Exit(1)
	}
	if err := buildGo(env, GO_ARCH_AMD64, commitHash, outZipAmd64); err != nil {
		fmt.Fprintf(os.Stderr, "Building GOARCH=amd64 failed : %v\n", err)
		os.Exit(1)
	}
}
