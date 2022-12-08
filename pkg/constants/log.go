package constants

import "k8s.io/klog/v2"

// klog level ref: https://kingye.me/study-go/docs/advanced/pkg/klog/
const (
	LevelGeneral klog.Level = iota
	LevelReasonable
	LevelNormal
	LevelDebug klog.Level = 4
	LevelTrace
	LevelRequest
	LevelHttpRequest
)
