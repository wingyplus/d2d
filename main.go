package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

func main() {
	dockerfile := os.Args[1]
	f, err := os.Open(dockerfile)
	if err != nil {
		log.Fatal(err)
	}

	r, err := parser.Parse(f)
	if err != nil {
		log.Fatal(err)
	}

	fn := Fn("Main", []Parameter{{"ctx", "context.Context"}, {"dag", "*dagger.Client"}}).
		AddLine("out, _ := dag.Container().")

	for _, children := range r.AST.Children {
		switch children.Value {
		case "FROM":
			fn.AddLine(fmt.Sprintf(`From("%s").`, children.Next.Value))

		case "RUN":
			fn.AddLine(fmt.Sprintf(`WithExec([]string{"sh", "-c", "%s"}).`, children.Next.Value))

		default:
			continue
		}
	}

	fn.AddLine("Stdout(ctx)").AddLine(`fmt.Println(out)`)

	src := &Source{
		Package:   "main",
		Imports:   []string{"context", "log", "fmt", "dagger.io/dagger"},
		Functions: []*Function{fn},
	}

	fmt.Println(src)
}

type Source struct {
	Package   string
	Imports   []string
	Functions []*Function
}

func (src *Source) String() string {
	builder := &strings.Builder{}
	builder.WriteString("package main\n\n")
	builder.WriteString("import (\n")

	for _, imp := range src.Imports {
		builder.WriteRune('"')
		builder.WriteString(imp)
		builder.WriteRune('"')
		builder.WriteRune('\n')
	}

	builder.WriteString("\n)\n\n")

	for _, fn := range src.Functions {
		builder.WriteString(fn.String())
		builder.WriteString("\n\n")
	}

	// TODO: write main somewhere.

	builder.WriteString(`func main() {
    dag, err := dagger.Connect(context.Background())
    if err != nil {
        log.Fatal(err)
    }
    defer dag.Close()
    Main(context.Background(), dag)
}`)

	return builder.String()
}

type Parameter struct {
	Name string
	Type string
}

type Function struct {
	Name       string
	Parameters []Parameter
	Body       []string
}

func (fn *Function) AddLine(line string) *Function {
	fn.Body = append(fn.Body, line)
	return fn
}

func (fn *Function) String() string {
	builder := &strings.Builder{}

	builder.WriteString("func ")
	builder.WriteString(fn.Name)
	builder.WriteRune('(')

	for i, param := range fn.Parameters {
		builder.WriteString(param.Name + " " + param.Type)

		if i < len(fn.Parameters)-1 {
			builder.WriteRune(',')
		}
	}

	builder.WriteRune(')')
	builder.WriteString(" {")
	builder.WriteRune('\n')
	for _, line := range fn.Body {
		builder.WriteString(line)
		builder.WriteRune('\n')
	}
	builder.WriteRune('\n')
	builder.WriteRune('}')
	return builder.String()
}

func Fn(name string, params []Parameter) *Function {
	return &Function{Name: name, Parameters: params}
}
