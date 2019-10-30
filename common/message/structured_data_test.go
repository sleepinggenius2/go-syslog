package message

import (
	"reflect"
	"testing"
)

func TestParseStructuredData(t *testing.T) {
	type args struct {
		in string
	}
	tests := []struct {
		name    string
		args    args
		want    StructuredData
		wantErr bool
	}{
		{"Empty input", args{``}, nil, false},
		{"Nil STRUCTURED-DATA", args{`-`}, nil, false},
		{"Invalid STRUCTURED-DATA", args{`a`}, nil, true},
		{"Empty SD-ELEMENT", args{`[]`}, nil, true},
		{"SD-ELEMENT without SD-PARAMs", args{`[id]`}, StructuredData{"id": SDParams{}}, false},
		{"Single SD-PARAM", args{`[id a="1"]`}, StructuredData{"id": SDParams{"a": "1"}}, false},
		{"Multiple SD-PARAMs", args{`[id a="1" b="2"]`}, StructuredData{"id": SDParams{"a": "1", "b": "2"}}, false},
		{"SD-ELEMENT missing opening bracket[", args{`id a="1"]`}, nil, true},
		{"SD-ELEMENT missing closing bracket", args{`[id a="1"`}, nil, true},
		{"SD-PARAM missing", args{`[id ]`}, nil, true},
		{"SD-PARAM missing opening quote", args{`[id a=1"]`}, nil, true},
		{"SD-PARAM missing closing quote", args{`[id a="1]`}, nil, true},
		{"PARAM-VALUE missing", args{`[id a]`}, nil, true},
		{"PARAM-VALUE missing with equals", args{`[id a=]`}, nil, true},
		{"PARAM-VALUE UTF-8", args{`[id a="☺"]`}, StructuredData{"id": SDParams{"a": "☺"}}, false},
		{"PARAM-VALUE escape quote", args{`[id a="\""]`}, StructuredData{"id": SDParams{"a": "\""}}, false},
		{"PARAM-VALUE escape backslash", args{`[id a="\\"]`}, StructuredData{"id": SDParams{"a": "\\"}}, false},
		{"PARAM-VALUE escape closing bracket", args{`[id a="\]"]`}, StructuredData{"id": SDParams{"a": "]"}}, false},
		{"PARAM-VALUE escape other", args{`[id a="\1"]`}, StructuredData{"id": SDParams{"a": "1"}}, false},
		{"PARAM-VALUE opening bracket", args{`[id a="["]`}, StructuredData{"id": SDParams{"a": "["}}, false},
		{"PARAM-VALUE equals", args{`[id a="="]`}, StructuredData{"id": SDParams{"a": "="}}, false},
		{"PARAM-VALUE escape equals", args{`[id a="\="]`}, StructuredData{"id": SDParams{"a": "="}}, false},
		{"SD-ID too long", args{`[123456789012345678901234567890123]`}, nil, true},
		{"PARAM-NAME too long", args{`[id 123456789012345678901234567890123="1"]`}, nil, true},
		{"Invalid quote", args{`[\"`}, nil, true},
		{"Invalid backslash", args{`[\`}, nil, true},
		{"Invalid SD-ID quote", args{`[id"`}, nil, true},
		{"Invalid SD-ID bracket", args{`[id[`}, nil, true},
		{"Invalid SD-ID equals", args{`[id=`}, nil, true},
		{"Invalid SD-ID space", args{`[ `}, nil, true},
		{"Invalid PARAM-NAME quote", args{`[id a"`}, nil, true},
		{"Invalid PARAM-NAME bracket", args{`[id a[`}, nil, true},
		{"Invalid PARAM-NAME space", args{`[id a `}, nil, true},
		{"Invalid SD-ELEMENT bracket", args{`[id a="b"[`}, nil, true},
		{"Invalid quote outside SD-ELEMENT", args{`[id]"`}, nil, true},
		{"Invalid bracket outside SD-ELEMENT", args{`[id]]`}, nil, true},
		{"Invalid equals outside SD-ELEMENT", args{`[id]=`}, nil, true},
		{"Invalid space outside SD-ELEMENT", args{`[id] `}, nil, true},
		{"Invalid character outside SD-ELEMENT", args{`[id]a`}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseStructuredData(tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseStructuredData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseStructuredData() = %v, want %v", got, tt.want)
			}
		})
	}
}
