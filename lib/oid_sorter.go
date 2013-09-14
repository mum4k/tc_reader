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


oid_sorter.go holds methods that sort OIDs in the order expected by SNMP daemon.

OIDs have to be sorter numerically even though they are stored as strings.
*/

package lib

import (
	"strconv"
	"strings"
)

// oidSorter is used to sort the OIDs. It implements the sort.Sort interface.
type oidSorter struct {
	oids *[]string
}

// Len determines the length of oids.
func (o *oidSorter) Len() int {
	return len(*o.oids)
}

// Swap swaps two oids.
func (o *oidSorter) Swap(i, j int) {
	(*o.oids)[i], (*o.oids)[j] = (*o.oids)[j], (*o.oids)[i]
}

// Less compares two oids.
func (o *oidSorter) Less(i, j int) bool {
	parts_first := strings.Split((*o.oids)[i], ".")
	parts_second := strings.Split((*o.oids)[j], ".")

	var inverse bool
	// Swap the parts if the first one is shorter. E.g. always compare the longer to the shorter.
	if len(parts_first) < len(parts_second) {
		parts_first, parts_second = parts_second, parts_first
		inverse = true
	}

	var int_second int64
	for x, value_first := range parts_first {
		int_first, _ := strconv.ParseInt(value_first, 10, 32)
		int_second = 0
		if len(parts_second) > x {
			value_second := parts_second[x]
			int_second, _ = strconv.ParseInt(value_second, 10, 32)
		}
		if int_first == int_second {
			continue
		} else {
			if inverse {
				return int_first > int_second
			} else {
				return int_first < int_second
			}
		}
	}
	return false
}
