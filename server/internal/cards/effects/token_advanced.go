package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// token_advanced.go — TokenSpec and the CreateTokenAdvanced
// primitive.
//
// CreateToken takes a bare game.Card template and puts N of it on the
// battlefield. That covers "create two 1/1 red Goblin creature
// tokens" and nothing else. The printed text of a token-maker
// routinely says more than the token's own characteristics:
//
//	"create a TAPPED Powerstone token"        (Stern Lesson)
//	"create a 1/1 Goblin token WITH HASTE"    (Legion Warboss)
//	"create an X/X Ooze WITH X +1/+1 COUNTERS"
//
// Those are properties of the CREATION, not of the token, so they do
// not belong in the template constructors in tokens.go — the same
// printed Powerstone is made tapped by one card and untapped by
// another, and forking the constructor per variant is how a token
// file turns into forty near-identical functions.
//
// TokenSpec is that structured description: one printed template plus
// the creation-time modifiers, built up with chainable helpers so a
// card reads like its oracle text. CreateTokenAdvanced applies it.

// TokenSpec is a structured token description: what to make, and how
// it enters.
//
// Build one with Token(template) and refine it with EntersTapped /
// WithCounters / WithKeywords. The zero modifiers are the ordinary
// case, so `Token(RedGoblinToken())` is exactly CreateToken's
// behaviour.
type TokenSpec struct {
	// Template is the printed token — one of the constructors in
	// tokens.go.
	Template game.Card

	// Tapped enters the tokens tapped.
	Tapped bool

	// Counters are the counters each token enters with, applied
	// before its ETB event fires.
	Counters map[string]int

	// Keywords are granted on top of the template's printed ones.
	Keywords []string
}

// Token starts a TokenSpec from a printed template.
func Token(template game.Card) TokenSpec {
	return TokenSpec{Template: template}
}

// EntersTapped makes the tokens enter tapped.
func (s TokenSpec) EntersTapped() TokenSpec {
	s.Tapped = true
	return s
}

// WithCounters adds n counters of the named kind to each token as it
// enters. Repeated calls accumulate different kinds.
func (s TokenSpec) WithCounters(kind string, n int) TokenSpec {
	if kind == "" || n <= 0 {
		return s
	}
	next := make(map[string]int, len(s.Counters)+1)
	for k, v := range s.Counters {
		next[k] = v
	}
	next[kind] += n
	s.Counters = next
	return s
}

// WithKeywords grants keywords on top of the template's printed ones
// — the "…with haste" half of a token-maker.
func (s TokenSpec) WithKeywords(kw ...string) TokenSpec {
	if len(kw) == 0 {
		return s
	}
	next := make([]string, 0, len(s.Keywords)+len(kw))
	next = append(next, s.Keywords...)
	next = append(next, kw...)
	s.Keywords = next
	return s
}

// options renders the spec's modifiers for the engine helper.
func (s TokenSpec) options() game.TokenEntryOptions {
	return game.TokenEntryOptions{
		Tapped:   s.Tapped,
		Counters: s.Counters,
		Keywords: s.Keywords,
	}
}

// CreateTokenAdvanced puts N tokens described by a TokenSpec onto the
// battlefield under Controller's control. It is CreateToken plus the
// creation-time modifiers; a card with none of them should keep using
// CreateToken, which reads better for the common case.
type CreateTokenAdvanced struct {
	Controller uuid.UUID
	Spec       TokenSpec
	N          int
}

func (c CreateTokenAdvanced) Apply(ctx *Context) error {
	controller := c.Controller
	if controller == uuid.Nil {
		controller = ctx.Controller()
	}
	_, err := ctx.Game.CreateTokensForEffect(controller, c.Spec.Template, c.N, c.Spec.options())
	return err
}
