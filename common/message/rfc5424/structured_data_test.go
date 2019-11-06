package rfc5424

import (
	"reflect"
	"testing"

	"github.com/sleepinggenius2/go-syslog/common/message"
)

func TestParseStructuredData(t *testing.T) {
	type args struct {
		buff string
	}
	tests := []struct {
		name    string
		args    args
		want    message.StructuredData
		wantErr bool
	}{
		{"Empty input", args{``}, nil, false},
		{"Nil STRUCTURED-DATA", args{`-`}, nil, false},
		{"Invalid STRUCTURED-DATA", args{`a`}, nil, true},
		{"Invalid Nil STRUCTURED-DATA", args{`-a`}, nil, true},
		{"Empty SD-ELEMENT", args{`[]`}, nil, true},
		{"SD-ELEMENT without SD-PARAMs", args{`[id]`}, message.StructuredData{"id": message.SDParams{}}, false},
		{"Single SD-PARAM", args{`[id a="1"]`}, message.StructuredData{"id": message.SDParams{"a": "1"}}, false},
		{"Multiple SD-PARAMs", args{`[id a="1" b="2"]`}, message.StructuredData{"id": message.SDParams{"a": "1", "b": "2"}}, false},
		{"Space after SD-ELEMENT", args{`[id] `}, message.StructuredData{"id": message.SDParams{}}, false},
		{"SD-ELEMENT missing opening bracket[", args{`id a="1"]`}, nil, true},
		{"SD-ELEMENT missing closing bracket", args{`[id a="1"`}, nil, true},
		{"SD-PARAM missing", args{`[id ]`}, nil, true},
		{"SD-PARAM missing opening quote", args{`[id a=1"]`}, nil, true},
		{"SD-PARAM missing closing quote", args{`[id a="1]`}, nil, true},
		{"PARAM-VALUE missing", args{`[id a]`}, nil, true},
		{"PARAM-VALUE missing with equals", args{`[id a=]`}, nil, true},
		{"PARAM-VALUE missing multi", args{`[id a= b="2"]`}, nil, true},
		{"PARAM-VALUE UTF-8", args{`[id a="☺"]`}, message.StructuredData{"id": message.SDParams{"a": "☺"}}, false},
		{"PARAM-VALUE escape quote", args{`[id a="\""]`}, message.StructuredData{"id": message.SDParams{"a": `"`}}, false},
		{"PARAM-VALUE escape backslash", args{`[id a="\\"]`}, message.StructuredData{"id": message.SDParams{"a": `\`}}, false},
		{"PARAM-VALUE escape closing bracket", args{`[id a="\]"]`}, message.StructuredData{"id": message.SDParams{"a": `]`}}, false},
		{"PARAM-VALUE escape other", args{`[id a="\1"]`}, message.StructuredData{"id": message.SDParams{"a": `\1`}}, false},
		{"PARAM-VALUE escape multiple", args{`[id a="\"\]"]`}, message.StructuredData{"id": message.SDParams{"a": `"]`}}, false},
		{"PARAM-VALUE opening bracket", args{`[id a="["]`}, message.StructuredData{"id": message.SDParams{"a": `[`}}, false},
		{"PARAM-VALUE equals", args{`[id a="="]`}, message.StructuredData{"id": message.SDParams{"a": `=`}}, false},
		{"PARAM-VALUE escape equals", args{`[id a="\="]`}, message.StructuredData{"id": message.SDParams{"a": `\=`}}, false},
		{"PARAM-VALUE space", args{`[id a=" "]`}, message.StructuredData{"id": message.SDParams{"a": ` `}}, false},
		{"PARAM-VALUE escape space", args{`[id a="\ "]`}, message.StructuredData{"id": message.SDParams{"a": `\ `}}, false},
		{"SD-ID too long", args{`[123456789012345678901234567890123]`}, nil, true},
		{"PARAM-NAME too long", args{`[id 123456789012345678901234567890123="1"]`}, nil, true},
		{"Invalid quote", args{`[\"`}, nil, true},
		{"Invalid backslash", args{`[\`}, nil, true},
		{"Invalid SD-ID quote", args{`[id"`}, nil, true},
		{"Invalid SD-ID bracket", args{`[id[`}, nil, true},
		{"Invalid SD-ID equals", args{`[id=`}, nil, true},
		{"Invalid SD-ID space", args{`[ `}, nil, true},
		{"Invalid SD-ID not PRINTUSASCII", args{`[	`}, nil, true},
		{"Invalid PARAM-NAME quote", args{`[id a"`}, nil, true},
		{"Invalid PARAM-NAME bracket", args{`[id a[`}, nil, true},
		{"Invalid PARAM-NAME space", args{`[id a `}, nil, true},
		{"Invalid PARAM-NAME not PRINTUSASCII", args{`[id a	`}, nil, true},
		{"Invalid SD-ELEMENT bracket", args{`[id a="b"[`}, nil, true},
		{"Invalid quote outside SD-ELEMENT", args{`[id]"`}, nil, true},
		{"Invalid bracket outside SD-ELEMENT", args{`[id]]`}, nil, true},
		{"Invalid equals outside SD-ELEMENT", args{`[id]=`}, nil, true},
		{"Invalid character outside SD-ELEMENT", args{`[id]a`}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cursor int
			got, err := parseStructuredData([]byte(tt.args.buff), &cursor, len(tt.args.buff))
			if (err != nil) != tt.wantErr {
				t.Errorf("parseStructuredData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseStructuredData() = %v, want %v", got, tt.want)
			}
		})
	}
}
