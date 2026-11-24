package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b015_helpers.go — shared pieces of the Secret Lair "Life gain and
// loss" cards (S82 OLD batch 15).

// OldB015PayLifeLabel is the option an opponent takes to pay the 3 life.
const OldB015PayLifeLabel = "Pay 3 life"

// OldB015PayThreeLifeUnless is "<do X> unless <a player> pays 3 life" asked of
// that player as the ability resolves (Athreos, God of Passage; Prince of
// Thralls): the first option lets X happen, the second pays 3 life and stops it.
// The payment is CR 119.4's, so a player with fewer than 3 life, or one who
// can't lose life, is not offered it and X happens. `decline` runs with the
// card the ability is about, as the asking item's Context.
type OldB015PayThreeLifeUnless struct {
	question, declineLabel string
	then                   game.OptionPickThen
}

// NewOldB015PayThreeLifeUnless registers the keyed answer: call it from a
// package-level var. The key is an on-disk identity ("option-pick/<card>-…").
func NewOldB015PayThreeLifeUnless(key, question, declineLabel string, decline func(ctx *Context, card uuid.UUID) error) *OldB015PayThreeLifeUnless {
	a := &OldB015PayThreeLifeUnless{question: question, declineLabel: declineLabel}
	a.then = OptionPickThen(key, func(ctx *Context, r game.OptionPicked) error {
		if len(r.Carry) == 0 {
			return nil
		}
		if r.Option != nil && r.Option.Label == OldB015PayLifeLabel {
			if err := ctx.Game.PayLifeForEffect(ctx.Source(), r.Chooser, 3); err == nil {
				return nil
			}
		}
		return decline(ctx, r.Carry[0])
	})
	return a
}

// Ask puts the question to `victim` about `card`.
func (a *OldB015PayThreeLifeUnless) Ask(ctx *Context, victim, card uuid.UUID) error {
	p := ctx.Game.PlayerByIDForEffect(victim)
	if p == nil || p.Eliminated {
		return nil
	}
	options := []game.ChoiceOption{{Label: a.declineLabel}}
	if p.Life >= 3 {
		options = append(options, game.ChoiceOption{Label: OldB015PayLifeLabel, LifeCost: 3})
	}
	return PickOption{
		Player:   victim,
		Question: a.question,
		Options:  options,
		ThenKey:  a.then,
		Carry:    []uuid.UUID{card},
		WithItem: true,
	}.Apply(ctx)
}
