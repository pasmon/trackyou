package main

const detachMarkerEnv = "TRACKYOU_DETACHED"
const detachMarkerValue = "1"

func shouldDetachForInteractiveLaunch(isInteractiveTTY bool, detachMarker string) bool {
	return isInteractiveTTY && detachMarker != detachMarkerValue
}
