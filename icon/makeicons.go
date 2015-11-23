// +build ignore

package main

import (
	"bytes"
	"encoding/xml"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
	"text/template"
)

const tmpl = `// generated by makeicons.go; DO NOT EDIT
package icon

type Icon int

func (ic Icon) Texcoords() (float32, float32) {
	tc := texcoords[int(ic)]
	return tc[0], tc[1]
}

const ({{range $i, $ic := .Icons}}
	{{if eq $i 0}}{{$ic.Name}} Icon = iota{{else}}{{$ic.Name}}{{end}}{{end}}
)

var texcoords = [][2]float32{
	{{range $i, $ic := .Icons}}{{"{"}}{{$ic.X}}, {{$ic.Y}}{{"}"}},{{end}}
}
`

type Icon struct {
	Name string
	Set  string
	X, Y float32
}

type Set struct {
	Name string
	Url  string
}

type ByName []Icon

func (a ByName) Len() int           { return len(a) }
func (a ByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool { return a[i].Name < a[j].Name }

func main() {
	gopath := os.Getenv("GOPATH")
	iconpath := gopath + "/src/github.com/google/material-design-icons/"
	glob := "*/drawable-mdpi/*black_48dp.png"
	name := "material-icons-black-mdpi"
	cmd := exec.Command("sprity", "create", ".", iconpath+glob, "--margin=0", "--orientation=binary-tree", "--name="+name, "--template=makeicons.hbs", "--style=makeicons.json", "--css-path=''", "--prefix=Foo")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	f, err := os.Open("makeicons.json")
	if err != nil {
		log.Fatal(err)
	}
	dec := xml.NewDecoder(f)

	ics := struct {
		Sets  []Set  `xml:"Set"`
		Icons []Icon `xml:"Icon"`
	}{}
	dec.Decode(&ics)

	for i, ic := range ics.Icons {
		name := strings.Split(ic.Name, "-")
		a := strings.Title(name[0])
		b := strings.Replace(name[3][3:], "_black_48dp", "", 1)
		c := strings.Split(b, "_")
		for i, s := range c {
			c[i] = strings.Title(s)
		}
		b = strings.Join(c, "")
		ic.Name = a + b
		ic.X, ic.Y = ic.X/2048, ic.Y/2048
		ics.Icons[i] = ic
	}

	sort.Sort(ByName(ics.Icons))

	buf := new(bytes.Buffer)
	t, err := template.New("").Parse(tmpl)
	if err != nil {
		log.Fatal(err)
	}
	t.Execute(buf, ics)
	if err := ioutil.WriteFile("icon_mdpi.go", buf.Bytes(), 0644); err != nil {
		log.Fatal(err)
	}
	os.Remove("makeicons.json")
}
