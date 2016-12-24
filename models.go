package main

import (
	"reflect"
)

type RAStatusRAW struct {
	RelayData      int `xml:"R"`
	RelayMaskOn    int `xml:"RON"`
	RelayMaskOff   int `xml:"ROFF"`
	ExpModule      int `xml:"EM"`
	ExpModule1     int `xml:"EM1"`
	RelayExpModule int `xml:"REM"`
	AlertFlags     int `xml:"AF"`
	StatusFlags    int `xml:"SF"`
}

type RAStatus struct {
	Source              string
	Id                  string `xml:"ID"`
	T1                  int    `xml:"T1"`
	T2                  int    `xml:"T2"`
	T3                  int    `xml:"T3"`
	PH                  int    `xml:"PH"`
	ATOLow              int    `xml:"ATOLOW"`
	ATOHigh             int    `xml:"ATOHIGH"`
	PWMActinic          int    `xml:"PWMA"`
	PWMDaylight         int    `xml:"PWMD"`
	PWMActinicOverride  int    `xml:"PWMAO"`
	PWMDaylightOverride int    `xml:"PWMDO"`
	WaterLevel          int    `xml:"WL"`
	BoardId             int    `xml:"BID"`
	Relays              map[string]RelayStatus
	ExpModules          *ExpModules
	RelaysExp           map[string]RelayStatus
	AlertFlag           *AlertFlag
	StatusFlag          *StatusFlag
}

type AlertFlag struct {
	ATOTimeOut int
	Overheat   int
	BusLock    int
	Leak       int
}

type StatusFlag struct {
	LightsOn    int
	Feeding     int
	WaterChange int
}

type ExpModules struct {
	PWMEbit   bool
	RFEbit    bool
	AIbit     bool
	Salbit    bool
	ORPbit    bool
	IObit     bool
	PHbit     bool
	WLbit     bool
	HUMbit    bool
	DCPumpbit bool
	Leakbit   bool
	PARbit    bool
	SCPWMbit  bool
}

func NewExpModules(n int) *ExpModules {
	expModules := ExpModules{}
	numField := reflect.ValueOf(expModules).NumField()
	v := reflect.ValueOf(&expModules)
	s := v.Elem()
	bits := parseIntToBits(n, numField)
	for i := 0; i < numField; i++ {
		field := s.Field(i)
		if field.CanSet() {
			field.SetBool(bits[i] > 0)
		} else {
		}
	}
	return &expModules
}

func ReadBits(n int, target interface{}, setBool bool) {
	v := reflect.ValueOf(target)
	s := v.Elem()
	numField := s.NumField()
	bits := parseIntToBits(n, numField)
	for i := 0; i < numField; i++ {
		field := s.Field(i)
		if field.CanSet() {
			if setBool {
				field.SetBool(bits[i] > 0)
			} else {
				field.SetInt(int64(bits[i]))
			}
		} else {
		}
	}
}

type RelayStatus struct {
	On      int
	MaskOn  int
	MaskOff int
	Name    string
}

func (r *RelayStatus) isOn() bool {
	return r.On > 0 || r.MaskOn > 0 && r.MaskOff > 0
}

var DefaultRelayNames = []string{"Relay1", "Relay2", "Relay3", "Relay4", "Relay5", "Relay6", "Relay7", "Relay8"}

func NewRelays(data, maskOn, maskOf int, names []string) map[string]RelayStatus {
	relayStatus := make(map[string]RelayStatus)
	relayData := parseIntToBits(data, 8)
	relayMaskOn := parseIntToBits(maskOn, 8)
	relayMaskOff := parseIntToBits(maskOf, 8)
	relayNames := DefaultRelayNames[0:]
	if names != nil {
		relayNames = names[0:]
	}
	for i := 0; i < 8; i++ {
		relayStatus[relayNames[i]] = RelayStatus{
			relayData[i],
			relayMaskOn[i],
			relayMaskOff[i],
			relayNames[i],
		}
	}
	return relayStatus
}

type InternalMemory struct {
}
