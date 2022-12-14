/*
 * Copyright © 2022 Andrey Kuvshinov. Contacts: <syslinux@protonmail.com>
 * Copyright © 2022 Eltaline OU. Contacts: <eltaline.ou@gmail.com>
 * Copyright © 2016 inhere
 *
 * This file is part of eCrond.
 *
 * eCrond is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * eCrond is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package main

import (
	"os"
	"os/exec"
	"strings"
)

// LineParser struct
type LineParser struct {
	parsed bool
	// Line the full input command line text
	// eg `kite top sub -a "the a message" --foo val1 --bar "val 2"`
	Line string
	// ParseEnv parse ENV var on the line.
	ParseEnv bool
	// the exploded nodes by space.
	nodes []string
	// the parsed args
	args []string
}

// NewParser
func NewParser(line string) *LineParser {
	return &LineParser{Line: line}
}

// ParseLine input command line text
func ParseLine(line string) []string {
	p := &LineParser{Line: line}
	return p.Parse()
}

// AlsoEnvParse input command line text to os.Args, will parse ENV var
func (p *LineParser) AlsoEnvParse() []string {
	p.ParseEnv = true
	return p.Parse()
}

// Parse input command line text to os.Args
func (p *LineParser) Parse() []string {
	if p.parsed {
		return p.args
	}

	p.parsed = true
	p.Line = strings.TrimSpace(p.Line)
	if p.Line == "" {
		return p.args
	}

	if p.ParseEnv {
		p.Line = os.ExpandEnv(p.Line)
	}

	p.nodes = strings.Split(p.Line, " ")
	if len(p.nodes) == 1 {
		p.args = p.nodes
		return p.args
	}

	var quoteChar, fullNode string
	for _, node := range p.nodes {
		if node == "" {
			continue
		}

		nodeLen := len(node)
		start, end := node[:1], node[nodeLen-1:]

		var clearTemp bool
		if start == "'" || start == `"` {
			noStart := node[1:]
			if quoteChar == "" { // start
				// only one words. eg: `-m "msg"`
				if end == start {
					p.args = append(p.args, node[1:nodeLen-1])
					continue
				}

				fullNode += noStart
				quoteChar = start
			} else if quoteChar == start { // invalid. eg: `-m "this is "message` `-m "this is "message"`
				p.appendWithPrefix(strings.Trim(node, quoteChar), fullNode)
				clearTemp = true // clear temp value
			} else if quoteChar == end { // eg: `"has inner 'quote'"`
				p.appendWithPrefix(node[:nodeLen-1], fullNode)
				clearTemp = true // clear temp value
			} else { // goon. eg: `-m "the 'some' message"`
				fullNode += " " + node
			}
		} else if end == "'" || end == `"` {
			noEnd := node[:nodeLen-1]
			if quoteChar == "" { // end
				p.appendWithPrefix(noEnd, fullNode)
				clearTemp = true // clear temp value
			} else if quoteChar == end { // end
				p.appendWithPrefix(noEnd, fullNode)
				clearTemp = true // clear temp value
			} else { // goon. eg: `-m "the 'some' message"`
				fullNode += " " + node
			}
		} else {
			if quoteChar != "" {
				fullNode += " " + node
			} else {
				p.args = append(p.args, node)
			}
		}

		if clearTemp {
			quoteChar, fullNode = "", ""
		}
	}

	if fullNode != "" {
		p.args = append(p.args, fullNode)
	}

	return p.args
}

// BinAndArgs get binName and args
func (p *LineParser) BinAndArgs() (bin string, args []string) {
	p.Parse() // ensure parsed.

	ln := len(p.args)
	if ln == 0 {
		return
	}

	bin = p.args[0]
	if ln > 1 {
		args = p.args[1:]
	}
	return
}

// NewExecCmd quick create exec.Cmd by cmdline string
func (p *LineParser) NewExecCmd() *exec.Cmd {
	// parse get bin and args
	binName, args := p.BinAndArgs()
	// create a new Cmd instance
	return exec.Command(binName, args...)
}

// Append prefix function
func (p *LineParser) appendWithPrefix(node, prefix string) {
	if prefix != "" {
		p.args = append(p.args, prefix+" "+node)
	} else {
		p.args = append(p.args, node)
	}
}
