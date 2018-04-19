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
*/

package lib

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func TestSnmpLogIfDebug(t *testing.T) {
	testData := []struct {
		debug   bool
		message string
		expInfo []string
	}{
		{false, "message", nil},
		{true, "message", []string{"message"}},
	}

	var o *SnmpOptions
	var s *snmp
	for i, params := range testData {
		fs := &fakeSyslog{}
		o = &SnmpOptions{
			Debug: params.debug,
		}
		s = &snmp{
			logger:  fs,
			options: o,
		}
		s.logIfDebug(params.message)
		if !reflect.DeepEqual(fs.info, params.expInfo) {
			t.Errorf("TestSnmpLogIfDebug(testCase %d) got: '%v' want: '%v'", i, fs.info, params.expInfo)
		}
	}
}

// compareSnmpData the received a map of OIDs to snmpData to the expected one, dereferencing pointers and making the output human readable.
func compareSnmpData(received map[string]*snmpData, expected map[string]snmpData) error {
	for k, v := range received {
		if expectedVal, ok := expected[k]; !ok {
			return fmt.Errorf("key '%v' in the result, but not in expected.", k)
		} else {
			if v == nil {
				return fmt.Errorf("key '%v' in result had a nil value.", k)
			}
			if *v != expectedVal {
				return fmt.Errorf("key '%v' had value '%v' in result, but '%v' in expected.", k, *v, expectedVal)
			}
		}
	}

	for k := range expected {
		if _, ok := received[k]; !ok {
			return fmt.Errorf("key '%v' found in expected, but not in the result.", k)
		}
	}
	return nil
}

func TestSnmpAddData(t *testing.T) {
	// These common OIDs are present in every test case.
	var commonOIDs map[string]snmpData = map[string]snmpData{
		".1.3.6.1.4.1.2021.255":    {".1.3.6.1.4.1.2021.255", "string", myName},
		".1.3.6.1.4.1.2021.255.1":  {".1.3.6.1.4.1.2021.255.1", "string", "tcIndexLeaf"},
		".1.3.6.1.4.1.2021.255.3":  {".1.3.6.1.4.1.2021.255.3", "string", "tcNameLeaf"},
		".1.3.6.1.4.1.2021.255.4":  {".1.3.6.1.4.1.2021.255.4", "string", "sentBytesLeaf"},
		".1.3.6.1.4.1.2021.255.5":  {".1.3.6.1.4.1.2021.255.5", "string", "sentPktLeaf"},
		".1.3.6.1.4.1.2021.255.6":  {".1.3.6.1.4.1.2021.255.6", "string", "droppedPktLeaf"},
		".1.3.6.1.4.1.2021.255.7":  {".1.3.6.1.4.1.2021.255.7", "string", "overLimitPktLeaf"},
		".1.3.6.1.4.1.2021.255.8":  {".1.3.6.1.4.1.2021.255.8", "string", "tcUserIndexLeaf"},
		".1.3.6.1.4.1.2021.255.10": {".1.3.6.1.4.1.2021.255.10", "string", "tcUserNameLeaf"},
		".1.3.6.1.4.1.2021.255.11": {".1.3.6.1.4.1.2021.255.11", "string", "tcUserDownBytesLeaf"},
		".1.3.6.1.4.1.2021.255.12": {".1.3.6.1.4.1.2021.255.12", "string", "tcUserDownPktLeaf"},
		".1.3.6.1.4.1.2021.255.13": {".1.3.6.1.4.1.2021.255.13", "string", "tcUserDownDroppedPktLeaf"},
		".1.3.6.1.4.1.2021.255.14": {".1.3.6.1.4.1.2021.255.14", "string", "tcUserDownOverLimitPktLeaf"},
		".1.3.6.1.4.1.2021.255.15": {".1.3.6.1.4.1.2021.255.15", "string", "tcUserUpBytesLeaf"},
		".1.3.6.1.4.1.2021.255.16": {".1.3.6.1.4.1.2021.255.16", "string", "tcUserUpPktLeaf"},
		".1.3.6.1.4.1.2021.255.17": {".1.3.6.1.4.1.2021.255.17", "string", "tcUserUpDroppedPktLeaf"},
		".1.3.6.1.4.1.2021.255.18": {".1.3.6.1.4.1.2021.255.18", "string", "tcUserUpOverLimitPktLeaf"},
	}

	testData := []struct {
		p                       []*parsedData
		expectedOIDData         map[string]snmpData
		expectedOIDs            []string
		expectedTcLastNameIndex int
		expectedNameToIndex     map[string]int
		expectedTcLastUserIndex int
		expectedUserToIndex     map[string]int
	}{

		// A test case without any parsed data - just testing erase().
		{
			nil,
			map[string]snmpData{},
			[]string{
				".1.3.6.1.4.1.2021.255",
				".1.3.6.1.4.1.2021.255.1",
				".1.3.6.1.4.1.2021.255.3",
				".1.3.6.1.4.1.2021.255.4",
				".1.3.6.1.4.1.2021.255.5",
				".1.3.6.1.4.1.2021.255.6",
				".1.3.6.1.4.1.2021.255.7",
				".1.3.6.1.4.1.2021.255.8",
				".1.3.6.1.4.1.2021.255.10",
				".1.3.6.1.4.1.2021.255.11",
				".1.3.6.1.4.1.2021.255.12",
				".1.3.6.1.4.1.2021.255.13",
				".1.3.6.1.4.1.2021.255.14",
				".1.3.6.1.4.1.2021.255.15",
				".1.3.6.1.4.1.2021.255.16",
				".1.3.6.1.4.1.2021.255.17",
				".1.3.6.1.4.1.2021.255.18",
			},
			0,
			map[string]int{},
			0,
			map[string]int{},
		},

		// A test case with single generic parsedData.
		{
			[]*parsedData{
				{"eth0:2:3", 1, 2, 3, 4, nil},
			},
			map[string]snmpData{
				".1.3.6.1.4.1.2021.255.1.1": {".1.3.6.1.4.1.2021.255.1.1", "integer", 1},
				".1.3.6.1.4.1.2021.255.2":   {".1.3.6.1.4.1.2021.255.2", "integer", 1},
				".1.3.6.1.4.1.2021.255.3.1": {".1.3.6.1.4.1.2021.255.3.1", "string", "eth0:2:3"},
				".1.3.6.1.4.1.2021.255.4.1": {".1.3.6.1.4.1.2021.255.4.1", "counter64", int64(1)},
				".1.3.6.1.4.1.2021.255.5.1": {".1.3.6.1.4.1.2021.255.5.1", "counter64", int64(2)},
				".1.3.6.1.4.1.2021.255.6.1": {".1.3.6.1.4.1.2021.255.6.1", "counter64", int64(3)},
				".1.3.6.1.4.1.2021.255.7.1": {".1.3.6.1.4.1.2021.255.7.1", "counter64", int64(4)},
			},
			[]string{
				".1.3.6.1.4.1.2021.255",
				".1.3.6.1.4.1.2021.255.1",
				".1.3.6.1.4.1.2021.255.1.1",
				".1.3.6.1.4.1.2021.255.2",
				".1.3.6.1.4.1.2021.255.3",
				".1.3.6.1.4.1.2021.255.3.1",
				".1.3.6.1.4.1.2021.255.4",
				".1.3.6.1.4.1.2021.255.4.1",
				".1.3.6.1.4.1.2021.255.5",
				".1.3.6.1.4.1.2021.255.5.1",
				".1.3.6.1.4.1.2021.255.6",
				".1.3.6.1.4.1.2021.255.6.1",
				".1.3.6.1.4.1.2021.255.7",
				".1.3.6.1.4.1.2021.255.7.1",
				".1.3.6.1.4.1.2021.255.8",
				".1.3.6.1.4.1.2021.255.10",
				".1.3.6.1.4.1.2021.255.11",
				".1.3.6.1.4.1.2021.255.12",
				".1.3.6.1.4.1.2021.255.13",
				".1.3.6.1.4.1.2021.255.14",
				".1.3.6.1.4.1.2021.255.15",
				".1.3.6.1.4.1.2021.255.16",
				".1.3.6.1.4.1.2021.255.17",
				".1.3.6.1.4.1.2021.255.18",
			},
			1,
			map[string]int{"eth0:2:3": 1},
			0,
			map[string]int{},
		},

		// A test case with single user parsedData (both upload and download).
		{
			[]*parsedData{
				{"eth0:2:3", 1, 2, 3, 4, &userClass{0, "username"}},
				{"eth1:2:3", 5, 6, 7, 8, &userClass{1, "username"}},
			},
			map[string]snmpData{
				".1.3.6.1.4.1.2021.255.8.1":  {".1.3.6.1.4.1.2021.255.8.1", "integer", 1},
				".1.3.6.1.4.1.2021.255.9":    {".1.3.6.1.4.1.2021.255.9", "integer", 1},
				".1.3.6.1.4.1.2021.255.10.1": {".1.3.6.1.4.1.2021.255.10.1", "string", "username"},
				".1.3.6.1.4.1.2021.255.11.1": {".1.3.6.1.4.1.2021.255.11.1", "counter64", int64(5)},
				".1.3.6.1.4.1.2021.255.12.1": {".1.3.6.1.4.1.2021.255.12.1", "counter64", int64(6)},
				".1.3.6.1.4.1.2021.255.13.1": {".1.3.6.1.4.1.2021.255.13.1", "counter64", int64(7)},
				".1.3.6.1.4.1.2021.255.14.1": {".1.3.6.1.4.1.2021.255.14.1", "counter64", int64(8)},
				".1.3.6.1.4.1.2021.255.15.1": {".1.3.6.1.4.1.2021.255.15.1", "counter64", int64(1)},
				".1.3.6.1.4.1.2021.255.16.1": {".1.3.6.1.4.1.2021.255.16.1", "counter64", int64(2)},
				".1.3.6.1.4.1.2021.255.17.1": {".1.3.6.1.4.1.2021.255.17.1", "counter64", int64(3)},
				".1.3.6.1.4.1.2021.255.18.1": {".1.3.6.1.4.1.2021.255.18.1", "counter64", int64(4)},
			},
			[]string{
				".1.3.6.1.4.1.2021.255",
				".1.3.6.1.4.1.2021.255.1",
				".1.3.6.1.4.1.2021.255.3",
				".1.3.6.1.4.1.2021.255.4",
				".1.3.6.1.4.1.2021.255.5",
				".1.3.6.1.4.1.2021.255.6",
				".1.3.6.1.4.1.2021.255.7",
				".1.3.6.1.4.1.2021.255.8",
				".1.3.6.1.4.1.2021.255.8.1",
				".1.3.6.1.4.1.2021.255.9",
				".1.3.6.1.4.1.2021.255.10",
				".1.3.6.1.4.1.2021.255.10.1",
				".1.3.6.1.4.1.2021.255.11",
				".1.3.6.1.4.1.2021.255.11.1",
				".1.3.6.1.4.1.2021.255.12",
				".1.3.6.1.4.1.2021.255.12.1",
				".1.3.6.1.4.1.2021.255.13",
				".1.3.6.1.4.1.2021.255.13.1",
				".1.3.6.1.4.1.2021.255.14",
				".1.3.6.1.4.1.2021.255.14.1",
				".1.3.6.1.4.1.2021.255.15",
				".1.3.6.1.4.1.2021.255.15.1",
				".1.3.6.1.4.1.2021.255.16",
				".1.3.6.1.4.1.2021.255.16.1",
				".1.3.6.1.4.1.2021.255.17",
				".1.3.6.1.4.1.2021.255.17.1",
				".1.3.6.1.4.1.2021.255.18",
				".1.3.6.1.4.1.2021.255.18.1",
			},
			0,
			map[string]int{},
			1,
			map[string]int{"username": 1},
		},

		// A test case with both generic and user parsedData (both upload and download).
		{
			[]*parsedData{
				{"eth0:2:3", 1, 2, 3, 4, &userClass{0, "username"}},
				{"eth1:2:3", 5, 6, 7, 8, &userClass{1, "username"}},
				{"eth0:1:3", 9, 10, 11, 12, nil},
			},
			map[string]snmpData{
				".1.3.6.1.4.1.2021.255.1.1":  {".1.3.6.1.4.1.2021.255.1.1", "integer", 1},
				".1.3.6.1.4.1.2021.255.2":    {".1.3.6.1.4.1.2021.255.2", "integer", 1},
				".1.3.6.1.4.1.2021.255.3.1":  {".1.3.6.1.4.1.2021.255.3.1", "string", "eth0:1:3"},
				".1.3.6.1.4.1.2021.255.4.1":  {".1.3.6.1.4.1.2021.255.4.1", "counter64", int64(9)},
				".1.3.6.1.4.1.2021.255.5.1":  {".1.3.6.1.4.1.2021.255.5.1", "counter64", int64(10)},
				".1.3.6.1.4.1.2021.255.6.1":  {".1.3.6.1.4.1.2021.255.6.1", "counter64", int64(11)},
				".1.3.6.1.4.1.2021.255.7.1":  {".1.3.6.1.4.1.2021.255.7.1", "counter64", int64(12)},
				".1.3.6.1.4.1.2021.255.8.1":  {".1.3.6.1.4.1.2021.255.8.1", "integer", 1},
				".1.3.6.1.4.1.2021.255.9":    {".1.3.6.1.4.1.2021.255.9", "integer", 1},
				".1.3.6.1.4.1.2021.255.10.1": {".1.3.6.1.4.1.2021.255.10.1", "string", "username"},
				".1.3.6.1.4.1.2021.255.11.1": {".1.3.6.1.4.1.2021.255.11.1", "counter64", int64(5)},
				".1.3.6.1.4.1.2021.255.12.1": {".1.3.6.1.4.1.2021.255.12.1", "counter64", int64(6)},
				".1.3.6.1.4.1.2021.255.13.1": {".1.3.6.1.4.1.2021.255.13.1", "counter64", int64(7)},
				".1.3.6.1.4.1.2021.255.14.1": {".1.3.6.1.4.1.2021.255.14.1", "counter64", int64(8)},
				".1.3.6.1.4.1.2021.255.15.1": {".1.3.6.1.4.1.2021.255.15.1", "counter64", int64(1)},
				".1.3.6.1.4.1.2021.255.16.1": {".1.3.6.1.4.1.2021.255.16.1", "counter64", int64(2)},
				".1.3.6.1.4.1.2021.255.17.1": {".1.3.6.1.4.1.2021.255.17.1", "counter64", int64(3)},
				".1.3.6.1.4.1.2021.255.18.1": {".1.3.6.1.4.1.2021.255.18.1", "counter64", int64(4)},
			},
			[]string{
				".1.3.6.1.4.1.2021.255",
				".1.3.6.1.4.1.2021.255.1",
				".1.3.6.1.4.1.2021.255.1.1",
				".1.3.6.1.4.1.2021.255.2",
				".1.3.6.1.4.1.2021.255.3",
				".1.3.6.1.4.1.2021.255.3.1",
				".1.3.6.1.4.1.2021.255.4",
				".1.3.6.1.4.1.2021.255.4.1",
				".1.3.6.1.4.1.2021.255.5",
				".1.3.6.1.4.1.2021.255.5.1",
				".1.3.6.1.4.1.2021.255.6",
				".1.3.6.1.4.1.2021.255.6.1",
				".1.3.6.1.4.1.2021.255.7",
				".1.3.6.1.4.1.2021.255.7.1",
				".1.3.6.1.4.1.2021.255.8",
				".1.3.6.1.4.1.2021.255.8.1",
				".1.3.6.1.4.1.2021.255.9",
				".1.3.6.1.4.1.2021.255.10",
				".1.3.6.1.4.1.2021.255.10.1",
				".1.3.6.1.4.1.2021.255.11",
				".1.3.6.1.4.1.2021.255.11.1",
				".1.3.6.1.4.1.2021.255.12",
				".1.3.6.1.4.1.2021.255.12.1",
				".1.3.6.1.4.1.2021.255.13",
				".1.3.6.1.4.1.2021.255.13.1",
				".1.3.6.1.4.1.2021.255.14",
				".1.3.6.1.4.1.2021.255.14.1",
				".1.3.6.1.4.1.2021.255.15",
				".1.3.6.1.4.1.2021.255.15.1",
				".1.3.6.1.4.1.2021.255.16",
				".1.3.6.1.4.1.2021.255.16.1",
				".1.3.6.1.4.1.2021.255.17",
				".1.3.6.1.4.1.2021.255.17.1",
				".1.3.6.1.4.1.2021.255.18",
				".1.3.6.1.4.1.2021.255.18.1",
			},
			1,
			map[string]int{"eth0:1:3": 1},
			1,
			map[string]int{"username": 1},
		},
	}

	// Not creating new snmp for every test case will also verify that erase() works.
	fs := &fakeSyslog{}
	o := &SnmpOptions{}
	s := &snmp{
		logger:  fs,
		options: o,
	}
	for i, params := range testData {
		s.lock()
		s.erase()
		for _, data := range params.p {
			s.addData(data)
		}
		s.unlock()

		// Add the common OIDs into the expected data.
		for k, v := range commonOIDs {
			params.expectedOIDData[k] = v
		}
		if err := compareSnmpData(s.oidData, params.expectedOIDData); err != nil {
			t.Errorf("TestSnmpAddData(testCase %d) snmpData: %s", i, err)
		}
		if !reflect.DeepEqual(s.oids, params.expectedOIDs) {
			t.Errorf("TestSnmpAddData(testCase %d) oids \n got: '%v', \nwant: '%v'.", i, s.oids, params.expectedOIDs)
		}
		if !reflect.DeepEqual(s.tcLastNameIndex, params.expectedTcLastNameIndex) {
			t.Errorf("TestSnmpAddData(testCase %d) tcLastNameIndex got: '%v', want: '%v'.", i, s.tcLastNameIndex, params.expectedTcLastNameIndex)
		}
		if !reflect.DeepEqual(s.nameToIndex, params.expectedNameToIndex) {
			t.Errorf("TestSnmpAddData(testCase %d) nameToIndex got: '%v', want: '%v'.", i, s.nameToIndex, params.expectedNameToIndex)
		}
		if !reflect.DeepEqual(s.tcLastUserIndex, params.expectedTcLastUserIndex) {
			t.Errorf("TestSnmpAddData(testCase %d) tcLastUserIndex got: '%v', want: '%v'.", i, s.tcLastUserIndex, params.expectedTcLastUserIndex)
		}
		if !reflect.DeepEqual(s.userToIndex, params.expectedUserToIndex) {
			t.Errorf("TestSnmpAddData(testCase %d) userToIndex got: '%v', want: '%v'.", i, s.userToIndex, params.expectedUserToIndex)
		}
	}
}

// testTalker implements snmpTalker and is used in tests.
type testTalker struct {
	// input is a list of strings that should be returned by getLine().
	input []string

	// output is the list of strings that were added by calling putLine()
	output []string
}

// getLine returns a single line from the predefined output.
func (tr *testTalker) getLine() string {
	if len(tr.input) < 1 {
		panic("getLine(): unexpected call to getLine, got no more input data.")
	}
	input := tr.input[0]
	tr.input = tr.input[1:]
	return input
}

// putLine stores a single line inside testTalker struct.
func (tr *testTalker) putLine(line string) {
	tr.output = append(tr.output, line)
}

// erase deletes all stored inputs and output.
func (tr *testTalker) erase() {
	tr.input = make([]string, 0)
	tr.output = make([]string, 0)
}

func TestSnmpListen(t *testing.T) {
	// Store some data.
	var p []*parsedData = []*parsedData{
		{"eth0:2:3", 1, 2, 3, 4, &userClass{0, "username"}},
		{"eth1:2:3", 5, 6, 7, 8, &userClass{1, "username"}},
		{"eth0:1:3", 9, 10, math.MaxInt32, math.MaxInt32 + 1, nil},
	}
	tr := &testTalker{}
	fs := &fakeSyslog{}
	o := &SnmpOptions{}
	s := &snmp{
		snmpTalker: tr,
		logger:     fs,
		options:    o,
	}
	s.lock()
	s.erase()
	for _, data := range p {
		s.addData(data)
	}
	s.unlock()

	testData := []struct {
		desc     string
		commands []string
		want     []string
	}{

		{
			desc:     "simple PING-PONG exchange",
			commands: []string{"PING", ""},
			want:     []string{"PONG"},
		},
		{
			desc:     "standard SNMP GET for a known OID - a string",
			commands: []string{"PING", "get", ".1.3.6.1.4.1.2021.255.3.1", ""},
			want:     []string{"PONG", ".1.3.6.1.4.1.2021.255.3.1", "string", "eth0:1:3"},
		},
		{
			desc:     "standard SNMP GET for a known OID - an integer",
			commands: []string{"PING", "get", ".1.3.6.1.4.1.2021.255.1.1", ""},
			want:     []string{"PONG", ".1.3.6.1.4.1.2021.255.1.1", "integer", "1"},
		},
		{
			desc:     "standard SNMP GET for a known OID - a counter64",
			commands: []string{"PING", "get", ".1.3.6.1.4.1.2021.255.4.1", ""},
			want:     []string{"PONG", ".1.3.6.1.4.1.2021.255.4.1", "counter64", "9"},
		},
		{
			desc:     "standard SNMP GET for a known OID - a counter64, and the value is at the math.MaxInt32",
			commands: []string{"PING", "get", ".1.3.6.1.4.1.2021.255.6.1", ""},
			want:     []string{"PONG", ".1.3.6.1.4.1.2021.255.6.1", "counter64", strconv.Itoa(math.MaxInt32)},
		},
		{
			desc:     "standard SNMP GET for a known OID - a counter64, and the value is larger than math.MaxInt32",
			commands: []string{"PING", "get", ".1.3.6.1.4.1.2021.255.7.1", ""},
			want:     []string{"PONG", ".1.3.6.1.4.1.2021.255.7.1", "counter64", strconv.Itoa(math.MaxInt32 + 1)},
		},
		{
			desc:     "standard SNMP GET for unknown OID",
			commands: []string{"PING", "get", ".1.3.7", ""},
			want:     []string{"PONG", ""},
		},
		{
			desc:     "standard SNMP GET-NEXT for two known OIDs",
			commands: []string{"PING", "getnext", ".1.3.6.1.4.1.2021.255.9", "getnext", ".1.3.6.1.4.1.2021.255.10", ""},
			want:     []string{"PONG", ".1.3.6.1.4.1.2021.255.10", "string", "tcUserNameLeaf", ".1.3.6.1.4.1.2021.255.10.1", "string", "username"},
		},
		{
			desc:     "standard SNMP GET-NEXT for the last OID",
			commands: []string{"PING", "getnext", ".1.3.6.1.4.1.2021.255.18.1", ""},
			want:     []string{"PONG", ""},
		},
		{
			desc:     "standard SNMP GET-NEXT for unknown OID",
			commands: []string{"PING", "getnext", ".1.3.7", ""},
			want:     []string{"PONG", ""},
		},
		{
			desc:     "unknown command",
			commands: []string{"PING", "set", ""},
			want:     []string{"PONG", ""},
		},
	}

	for _, tc := range testData {
		t.Run(tc.desc, func(t *testing.T) {
			tr.erase()
			tr.input = tc.commands
			s.Listen()
			if diff := pretty.Compare(tc.want, tr.output); diff != "" {
				t.Errorf("Listen => unexpected output, diff (-want, +got)\n%s", diff)
			}
		})
	}
}
