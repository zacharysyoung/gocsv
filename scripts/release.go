package main

import (
	"archive/zip"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	distDir    = "dist"
	executable = "gozcsv"
)

var (
	osArches = [][]string{
		{"darwin", "amd64"},
		{"darwin", "arm64"},
		{"freebsd", "amd64"},
		{"freebsd", "arm64"},
		{"linux", "386"},
		{"linux", "amd64"},
		{"linux", "arm"},
		{"linux", "arm64"},
		{"linux", "ppc64le"},
		{"linux", "riscv64"},
		{"windows", "amd64"},
		{"windows", "arm64"},
	}
)

func main() {
	rmdir(distDir)
	mkdir(distDir)

	for _, osArch := range osArches {
		var (
			goos   = osArch[0]
			goarch = osArch[1]

			goosVar   = "GOOS=" + goos
			goarchVar = "GOARCH=" + goarch
		)

		fmt.Printf("creating %s %s\n", goosVar, goarchVar)

		osArchDir := fmt.Sprintf("%s-%s", goos, goarch)
		osArchDir = filepath.Join(distDir, osArchDir)

		bin := executable
		if goos == "windows" {
			bin += ".exe"
		}
		binPath := filepath.Join(osArchDir, bin)

		cmd := exec.Command("go", "build", "-o", binPath, "./cmd/cli")
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, goosVar)
		cmd.Env = append(cmd.Env, goarchVar)

		rmdir(osArchDir)
		mkdir(osArchDir)
		err := cmd.Run()
		if err != nil {
			fatalf("could not build %s %s: %v", goosVar, goarchVar, err)
		}
		zipBin(binPath)
		rmdir(osArchDir)
	}
}

func mkdir(path string) {
	err := os.Mkdir(path, 0755)
	if err != nil {
		fatalf("", err)
	}
}

func rmdir(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		fatalf("", err)
	}
}

func zipBin(binPath string) {
	// fmt.Println(binPath)
	// dir, bin := filepath.Split(binPath)
	dir, _ := filepath.Split(binPath)
	dir = filepath.Clean(dir)
	// fmt.Println(dir, bin)
	// return
	fIn, err := os.Open(binPath)
	if err != nil {
		fatalf("could not open bin %s: %v", binPath, err)
	}
	defer fIn.Close()

	fOut, err := os.Create(dir + ".zip")
	if err != nil {
		fatalf("could not create zip: %v", err)
	}
	defer fOut.Close()

	w := zip.NewWriter(fOut)
	// f, err := w.Create(bin)
	// if err != nil {
	// 	fatalf("could not create %s in archive: %v", bin, err)
	// }
	// if _, err := io.Copy(f, fIn); err != nil {
	// 	fatalf("could not write to archive: %v", err)
	// }

	if err := w.AddFS(os.DirFS(dir)); err != nil {
		fatalf("could not add dir to archive: %v", err)
	}
	if err := w.Close(); err != nil {
		fatalf("problem closing archive: %v", err)
	}
}

func fatalf(format string, args ...any) {
	if format == "" {
		format = "%v"
	}
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

// # rm -rf ${DIST_DIR}
// # mkdir ${DIST_DIR}

// # for op
