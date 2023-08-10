package notification

import (
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	notifier := NewNotifier(WithText("Hi, there!"), WithSound("Submarine"))
	fmt.Println(notifier.Push())
	// notifier.Push(WithText("Hi again"))
}
