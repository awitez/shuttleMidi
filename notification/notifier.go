package notification

import (
	"errors"
	"fmt"
	"strings"
)

type Notifier struct {
	Title    string
	Subtitle string
	Text     string
	Sound    string
	cmdr     *commander
}

type Option func(n *Notifier)

func NewNotifier(opts ...Option) (notifier *Notifier) {
	notifier = &Notifier{
		cmdr:  newCommander(),
		Title: "Notifier",
	}
	for _, opt := range opts {
		opt(notifier)
	}
	return
}

func (n *Notifier) Push(opts ...Option) error {
	for _, opt := range opts {
		opt(n)
	}
	quote := func(s string) string {
		s = strings.Replace(s, `\`, `\\`, -1)
		s = strings.Replace(s, `'`, `"`, -1)
		s = strings.Replace(s, `"`, `\"`, -1)
		return s
	}
	elems := make([]string, 0, 5)
	elems = append(elems, "display notification")
	elems = append(elems, `"`+quote(n.Text)+`"`)
	elems = append(elems, `with title "`+quote(n.Title)+`"`)
	if len(n.Subtitle) != 0 {
		elems = append(elems, `subtitle "`+quote(n.Subtitle)+`"`)
	}
	if len(n.Sound) != 0 {
		elems = append(elems, `sound name "`+quote(n.Sound)+`"`)
	}

	output, err := n.cmdr.exec(`osascript -e '` + strings.Join(elems, " ") + `'`)
	if err != nil {
		return fmt.Errorf("%w: %s", err, output)
	}
	if len(output) != 0 {
		return errors.New(output)
	}
	return nil
}

func WithTitle(title string) Option {
	return func(n *Notifier) {
		n.Title = title
	}
}

func WithSubtitle(subtitle string) Option {
	return func(n *Notifier) {
		n.Subtitle = subtitle
	}
}

func WithText(text string) Option {
	return func(n *Notifier) {
		n.Text = text
	}
}

func WithSound(sound string) Option {
	return func(n *Notifier) {
		n.Sound = sound
	}
}
