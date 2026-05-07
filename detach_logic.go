package main

const detachMarkerEnv = "TRACKYOU_DETACHED"

func shouldDetachForInteractiveLaunch(isInteractiveTTY bool, detachMarker string) bool {
	return isInteractiveTTY && detachMarker != "1"
}
