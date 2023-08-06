package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"go.olapie.com/utils"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s handler...\n", os.Args[0])
	}
	const codeTemplate = `
package httphandler

import "net/http"

type {{.HandlerName}} struct {
}

var _ http.Handler = (*{{.HandlerName}})(nil)

func New{{.HandlerName}}() *{{.HandlerName}} {
	h := &{{.HandlerName}} {
	}
	return h
}

func (h *{{.HandlerName}}) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

}

`
	for _, name := range os.Args[1:] {
		handlerName := name
		if strings.HasSuffix(name, "Handler") {
			name = name[0 : len(name)-7]
		} else {
			handlerName = name + "Handler"
		}
		fileName := utils.ToSnake(name) + ".go"
		code := strings.ReplaceAll(codeTemplate, "{{.HandlerName}}", handlerName)
		err := os.WriteFile(fileName, []byte(code), 0644)
		if err != nil {
			log.Fatalln(err)
		} else {
			log.Println("Done")
		}
	}
}
