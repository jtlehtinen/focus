package focus

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/urfave/cli/v2"
)

type sessionType int

type countdown struct {
	t int
	m int
	s int
}

const (
	pomodoro sessionType = iota
	shortBreak
	longBreak
)

type sessionStatus string

const (
	STARTED   sessionStatus = "STARTED"
	STOPPED   sessionStatus = "STOPPED"
	COMPLETED sessionStatus = "COMPLETED"
	SKIPPED   sessionStatus = "SKIPPED"
)

type event struct {
	session         sessionType
	status          sessionStatus
	duration        int
	startTime       time.Time
	expectedEndTime time.Time
	actualEndTime   time.Time
}

type kind map[sessionType]int

type message map[sessionType]string

type timer struct {
	currentSession    sessionType
	kind              kind
	autoStartPomodoro bool
	autoStartBreak    bool
	longBreakInterval int
	maxPomodoros      int
	iteration         int
	msg               message
	showNotification  bool
}

// nextSession retrieves the next session.
func (t *timer) nextSession() sessionType {
	var next sessionType

	switch t.currentSession {
	case pomodoro:
		if t.iteration == t.longBreakInterval {
			next = longBreak
		} else {
			next = shortBreak
		}
	case shortBreak, longBreak:
		next = pomodoro
	}

	return next
}

// getTimeRemaining subtracts the endTime from the currentTime
// and returns the total number of hours, minutes and seconds
// left.
func (t *timer) getTimeRemaining(endTime time.Time) countdown {
	currentTime := time.Now()
	difference := endTime.Sub(currentTime)
	total := int(difference.Seconds())
	minutes := total / 60
	seconds := total % 60

	return countdown{
		t: total,
		m: minutes,
		s: seconds,
	}
}

func (t *timer) printSession(endTime time.Time) {
	var text string

	switch t.currentSession {
	case pomodoro:
		text = fmt.Sprintf(printColor(green, "[Pomodoro %d/%d]"), t.iteration, t.longBreakInterval) + ": " + t.msg[pomodoro]
	case shortBreak:
		text = printColor(yellow, "[Short break]") + ": " + t.msg[shortBreak]
	case longBreak:
		text = printColor(blue, "[Long break]") + ": " + t.msg[longBreak]
	}

	fmt.Printf("%s (until %s)\n", text, endTime.Format("03:04:05 PM"))
}

func (t *timer) notify() {
	fmt.Printf("Session completed!\n\n")

	m := map[sessionType]string{
		pomodoro:   "Pomodoro",
		shortBreak: "Short break",
		longBreak:  "Long break",
	}

	if t.showNotification {
		msg := m[t.currentSession] + " is finished"

		// TODO: Handle error
		_ = beeep.Notify(msg, t.msg[t.nextSession()], "")
	}
}

// start begins a new session.
func (t *timer) start(session sessionType) {
	t.currentSession = session

	if session == pomodoro {
		if t.iteration == t.longBreakInterval {
			t.iteration = 1
		} else {
			t.iteration++
		}
	}

	endTime := time.Now().Add(time.Duration(t.kind[session]) * time.Minute)

	t.printSession(endTime)

	fmt.Print("\033[s")

	timeRemaining := t.getTimeRemaining(endTime)

	t.countdown(timeRemaining)

	ticker := time.NewTicker(time.Second)
	for range ticker.C {
		fmt.Print("\033[u\033[K")

		timeRemaining = t.getTimeRemaining(endTime)

		if timeRemaining.t <= 0 {
			t.notify()
			break
		}

		t.countdown(timeRemaining)
	}

	if t.currentSession != pomodoro && !t.autoStartPomodoro || t.currentSession == pomodoro && !t.autoStartBreak {
		// Block until user input before beginning next session
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Press ENTER to start the next session\n")

		_, _ = reader.ReadString('\n')
	}

	t.start(t.nextSession())
}

// countdown prints.
func (t *timer) countdown(timeRemaining countdown) {
	fmt.Printf("Minutes: %02d Seconds: %02d", timeRemaining.m, timeRemaining.s)
}

// newTimer returns a new timer constructed from
// command line arguments.
func newTimer(ctx *cli.Context, c *config) *timer {
	t := &timer{
		kind: kind{
			pomodoro:   c.PomodoroMinutes,
			shortBreak: c.ShortBreakMinutes,
			longBreak:  c.LongBreakMinutes,
		},
		longBreakInterval: c.LongBreakInterval,
		msg: message{
			pomodoro:   c.PomodoroMessage,
			shortBreak: c.ShortBreakMessage,
			longBreak:  c.LongBreakMessage,
		},
		showNotification:  c.Notify,
		autoStartPomodoro: c.AutoStartPomorodo,
		autoStartBreak:    c.AutoStartBreak,
	}

	// Command-line flags will override the configuration
	// file
	if ctx.Uint("pomodoro") > 0 {
		t.kind[pomodoro] = int(ctx.Uint("pomodoro"))
	}

	if ctx.Uint("short-break") > 0 {
		t.kind[shortBreak] = int(ctx.Uint("short-break"))
	}

	if ctx.Uint("long-break") > 0 {
		t.kind[longBreak] = int(ctx.Uint("long-break"))
	}

	if ctx.Uint("long-break-interval") > 0 {
		t.longBreakInterval = int(ctx.Uint("long-break-interval"))
	}

	if ctx.Bool("auto-pomodoro") {
		t.autoStartPomodoro = true
	}

	if ctx.Bool("auto-break") {
		t.autoStartBreak = true
	}

	if ctx.Bool("disable-notifications") {
		t.showNotification = false
	}

	if t.longBreakInterval <= 0 {
		t.longBreakInterval = 4
	}

	return t
}
