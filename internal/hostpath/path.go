// Package hostpath resolves user-supplied paths for the current execution environment.
package hostpath

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

var ErrInvalidPath = errors.New("expected an absolute local path")

type Converter func(context.Context, string) (string, error)

type Resolver struct {
	convertWindows Converter
}

func Native() Resolver { return Resolver{} }

func WSL(convert Converter) Resolver {
	if convert == nil {
		convert = WSLPath
	}
	return Resolver{convertWindows: convert}
}

func (r Resolver) Resolve(ctx context.Context, input string) (string, error) {
	if strings.TrimSpace(input) == "" || strings.ContainsAny(input, "\x00\r\n") || strings.HasPrefix(input, "//") || strings.HasPrefix(input, `\\`) {
		return "", ErrInvalidPath
	}
	windows := len(input) >= 3 && ((input[0] >= 'A' && input[0] <= 'Z') || (input[0] >= 'a' && input[0] <= 'z')) && input[1] == ':' && (input[2] == '\\' || input[2] == '/')
	if windows {
		if r.convertWindows == nil {
			return "", fmt.Errorf("%w: Windows drive paths require WSL", ErrInvalidPath)
		}
		output, err := r.convertWindows(ctx, input)
		if err != nil {
			return "", err
		}
		input = strings.TrimSuffix(strings.TrimSuffix(output, "\n"), "\r")
	}
	if !filepath.IsAbs(input) || strings.ContainsAny(input, "\x00\r\n") {
		return "", ErrInvalidPath
	}
	return filepath.Clean(input), nil
}
