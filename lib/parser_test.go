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

	"github.com/kylelemons/godebug/pretty"
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
	info []string

	// err is the message logged using the Err function call.
	err []string
}

func (fs *fakeSyslog) Info(m string) (err error) {
	fs.info = append(fs.info, m)
	return nil
}

func (fs *fakeSyslog) Err(m string) (err error) {
	fs.err = append(fs.err, m)
	return nil
}

func TestTcParserLogIfDebug(t *testing.T) {
	testData := []struct {
		debug   bool
		message string
		expInfo []string
	}{
		{false, "message", nil},
		{true, "message", []string{"message"}},
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
		if diff := pretty.Compare(params.expInfo, fs.info); diff != "" {
			t.Errorf("TestTcParserLogIfDebug(testCase %d) unexpected log, diff (-want, +got):\n%s", i, diff)
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

func TestTcParserParse(t *testing.T) {
	testData := []struct {
		desc            string
		qdiscOutputFile string
		classOutputFile string
		qdiscExecError  error
		classExecError  error
		userNameClass   map[string]userClass
		wantLog         []string
		want            []parsedData
		wantLockCount   int
		wantUnlockCount int
		wantEraseCount  int
	}{
		{
			desc:            "custom Qdiscs and Classes configured, no user names are configured",
			qdiscOutputFile: "testdata/tc_qdisc_custom",
			classOutputFile: "testdata/tc_class_custom",
			qdiscExecError:  nil,
			classExecError:  nil,
			userNameClass:   map[string]userClass{"1": {1, "username"}},
			want: []parsedData{
				{"eth0:1:0", 12548819, 124105, 13, 25, nil},
				{"eth0:2:0", 12548819, 24106, 128, 29, nil},
				{"eth0:a:0", 123432, 1027, 11, 2048, nil},
				{"eth0:6e:0", 9397865, 102745, 0, 0, nil},
				{"eth0:2:1", 931528, 9571, 127, 25, nil},
				{"eth0:2:2", 11630676, 114607, 13, 5211, nil},
				{"eth0:4:1", 11601665, 114364, 0, 0, nil},
				{"eth0:4:a", 1096857, 7059, 0, 0, nil},
				{"eth0:4:6e", 256, 13, 7, 0, nil},
			},
			wantLockCount:   1,
			wantUnlockCount: 1,
			wantEraseCount:  1,
		},
		{
			desc:            "large values are parsed correctly",
			qdiscOutputFile: "testdata/tc_qdisc_large_values",
			classOutputFile: "testdata/tc_class_large_values",
			qdiscExecError:  nil,
			classExecError:  nil,
			userNameClass:   map[string]userClass{"1": {1, "username"}},
			want: []parsedData{
				{"eth0:1:0", 4791659924490, 4791659924491, 4791659924492, 4791659924493, nil},
				{"eth0:2:1", 4791659924495, 4791659924496, 4791659924497, 4791659924498, nil},
			},
			wantLockCount:   1,
			wantUnlockCount: 1,
			wantEraseCount:  1,
		},
		{
			desc:            "custom Qdiscs and Classes configured, one user name is configured",
			qdiscOutputFile: "testdata/tc_qdisc_custom",
			classOutputFile: "testdata/tc_class_custom",
			qdiscExecError:  nil,
			classExecError:  nil,
			userNameClass: map[string]userClass{
				"eth0:4:1": {0, "username"},
				"eth0:4:a": {1, "username"},
			},
			want: []parsedData{
				{"eth0:1:0", 12548819, 124105, 13, 25, nil},
				{"eth0:2:0", 12548819, 24106, 128, 29, nil},
				{"eth0:a:0", 123432, 1027, 11, 2048, nil},
				{"eth0:6e:0", 9397865, 102745, 0, 0, nil},
				{"eth0:2:1", 931528, 9571, 127, 25, nil},
				{"eth0:2:2", 11630676, 114607, 13, 5211, nil},
				{"eth0:4:1", 11601665, 114364, 0, 0, nil},
				{"eth0:4:1", 11601665, 114364, 0, 0, &userClass{0, "username"}},
				{"eth0:4:a", 1096857, 7059, 0, 0, nil},
				{"eth0:4:a", 1096857, 7059, 0, 0, &userClass{1, "username"}},
				{"eth0:4:6e", 256, 13, 7, 0, nil},
			},
			wantLockCount:   1,
			wantUnlockCount: 1,
			wantEraseCount:  1,
		},
		{
			desc:            "the default Qdiscs and no classes",
			qdiscOutputFile: "testdata/tc_qdisc_default",
			classOutputFile: "testdata/tc_no_output",
			qdiscExecError:  nil,
			classExecError:  nil,
			userNameClass: map[string]userClass{
				"eth0:4:1":  {0, "username"},
				"eth0:4:10": {1, "username"},
			},
			want: []parsedData{
				{"eth0:0:0", 8214, 48, 0, 10, nil},
			},
			wantLockCount:   1,
			wantUnlockCount: 1,
			wantEraseCount:  1,
		},
		{
			desc:            "unable to execute the TC command",
			qdiscOutputFile: "testdata/tc_qdisc_custom",
			classOutputFile: "testdata/tc_class_custom",
			qdiscExecError:  fmt.Errorf("cannot execute"),
			classExecError:  nil,
			userNameClass: map[string]userClass{
				"eth0:4:1":  {0, "username"},
				"eth0:4:10": {1, "username"},
			},
			wantLog: []string{
				"parseTc(): Unable to get TC command output, error: cannot execute",
			},
			want:            []parsedData{},
			wantLockCount:   1,
			wantUnlockCount: 1,
			wantEraseCount:  1,
		},
		{
			desc:            "no output from the TC command",
			qdiscOutputFile: "testdata/tc_no_output",
			classOutputFile: "testdata/tc_no_output",
			qdiscExecError:  nil,
			classExecError:  nil,
			userNameClass: map[string]userClass{
				"eth0:4:1":  {0, "username"},
				"eth0:4:10": {1, "username"},
			},
			want:            []parsedData{},
			wantLockCount:   1,
			wantUnlockCount: 1,
			wantEraseCount:  1,
		},
	}

	var p *tcParser
	for _, tc := range testData {
		t.Run(tc.desc, func(t *testing.T) {
			fs := &fakeSyslog{}
			fsn := &fakeSnmp{}

			qdiscFile, err := ioutil.ReadFile(tc.qdiscOutputFile)
			if err != nil {
				t.Fatalf("ReadFile %s => unexpected err: %s", tc.qdiscOutputFile, err)
			}
			classFile, err := ioutil.ReadFile(tc.classOutputFile)
			if err != nil {
				t.Fatalf("ReadFile %s => unexpected err: %s", tc.classOutputFile, err)
			}
			var outputs []string = []string{string(qdiscFile), string(classFile)}
			var errors []error = []error{tc.qdiscExecError, tc.classExecError}

			o := &TcParserOptions{
				Ifaces:        []string{"eth0"},
				UserNameClass: tc.userNameClass,
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
			if !reflect.DeepEqual(fs.err, tc.wantLog) {
				t.Errorf("parseTc => wantLog got: '%v' want: '%v'", fs.err, tc.wantLog)
			}
			if diff := pretty.Compare(tc.want, fsn.data); diff != "" {
				t.Errorf("parseTc => unexpected data, diff(-want, +got):\n%s", diff)
			}
			if !reflect.DeepEqual(fsn.lockCount, tc.wantLockCount) {
				t.Errorf("parseTc => wantLockCount got: '%v' want: '%v'", fsn.lockCount, tc.wantLockCount)
			}
			if !reflect.DeepEqual(fsn.unlockCount, tc.wantUnlockCount) {
				t.Errorf("parseTc => wantUnlockCount got: '%v' want: '%v'", fsn.unlockCount, tc.wantUnlockCount)
			}
			if !reflect.DeepEqual(fsn.eraseCount, tc.wantEraseCount) {
				t.Errorf("parseTc => wantEraseCount got: '%v' want: '%v'", fsn.eraseCount, tc.wantEraseCount)
			}
		})
	}
}
