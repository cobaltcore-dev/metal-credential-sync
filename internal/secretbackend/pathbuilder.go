// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and CobaltCore contributors
// SPDX-License-Identifier: Apache-2.0

package secretbackend

import (
	"bytes"
	"fmt"
	"text/template"
)

// PathBuilder builds secret paths from templates
type PathBuilder struct {
	template *template.Template
}

// PathVariables holds the variables for path template expansion
type PathVariables struct {
	Region   string
	Hostname string
	Username string
}

// NewPathBuilder creates a new PathBuilder with the given template string
func NewPathBuilder(templateStr string) (*PathBuilder, error) {
	tmpl, err := template.New("path").Parse(templateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path template: %w", err)
	}

	return &PathBuilder{
		template: tmpl,
	}, nil
}

// Build constructs a path using the provided variables
func (pb *PathBuilder) Build(vars PathVariables) (string, error) {
	var buf bytes.Buffer
	if err := pb.template.Execute(&buf, vars); err != nil {
		return "", fmt.Errorf("failed to execute path template: %w", err)
	}
	return buf.String(), nil
}
