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


parser.go contains structs and methods used to execute the TC command and parse information about Queues and Classes.
*/

package lib

import (
	"fmt"
	"log/syslog"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Package constants.
const (
	// emptyString is an empty string.
	emptyString = ""

	// newLine is the new line character.
	newLine = "\n"

	// reQdiscHeaderStr is string version of the RE to match header of a Qdisc in Qdisc statistics.
	reQdiscHeaderStr = "qdisc (?P<qdiscName>[a-zA-Z_]+) (?P<qdiscHandle>[0-9a-f]+):.*"

	// reClassHeaderStr is string version of the RE to match header of a Class in Class statistics.
	reClassHeaderStr = "class (?P<className>[a-zA-Z_]+) (?P<qdiscHandle>[0-9a-f]+):(?P<classHandle>[0-9a-f]+).*"

	// reStatsStr is string version of the RE to match the Qdisc and Class statisticsin TC output.
	reStatsStr = " Sent (?P<sentBytes>[0-9]+) bytes (?P<sentPkt>[0-9]+) pkt .dropped (?P<droppedPkt>[0-9]+), overlimits (?P<overLimitPkt>[0-9]+) requeues 0."
)

// These variables are the default options used by tcParser.
var (
	// tcCmdPath is the default path to the TC binary.
	tcCmdPath = "/sbin/tc"

	// parseInterval is the period in seconds how often we run TC and parse its output which is then stored in SNMP for retrieval.
	parseInterval = 5

	// tcQdiscStats are the default arguments that should be passed to TC in order to get Qdisc statistics.
	tcQdiscStats = []string{"-s", "qdisc", "show", "dev"}

	// tcClassStats are the default arguments that should be passed to TC in order to get Class statistics.
	tcClassStats = []string{"-s", "class", "show", "dev"}

	// ifaces is the default slice of interface names that should be monitored.
	ifaces = []string{"eth0"}
)

// sysLogger is an interface to Syslog.
type sysLogger interface {
	Info(m string) (err error)
	Err(m string) (err error)
}

// commandExecuter is an interface that executes system commands.
type commandExecuter interface {
	Execute(name string, arg ...string) (string, error)
}

// systemCommand implements commandExecuter.
type systemCommand struct {
}

// Execute runs a system command and returns its standard output.
func (sc *systemCommand) Execute(name string, arg ...string) (string, error) {
	cmdOutput, err := exec.Command(name, arg...).Output()
	if err != nil {
		return emptyString, err
	}
	outputString := string(cmdOutput)
	return outputString, nil
}

// TcParserOptions holds the configurable options for the tcParser.
type TcParserOptions struct {
	// TcCmdPath is the path to the TC binary.
	TcCmdPath string

	// ParseInterval is the number of seconds during which we consider data fresh and do not rerun TC command.
	ParseInterval int

	// TcQdiscStats are the arguments that should be passed to TC in order to get Qdisc statistics.
	TcQdiscStats []string

	// TcClassStats are the arguments that should be passed to TC in order to get Class statistics.
	TcClassStats []string

	// Ifaces is a slice of interface names that should be monitored.
	Ifaces []string

	// UserNameClass is a map of the tcNames (see parseData()) to userClass definitions.
	UserNameClass map[string]userClass

	// Debug determines whether we perform extensive logging to Syslog.
	Debug bool
}

// tcCmdPath returns the configured tcCmdPath, or the default one if it wasn't set.
func (o *TcParserOptions) tcCmdPath() string {
	if o != nil && o.TcCmdPath != "" {
		return o.TcCmdPath
	}
	return tcCmdPath
}

// parseInterval returns the configured parseInterval, or the default one if it wasn't set.
func (o *TcParserOptions) parseInterval() int {
	if o != nil && o.ParseInterval != 0 {
		return o.ParseInterval
	}
	return parseInterval
}

// tcQdiscStats returns the configured tcQdiscStats, or the default one if it wasn't set.
func (o *TcParserOptions) tcQdiscStats() []string {
	if o != nil && o.TcQdiscStats != nil {
		return o.TcQdiscStats
	}
	return tcQdiscStats
}

// tcClassStats returns the configured tcClassStats, or the default one if it wasn't set.
func (o *TcParserOptions) tcClassStats() []string {
	if o != nil && o.TcClassStats != nil {
		return o.TcClassStats
	}
	return tcClassStats
}

// ifaces returns the configured ifaces, or the default one if it wasn't set.
func (o *TcParserOptions) ifaces() []string {
	if o != nil && o.Ifaces != nil {
		return o.Ifaces
	}
	return ifaces
}

// userNameClass returns the configured userNameClass, or the default one if it wasn't set.
func (o *TcParserOptions) userNameClass() map[string]userClass {
	if o != nil && o.UserNameClass != nil {
		return o.UserNameClass
	}
	userNameClass := make(map[string]userClass)
	return userNameClass
}

// tcParser reads qdisc and class stats from TC command output and provides them to SNMPD.
type tcParser struct {
	// logger is the Writer used to log messages to Syslog.
	logger sysLogger

	// parserOptions stores the configuration options provided to the tcParser.
	options *TcParserOptions

	// reQdiscHeader is the compiled version of reQdiscHeaderStr.
	reQdiscHeader *regexp.Regexp

	// reClassHeader is the compiled version of reClassHeaderStr.
	reClassHeader *regexp.Regexp

	// reStats is the compiled version of reStatsStr.
	reStats *regexp.Regexp

	// snmp is the SNMP handler that will store our parsed data and deliver them to the SNMP daemon.
	snmp snmpHandler

	// executer is interface that runs system commands.
	executer commandExecuter
}

// NewTcParser creates new tcParser.
func NewTcParser(options *TcParserOptions, snmp *snmp, logger *syslog.Writer) *tcParser {
	tp := &tcParser{
		logger:        logger,
		options:       options,
		reQdiscHeader: regexp.MustCompile(reQdiscHeaderStr),
		reClassHeader: regexp.MustCompile(reClassHeaderStr),
		reStats:       regexp.MustCompile(reStatsStr),
		snmp:          snmp,
		executer:      &systemCommand{},
	}
	tp.start()
	return tp
}

// logIfDebug logs a message into Syslog if the debug option is set.
func (t *tcParser) logIfDebug(message string) {
	if t.options.Debug {
		t.logger.Info(message)
	}
}

// start starts the periodic execution of TC every ParseInterval.
func (t *tcParser) start() {
	t.logger.Info("start(): Starting the tc_reader.")
	configTemplate := "tc_reader configuration:  tcCmdPath: %s  parseInterval: %d  tcQdiscStats: %s  tcClassStats: %s  ifaces: %s  userNameClass: %v"
	t.logIfDebug(fmt.Sprintf(configTemplate, t.options.tcCmdPath(), t.options.parseInterval(), t.options.tcQdiscStats(), t.options.tcClassStats(), t.options.ifaces(), t.options.userNameClass()))
	// One initial run of TC execution and parsing.
	t.parseTc()

	go func() {
		for _ = range time.Tick(time.Duration(t.options.parseInterval()) * time.Second) {
			t.parseTc()
		}
	}()
}

// executeTc executes the TC commands for an interface and returns the command output.
func (t *tcParser) executeTc(iface string) (string, string, error) {
	qdiscStats := append(t.options.tcQdiscStats(), iface)
	qdiscOutput, err := t.executer.Execute(t.options.tcCmdPath(), qdiscStats...)
	if err != nil {
		return emptyString, emptyString, err
	}

	clasStats := append(t.options.tcClassStats(), iface)
	classOutput, err := t.executer.Execute(t.options.tcCmdPath(), clasStats...)
	if err != nil {
		return emptyString, emptyString, err
	}
	return qdiscOutput, classOutput, nil
}

// Executes the TC command to get statistics for Qdiscs and Classes on a interfaces and parses the output.
//
// Example output of 'tc -s qdisc show dev eth0':
// qdisc dsmark 1: root refcnt 2 indices 0x0010 default_index 0x0000
//  Sent 8165477580 bytes 5927092 pkt (dropped 49112, overlimits 0 requeues 0)
//  rate 0bit 0pps backlog 0b 0p requeues 0
// qdisc htb 2: parent 1: r2q 10 default 0 direct_packets_stat 42920
//  Sent 8165477220 bytes 5927088 pkt (dropped 49112, overlimits 9389236 requeues 0)
//  rate 0bit 0pps backlog 0b 0p requeues 0
//
// Example output of 'tc -s class show dev eth0':
// class htb 2:1 root rate 3072Kbit ceil 3072Kbit burst 3141b cburst 3141b
//  Sent 8092853284 bytes 5693309 pkt (dropped 0, overlimits 0 requeues 0)
//  rate 22528bit 34pps backlog 0b 0p requeues 0
//  lended: 4348128 borrowed: 0 giants: 0
//  tokens: 124922 ctokens: 124922
//
// class htb 2:2 parent 2:1 leaf 3: prio 0 rate 614400bit ceil 614400bit burst 1907b cburst 1907b
//  Sent 0 bytes 0 pkt (dropped 0, overlimits 0 requeues 0)
//  rate 0bit 0pps backlog 0b 0p requeues 0
//  lended: 0 borrowed: 0 giants: 0
//  tokens: 388171 ctokens: 388171
func (t *tcParser) parseTc() {
	t.snmp.lock()
	defer t.snmp.unlock()

	// Erase any previous data.
	t.snmp.erase()

	for _, iface := range t.options.ifaces() {
		qdiscOutput, classOutput, err := t.executeTc(iface)
		if err != nil {
			t.logger.Err(fmt.Sprintf("parseTc(): Unable to get TC command output, error: %s", err))
			return
		}

		err = t.parseData(qdiscOutput, iface, t.reQdiscHeader, t.reStats)
		if err != nil {
			t.logger.Err(fmt.Sprintf("parseTc(): Unable to parse the output of TC commands while getting Qdisc statistics, error: %s", err))
			return
		}

		err = t.parseData(classOutput, iface, t.reClassHeader, t.reStats)
		if err != nil {
			t.logger.Err(fmt.Sprintf("parseTc(): Unable to parse the output of TC commands while getting Class statistics, error: %s", err))
			return
		}
	}
}

// parseData parses data received from the TC command output.
func (t *tcParser) parseData(cmdOutput string, ifaceName string, reHeader, reData *regexp.Regexp) error {

	// haveHeader indicates that parseData saw the header line for a Qdisc / Class.
	var haveHeader bool

	// haveData indicates that parseData saw the data line for a Qdisc / Class.
	var haveData bool

	var qdiscHandle int64
	var classHandle int64
	var sentBytes int64
	var sentPkt int64
	var droppedPkt int64
	var overLimitPkt int64
	var err error

	for _, line := range strings.Split(cmdOutput, newLine) {
		// Does this line contain the header ?
		if match := reHeader.FindAllStringSubmatch(line, -1); match != nil {
			matchSlice := match[0]
			qdiscHandle, err = strconv.ParseInt(matchSlice[2], 16, 32)
			if err != nil {
				return err
			}
			// Class handle is only present in the output for a Class. We assume zero in the output for a Qdisc.
			if len(matchSlice) == 4 {
				classHandle, err = strconv.ParseInt(matchSlice[3], 16, 32)
				if err != nil {
					return err
				}
			}
			haveHeader = true
		}

		// Does this line contain the data ?
		if match := reData.FindAllStringSubmatch(line, -1); match != nil {
			matchSlice := match[0]
			sentBytes, err = strconv.ParseInt(matchSlice[1], 10, 64)
			if err != nil {
				return err
			}
			sentPkt, err = strconv.ParseInt(matchSlice[2], 10, 32)
			if err != nil {
				return err
			}
			droppedPkt, err = strconv.ParseInt(matchSlice[3], 10, 32)
			if err != nil {
				return err
			}
			overLimitPkt, err = strconv.ParseInt(matchSlice[4], 10, 32)
			if err != nil {
				return err
			}
			haveData = true
		}

		// Store the data once we parsed both the header and the data.
		if haveHeader && haveData {
			haveHeader = false
			haveData = false

			// tcName is the internal name for this Qdisc / Class on an interface. Example: "eth0:2:3" is Class 3, Qdisc 2 on interface eth0.
			tcName := fmt.Sprintf("%s:%s:%s", ifaceName, strconv.FormatInt(qdiscHandle, 10), strconv.FormatInt(classHandle, 10))
			data := &parsedData{
				name:         tcName,
				sentBytes:    sentBytes,
				sentPkt:      sentPkt,
				droppedPkt:   droppedPkt,
				overLimitPkt: overLimitPkt,
			}
			t.snmp.addData(data)

			// Store information for an user if this tcName is configured as belonging to an user.
			if userClass, ok := t.options.userNameClass()[tcName]; ok {
				userData := &parsedData{
					name:         tcName,
					sentBytes:    sentBytes,
					sentPkt:      sentPkt,
					droppedPkt:   droppedPkt,
					overLimitPkt: overLimitPkt,
					userClass:    &userClass,
				}
				t.snmp.addData(userData)
			}
		}
	}
	return nil
}
