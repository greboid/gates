package main

import "machine"

type Gate struct {
	Name              string
	openRequestInput  machine.Pin
	closedInput       machine.Pin
	closedOutput      machine.Pin
	openRequestOutput machine.Pin
	workingOutput     machine.Pin
	enabledInput      machine.Pin
	working           bool
}

type Controller struct {
	stuckStarted      bool
	stuckGate         *Gate
	singleGate        *Gate
	inbound           bool
	outbound          bool
	gate1Opened       bool
	gate1Closed       bool
	gate2Opened       bool
	gate2Closed       bool
	outerGate         *Gate
	innerGate         *Gate
	inboundCycling    machine.Pin
	outboundCycling   machine.Pin
	gate1             *Gate
	gate2             *Gate
	cycleTicks        int
	stuckRequestInput machine.Pin
	gatesOpenOutput   machine.Pin
	debug             machine.Pin
}
