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


Package main starts tc_reader that acts as a pass_persist script for the Net-SNMP daemon.

It extends SNMP information by adding data under the configured myOID (defaults to .1.3.6.1.4.1.2021.255)

myOID data will be in this hierarchy:
myOID.1 - tcIndexLeaf                   - Stores integers, the SNMP indexes assigned to Qdiscs and Classes.
myOID.2 - tcNumIndexLeaf                - Stores an integer, the count of indexes assigned to Qdiscs and Classes.
myOID.3 - tcNameLeaf                    - Stores strings, the names of Qdiscs and Classes. Names are in the form "eth0:2:3", which means interface eth0, Qdisc 2, Class 3.
myOID.4 - sentBytesLeaf                 - Stores counter32, the sent bytes for each tcIndex.
myOID.5 - sentPktLeaf                   - Stores counter32, the sent packets for each tcIndex.
myOID.6 - droppedPktLeaf                - Stores counter32, the dropped packets for each tcIndex.
myOID.7 - overLimitPktLeaf              - Stores counter32, the over limit packets for each tcIndex.

You can further configure user names, by assigning two specific tcNames to user names. One as upload and the other one as download direction. If this is configured, the output will further contain:
myOID.8 - tcUserIndexLeaf               - Stores integers, the SNMP indexes assigned to the configured user names.
myOID.9 - tcUserNumIndexLeaf            - Stores an integer, the count of indexes assigned to the configured user names.
myOID.10 - tcUserNameLeaf               - Stores strings, the names of the configured user names.
myOID.11 - tcUserDownBytesLeaf          - Stores counter32, the downloaded bytes for each tcUserIndex.
myOID.12 - tcUserDownPktLeaf            - Stores counter32, the downloaded packets for each tcUserIndex.
myOID.13 - tcUserDownDroppedPktLeaf     - Stores counter32, the dropped packets in download direction for each tcUserIndex.
myOID.14 - tcUserDownOverLimitPktLeaf   - Stores counter32, the over limit packets in download direction for each tcUserIndex.
myOID.15 - tcUserUpBytesLeaf            - Stores counter32, the uploaded bytes for each tcUserIndex.
myOID.16 - tcUserUpPktLeaf              - Stores counter32, the uploaded packets for each tcUserIndex.
myOID.17 - tcUserUpDroppedPktLeaf       - Stores counter32, the dropped packets in upload direction for each tcUserIndex.
myOID.18 - tcUserUpOverLimitPktLeaf     - Stores counter32, the over limit packets in upload direction for each tcUserIndex.

tc_reader reads configuration from file named tc_reader.conf
This configuration file should be located in one of these directories (sorted by order of preference):
1) ./tc_reader.conf (e.g the current working directory)
2) /etc/tc_reader.conf
If the configuration file cannot be located in any of these two locations, tc_reader will use its internal defaults, which probably isn't what you want.

Example output:
user@host:~# snmpwalk -v2c -c public localhost .1.3.6.1.4.1.2021.255
iso.3.6.1.4.1.2021.255.1 = STRING: "tcIndexLeaf"
iso.3.6.1.4.1.2021.255.1.1 = INTEGER: 1
iso.3.6.1.4.1.2021.255.1.2 = INTEGER: 2
iso.3.6.1.4.1.2021.255.2 = INTEGER: 2
iso.3.6.1.4.1.2021.255.3 = STRING: "tcNameLeaf"
iso.3.6.1.4.1.2021.255.3.1 = STRING: "eth0:1:0"
iso.3.6.1.4.1.2021.255.3.2 = STRING: "eth0:2:0"
iso.3.6.1.4.1.2021.255.4 = STRING: "sentBytesLeaf"
iso.3.6.1.4.1.2021.255.4.1 = Counter32: 1346379934
iso.3.6.1.4.1.2021.255.4.2 = Counter32: 1346379670
iso.3.6.1.4.1.2021.255.5 = STRING: "sentPktLeaf"
iso.3.6.1.4.1.2021.255.5.1 = Counter32: 13373109
iso.3.6.1.4.1.2021.255.5.2 = Counter32: 13373105
iso.3.6.1.4.1.2021.255.6 = STRING: "droppedPktLeaf"
iso.3.6.1.4.1.2021.255.6.1 = Counter32: 111
iso.3.6.1.4.1.2021.255.6.2 = Counter32: 111
iso.3.6.1.4.1.2021.255.7 = STRING: "overLimitPktLeaf"
iso.3.6.1.4.1.2021.255.7.1 = Counter32: 0
iso.3.6.1.4.1.2021.255.7.2 = Counter32: 0
iso.3.6.1.4.1.2021.255.8 = STRING: "tcUserIndexLeaf"
iso.3.6.1.4.1.2021.255.8.1 = INTEGER: 1
iso.3.6.1.4.1.2021.255.8.2 = INTEGER: 2
iso.3.6.1.4.1.2021.255.9 = INTEGER: 2
iso.3.6.1.4.1.2021.255.10 = STRING: "tcUserNameLeaf"
iso.3.6.1.4.1.2021.255.10.1 = STRING: "user1"
iso.3.6.1.4.1.2021.255.10.2 = STRING: "user2"
iso.3.6.1.4.1.2021.255.11 = STRING: "tcUserDownBytesLeaf"
iso.3.6.1.4.1.2021.255.11.1 = Counter32: 761936079
iso.3.6.1.4.1.2021.255.11.2 = Counter32: 515910479
iso.3.6.1.4.1.2021.255.12 = STRING: "tcUserDownPktLeaf"
iso.3.6.1.4.1.2021.255.12.1 = Counter32: 9214762
iso.3.6.1.4.1.2021.255.12.2 = Counter32: 3426902
iso.3.6.1.4.1.2021.255.13 = STRING: "tcUserDownDroppedPktLeaf"
iso.3.6.1.4.1.2021.255.13.1 = Counter32: 81459
iso.3.6.1.4.1.2021.255.13.2 = Counter32: 6546
iso.3.6.1.4.1.2021.255.14 = STRING: "tcUserDownOverLimitPktLeaf"
iso.3.6.1.4.1.2021.255.14.1 = Counter32: 0
iso.3.6.1.4.1.2021.255.14.2 = Counter32: 0
iso.3.6.1.4.1.2021.255.15 = STRING: "tcUserUpBytesLeaf"
iso.3.6.1.4.1.2021.255.15.1 = Counter32: 719988341
iso.3.6.1.4.1.2021.255.15.2 = Counter32: 200465101
iso.3.6.1.4.1.2021.255.16 = STRING: "tcUserUpPktLeaf"
iso.3.6.1.4.1.2021.255.16.1 = Counter32: 7076236
iso.3.6.1.4.1.2021.255.16.2 = Counter32: 2036062
iso.3.6.1.4.1.2021.255.17 = STRING: "tcUserUpDroppedPktLeaf"
iso.3.6.1.4.1.2021.255.17.1 = Counter32: 0
iso.3.6.1.4.1.2021.255.17.2 = Counter32: 0
iso.3.6.1.4.1.2021.255.18 = STRING: "tcUserUpOverLimitPktLeaf"
iso.3.6.1.4.1.2021.255.18.1 = Counter32: 0
iso.3.6.1.4.1.2021.255.18.2 = Counter32: 0
*/

package main

import (
	"fmt"
	"log/syslog"
	"os"
	"path/filepath"

	"github.com/mum4k/tc_reader/lib"
)

const (
	// syslogTag is the TAG used in syslog messages.
	syslogTag = "tc_reader"

	// configName is the file name that holds configuration for tc_reader.
	configName = "tc_reader.conf"

	// configPath is the defaut paths to the directory that contains the config file.
	configPath = "/etc"
)

// The exit codes.
const (
	exitOk = iota
	exitSyslogError
)

// main starts up tc_reader.
func main() {
	logger, err := syslog.New(syslog.LOG_INFO, syslogTag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: Cannot open connection to Syslog, err: %s", syslogTag, err)
		os.Exit(exitSyslogError)
	}

	// Try to load the config file.
	c, err := lib.NewConfig(configName)
	if err != nil {
		fileName := filepath.Join(configPath, configName)
		c, err = lib.NewConfig(fileName)
		if err != nil {
			logger.Info(fmt.Sprintf("Cannot locate tc_reader config file. Tried %s and %s. Using the defaults.", configName, fileName))
		}
	}

	// Configure the SNMP handler.
	so := &lib.SnmpOptions{
		Debug: c.Debug,
	}
	s := lib.NewSnmp(so, logger)

	// Configure the TC parser.
	tpo := &lib.TcParserOptions{
		TcCmdPath:     c.TcCmdPath,
		ParseInterval: c.ParseInterval,
		TcQdiscStats:  c.TcQdiscStats,
		TcClassStats:  c.TcClassStats,
		Ifaces:        c.Ifaces,
		UserNameClass: c.UserNameClass,
		Debug:         c.Debug,
	}
	lib.NewTcParser(tpo, s, logger)

	// Listen to commands from SNMP daemon.
	s.Listen()
	os.Exit(exitOk)
}
