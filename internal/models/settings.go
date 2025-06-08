package models

import (
	"strconv"
	"time"
)

type SettingGroup struct {
	id          int
	description string
}

func (g *SettingGroup) Id() int {
	return g.id
}

func (g *SettingGroup) Description() string {
	return g.description
}

const (
	SETTING_GROUP_USER      = "user"
	SETTING_GROUP_GUESTBOOK = "guestbook"
)

type SettingDataType struct {
	id          int
	description string
}

func (d *SettingDataType) Id() int {
	return d.id
}

func (d *SettingDataType) Description() string {
	return d.description
}

const (
	SETTING_TYPE_INTEGER = "integer"
	SETTING_TYPE_STRING  = "alphanumeric"
	SETTING_TYPE_DATE    = "datetime"
	SETTING_TYPE_BOOL    = "boolean"
)

type Setting struct {
	id           int
	description  string
	constrained  bool
	dataType     SettingDataType
	settingGroup SettingGroup
	minValue     string
	maxValue     string
}

func (s *Setting) Id() int {
	return s.id
}

func (s *Setting) Description() string {
	return s.description
}

func (s *Setting) Constrained() bool {
	return s.constrained
}

func (s *Setting) DataType() SettingDataType {
	return s.dataType
}

func (s *Setting) SettingGroup() SettingGroup {
	return s.settingGroup
}

func (s *Setting) MinValue() string {
	return s.minValue
}

func (s *Setting) MaxValue() string {
	return s.maxValue
}

func (s *Setting) Validate(value string) bool {
	switch s.dataType.description {
	case SETTING_TYPE_INTEGER:
		return s.validateInt(value)
	case SETTING_TYPE_STRING:
		return s.validateAlphanum(value)
	case SETTING_TYPE_DATE:
		return s.validateDatetime(value)
	case SETTING_TYPE_BOOL:
		return s.validateBool(value)
	}
	return false
}

func (s *Setting) validateInt(value string) bool {
	v, err := strconv.ParseInt(value, 10, 0)
	if err != nil {
		return false
	}
	var min int64
	var max int64
	if len(s.minValue) > 0 {
		min, err = strconv.ParseInt(s.minValue, 10, 0)
		if err != nil {
			return false
		}
		if v < min {
			return false
		}
	}
	if len(s.maxValue) > 0 {
		max, err = strconv.ParseInt(s.maxValue, 10, 0)
		if err != nil {
			return false
		}
		if v < max {
			return false
		}
	}
	return true
}

func (s *Setting) validateDatetime(value string) bool {
	v, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return false
	}
	var min time.Time
	var max time.Time

	if len(s.minValue) > 0 {
		min, err = time.Parse(time.RFC3339, s.minValue)
		if err != nil {
			return false
		}
		if v.Before(min) {
			return false
		}
	}
	if len(s.maxValue) > 0 {
		max, err = time.Parse(time.RFC3339, s.maxValue)
		if err != nil {
			return false
		}
		if v.After(max) {
			return false
		}
	}
	return false
}

func (s *Setting) validateAlphanum(value string) bool {
	return len(value) >= 0
}

func (s *Setting) validateBool(value string) bool {
	_, err := strconv.ParseBool(value)
	if err != nil {
		return false
	}
	return true
}
