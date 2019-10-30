package message

import (
	"github.com/pkg/errors"
)

type SDID string
type SDParams map[string]string
type StructuredData map[SDID]SDParams

func ParseStructuredData(in string) (StructuredData, error) {
	if len(in) == 0 || in == "-" {
		return nil, nil
	}
	if in[0] != '[' {
		return nil, errors.New("Invalid STRUCTURED-DATA")
	}
	out := make(StructuredData)
	var (
		cursor                                              int
		inElement, inID, inParam, inName, inValue, inEscape bool
		currID                                              SDID
		currName, currValue                                 string
	)
	for i, r := range in {
		switch r {
		case '\\':
			if !inValue {
				return nil, errors.New("Invalid '\\'")
			}
			if !inEscape {
				currValue += in[cursor:i]
				cursor = i + 1
			}
			inEscape = !inEscape
			continue
		case '[':
			switch {
			case inValue:
				break
			case inID:
				// Technically not to RFC spec?
				return nil, errors.New("SD-ID cannot contain '['")
			case inName:
				// Technically not to RFC spec?
				return nil, errors.New("PARAM-NAME cannot contain '['")
			case inElement:
				return nil, errors.New("Invalid '['")
			default:
				inElement = true
				inID = true
				cursor = i + 1
			}
		case ']':
			switch {
			case inEscape:
				break
			case inID && i > cursor:
				currID = SDID(in[cursor:i])
				out[currID] = make(SDParams)
				inID = false
				inElement = false
			case inValue && currName != "":
				return nil, errors.New("Must escape ']' inside of PARAM-VALUE")
			case inParam:
				if i == cursor {
					return nil, errors.New("Missing SD-PARAM")
				}
				if inName || currName != "" {
					return nil, errors.New("Missing PARAM-VALUE")
				}
				fallthrough
			case inElement:
				if i == cursor {
					return nil, errors.New("Empty SD-ELEMENT")
				}
				currID = ""
				inElement = false
			default:
				return nil, errors.New("Cannot have ']' outside of SD-ELEMENT")
			}
		case '=':
			switch {
			case inEscape || inValue:
				break
			case inName:
				currName = in[cursor:i]
				inName = false
				cursor = i + 1
			case inID:
				return nil, errors.New("SD-ID cannot contain '='")
			case !inElement:
				return nil, errors.New("Cannot have '=' outside of SD-ELEMENT")
			}
		case ' ':
			switch {
			case inID:
				if i == cursor {
					return nil, errors.New("Missing SD-ID")
				}
				currID = SDID(in[cursor:i])
				out[currID] = make(SDParams)
				inID = false
				inParam = true
				inName = true
				cursor = i + 1
			case inName:
				return nil, errors.New("PARAM-NAME cannot contain ' '")
			case inParam:
				inName = true
				cursor = i + 1
			case !inElement:
				return nil, errors.New("Cannot have ' ' outside of SD-ELEMENT")
			}
		case '"':
			switch {
			case inName:
				return nil, errors.New("PARAM-NAME cannot contain '\"'")
			case inID:
				return nil, errors.New("SD-ID cannot contain '\"'")
			case inEscape:
				break
			case inValue:
				out[currID][currName] = currValue + in[cursor:i]
				currName = ""
				currValue = ""
				inValue = false
			case inParam:
				inValue = true
				cursor = i + 1
			case !inElement:
				return nil, errors.New("Cannot have '\"' outside of SD-ELEMENT")
			}
		default:
			if !inElement {
				return nil, errors.New("Invalid character outside of SD-ELEMENT")
			}
			if inID && i-cursor == 32 {
				return nil, errors.New("SD-ID length must be <= 32")
			}
			if inName && i-cursor == 32 {
				return nil, errors.New("PARAM-NAME length must be <= 32")
			}
		}
		inEscape = false
	}
	if inElement {
		return nil, errors.New("Unterminated SD-ELEMENT")
	}
	return out, nil
}
