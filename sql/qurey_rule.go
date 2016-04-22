package gosql

import . "github.com/suboat/go-sql-kit"

type ValueFormatFunc func(string, string, interface{}) string

type SQLRule struct {
	RuleMapping
	ValueFormat ValueFormatFunc
}

func (s *SQLRule) SetValueFormat(f ValueFormatFunc) {
	s.ValueFormat = f
}

func (s *SQLRule) ValueString(v *QueryValue) string {
	if v == nil {
	} else if s.ValueFormat == nil {
	} else if !s.IsAllowed(v.Field) {
	} else {
		return s.ValueFormat(v.Key, s.GetMapping(v.Field), v.Value)
	}
	return ""
}
