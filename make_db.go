// +build ignore

package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mmcloughlin/cpuidb"
)

var (
	src = flag.String("src", "./source", "source directory containing mirror of InstLatx64")
	out = flag.String("out", "db.go", "output file")
)

// CPUIDFiles returns a list of all files in the given directory containing CPUID data.
func CPUIDFiles(dir string) ([]string, error) {
	return filepath.Glob(filepath.Join(dir, "*CPUID*.txt"))
}

// ParseCPUIDFiles parses all CPUID files in a directory. Errors are logged.
func ParseCPUIDFiles(dir string) []*cpuidb.CPU {
	filenames, err := CPUIDFiles(*src)
	if err != nil {
		log.Fatal(err)
	}

	cpus := make([]*cpuidb.CPU, 0, len(filenames))
	for _, filename := range filenames {
		cpu, err := cpuidb.ParseCPUFile(filename)
		if err != nil {
			log.Printf("failed to parse %s: %s", filename, err)
			continue
		}
		cpus = append(cpus, cpu)
	}

	return cpus
}

// CPUGoSyntax returns Go code for the given struct. Intended to be used inside an array initializer.
func CPUGoSyntax(cpu *cpuidb.CPU) string {
	return strings.Replace(fmt.Sprintf("%#v", *cpu), "cpuidb.", "", -1)
}

// Build Go source code that defines the given list of CPUs. Output not formatted.
func Build(cpus []*cpuidb.CPU) []byte {
	w := new(bytes.Buffer)

	_, self, _, _ := runtime.Caller(0)
	fmt.Fprintf(w, "// Code generated by go run %s. DO NOT EDIT.\n\n", filepath.Base(self))
	fmt.Fprintf(w, "package cpuidb\n\n")

	fmt.Fprintf(w, "var CPUs = []CPU{\n")
	for _, cpu := range cpus {
		fmt.Fprintf(w, "%s,\n", CPUGoSyntax(cpu))
	}
	fmt.Fprintf(w, "}\n")

	return w.Bytes()
}

// Write formatted Go code that initializes the given CPU list to a Writer.
func Output(w io.Writer, cpus []*cpuidb.CPU) error {
	code, err := format.Source(Build(cpus))
	if err != nil {
		return err
	}

	_, err = w.Write(code)
	return err
}

func main() {
	flag.Parse()

	cpus := ParseCPUIDFiles(*src)

	f, err := os.Create(*out)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	err = Output(f, cpus)
	if err != nil {
		log.Fatal(err)
	}
}