// Code generated by "stringer -type=Status"; DO NOT EDIT.

package common

import "strconv"

const _Status_name = "OKKODeactivated"

var _Status_index = [...]uint8{0, 2, 4, 15}

func (i Status) String() string {
	if i < 0 || i >= Status(len(_Status_index)-1) {
		return "Status(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Status_name[_Status_index[i]:_Status_index[i+1]]
}
