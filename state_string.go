// Code generated by "stringer -type=State"; DO NOT EDIT.

package main

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[InboundCycleStarted-0]
	_ = x[InboundCycleFirstWaiting-1]
	_ = x[InboundCycleFirstOpened-2]
	_ = x[InboundCycleFirstClosed-3]
	_ = x[InboundCycleSecondWaiting-4]
	_ = x[InboundCycleSecondOpened-5]
	_ = x[InboundCycleCompleted-6]
	_ = x[OutboundCycleStarted-7]
	_ = x[OutboundCycleFirstWaiting-8]
	_ = x[OutboundCycleFirstOpened-9]
	_ = x[OutboundCycleFirstClosed-10]
	_ = x[OutboundCycleSecondWaiting-11]
	_ = x[OutboundCycleSecondOpened-12]
	_ = x[OutboundCycleCompleted-13]
	_ = x[StuckCycleStarted-14]
	_ = x[StuckCycleOuterWaiting-15]
	_ = x[StuckCycleOuterOpened-16]
	_ = x[StuckCycleInnerWaiting-17]
	_ = x[StuckCycleInnerOpened-18]
	_ = x[StuckCycleComplete-19]
	_ = x[Idle-20]
}

const _State_name = "InboundCycleStartedInboundCycleFirstWaitingInboundCycleFirstOpenedInboundCycleFirstClosedInboundCycleSecondWaitingInboundCycleSecondOpenedInboundCycleCompletedOutboundCycleStartedOutboundCycleFirstWaitingOutboundCycleFirstOpenedOutboundCycleFirstClosedOutboundCycleSecondWaitingOutboundCycleSecondOpenedOutboundCycleCompletedStuckCycleStartedStuckCycleOuterWaitingStuckCycleOuterOpenedStuckCycleInnerWaitingStuckCycleInnerOpenedStuckCycleCompleteIdle"

var _State_index = [...]uint16{0, 19, 43, 66, 89, 114, 138, 159, 179, 204, 228, 252, 278, 303, 325, 342, 364, 385, 407, 428, 446, 450}

func (i State) String() string {
	if i < 0 || i >= State(len(_State_index)-1) {
		return "State(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _State_name[_State_index[i]:_State_index[i+1]]
}