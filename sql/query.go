package gosql

import (
	"bytes"
	"fmt"
	. "github.com/suboat/go-sql-kit"
	"strings"
)

type ValueStringFunc func(*QueryValue) string

type SQLQuery struct {
	RuleMapping
	QueryRoot
	valueStringFunc ValueStringFunc
}

func NewSQLQuery() *SQLQuery {
	return new(SQLQuery).AllowCommonKey().SetValueFormat(nil)
}

func (s *SQLQuery) Reset() *SQLQuery {
	s.RuleMapping.Reset()
	return s.AllowCommonKey()
}

func (s *SQLQuery) AllowCommonKey() *SQLQuery {
	s.Allow(QueryKeyAnd, QueryKeyOr,
		QueryKeyEq, QueryKeyNe,
		QueryKeyLt, QueryKeyLte,
		QueryKeyGt, QueryKeyGte,
		QueryKeyIn, QueryKeyBetween, QueryKeyNotBetween,
	)
	return s
}

func (s *SQLQuery) SetValueFormat(f ValueStringFunc) *SQLQuery {
	if f != nil {
		s.valueStringFunc = f
	} else {
		s.valueStringFunc = s.ValueString
	}
	return s
}

func (s *SQLQuery) String(alias ...string) string {
	if s.Values == nil || len(s.Values) == 0 {
		return ""
	}
	set := make([]string, 0, len(s.Values))
	for _, iv := range s.Values {
		if v, ok := iv.(*QueryElem); ok {
			if str := s.elemString(v, alias...); len(str) != 0 {
				set = append(set, str)
			}
		}
	}
	if len(set) != 0 {
		return "WHERE " + strings.Join(set, " AND ")
	}
	return ""
}

func (s *SQLQuery) elemString(elem *QueryElem, alias ...string) string {
	if !s.IsAllowed(elem.Key) {
		return ""
	}
	set := make([]string, 0, len(elem.Values))
	for _, iv := range elem.Values {
		if v, ok := iv.(*QueryElem); ok {
			if str := s.elemString(v, alias...); len(str) != 0 {
				set = append(set, str)
			}
		} else if v, ok := iv.(*QueryValue); ok {
			if str := s.valueString(v, alias...); len(str) != 0 {
				set = append(set, str)
			}
		}
	}
	if len(set) == 0 {
		return ""
	} else if elem.IsAnonymous() {
		if len(set) == 1 {
			return set[0]
		} else if elem.RelKey == QueryKeyOr {
			return fmt.Sprintf("(%v)", strings.Join(set, " OR "))
		}
		return fmt.Sprintf("(%v)", strings.Join(set, " AND "))
	} else {
		switch elem.Key {
		case QueryKeyAnd:
			return strings.Join(set, " AND ")
		case QueryKeyOr:
			if len(set) == 1 {
				return set[0]
			}
			return fmt.Sprintf("(%v)", strings.Join(set, " OR "))
		}
	}
	return ""
}

func (s *SQLQuery) valueString(v *QueryValue, alias ...string) string {
	if v == nil {
	} else if !s.IsAllowed(v.Field) {
	} else if v.Field = s.GetMapping(v.Field); len(v.Field) != 0 {
		if f, ok := s.GetRuleMappingResult(v.Field); ok {
			if result, ok := f(v.Field, v.Value, v.Key, alias...); ok {
				if str, ok := result.(string); ok {
					return str
				}
			}
		}
		if f, ok := s.GetMappingFunc(v.Field); ok {
			if v.Field, v.Value, ok = f(v.Field, v.Value); !ok {
				return ""
			}
		}
		if len(alias) != 0 {
			v.Field = fmt.Sprintf("%v.%v", alias[0], v.Field)
		}
		return s.valueStringFunc(v)
	}
	return ""

}

func (s *SQLQuery) ValueString(v *QueryValue) string {
	opera := ""
	switch v.Key {
	case QueryKeyEq:
		opera = "="
	case QueryKeyNe:
		opera = "<>"
	case QueryKeyLt:
		opera = "<"
	case QueryKeyLte:
		opera = "<="
	case QueryKeyGt:
		opera = ">"
	case QueryKeyGte:
		opera = ">="
	case QueryKeyLike:
		return fmt.Sprintf("%v LIKE '%%%v%%'", v.Field, v.Value)
	case QueryKeyIn:
		if vs, ok := v.Value.([]interface{}); ok {
			if vs == nil || len(vs) == 0 {
				return ""
			}
			var sb bytes.Buffer
			sb.WriteString(fmt.Sprintf("%v IN (", v.Field))
			l := len(vs)
			for i, vi := range vs {
				switch vi.(type) {
				case int, int8, int16, int32, int64, float32, float64:
					sb.WriteString(fmt.Sprintf("%v", vi))
				default:
					sb.WriteString(fmt.Sprintf("'%v'", vi))
				}
				if i+1 < l {
					sb.WriteString(", ")
				}
			}
			sb.WriteString(")")
			return sb.String()
		}
		return ""
	case QueryKeyBetween:
		if vs, ok := v.Value.([]interface{}); ok {
			if vs == nil || len(vs) < 2 {
				return ""
			}
			switch vs[0].(type) {
			case int, int8, int16, int32, int64, float32, float64:
				return fmt.Sprintf("%v BETWEEN %v AND %v", v.Field, vs[0], vs[1])
			default:
				return fmt.Sprintf("%v BETWEEN '%v' AND '%v'", v.Field, vs[0], vs[1])
			}
		}
		return ""
	case QueryKeyNotBetween:
		if vs, ok := v.Value.([]interface{}); ok {
			if vs == nil || len(vs) < 2 {
				return ""
			}
			switch vs[0].(type) {
			case int, int8, int16, int32, int64, float32, float64:
				return fmt.Sprintf("%v NOT BETWEEN %v AND %v", v.Field, vs[0], vs[1])
			default:
				return fmt.Sprintf("%v NOT BETWEEN '%v' AND '%v'", v.Field, vs[0], vs[1])
			}
		}
		return ""
	default:
		return ""
	}
	switch v.Value.(type) {
	case int, int8, int16, int32, int64, float32, float64:
		return fmt.Sprintf("%v%v%v", v.Field, opera, v.Value)
	default:
		return fmt.Sprintf("%v%v'%v'", v.Field, opera, v.Value)
	}
}

func (s *SQLQuery) JSONtoSQLString(str string, alias ...string) (string, error) {
	if err := s.ParseJSONString(str); err != nil {
		return "", err
	}
	return s.String(alias...), nil
}

func (s *SQLQuery) SQLString(m map[string]interface{}, alias ...string) (string, error) {
	if err := s.Parse(m); err != nil {
		return "", err
	}
	return s.String(alias...), nil
}
