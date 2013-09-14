/*
Copyright 2013 Google Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.


config.go reads the config file.
*/

package lib

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
)

const (
	// reComment is regexp that matches a comment line.
	reComment = "^#.*"

	// reEmpty is regexp that matches an empty line.
	reEmpty = "^$"

	// reTcCmdPath is regexp that matches line that defines tcCmdPath.
	reTcCmdPath = "^tcCmdPath = \"(?P<tcCmdPath>.*)\"$"

	// reParseInterval is regexp that matches line that defines parseInterval.
	reParseInterval = "^parseInterval = (?P<parseInterval>[0-9]+)$"

	// reTcQdiscStats is regexp that matches line that defines tcQdiscStats.
	reTcQdiscStats = "^tcQdiscStats = \"(?P<tcQdiscStats>.*)\"$"

	// reTcClassStats is regexp that matches line that defines tcClassStats.
	reTcClassStats = "^tcClassStats = \"(?P<tcClassStats>.*)\"$"

	// reIfaces is regexp that matches line that defines ifaces.
	reIfaces = "^ifaces = \"(?P<ifaces>.*)\"$"

	// reUserNameClass is regexp that matches line that defines user name.
	reUserNameClass = "^user[\t ]+=[\t ]+\"(?P<userName>.*)\"[\t ]+\"(?P<uploadClass>.*)\"[\t ]+\"(?P<downloadClass>.*)\"$"

	// reDebug is regexp that matches line that defines debug..
	reDebug = "^debug = (?P<debug>true|false)$"

	// trueString is the string representation of true.
	trueString = "true"

	// falseString is the string representatiopn of false.
	falseString = "false"
)

// config parses the configuration file and stores the parsed values.
type config struct {
	// TcCmdPath is the parsed tcCmdPath, defaults to empty string so that parser will use its internal default.
	TcCmdPath string

	// ParseInterval is the parsed ParseInterval, defaults to zero so that parser will use its internal default.
	ParseInterval int

	// TcQdiscStats is the parsed TcQdiscStats, defaults to nil so that parser will use its internal default.
	TcQdiscStats []string

	// TcClassStats is the parsed TcClassStats, defaults to nil so that parser will use its internal default.
	TcClassStats []string

	// Ifaces is the parsed Ifaces, defaults to nil so that parser will use its internal default.
	Ifaces []string

	// UserNameClass are the parsed user definitions, defaults to nil so that parser will use its internal default.
	UserNameClass map[string]userClass

	// Debug is the parsed Debug, defaults to false.
	Debug bool

	// filename is the config file name.
	filename string

	// reComment is the compiled version of reComment constant.
	reComment *regexp.Regexp

	// reEmpty is the compiled version of reEmpty constant.
	reEmpty *regexp.Regexp

	// reTcCmdPath is the compiled version of reTcCmdPath constant.
	reTcCmdPath *regexp.Regexp

	// reParseInterval is the compiled version of reParseInterval constant.
	reParseInterval *regexp.Regexp

	// reTcQdiscStats is the compiled version of reTcQdiscStats constant.
	reTcQdiscStats *regexp.Regexp

	// reTcClassStats is the compiled version of reTcClassStats constant.
	reTcClassStats *regexp.Regexp

	// reIfaces is the compiled version of reIfaces constant.
	reIfaces *regexp.Regexp

	// reUserNameClass is the compiled version of reUserNameClass constant.
	reUserNameClass *regexp.Regexp

	// reDebug is the compiled version of reDebug constant.
	reDebug *regexp.Regexp
}

// readConfig reads the configuration file and parses its content.
func (c *config) readConfig() error {
	content, err := ioutil.ReadFile(c.filename)
	if err != nil {
		return err
	}
	err = c.parseConfig(string(content))
	if err != nil {
		return err
	}
	return nil
}

// parseContent parses the content of the config file.
func (c *config) parseConfig(content string) error {
	lines := strings.Split(content, "\n")
	var err error
	for n, line := range lines {
		lineNumber := n + 1
		switch {
		// Ignore empty lines.
		case c.reEmpty.MatchString(line):
			continue

		// Ignore comments,
		case c.reComment.MatchString(line):
			continue

		// Line that defines path to the TC command.
		case c.reTcCmdPath.MatchString(line):
			err = c.getTcCmdPath(n, line)
			if err != nil {
				return err
			}

		// Line that defines parse interval.
		case c.reParseInterval.MatchString(line):
			err = c.getParseInterval(lineNumber, line)
			if err != nil {
				return err
			}

		// Line that defines TC parameters to get Qdisc statistics.
		case c.reTcQdiscStats.MatchString(line):
			err = c.getListOfStrings(&c.TcQdiscStats, c.reTcQdiscStats, lineNumber, line)
			if err != nil {
				return err
			}

		// Line that defines TC parameters to get Class statistics.
		case c.reTcClassStats.MatchString(line):
			err = c.getListOfStrings(&c.TcClassStats, c.reTcClassStats, lineNumber, line)
			if err != nil {
				return err
			}

		// Line that defines interfaces.
		case c.reIfaces.MatchString(line):
			err = c.getListOfStrings(&c.Ifaces, c.reIfaces, lineNumber, line)
			if err != nil {
				return err
			}

		// Line that defines an user.
		case c.reUserNameClass.MatchString(line):
			err = c.getUserName(lineNumber, line)
			if err != nil {
				return err
			}

		// Line that defines debug.
		case c.reDebug.MatchString(line):
			err = c.getDebug(lineNumber, line)
			if err != nil {
				return err
			}

		// Any other line.
		default:
			return fmt.Errorf("Error in config file %s on line %d: cannot parse this line: '%s'", c.filename, n, line)

		}
	}
	return nil
}

// getTcCmdPath parses line that contains tcCmdPath.
func (c *config) getTcCmdPath(lineNumber int, line string) error {
	if c.TcCmdPath != "" {
		return fmt.Errorf("Error in config file %s on line %d: found duplicate entry for tcCmdPath. Line: '%s'", c.filename, lineNumber, line)
	}
	if match := c.reTcCmdPath.FindAllStringSubmatch(line, -1); match != nil {
		matchSlice := match[0]
		c.TcCmdPath = matchSlice[1]
	} else {
		return fmt.Errorf("Error in config file %s on line %d: cannot parse this line: '%s'", c.filename, lineNumber, line)
	}
	return nil
}

// getParseInterval parses line that contains parseInterval.
func (c *config) getParseInterval(lineNumber int, line string) error {
	if c.ParseInterval != 0 {
		return fmt.Errorf("Error in config file %s on line %d: found duplicate entry for tcParseInterval. Line: '%s'", c.filename, lineNumber, line)
	}
	if match := c.reParseInterval.FindAllStringSubmatch(line, -1); match != nil {
		matchSlice := match[0]
		parseInterval, err := strconv.ParseInt(matchSlice[1], 10, 32)
		if err != nil {
			return fmt.Errorf("Error in config file %s on line %d: unable to parse the parseInterval value. Line: '%s', err: %s", c.filename, lineNumber, line, err)
		}
		c.ParseInterval = int(parseInterval)
	} else {
		return fmt.Errorf("Error in config file %s on line %d: cannot parse this line: '%s'", c.filename, lineNumber, line)
	}
	return nil
}

// getListOfStrings parses line that contains list of strings.
func (c *config) getListOfStrings(target *[]string, re *regexp.Regexp, lineNumber int, line string) error {
	if *target != nil {
		return fmt.Errorf("Error in config file %s on line %d: found duplicate entry. Line: '%s'", c.filename, lineNumber, line)
	}
	if match := re.FindAllStringSubmatch(line, -1); match != nil {
		matchSlice := match[0]
		*target = append(*target, strings.Split(matchSlice[1], " ")...)
	} else {
		return fmt.Errorf("Error in config file %s on line %d: cannot parse this line: '%s'", c.filename, lineNumber, line)
	}
	return nil
}

// getUserName parses line that contains user name definition.
func (c *config) getUserName(lineNumber int, line string) error {
	if match := c.reUserNameClass.FindAllStringSubmatch(line, -1); match != nil {
		matchSlice := match[0]
		name := matchSlice[1]
		uploadClass := matchSlice[2]
		downloadClass := matchSlice[3]
		// Is this a duplicate entry for this Class name ?
		if _, ok := c.UserNameClass[uploadClass]; ok {
			return fmt.Errorf("Error in config file %s on line %d: found duplicate definition of class %s. Line: '%s'", c.filename, lineNumber, uploadClass, line)
		}
		if _, ok := c.UserNameClass[downloadClass]; ok {
			return fmt.Errorf("Error in config file %s on line %d: found duplicate definition of class %s. Line: '%s'", c.filename, lineNumber, downloadClass, line)
		}

		if c.UserNameClass == nil {
			c.UserNameClass = make(map[string]userClass)
		}
		c.UserNameClass[uploadClass] = userClass{
			direction: uploadDirection,
			name:      name,
		}
		c.UserNameClass[downloadClass] = userClass{
			direction: downloadDirection,
			name:      name,
		}
	} else {
		return fmt.Errorf("Error in config file %s on line %d: cannot parse this line: '%s'", c.filename, lineNumber, line)
	}
	return nil
}

// getDebug parses line that contains debug.
func (c *config) getDebug(lineNumber int, line string) error {
	if match := c.reDebug.FindAllStringSubmatch(line, -1); match != nil {
		matchSlice := match[0]
		if matchSlice[1] == trueString {
			c.Debug = true
		}
	} else {
		return fmt.Errorf("Error in config file %s on line %d: cannot parse this line: '%s'", c.filename, lineNumber, line)
	}
	return nil
}

// NewConfig returns new config.
func NewConfig(filename string) (*config, error) {
	c := &config{
		filename:        filename,
		reComment:       regexp.MustCompile(reComment),
		reEmpty:         regexp.MustCompile(reEmpty),
		reTcCmdPath:     regexp.MustCompile(reTcCmdPath),
		reParseInterval: regexp.MustCompile(reParseInterval),
		reTcQdiscStats:  regexp.MustCompile(reTcQdiscStats),
		reTcClassStats:  regexp.MustCompile(reTcClassStats),
		reIfaces:        regexp.MustCompile(reIfaces),
		reUserNameClass: regexp.MustCompile(reUserNameClass),
		reDebug:         regexp.MustCompile(reDebug),
	}
	err := c.readConfig()
	return c, err
}
