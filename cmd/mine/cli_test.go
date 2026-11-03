package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/TechieQuokka/minesweeper/internal/game"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		args   []string
		action action
		level  string
		err    string
	}{
		{nil, actionPlay, "medium", ""},
		{[]string{"easy"}, actionPlay, "easy", ""},
		{[]string{"medium"}, actionPlay, "medium", ""},
		{[]string{"hard"}, actionPlay, "hard", ""},
		{[]string{"help"}, actionHelp, "", ""},
		{[]string{"--version"}, actionVersion, "", ""},
		{[]string{"-V"}, actionVersion, "", ""},
		{[]string{"-h"}, actionError, "", `unknown argument: "-h"`},
		{[]string{"--help"}, actionError, "", `unknown argument: "--help"`},
		{[]string{"-v"}, actionError, "", `unknown argument: "-v"`},
		{[]string{"extreme"}, actionError, "", `unknown argument: "extreme"`},
		{[]string{"easy", "hard"}, actionError, "", "too many arguments"},
	}
	for _, tt := range tests {
		got := parseArgs(tt.args)
		if got.action != tt.action || got.err != tt.err {
			t.Errorf("parseArgs(%q) = %+v, want action %v err %q", tt.args, got, tt.action, tt.err)
		}
		if tt.level != "" && got.level.Name != tt.level {
			t.Errorf("parseArgs(%q) level = %q, want %q", tt.args, got.level.Name, tt.level)
		}
	}
}

func TestHandleNonPlay(t *testing.T) {
	var out, errOut bytes.Buffer

	code, exit := handleNonPlay(parseArgs([]string{"help"}), &out, &errOut)
	if code != 0 || !exit || !strings.HasPrefix(out.String(), "usage: mine [easy|medium|hard]") {
		t.Errorf("help: code %d exit %v out %q", code, exit, out.String())
	}

	out.Reset()
	code, exit = handleNonPlay(parseArgs([]string{"-V"}), &out, &errOut)
	if code != 0 || !exit || out.String() != "mine "+version+"\n" {
		t.Errorf("version: code %d exit %v out %q", code, exit, out.String())
	}

	out.Reset()
	code, exit = handleNonPlay(parseArgs([]string{"-h"}), &out, &errOut)
	want := "unknown argument: \"-h\"\nrun 'mine help' for usage\n"
	if code != 2 || !exit || errOut.String() != want || out.Len() != 0 {
		t.Errorf("error: code %d exit %v stderr %q", code, exit, errOut.String())
	}

	if _, exit := handleNonPlay(command{action: actionPlay, level: game.Easy}, &out, &errOut); exit {
		t.Error("play should not exit")
	}
}

func TestResultLine(t *testing.T) {
	if got := resultLine(true, "hard", 142); got != "cleared hard in 142s" {
		t.Errorf("win line = %q", got)
	}
	if got := resultLine(false, "hard", 87); got != "boom! hard, 87s" {
		t.Errorf("loss line = %q", got)
	}
}
