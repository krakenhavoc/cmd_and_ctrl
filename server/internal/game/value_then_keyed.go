package game

// value_then_keyed.go — a string-answer prompt whose "then" is a registered
// tail rather than a closure (#4253): the card name (slice 10c), and since the
// sweep's second pass the resolution-time colour (ColorPrompt.ThenTail) and
// creature type (CreatureTypePrompt.ThenTail).
//
// All three are "a word, then the rest of the card". The keyed form holds no
// chooseValueResume frame: the tail rides PendingChoice.nameThen (written to a
// restore point as pendingChoiceSnapshot.NameThen, a field these kinds share)
// with its key on PickThen, so a binary that lacks the tail refuses the file
// rather than restoring a prompt whose answer lands nowhere. The answer is the
// tail's Word. The prompt takes the cause it was queued under, as a keyed
// number prompt does: the answer may come after a restore, when nothing is
// resolving.

// carriesValueThen reports whether a prompt kind may carry a keyed value
// "then" (nameThen).
func carriesValueThen(k PendingChoiceKind) bool {
	return k == PendingChoiceCardName || k == PendingChoiceColor || k == PendingChoiceCreatureType
}

// keyValueThenLocked makes c a keyed value prompt answered through `then`.
//
// Caller must hold g.mu.
func (g *Game) keyValueThenLocked(c *PendingChoice, then KeyedTail) {
	c.PickThen = then.Ref.key
	c.nameThen = copyKeyedTail(then)
	cause := g.resolutionCauseLocked()
	c.PayCauseController, c.PayCauseItem = cause.Controller, cause.Item
}

// valueTailFrameOf is the continuation behind a keyed colour or creature-type
// prompt, built from its data: the tail, told the answer (or "" when the
// chooser left) as its Word. Nil for any other prompt. The card-name resolver
// runs its tail itself (runValueTail).
func valueTailFrameOf(c *PendingChoice) *chooseValueFrame {
	if c == nil || c.nameThen.IsZero() || (c.Kind != PendingChoiceColor && c.Kind != PendingChoiceCreatureType) {
		return nil
	}
	tail, cause := copyKeyedTail(c.nameThen), carriedCauseOfChoice(c)
	return &chooseValueFrame{then: func(g *Game, value string) error {
		return g.runValueTail(tail, cause, value)
	}}
}

// runValueTail finishes a keyed value prompt: its tail, told the answer as
// its Word, under the cause the prompt was queued under.
//
// Caller must hold g.mu.
func (g *Game) runValueTail(then KeyedTail, cause carriedCause, value string) error {
	t := copyKeyedTail(then)
	t.Word = value
	return g.withCarriedCauseLocked(cause, func() error {
		return g.runKeyedTailLocked(t, nil, 0)
	})
}
