//go:build !darwin

package system

func CheckMicrophonePermission() string {
	return "granted"
}

func RequestMicrophonePermission() string {
	return "granted"
}
