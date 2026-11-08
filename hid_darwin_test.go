package hid

import (
	"runtime"
	"strings"
	"testing"
)

func TestCFStringRoundTrip(t *testing.T) {
	for _, tc := range []struct {
		name string
		text string
	}{
		{name: "empty", text: ""},
		{name: "ASCII", text: "USB HID device"},
		{name: "Chinese", text: "设备名称"},
		{name: "emoji", text: "灯💡👩🏽‍💻"},
		{name: "embedded NUL", text: "before\x00after"},
		{name: "Unicode and NUL", text: "设备\x00💡"},
		{name: "combining marks", text: "Cafe\u0301"},
		{name: "long", text: strings.Repeat("设备💡", 1000)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfs := cfstring(tc.text)
			if cfs == 0 {
				t.Fatal("CFString creation failed")
			}
			got := gostring(cfs)
			releaseCFString(cfs)
			// The returned Go string must remain valid after its CF source is released.
			runtime.GC()
			if got != tc.text {
				t.Errorf("round trip = %q, want %q", got, tc.text)
			}
		})
	}
}

func TestNilCFString(t *testing.T) {
	if got := gostring(0); got != "" {
		t.Errorf("nil CFString = %q, want empty string", got)
	}
	releaseCFString(0)
}
