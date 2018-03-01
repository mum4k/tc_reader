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
	"sort"
	"testing"
)

func TestOidSorter(t *testing.T) {
	testData := []struct {
		oids     []string
		expected []string
	}{

		// A test case with no oids to sort.
		{
			[]string{},
			[]string{},
		},

		// A test case with multiple oids to sort.
		{
			[]string{
				".1.3.6.1.4.1.2021.255.13",
				".1.3.6.1.4.1.2021.255",
				".1.3.6.1.4.1.2021.255.1",
				".1.3.6.1.4.1.2021.255.3",
				".1.3.6.1.4.1.2021.255.10",
				".1.3.6.1.4.1.2021.255.5",
				".1.3.6.1.4.1.2021.255.6",
				".1.3.6.1.4.1.2021.255.7",
				".1.3.6.1.4.1.2021.255.8",
				".1.3.6.1.4.1.2021.255.10.1",
				".1.3.6.1.4.1.2021.255.11",
				".1.3.6.1.4.1.2021.255.11.1",
				".1.3.6.1.4.1.2021.255.12",
				".1.3.6.1.4.1.2021.255.12.1",
				".1.3.6.1.4.1.2021.255.8.1",
				".1.3.6.1.4.1.2021.255.4",
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
				".1.3.6",
				".1.3.6.1.4.1.2021.255.9",
				".1.3.254",
			},
			[]string{
				".1.3.6",
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
				".1.3.254",
			},
		},
	}

	fs := &fakeSyslog{}
	o := &SnmpOptions{}
	for i, params := range testData {
		s := &snmp{
			logger:  fs,
			options: o,
			oids:    params.oids,
		}
		sorter := &oidSorter{
			oids: &s.oids,
		}
		sort.Sort(sorter)
		if !reflect.DeepEqual(s.oids, params.expected) {
			t.Errorf("TestOidSorter(testCase %d) got: '%v' want: '%v'", i, s.oids, params.expected)
		}
	}
}
