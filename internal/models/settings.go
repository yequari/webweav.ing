package models

import (
	"database/sql"
	"maps"
	"strconv"
	"time"
)

type SettingGroup int

const (
	SettingGroupUser      SettingGroup = 1
	SettingGroupGuestbook SettingGroup = 2
)

type SettingDataType int

const (
	SettingTypeInt      SettingDataType = 1
	SettingTypeDate     SettingDataType = 2
	SettingTypeAlphanum SettingDataType = 3
)

type SettingValue struct {
	Id                 int
	SettingId          int
	AllowedSettingId   int
	UnconstrainedValue string
}

type AllowedSettingValue struct {
	id        int
	itemValue string
	caption   string
}

func (s *AllowedSettingValue) Id() int {
	return s.id
}

func (s *AllowedSettingValue) ItemValue() string {
	return s.itemValue
}

func (s *AllowedSettingValue) Caption() string {
	return s.caption
}

type Setting struct {
	id            int
	description   string
	constrained   bool
	dataType      SettingDataType
	settingGroup  SettingGroup
	minValue      string // TODO: Maybe should be int?
	maxValue      string
	allowedValues map[int]AllowedSettingValue
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

func (s *Setting) AllowedValues() map[int]AllowedSettingValue {
	result := make(map[int]AllowedSettingValue, 50)
	maps.Copy(result, s.allowedValues)
	return result
}

func (s *Setting) ValidateUnconstrained(value string) bool {
	switch s.dataType {
	case SettingTypeInt:
		return s.validateInt(value)
	case SettingTypeAlphanum:
		return s.validateAlphanum(value)
	case SettingTypeDate:
		return s.validateDatetime(value)
	}
	return false
}

func (s *Setting) ValidateConstrained(value *AllowedSettingValue) bool {
	_, exists := s.allowedValues[value.id]
	if s.constrained && exists {
		return true
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
	v, err := time.Parse(time.DateTime, value)
	if err != nil {
		return false
	}
	var min time.Time
	var max time.Time

	if len(s.minValue) > 0 {
		min, err = time.Parse(time.DateTime, s.minValue)
		if err != nil {
			return false
		}
		if v.Before(min) {
			return false
		}
	}
	if len(s.maxValue) > 0 {
		max, err = time.Parse(time.DateTime, s.maxValue)
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
	return true
}

func validateSetting(db *sql.DB, settingId int, value string) (bool, error) {
	stmt := `SELECT s.Id, Description, Constrained, DataType, SettingGroup, MinValue, MaxValue, a.Id 
	FROM settings AS s LEFT JOIN allowed_setting_values AS a ON s.Id = a.SettingId AND a.ItemValue = ?
	WHERE s.Id = ?`
	result := db.QueryRow(stmt, value, settingId)

	var s Setting
	var minval sql.NullString
	var maxval sql.NullString
	var allowedId sql.NullInt64
	err := result.Scan(&s.id, &s.description, &s.constrained, &s.dataType, &s.settingGroup, &minval, &maxval, &allowedId)
	if err != nil {
		return false, err
	}
	if s.constrained && allowedId.Valid {
		return true, nil
	}
	switch s.dataType {
	case SettingTypeInt:
		return s.validateInt(value), nil
	case SettingTypeAlphanum:
		return s.validateAlphanum(value), nil
	case SettingTypeDate:
		return s.validateDatetime(value), nil
	}
	return false, nil
}
