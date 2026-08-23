//go:build darwin

// Package fsevents 是 fsnotify/fsevents 的最小编译期 stub。
//
// 背景：docker/compose 的 darwin 文件监听器（pkg/watch/watcher_darwin.go）
// 无条件引入 fsnotify/fsevents，而其上游实现依赖 cgo（macOS CoreServices
// 框架）。发布时以 CGO_ENABLED=0 交叉编译 darwin 目标，cgo 实现文件被排除，
// 上游 fsevents.go 引用的符号缺失，导致编译失败。
//
// 本项目不使用 compose 的 watch 功能（仅用其部署 API），因此用该 no-op
// stub 替换，仅满足编译期 API 契约，运行时不会被调用。
package fsevents

import "time"

// EventFlags 是 FSEventStreamEventFlags 位掩码。
type EventFlags uint32

// CreateFlags 是 FSEventStreamCreateFlags 位掩码。
type CreateFlags uint32

// 事件/创建标志（值与上游一致：映射 macOS CoreServices 常量）。
const (
	ItemCreated EventFlags = 1 << 8  // kFSEventStreamEventFlagItemCreated
	ItemIsDir   EventFlags = 1 << 17 // kFSEventStreamEventFlagItemIsDir

	FileEvents CreateFlags = 1 << 4 // kFSEventStreamCreateFlagFileEvents
	IgnoreSelf CreateFlags = 1 << 3 // kFSEventStreamCreateFlagIgnoreSelf
)

// Event 是单个文件系统事件。
type Event struct {
	ID    uint64
	Path  string
	Flags EventFlags
}

// EventStream 表示一个 FSEvents 事件流（no-op stub）。
type EventStream struct {
	Events  chan []Event
	Paths   []string
	Flags   CreateFlags
	Resume  bool
	EventID uint64
	Latency time.Duration
	Device  int32
}

// Start 启动事件流（stub：空操作）。
func (es *EventStream) Start() error { return nil }

// Stop 停止事件流（stub：空操作）。
func (es *EventStream) Stop() {}

// Flush 刷新事件（stub：空操作）。
func (es *EventStream) Flush(sync bool) {}

// Restart 重启事件流（stub：空操作）。
func (es *EventStream) Restart() error { return nil }

// LatestEventID 返回最新事件 ID（stub：返回 0）。
func LatestEventID() uint64 { return 0 }

// DeviceForPath 返回路径所在设备的设备号（stub：返回 0）。
func DeviceForPath(path string) (int32, error) { return 0, nil }
