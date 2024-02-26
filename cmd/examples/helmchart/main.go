package main

import (
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/davidmdm/halloumi/pkg/helm"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

//go:embed postgresql-14.2.3.tgz
var pg []byte

//go:embed redis-18.16.1.tgz
var redis []byte

func run() error {
	withRedis := flag.Bool("redis", false, "with redis instance")
	flag.Parse()

	var values map[string]any

	if !term.IsTerminal(int(os.Stdin.Fd())) {
		if err := yaml.NewDecoder(os.Stdin).Decode(&values); err != nil && err != io.EOF {
			return fmt.Errorf("failed to decode stdin: %w", err)
		}
	}

	chart, err := helm.LoadChartFromZippedArchive(pg)
	if err != nil {
		return fmt.Errorf("failed to load chart: %w", err)
	}

	resources, err := chart.Render(os.Args[0], "default", values)
	if err != nil {
		return fmt.Errorf("failed to render chart resources: %w", err)
	}

	if *withRedis {
		rc, err := helm.LoadChartFromZippedArchive(redis)
		if err != nil {
			return fmt.Errorf("failed to load redis: %w", err)
		}

		rr, err := rc.Render(os.Args[0], "default", nil)
		if err != nil {
			return err
		}

		resources = append(resources, rr...)
	}

	return json.NewEncoder(os.Stdout).Encode(resources)
}
