package main

import (
	"context"
	"dagger.io/dagger"
	"fmt"
	"log"
)

func Main(ctx context.Context, dag *dagger.Client) {
	out, _ := dag.Container().
		From("alpine").
		WithExec([]string{"sh", "-c", "echo 'hello, world'"}).
		Stdout(ctx)
	fmt.Println(out)

}

func main() {
	dag, err := dagger.Connect(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer dag.Close()
	Main(context.Background(), dag)
}
