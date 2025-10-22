package app

import (
	"bufio"
	"errors"
	"io"
	"strings"
)

type prompter struct {
	r *bufio.Reader
	w io.Writer
}

func (p *prompter) readLine() (string, error) {
	line, err := p.r.ReadString('\n')
	if err != nil {
		if errors.Is(err, io.EOF) && len(line) > 0 {
			return strings.TrimSpace(line), nil
		}
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func (p *prompter) readLineLower() (string, error) {
	s, err := p.readLine()
	return strings.ToLower(s), err
}

func (p *prompter) readLineLowerDefault(def string) (string, error) {
	s, err := p.readLine()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return strings.ToLower(def), nil
		}
		return "", err
	}
	if s == "" {
		return strings.ToLower(def), nil
	}
	return strings.ToLower(s), nil
}
