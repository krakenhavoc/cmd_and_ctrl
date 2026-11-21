package effectstest

// answers_guard_test.go — ADR 0142 decision 1: what effects.Register
// refuses in a declared Purpose.Answers or ManaAbility.Answers, and
// (#2877) on a spell's slots.

// refusal runs fn and returns its panic message, "" for none.
func Refusal(fn func()) (msg string) {
	defer func() {
		if r := recover(); r != nil {
			msg, _ = r.(string)
		}
	}()
	fn()
	return ""
}
