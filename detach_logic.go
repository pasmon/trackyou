package main

const detachMarkerEnv = "TRACKYOU_DETACHED"
const detachMarkerValue = "1"

// detachEnabledEnv enables opt-in auto-detach for interactive terminal
// launches when set to "1". Auto-detach stays disabled when unset or set
// to any other value.
const detachEnabledEnv = "TRACKYOU_AUTODETACH"
const detachEnabledValue = "1"

func shouldDetachForInteractiveLaunch(isInteractiveTTY bool, detachMarker, detachEnabled string) bool {
	return isInteractiveTTY && detachMarker == "" && detachEnabled == detachEnabledValue
}
