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


snmp.go contains structs and methods used to store data in SNMP structures and talk to the NET-SNMP daemon.
*/

package lib

import (
	"bufio"
	"fmt"
	"log/syslog"
	"math"
	"os"
	"sort"
	"strconv"
	"sync"
)

// Package constants.
const (
	// emptyLine represents an empty line received on the input.
	emptyLine = ""

	// pingRequst si the string that SNMPD sends to us to figure out if we are alive.
	pingRequst = "PING"

	// pingResponse is the response that we should send to pingRequest.
	pingResponse = "PONG"

	// getCommand is the command that SNMPD sends on a GET request.
	getCommand = "get"

	// getNextCommand is the command that SNMPD sends on a GET-NEXT request.
	getNextCommand = "getnext"

	// myName is the identification of this process in the SNMP tree.
	myName = "tc_reader by mumak@"

	// myOID is the beginning of our OID tree.
	myOID = ".1.3.6.1.4.1.2021.255"

	// tcIndexLeaf is the SNMP leaf number where the list of indexes assigned to Qdiscs and Classes is stored.
	tcIndexLeaf = 1

	// tcNumIndexLeaf is the SNMP leaf number where the total number of available indexes is stored.
	tcNumIndexLeaf = 2

	// tcNameLeaf is the SNMP leaf number where names for indexes are stored. Names are in the form "eth0:2:1",
	// this means interface eth0, Qdisc 2, Class 1.
	tcNameLeaf = 3

	// sentBytesLeaf is the SNMP leaf number where sentBytes will be stored.
	sentBytesLeaf = 4

	// sentPktLeaf is the SNMP leaf number where sentPkt will be stored.
	sentPktLeaf = 5

	// droppedPktLeaf is the SNMP leaf number where droppedPkt will be stored.
	droppedPktLeaf = 6

	// overLimitPktLeaf is the SNMP leaf number where overLimitPkt will be stored.
	overLimitPktLeaf = 7

	// tcUserIndexLeaf is the SNMP leaf number where list of indexes assigned to users is stored.
	tcUserIndexLeaf = 8

	// tcUserNumIndexLeaf is the SNMP leaf number where the total number of available user indexes is stored.
	tcUserNumIndexLeaf = 9

	// tcUserNameLeaf is the SNMP leaf number where names for user indexes are stored.
	tcUserNameLeaf = 10

	// tcUserDownBytesLeaf is the SNMP leaf number where we store user bytes count in the download direction.
	tcUserDownBytesLeaf = 11

	// tcUserDownPktLeaf is the SNMP leaf number where we store user packet count in the download direction.
	tcUserDownPktLeaf = 12

	// tcUserDownDroppedPktLeaf is the SNMP leaf number where we store user dropped packet in the download direction.
	tcUserDownDroppedPktLeaf = 13

	// tcUserDownOverLimitPktLeaf is the SNMP leaf number where we store user overlimit packets in the download direction.
	tcUserDownOverLimitPktLeaf = 14

	// tcUserUpBytesLeaf is the SNMP leaf number where we store user bytes count in the upload direction.
	tcUserUpBytesLeaf = 15

	// tcUserUpPktLeaf is the SNMP leaf number where we store user packet count in the upload direction.
	tcUserUpPktLeaf = 16

	// tcUserUpDroppedPktLeaf is the SNMP leaf number where we store user dropped packet in the upload direction.
	tcUserUpDroppedPktLeaf = 17

	// tcUserUpOverLimitPktLeaf is the SNMP leaf number where we store user overlimit packets in the upload direction.
	tcUserUpOverLimitPktLeaf = 18
)

// The enumerated direction of traffic used in userClass.
const (
	uploadDirection = iota
	downloadDirection
)

// userClass stores the direction of traffic and the name of the user.
type userClass struct {
	direction int
	name      string
}

// snmpHandler stores parsed data from tcParser and and serves them to the SNMP daemon.
type snmpHandler interface {
	// lock should be called by the tcParser before it starts adding newly parsed data.
	lock()

	// unlock should be called by the tcParser after it finished adding parsed data. Note, call to this method will also sort internally stored OIDs.
	unlock()

	// erase clears out all stored data.
	erase()

	// addData adds parsed data.
	addData(data *parsedData)
}

// snmpTalker reads one line from an input.
type snmpTalker interface {
	// getLine returns a single line from the input.
	getLine() string

	// putLine writes a single line to the output.
	putLine(line string)
}

// stdinTalker implements snmpTalker.
type stdinTalker struct {
	// reader reads lines from standard input.
	reader *bufio.Reader
}

// getLine reads a single line from the standard input.
func (s *stdinTalker) getLine() string {
	line, isPrefix, err := s.reader.ReadLine()
	var additional []byte
	for isPrefix {
		additional, isPrefix, err = s.reader.ReadLine()
		if err != nil {
			return emptyLine
		}
		line = append(line, additional...)
	}
	if err != nil {
		return emptyLine
	}
	return string(line)
}

// putLine srites a single line to the standard output.
func (s *stdinTalker) putLine(line string) {
	fmt.Println(line)
}

// newStdinTalker creates new stdinTalker.
func newStdinTalker() *stdinTalker {
	s := &stdinTalker{
		reader: bufio.NewReader(os.Stdin),
	}
	return s
}

// parsedData is used to add parsed data by the tcParser.
type parsedData struct {
	// name is name of the handle, e.g: "eth0:2:3" means interface eth0, Qdisc 2 and Class 3.
	name string

	// sentBytes is the number of bytes that were sent out via this Qdisc / Class.
	sentBytes int64

	// sentPkt is the number of packets that were sent out via this Qdisc / Class.
	sentPkt int64

	// droppedPkt is the number of packets that were dropped out of this Qdisc / Class.
	droppedPkt int64

	// overLimitPkt is the number of packets that were over the configured limit of this Qdisc / Class.
	overLimitPkt int64

	// userClass if present indicates that this parsedData holds information for a configured user name and not just generic Qdisc / Class.
	userClass *userClass
}

// snmpData represents data stored in the SNMP tree.
type snmpData struct {
	// oid is the OID of this SNMP data.
	oid string

	// objectType is the SNMP object type, one of: integer, gauge, counter, timeticks, ipaddress, objectid, or string
	objectType string

	// objectValue is the value stored in this OID.
	objectValue interface{}
}

type SnmpOptions struct {
	// Debug determines whether we perform extensive logging to Syslog.
	Debug bool
}

// snmp implements snmpHandler.
type snmp struct {
	// l is the lock surrounding access to the stored data.
	l sync.Mutex

	// snmpTalker reads lines from standard input, this is how SNMP daemon talks to us.
	snmpTalker snmpTalker

	// logger is the Writer used to log messages to Syslog.
	logger sysLogger

	// oidData is a map of OIDs to the snmpData objects holding the parsed data from TC.
	oidData map[string]*snmpData

	// oids is an ordered list of OIDs with data from tcParser.
	oids []string

	// options holds the configurable options.
	options *SnmpOptions

	// tcLastNameIndex is the last assigned SNMP index to TC Queue / Class.
	tcLastNameIndex int

	// nameToIndex maps handle names to the assigned tcLastNameIndex.
	nameToIndex map[string]int

	// tcLastUserIndex is the last assigned SNMP index to an user name.
	tcLastUserIndex int

	// userToIndex maps user names to the assigned tcLastUserIndex.
	userToIndex map[string]int
}

// NewSnmp creates new snmp.
func NewSnmp(options *SnmpOptions, logger *syslog.Writer) *snmp {
	s := &snmp{
		snmpTalker: newStdinTalker(),
		logger:     logger,
		options:    options,
	}
	// Erase and initialize.
	s.erase()
	return s
}

// logIfDebug logs a message into Syslog if the debug option is set.
func (s *snmp) logIfDebug(message string) {
	if s.options.Debug {
		s.logger.Info(message)
	}
}

// lock locks access to the stored data, no read will be allowed. This is used while updating the stored data.
func (s *snmp) lock() {
	s.l.Lock()
}

// unlock releases the lock that disallows access to the stored data and sorts the stored OIDs to the order expected by the SNMP daemon.
func (s *snmp) unlock() {
	// Sort the OIDs so that the SNMP daemon does not bark at us ...
	s.sortOIDs()
	s.l.Unlock()
}

// erase removes all stored data. Lock should be acquired by the caller before calling erase.
func (s *snmp) erase() {
	// Erase any old data.
	s.oidData = make(map[string]*snmpData)
	s.oids = make([]string, 0)
	s.tcLastNameIndex = 0
	s.nameToIndex = make(map[string]int)
	s.tcLastUserIndex = 0
	s.userToIndex = make(map[string]int)

	// Identify ourselves.
	s.addSnmpData(myOID, "string", myName)

	// Identify the main parts of the output.
	s.addSnmpData(fmt.Sprintf("%s.%d", myOID, tcIndexLeaf), "string", "tcIndexLeaf")
	s.addSnmpData(fmt.Sprintf("%s.%d", myOID, tcNameLeaf), "string", "tcNameLeaf")
	s.addSnmpData(fmt.Sprintf("%s.%d", myOID, sentBytesLeaf), "string", "sentBytesLeaf")
	s.addSnmpData(fmt.Sprintf("%s.%d", myOID, sentPktLeaf), "string", "sentPktLeaf")
	s.addSnmpData(fmt.Sprintf("%s.%d", myOID, droppedPktLeaf), "string", "droppedPktLeaf")
	s.addSnmpData(fmt.Sprintf("%s.%d", myOID, overLimitPktLeaf), "string", "overLimitPktLeaf")
	s.addSnmpData(fmt.Sprintf("%s.%d", myOID, tcUserIndexLeaf), "string", "tcUserIndexLeaf")
	s.addSnmpData(fmt.Sprintf("%s.%d", myOID, tcUserNameLeaf), "string", "tcUserNameLeaf")
	s.addSnmpData(fmt.Sprintf("%s.%d", myOID, tcUserDownBytesLeaf), "string", "tcUserDownBytesLeaf")
	s.addSnmpData(fmt.Sprintf("%s.%d", myOID, tcUserDownPktLeaf), "string", "tcUserDownPktLeaf")
	s.addSnmpData(fmt.Sprintf("%s.%d", myOID, tcUserDownDroppedPktLeaf), "string", "tcUserDownDroppedPktLeaf")
	s.addSnmpData(fmt.Sprintf("%s.%d", myOID, tcUserDownOverLimitPktLeaf), "string", "tcUserDownOverLimitPktLeaf")
	s.addSnmpData(fmt.Sprintf("%s.%d", myOID, tcUserUpBytesLeaf), "string", "tcUserUpBytesLeaf")
	s.addSnmpData(fmt.Sprintf("%s.%d", myOID, tcUserUpPktLeaf), "string", "tcUserUpPktLeaf")
	s.addSnmpData(fmt.Sprintf("%s.%d", myOID, tcUserUpDroppedPktLeaf), "string", "tcUserUpDroppedPktLeaf")
	s.addSnmpData(fmt.Sprintf("%s.%d", myOID, tcUserUpOverLimitPktLeaf), "string", "tcUserUpOverLimitPktLeaf")
}

// addSnmpData adds data stored in snmpData struct.
func (s *snmp) addSnmpData(oid, objectType string, objectValue interface{}) {
	data := &snmpData{
		oid:         oid,
		objectType:  objectType,
		objectValue: objectValue,
	}
	s.oids = append(s.oids, oid)
	s.oidData[oid] = data
}

// addGenericData stores the data from parsedData as data for generic Qdisc / Class.
func (s *snmp) addGenericData(data *parsedData) {
	tcIndex, ok := s.nameToIndex[data.name]
	if !ok {
		// Popullate tcIndexLeaf.
		s.tcLastNameIndex += 1
		tcIndex = s.tcLastNameIndex
		s.nameToIndex[data.name] = tcIndex
		// Popullate tcIndexLeaf.
		tcIndexOID := fmt.Sprintf("%s.%d.%d", myOID, tcIndexLeaf, tcIndex)
		s.addSnmpData(tcIndexOID, "integer", tcIndex)

		// Popullate tcNameLeaf.
		tcNameOID := fmt.Sprintf("%s.%d.%d", myOID, tcNameLeaf, tcIndex)
		s.addSnmpData(tcNameOID, "string", data.name)

		// Popullate tcNumIndexLeaf.
		s.addSnmpData(fmt.Sprintf("%s.%d", myOID, tcNumIndexLeaf), "integer", s.tcLastNameIndex)
	}

	// Popullate sentBytesLeaf.
	tcSentBytesOID := fmt.Sprintf("%s.%d.%d", myOID, sentBytesLeaf, tcIndex)
	s.addSnmpData(tcSentBytesOID, "counter", data.sentBytes)

	// Popullate sentPktLeaf.
	tcSentPktOID := fmt.Sprintf("%s.%d.%d", myOID, sentPktLeaf, tcIndex)
	s.addSnmpData(tcSentPktOID, "counter", data.sentPkt)

	// Popullate droppedPktLeaf.
	tcDroppedPktOID := fmt.Sprintf("%s.%d.%d", myOID, droppedPktLeaf, tcIndex)
	s.addSnmpData(tcDroppedPktOID, "counter", data.droppedPkt)

	// Popullate overLimitPktLeaf.
	tcOverlimitPktOID := fmt.Sprintf("%s.%d.%d", myOID, overLimitPktLeaf, tcIndex)
	s.addSnmpData(tcOverlimitPktOID, "counter", data.overLimitPkt)
}

// addUserData stores the data from parsedData as data for a configured user name.
func (s *snmp) addUserData(data *parsedData) {
	// Create new index for this user if we don't have it already.
	tcUserIndex, ok := s.userToIndex[data.userClass.name]
	if !ok {
		// Popullate tcUserIndexLeaf.
		s.tcLastUserIndex += 1
		tcUserIndex = s.tcLastUserIndex
		s.userToIndex[data.userClass.name] = s.tcLastUserIndex
		tcUserIndexOID := fmt.Sprintf("%s.%d.%d", myOID, tcUserIndexLeaf, tcUserIndex)
		s.addSnmpData(tcUserIndexOID, "integer", tcUserIndex)

		// Popullate tcUserNameLeaf.
		tcUserNameOID := fmt.Sprintf("%s.%d.%d", myOID, tcUserNameLeaf, tcUserIndex)
		s.addSnmpData(tcUserNameOID, "string", data.userClass.name)

		// Export the number of user indexes.
		s.addSnmpData(fmt.Sprintf("%s.%d", myOID, tcUserNumIndexLeaf), "integer", s.tcLastUserIndex)
	}
	var tcUserBytesOID, tcUserPktOID, tcUserDroppedPktOID, tcUserOverLimitPktOID string
	switch data.userClass.direction {
	case uploadDirection:
		tcUserBytesOID = fmt.Sprintf("%s.%d.%d", myOID, tcUserUpBytesLeaf, tcUserIndex)
		tcUserPktOID = fmt.Sprintf("%s.%d.%d", myOID, tcUserUpPktLeaf, tcUserIndex)
		tcUserDroppedPktOID = fmt.Sprintf("%s.%d.%d", myOID, tcUserUpDroppedPktLeaf, tcUserIndex)
		tcUserOverLimitPktOID = fmt.Sprintf("%s.%d.%d", myOID, tcUserUpOverLimitPktLeaf, tcUserIndex)

	case downloadDirection:
		tcUserBytesOID = fmt.Sprintf("%s.%d.%d", myOID, tcUserDownBytesLeaf, tcUserIndex)
		tcUserPktOID = fmt.Sprintf("%s.%d.%d", myOID, tcUserDownPktLeaf, tcUserIndex)
		tcUserDroppedPktOID = fmt.Sprintf("%s.%d.%d", myOID, tcUserDownDroppedPktLeaf, tcUserIndex)
		tcUserOverLimitPktOID = fmt.Sprintf("%s.%d.%d", myOID, tcUserDownOverLimitPktLeaf, tcUserIndex)
	}
	// Popullate tcUser*BytesLeaf.
	if tcUserBytesOID != "" {
		s.addSnmpData(tcUserBytesOID, "counter", data.sentBytes)
	}

	// Popullate tcUser*PktLeaf.
	if tcUserPktOID != "" {
		s.addSnmpData(tcUserPktOID, "counter", data.sentPkt)
	}

	// Popullate tcUser*DroppedPktLeaf.
	if tcUserDroppedPktOID != "" {
		s.addSnmpData(tcUserDroppedPktOID, "counter", data.droppedPkt)
	}

	// Popullate tcUser*OverLimitPktLeaf.
	if tcUserOverLimitPktOID != "" {
		s.addSnmpData(tcUserOverLimitPktOID, "counter", data.overLimitPkt)
	}
}

// addData stores the content of parsedData so it can be served to the SNMP daemon.
func (s *snmp) addData(data *parsedData) {
	switch data.userClass {
	// The data holds information about a generic Qdisc / Class.
	case nil:
		s.addGenericData(data)

	// The data holds information about a configured user.
	default:
		s.addUserData(data)
	}
}

// snmpGet performs a SNMP get for the SNMP daemon.
func (s *snmp) snmpGet(oid string) {
	s.l.Lock()
	defer s.l.Unlock()

	if snmpData, ok := s.oidData[oid]; ok {
		s.printData(snmpData)
	} else {
		s.snmpTalker.putLine(emptyLine)
	}
}

// snmpGet performs a SNMP walk for the SNMP daemon.
func (s *snmp) snmpGetNext(oid string) {
	s.l.Lock()
	defer s.l.Unlock()

	// Do we have the requested OID?
	if _, ok := s.oidData[oid]; !ok {
		s.snmpTalker.putLine(emptyLine)
		return
	}

	var targetPosition int
	for i, storedOID := range s.oids {
		if oid == storedOID {
			// snmpGetNext should get the next value after the requested OID.
			targetPosition = i + 1
		}
	}

	// Do we have the next OID?
	nextPosition := targetPosition + 1
	if len(s.oids) >= nextPosition {
		requestedOID := s.oids[targetPosition]
		s.printData(s.oidData[requestedOID])
	} else {
		s.snmpTalker.putLine(emptyLine)
	}
}

// printData prints out data for a single OID in format understandable by the SNMP daemon.
func (s *snmp) printData(data *snmpData) {
	s.snmpTalker.putLine(data.oid)
	s.snmpTalker.putLine(data.objectType)

	switch objectType := data.objectType; objectType {
	case "string":
		if value, ok := data.objectValue.(string); !ok {
			s.snmpTalker.putLine(emptyLine)
		} else {
			s.snmpTalker.putLine(value)
		}
	case "counter":
		if value, ok := data.objectValue.(int64); !ok {
			s.snmpTalker.putLine(emptyLine)
		} else {
			// Unfortunatelly SNMP daemon does not support counter64 for pass_persist scripts yet. Need to rotate this around at math.MaxInt32.
			rotated := math.Mod(float64(value), float64(math.MaxInt32))
			s.snmpTalker.putLine(strconv.FormatInt(int64(rotated), 10))
		}
	case "integer":
		if value, ok := data.objectValue.(int); !ok {
			s.snmpTalker.putLine(emptyLine)
		} else {
			s.snmpTalker.putLine(strconv.FormatInt(int64(value), 10))
		}
	default:
		s.snmpTalker.putLine(emptyLine)
	}
}

// Start starts listening to commands from the SNMP daemon and performing the necessary actions.
func (s *snmp) Listen() {
	// We are persistent so this goes forever until we receive an empty command.
	for {
		switch command := s.snmpTalker.getLine(); command {
		case emptyLine:
			// emptyLine means that we should exit.
			s.logIfDebug("Listen(): received an empty line from the SNMP daemon, exiting ...")
			return

		case pingRequst:
			s.logIfDebug("Listen(): received a PING.")
			s.snmpTalker.putLine(pingResponse)

		case getCommand:
			oid := s.snmpTalker.getLine()
			s.logIfDebug(fmt.Sprintf("Listen(): processing SNMP GET for oid %s", oid))
			s.snmpGet(oid)

		case getNextCommand:
			oid := s.snmpTalker.getLine()
			s.logIfDebug(fmt.Sprintf("Listen(): processing SNMP GET-NEXT for oid %s", oid))
			s.snmpGetNext(oid)

		default:
			s.logger.Info(fmt.Sprintf("Listen(): got an unexpected command %s", command))
			s.snmpTalker.putLine(emptyLine)
		}

	}
}

// sortOIDs sorts the SNMP OIDs.
func (s *snmp) sortOIDs() {
	sorter := &oidSorter{
		oids: &s.oids,
	}
	sort.Sort(sorter)
}
