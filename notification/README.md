# macOS Notification
[![go doc](https://godoc.org/github.com/electricbubble/mac-notification?status.svg)](https://pkg.go.dev/github.com/electricbubble/mac-notification?tab=doc#pkg-index)
[![go report](https://goreportcard.com/badge/github.com/electricbubble/mac-notification)](https://goreportcard.com/report/github.com/electricbubble/mac-notification)
[![license](https://img.shields.io/github/license/electricbubble/mac-notification)](https://github.com/electricbubble/mac-notification/blob/master/LICENSE)

Display a notification (macOS)

![](https://cdn.jsdelivr.net/gh/electricbubble/ImageHosting/img/20201123160223.png)

## Installation

```shell script
go get github.com/electricbubble/mac-notification
```

## Usage

> The sound can be one of the files in `/System/Library/Sounds` or in `~/Library/Sounds`.

```go
package main

import (
	macNotification "github.com/electricbubble/mac-notification"
)

func main() {
	notifier := macNotification.NewNotifier(macNotification.WithText("Hi, there!"), macNotification.WithSound("Submarine"))
	notifier.Push()
	notifier.Push(macNotification.WithText("Hi again"))
	// notifier.Text = "hey"
	// notifier.Push()
}

```

## Thanks

Thank you [JetBrains](https://www.jetbrains.com/?from=gwda) for providing free open source licenses
