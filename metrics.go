package main

import (
	"bytes"
	"fmt"
	"maps"
	"os"
	"slices"

	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"

	dto "github.com/prometheus/client_model/go"
)

func readMetrics(path string) ([]*dto.MetricFamily, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	parser := expfmt.NewTextParser(model.UTF8Validation)

	families, err := parser.TextToMetricFamilies(bytes.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("parsing: %w", err)
	}

	return slices.Collect(maps.Values(families)), nil
}

type gatherer []*dto.MetricFamily

func (g gatherer) Gather() ([]*dto.MetricFamily, error) {
	return g, nil
}
