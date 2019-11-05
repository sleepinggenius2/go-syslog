package format

import (
	"reflect"
	"testing"
)

func Test_rfc5425ScannerSplit(t *testing.T) {
	type args struct {
		data  []byte
		atEOF bool
	}
	tests := []struct {
		name        string
		args        args
		wantAdvance int
		wantToken   []byte
		wantErr     bool
	}{
		{
			"No data",
			args{nil, true},
			0,
			nil,
			false,
		},
		{
			"Valid atEOF",
			args{[]byte(`99 <165>1 2003-08-24T05:14:15.000003-07:00 192.0.2.1 myproc 8710 - - %% It's time to make the do-nuts.`), true},
			102,
			[]byte(`<165>1 2003-08-24T05:14:15.000003-07:00 192.0.2.1 myproc 8710 - - %% It's time to make the do-nuts.`),
			false,
		},
		{
			"Valid not atEOF",
			args{[]byte(`99 <165>1 2003-08-24T05:14:15.000003-07:00 192.0.2.1 myproc 8710 - - %% It's time to make the do-nuts.`), false},
			102,
			[]byte(`<165>1 2003-08-24T05:14:15.000003-07:00 192.0.2.1 myproc 8710 - - %% It's time to make the do-nuts.`),
			false,
		},
		{
			"Not enough data atEOF",
			args{[]byte(`99 <165>1 2003-08-24T05:14:15.000003-07:00 192.0.2.1 myproc 8710 - - %% It's time to make the do-nuts`), true},
			0,
			nil,
			true,
		},
		{
			"Not enough data not atEOF",
			args{[]byte(`99 <165>1 2003-08-24T05:14:15.000003-07:00 192.0.2.1 myproc 8710 - - %% It's time to make the do-nuts`), false},
			0,
			nil,
			false,
		},
		{
			"Short input atEOF",
			args{[]byte{'9'}, true},
			0,
			nil,
			true,
		},
		{
			"Short input not atEOF",
			args{[]byte{'9'}, false},
			0,
			nil,
			false,
		},
		{
			"Invalid first digit less",
			args{[]byte{'0'}, false},
			0,
			nil,
			true,
		},
		{
			"Invalid first digit greater",
			args{[]byte{':'}, false},
			0,
			nil,
			true,
		},
		{
			"Invalid digit less",
			args{[]byte{'9', '/'}, false},
			0,
			nil,
			true,
		},
		{
			"Invalid digit greater",
			args{[]byte{'9', ':'}, false},
			0,
			nil,
			true,
		},
		{
			"Too large",
			args{[]byte(`65537`), false},
			0,
			nil,
			true,
		},
		{
			"Too large characters",
			args{[]byte(`999999`), false},
			0,
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAdvance, gotToken, err := rfc5425ScannerSplit(tt.args.data, tt.args.atEOF)
			if (err != nil) != tt.wantErr {
				t.Errorf("rfc5425ScannerSplit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotAdvance != tt.wantAdvance {
				t.Errorf("rfc5425ScannerSplit() gotAdvance = %v, want %v", gotAdvance, tt.wantAdvance)
			}
			if !reflect.DeepEqual(gotToken, tt.wantToken) {
				t.Errorf("rfc5425ScannerSplit() gotToken = %v, want %v", gotToken, tt.wantToken)
			}
		})
	}
}
