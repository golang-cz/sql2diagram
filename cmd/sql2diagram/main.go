package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/golang-cz/sql2diagram/internal/diagram"
	"github.com/golang-cz/sql2diagram/internal/schema"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var input []byte
	var err error

	switch len(os.Args) {
	case 1:
		// No arguments - read from stdin
		input, err = io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(fmt.Errorf("read from stdin: %w", err))
		}

	case 2:
		// One argument - read from file
		filePath := os.Args[1]
		input, err = os.ReadFile(filePath)
		if err != nil {
			log.Fatal(fmt.Errorf("read file %s: %w", filePath, err))
		}

	default:
		log.Fatal("Usage: sql2diagram [schema_file] > schema.svg")
	}

	if err := generateDiagram(ctx, string(input)); err != nil {
		log.Fatal(err)
	}
}

func generateDiagram(ctx context.Context, input string) error {
	schemaSQL := strings.TrimSpace(input)

	if schemaSQL == "" {
		return fmt.Errorf("schema was not provided, input is empty")
	}

	schemaDef, err := schema.Parse(schemaSQL)
	if err != nil {
		return err
	}

	out, err := diagram.Render(ctx, schemaDef)
	if err != nil {
		return err
	}

	if _, err = os.Stdout.Write(out); err != nil {
		return fmt.Errorf("write to file: %w", err)
	}

	return nil
}
