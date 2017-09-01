package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"io"
	"os"
	"regexp"

	"golang.org/x/tools/go/loader"
)

func loadPkgs() (*loader.Program, error) {
	var ldr loader.Config
	ldr.ParserMode = parser.ParseComments
	for _, pkg := range os.Args[1:] {
		ldr.Import(pkg)
	}
	return ldr.Load()
}

func parse(prog *loader.Program) error {
	for _, pkg := range prog.InitialPackages() {
		err := parsePkg(prog, pkg)
		if err != nil {
			return err
		}
	}
	return nil
}

func parsePkg(prog *loader.Program, pkg *loader.PackageInfo) error {
	for _, f := range pkg.Files {
		for _, object := range f.Scope.Objects {
			if object.Kind == ast.Fun {
				err := handleFuncs(prog, f, object, pkg)
				if err != nil {
					fmt.Printf("error handling funcs for %s: %s\n", object.Name, err)
				}
			}
		}
	}
	return nil
}

var emptyLineRegex = regexp.MustCompile(`\n\n`)

func handleFuncs(prog *loader.Program, file *ast.File, object *ast.Object, pkg *loader.PackageInfo) error {
	tokenFile := prog.Fset.File(file.Pos())
	osFile, err := os.Open(tokenFile.Name())
	if err != nil {
		return err
	}
	defer osFile.Close()
	start := tokenFile.Position(object.Decl.(*ast.FuncDecl).Pos()).Offset
	end := tokenFile.Position(object.Decl.(*ast.FuncDecl).End()).Offset
	osFile.Seek(int64(start), os.SEEK_SET)
	data := make([]byte, end-start)
	_, err = io.ReadFull(osFile, data)
	if err != nil {
		return err
	}
	line := tokenFile.Line(object.Pos())
	if emptyLineRegex.Match(data) {
		fmt.Printf("%s.%s - %s:%d\n", pkg.String(), object.Name, tokenFile.Name(), line)
	}
	return nil
}

func main() {
	prog, err := loadPkgs()
	if err != nil {
		fmt.Println("error loading code ", err)
		return
	}
	err = parse(prog)
	if err != nil {
		fmt.Println("error parsing api ", err)
		return
	}
}
