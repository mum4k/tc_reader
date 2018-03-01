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
	"io/ioutil"
	"reflect"
	"regexp"
	"testing"
)

func TestTcParserOptionsTcCmdPath(t *testing.T) {
	testData := []struct {
		in  string
		out string
	}{
		{"", tcCmdPath},
		{"/some/path", "/some/path"},
	}

	var o *TcParserOptions
	for i, params := range testData {
		o = &TcParserOptions{
			TcCmdPath: params.in,
		}
		if !reflect.DeepEqual(o.tcCmdPath(), params.out) {
			t.Errorf("TestTcParserOptionsTcCmdPath(testCase %d) got: '%v' want: '%v'", i, o.tcCmdPath(), params.out)
		}
	}
}

func TestTcParserOptionsParseInterval(t *testing.T) {
	testData := []struct {
		in  int
		out int
	}{
		{0, parseInterval},
		{13, 13},
	}

	var o *TcParserOptions
	for i, params := range testData {
		o = &TcParserOptions{
			ParseInterval: params.in,
		}
		if !reflect.DeepEqual(o.parseInterval(), params.out) {
			t.Errorf("TestTcParserOptionsParseInterval(testCase %d) got: '%v' want: '%v'", i, o.parseInterval(), params.out)
		}
	}
}

func TestTcParserOptionsTcQdiscStats(t *testing.T) {
	testData := []struct {
		in  []string
		out []string
	}{
		{nil, tcQdiscStats},
		{[]string{"1", "2"}, []string{"1", "2"}},
	}

	var o *TcParserOptions
	for i, params := range testData {
		o = &TcParserOptions{
			TcQdiscStats: params.in,
		}
		if !reflect.DeepEqual(o.tcQdiscStats(), params.out) {
			t.Errorf("TestTcParserOptionsTcQdiscStats(testCase %d) got: '%v' want: '%v'", i, o.tcQdiscStats(), params.out)
		}
	}
}

func TestTcParserOptionsTcClassStats(t *testing.T) {
	testData := []struct {
		in  []string
		out []string
	}{
		{nil, tcClassStats},
		{[]string{"1", "2"}, []string{"1", "2"}},
	}

	var o *TcParserOptions
	for i, params := range testData {
		o = &TcParserOptions{
			TcClassStats: params.in,
		}
		if !reflect.DeepEqual(o.tcClassStats(), params.out) {
			t.Errorf("TestTcParserOptionsTcClassStats(testCase %d) got: '%v' want: '%v'", i, o.tcClassStats(), params.out)
		}
	}
}

func TestTcParserOptionsIfaces(t *testing.T) {
	testData := []struct {
		in  []string
		out []string
	}{
		{nil, ifaces},
		{[]string{"1", "2"}, []string{"1", "2"}},
	}

	var o *TcParserOptions
	for i, params := range testData {
		o = &TcParserOptions{
			Ifaces: params.in,
		}
		if !reflect.DeepEqual(o.ifaces(), params.out) {
			t.Errorf("TestTcParserOptionsIfaces(testCase %d) got: '%v' want: '%v'", i, o.ifaces(), params.out)
		}
	}
}

func TestTcParserOptionsUserNameClass(t *testing.T) {
	testData := []struct {
		in  map[string]userClass
		out map[string]userClass
	}{
		{nil, map[string]userClass{}},
		{
			map[string]userClass{"1": {1, "username"}},
			map[string]userClass{"1": {1, "username"}},
		},
	}

	var o *TcParserOptions
	for i, params := range testData {
		o = &TcParserOptions{
			UserNameClass: params.in,
		}
		if !reflect.DeepEqual(o.userNameClass(), params.out) {
			t.Errorf("TestTcParserOptionsUserNameClass(testCase %d) got: '%v' want: '%v'", i, o.userNameClass(), params.out)
		}
	}
}

// fakeSyslog implements the sysLogger interface and is used in tests.
type fakeSyslog struct {
	// info is the message logged using the Info() function call.
	info string

	// err is the message logged using the Err function call.
	err string
}

func (fs *fakeSyslog) Info(m string) (err error) {
	fs.info = m
	return nil
}

func (fs *fakeSyslog) Err(m string) (err error) {
	fs.err = m
	return nil
}

func TestTcParserLogIfDebug(t *testing.T) {
	testData := []struct {
		debug   bool
		message string
		expInfo string
	}{
		{false, "message", ""},
		{true, "message", "message"},
	}

	var o *TcParserOptions
	var p *tcParser
	for i, params := range testData {
		fs := &fakeSyslog{}
		o = &TcParserOptions{
			Debug: params.debug,
		}
		p = &tcParser{
			logger:  fs,
			options: o,
		}
		p.logIfDebug(params.message)
		if !reflect.DeepEqual(fs.info, params.expInfo) {
			t.Errorf("TestTcParserLogIfDebug(testCase %d) got: '%v' want: '%v'", i, fs.info, params.expInfo)
		}
	}
}

// fakeExecuter implements the commandExecuter interface and is used in tests.
type fakeExecuter struct {
	// command are the paths to the commands that were executed.
	command []string

	// args are the command arguments used in function calls to Execute().
	args [][]string

	// output is the output that should be returned after the call to Execute(). First call will return the first entry and delete it.
	output []string

	// err is the error that should be returned after the call to Execute(). First call will return the first entry and delete it.
	err []error
}

func (fe *fakeExecuter) Execute(name string, arg ...string) (string, error) {
	fe.command = append(fe.command, name)
	fe.args = append(fe.args, arg)
	if len(fe.output) < 1 || len(fe.err) < 1 {
		panic("Execute(): unexpected call to Execute, got no more output data.")
	}
	output := fe.output[0]
	err := fe.err[0]
	fe.output = fe.output[1:]
	fe.err = fe.err[1:]
	return output, err
}

func TestTcParserExecuteTc(t *testing.T) {
	testData := []struct {
		output              []string
		err                 []error
		iface               string
		expectedCommand     []string
		expectedArgs        [][]string
		expectedQdiscOutput string
		expectedClassOutput string
		expectedErr         error
	}{
		// A test case where everything works.
		{
			[]string{"qdiscOutput", "classOutput"},
			[]error{nil, nil},
			"eth0",
			[]string{"/sbin/tc", "/sbin/tc"},
			[][]string{{"-s", "qdisc", "show", "dev", "eth0"}, {"-s", "class", "show", "dev", "eth0"}},
			"qdiscOutput",
			"classOutput",
			nil,
		},

		// A test case where the second command execution returns an Error.
		{
			[]string{"qdiscOutput", ""},
			[]error{nil, fmt.Errorf("an error")},
			"eth0",
			[]string{"/sbin/tc", "/sbin/tc"},
			[][]string{{"-s", "qdisc", "show", "dev", "eth0"}, {"-s", "class", "show", "dev", "eth0"}},
			"",
			"",
			fmt.Errorf("an error"),
		},

		// A test case where the first command execution returns an Error.
		{
			[]string{"qdiscOutput", "classOutput"},
			[]error{fmt.Errorf("an error"), nil},
			"eth0",
			[]string{"/sbin/tc"},
			[][]string{{"-s", "qdisc", "show", "dev", "eth0"}},
			"",
			"",
			fmt.Errorf("an error"),
		},
	}

	o := &TcParserOptions{}
	var p *tcParser
	var qdiscOutput, classOutput string
	var err error
	for i, params := range testData {
		fs := &fakeSyslog{}
		fe := &fakeExecuter{
			output: params.output,
			err:    params.err,
		}
		p = &tcParser{
			logger:   fs,
			options:  o,
			executer: fe,
		}
		qdiscOutput, classOutput, err = p.executeTc(params.iface)
		if !reflect.DeepEqual(qdiscOutput, params.expectedQdiscOutput) {
			t.Errorf("TestTcParseriExecuteTc(testCase %d) qdiscOutput got: '%v' want: '%v'", i, qdiscOutput, params.expectedQdiscOutput)
		}
		if !reflect.DeepEqual(classOutput, params.expectedClassOutput) {
			t.Errorf("TestTcParseriExecuteTc(testCase %d) classOutput got: '%v' want: '%v'", i, classOutput, params.expectedClassOutput)
		}
		if !reflect.DeepEqual(err, params.expectedErr) {
			t.Errorf("TestTcParseriExecuteTc(testCase %d) err got: '%v' want: '%v'", i, err, params.expectedErr)
		}
		if !reflect.DeepEqual(fe.command, params.expectedCommand) {
			t.Errorf("TestTcParseriExecuteTc(testCase %d) fe.command got: '%v' want: '%v'", i, fe.command, params.expectedCommand)
		}
		if !reflect.DeepEqual(fe.args, params.expectedArgs) {
			t.Errorf("TestTcParseriExecuteTc(testCase %d) fe.args got: '%v' want: '%v'", i, fe.args, params.expectedArgs)
		}
	}
}

// fakeSnmp implements snmpHandler interface and is used in tests.
type fakeSnmp struct {
	// lockCount is the number of times that lock() was called.
	lockCount int

	// unlockCount is the number of times that unlock() was called.
	unlockCount int

	// eraseCount is the number of times that erase() was called.
	eraseCount int

	// data contains the stored data added via addData().
	data []parsedData
}

func (fs *fakeSnmp) lock() {
	fs.lockCount += 1
}

func (fs *fakeSnmp) unlock() {
	fs.unlockCount += 1
}

func (fs *fakeSnmp) erase() {
	fs.eraseCount += 1
}

func (fs *fakeSnmp) addData(data *parsedData) {
	fs.data = append(fs.data, *data)
}

// formatParsedData formats parsedData to a human readable format.
func formatParsedData(d parsedData) string {
	var direction int
	var user string
	if d.userClass != nil {
		direction = d.userClass.direction
		user = d.userClass.name
	}
	return fmt.Sprintf("n:%s b:%d, p:%d, d:%d, o:%d, direction:%d, user:%s", d.name, d.sentBytes, d.sentPkt, d.droppedPkt, d.overLimitPkt, direction, user)
}

// compareParsedData compares the parsed data to the expected values element by element, dereferencing pointers and providing readable output.
func compareParsedData(received, expected []parsedData) error {
	if len(expected) < len(received) {
		unexpected := received[len(expected)+1]
		return fmt.Errorf("unexpected element in received data: %s", formatParsedData(unexpected))
	}

	for i, want := range expected {
		if len(received) < i+1 {
			return fmt.Errorf("received data does not contain element: %s", formatParsedData(want))
		}

		if !reflect.DeepEqual(want, received[i]) {
			return fmt.Errorf("difference in element %d\n got: %s\nwant: %s", i, formatParsedData(received[i]), formatParsedData(want))
		}
	}
	return nil

}

func TestTcParserParse(t *testing.T) {
	testData := []struct {
		qdiscOutputFile     string
		classOutputFile     string
		qdiscExecError      error
		classExecError      error
		userNameClass       map[string]userClass
		expectedLogErr      string
		expectedData        []parsedData
		expectedLockCount   int
		expectedUnlockCount int
		expectedEraseCount  int
	}{
		// A test case on a system that has custom Qdiscs and Classes configured. No user names are configured.
		{
			"testdata/tc_qdisc_custom",
			"testdata/tc_class_custom",
			nil,
			nil,
			map[string]userClass{"1": {1, "username"}},
			"",
			[]parsedData{
				{"eth0:1:0", 12548819, 124105, 13, 25, nil},
				{"eth0:2:0", 12548819, 24106, 128, 29, nil},
				{"eth0:10:0", 123432, 1027, 11, 2048, nil},
				{"eth0:110:0", 9397865, 102745, 0, 0, nil},
				{"eth0:2:1", 931528, 9571, 127, 25, nil},
				{"eth0:2:2", 11630676, 114607, 13, 5211, nil},
				{"eth0:4:1", 11601665, 114364, 0, 0, nil},
				{"eth0:4:10", 1096857, 7059, 0, 0, nil},
				{"eth0:4:110", 256, 13, 7, 0, nil},
			},
			1,
			1,
			1,
		},

		// A test case on a system that has custom Qdiscs and Classes configured. One user name is configured.
		{
			"testdata/tc_qdisc_custom",
			"testdata/tc_class_custom",
			nil,
			nil,
			map[string]userClass{
				"eth0:4:1":  {0, "username"},
				"eth0:4:10": {1, "username"},
			},
			"",
			[]parsedData{
				{"eth0:1:0", 12548819, 124105, 13, 25, nil},
				{"eth0:2:0", 12548819, 24106, 128, 29, nil},
				{"eth0:10:0", 123432, 1027, 11, 2048, nil},
				{"eth0:110:0", 9397865, 102745, 0, 0, nil},
				{"eth0:2:1", 931528, 9571, 127, 25, nil},
				{"eth0:2:2", 11630676, 114607, 13, 5211, nil},
				{"eth0:4:1", 11601665, 114364, 0, 0, nil},
				{"eth0:4:1", 11601665, 114364, 0, 0, &userClass{0, "username"}},
				{"eth0:4:10", 1096857, 7059, 0, 0, nil},
				{"eth0:4:10", 1096857, 7059, 0, 0, &userClass{1, "username"}},
				{"eth0:4:110", 256, 13, 7, 0, nil},
			},
			1,
			1,
			1,
		},

		// A test case on a system that has the default Qdiscs and no classes.
		{
			"testdata/tc_qdisc_default",
			"testdata/tc_no_output",
			nil,
			nil,
			map[string]userClass{
				"eth0:4:1":  {0, "username"},
				"eth0:4:10": {1, "username"},
			},
			"",
			[]parsedData{
				{"eth0:0:0", 8214, 48, 0, 10, nil},
			},
			1,
			1,
			1,
		},

		// A test case where we are unable to execute the TC command.
		{
			"testdata/tc_qdisc_custom",
			"testdata/tc_class_custom",
			fmt.Errorf("cannot execute"),
			nil,
			map[string]userClass{
				"eth0:4:1":  {0, "username"},
				"eth0:4:10": {1, "username"},
			},
			"parseTc(): Unable to get TC command output, error: cannot execute",
			[]parsedData{},
			1,
			1,
			1,
		},

		// A test case where we get no output from the TC command.
		{
			"testdata/tc_no_output",
			"testdata/tc_no_output",
			nil,
			nil,
			map[string]userClass{
				"eth0:4:1":  {0, "username"},
				"eth0:4:10": {1, "username"},
			},
			"",
			[]parsedData{},
			1,
			1,
			1,
		},
	}

	var p *tcParser
	for i, params := range testData {
		fs := &fakeSyslog{}
		fsn := &fakeSnmp{}

		qdiscFile, err := ioutil.ReadFile(params.qdiscOutputFile)
		if err != nil {
			t.Errorf("TestTcParserParse(testcase %d), ReadFile %s, err: %s", i, params.qdiscOutputFile, err)
		}
		classFile, err := ioutil.ReadFile(params.classOutputFile)
		if err != nil {
			t.Errorf("TestTcParserParse(testcase %d), ReadFile %s, err: %s", i, params.classOutputFile, err)
		}
		var outputs []string = []string{string(qdiscFile), string(classFile)}
		var errors []error = []error{params.qdiscExecError, params.classExecError}

		o := &TcParserOptions{
			Ifaces:        []string{"eth0"},
			UserNameClass: params.userNameClass,
		}
		fe := &fakeExecuter{
			output: outputs,
			err:    errors,
		}
		p = &tcParser{
			logger:        fs,
			options:       o,
			snmp:          fsn,
			executer:      fe,
			reQdiscHeader: regexp.MustCompile(reQdiscHeaderStr),
			reClassHeader: regexp.MustCompile(reClassHeaderStr),
			reStats:       regexp.MustCompile(reStatsStr),
		}
		p.parseTc()
		if !reflect.DeepEqual(fs.err, params.expectedLogErr) {
			t.Errorf("TestTcParserParse(testCase %d) expectedLogErr got: '%v' want: '%v'", i, fs.err, params.expectedLogErr)
		}
		err = compareParsedData(fsn.data, params.expectedData)
		if err != nil {
			t.Errorf("TestTcParserParse(testCase %d) expectedData: %s", i, err)
		}
		if !reflect.DeepEqual(fsn.lockCount, params.expectedLockCount) {
			t.Errorf("TestTcParserParse(testCase %d) expectedLockCount got: '%v' want: '%v'", i, fsn.lockCount, params.expectedLockCount)
		}
		if !reflect.DeepEqual(fsn.unlockCount, params.expectedUnlockCount) {
			t.Errorf("TestTcParserParse(testCase %d) expectedUnlockCount got: '%v' want: '%v'", i, fsn.unlockCount, params.expectedUnlockCount)
		}
		if !reflect.DeepEqual(fsn.eraseCount, params.expectedEraseCount) {
			t.Errorf("TestTcParserParse(testCase %d) expectedEraseCount got: '%v' want: '%v'", i, fsn.eraseCount, params.expectedEraseCount)
		}
	}
}
