package main

import "machine"

//go:generate go run golang.org/x/tools/cmd/stringer@latest -type=State

const (
	InboundCycleStarted State = iota
	InboundCycleFirstWaiting
	InboundCycleFirstOpened
	InboundCycleFirstClosed
	InboundCycleSecondWaiting
	InboundCycleSecondOpened
	InboundCycleCompleted
	OutboundCycleStarted
	OutboundCycleFirstWaiting
	OutboundCycleFirstOpened
	OutboundCycleFirstClosed
	OutboundCycleSecondWaiting
	OutboundCycleSecondOpened
	OutboundCycleCompleted
	StuckCycleStarted
	StuckCycleOuterWaiting
	StuckCycleOuterOpened
	StuckCycleInnerWaiting
	StuckCycleInnerOpened
	StuckCycleComplete
	Idle
)

type (
	State int
	Door  struct {
		Open                bool
		Enabled             bool
		requestOpenPin      machine.Pin
		inputRequestOpenPin machine.Pin
		isOpenPin           machine.Pin
		isEnabledPin        machine.Pin
	}
)

type Context struct {
	State           State
	Inner           *Door
	Outer           *Door
	InboundRequest  bool
	OutboundRequest bool
	StuckRequest    bool
	Ticks           int
	inboundPin      machine.Pin
	outboundPin     machine.Pin
	stuckCyclePin   machine.Pin
	isClosedPin     machine.Pin
}
