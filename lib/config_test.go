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
	"reflect"
	"testing"
)

func TestConfig(t *testing.T) {
	testData := []struct {
		configFile            string
		expectedErr           string
		expectedTcCmdPath     string
		expectedParseInterval int
		expectedTcQdiscStats  []string
		expectedTcClassStats  []string
		expectedIfaces        []string
		expectedUserNameClass map[string]userClass
		expectedDebug         bool
	}{

		// A test case with completely valid config file.
		{
			"testdata/config_valid",
			"",
			"/sbin/tc",
			5,
			[]string{"-s", "qdisc", "show", "dev"},
			[]string{"-s", "class", "show", "dev"},
			[]string{"eth0", "eth1", "eth2"},
			map[string]userClass{
				"eth0:2:3": {uploadDirection, "user1"},
				"eth0:2:4": {uploadDirection, "user2"},
				"eth1:2:3": {downloadDirection, "user1"},
				"eth1:2:4": {downloadDirection, "user2"},
			},
			true,
		},

		// A test case with empty config file.
		{
			"testdata/config_empty",
			"",
			"",
			0,
			nil,
			nil,
			nil,
			nil,
			false,
		},

		// A test case with config file that does not exist.
		{
			"testdata/config_not_existing",
			"open testdata/config_not_existing: no such file or directory",
			"",
			0,
			nil,
			nil,
			nil,
			nil,
			false,
		},

		// A test case with config file that contains unexpected line.
		{
			"testdata/config_unexpected_line",
			"Error in config file testdata/config_unexpected_line on line 41: cannot parse this line: 'garbage'",
			"",
			0,
			nil,
			nil,
			nil,
			nil,
			false,
		},

		// A test case with config file that contains duplicate entry.
		{
			"testdata/config_duplicate",
			"Error in config file testdata/config_duplicate on line 42: found duplicate entry. Line: 'ifaces = \"eth0 eth1 eth2\"'",
			"",
			0,
			nil,
			nil,
			nil,
			nil,
			false,
		},
	}

	for i, params := range testData {
		c, err := NewConfig(params.configFile)
		if err != nil && !reflect.DeepEqual(err.Error(), params.expectedErr) {
			t.Errorf("TestConfig(testcase %d), err \n got: '%s', \nwant: '%s'", i, err.Error(), params.expectedErr)
		}
		if err == nil {
			if !reflect.DeepEqual(c.TcCmdPath, params.expectedTcCmdPath) {
				t.Errorf("TestConfig(testCase %d) TcCmdPath got: '%v' want: '%v'", i, c.TcCmdPath, params.expectedTcCmdPath)
			}
			if !reflect.DeepEqual(c.ParseInterval, params.expectedParseInterval) {
				t.Errorf("TestConfig(testCase %d) ParseInterval got: '%v' want: '%v'", i, c.ParseInterval, params.expectedParseInterval)
			}
			if !reflect.DeepEqual(c.TcQdiscStats, params.expectedTcQdiscStats) {
				t.Errorf("TestConfig(testCase %d) TcQdiscStats got: '%v' want: '%v'", i, c.TcQdiscStats, params.expectedTcQdiscStats)
			}
			if !reflect.DeepEqual(c.TcClassStats, params.expectedTcClassStats) {
				t.Errorf("TestConfig(testCase %d) TcClassStats got: '%v' want: '%v'", i, c.TcClassStats, params.expectedTcClassStats)
			}
			if !reflect.DeepEqual(c.Ifaces, params.expectedIfaces) {
				t.Errorf("TestConfig(testCase %d) Ifaces got: '%v' want: '%v'", i, c.Ifaces, params.expectedIfaces)
			}
			if !reflect.DeepEqual(c.UserNameClass, params.expectedUserNameClass) {
				t.Errorf("TestConfig(testCase %d) UserNameClass \n got: '%v' \nwant: '%v'", i, c.UserNameClass, params.expectedUserNameClass)
			}
			if !reflect.DeepEqual(c.Debug, params.expectedDebug) {
				t.Errorf("TestConfig(testCase %d) Debug got: '%v' want: '%v'", i, c.Debug, params.expectedDebug)
			}
		}
	}
}
