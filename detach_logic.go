package main

const detachMarkerEnv = "TRACKYOU_DETACHED"
const detachMarkerValue = "1"
const detachEnabledEnv = "TRACKYOU_AUTODETACH"
const detachEnabledValue = "1"

func shouldDetachForInteractiveLaunch(isInteractiveTTY bool, detachMarker, detachEnabled string) bool {
	return isInteractiveTTY && detachMarker == "" && detachEnabled == detachEnabledValue
}
