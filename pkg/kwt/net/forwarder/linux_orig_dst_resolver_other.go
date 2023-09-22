//go:build !linux
// +build !linux

package forwarder

// Define dummy const so that LinuxOriginalDstResolver can be compiled on non-linux systems
const solIP = 0
