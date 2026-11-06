package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gray and Adira, Forever Joined — Legendary Creature — Trill Human
// Officer {2}{U/R}{U/R}, 2/3 (Star Trek, #4973):
//
//	"At the beginning of combat on your turn, if you've cast a
//	 noncreature spell this turn, create a token that's a copy of target
//	 creature you control, except it isn't legendary. It gains haste
//	 until end of turn. Sacrifice it at the beginning of the next end
//	 step."
//
// Mister Fantastic's intervening if (CR 603.4): with no noncreature spell
// cast this turn the trigger never goes on the stack, and it checks again
// as it resolves. The copy's one exception is the missing Legendary
// supertype (CR 707.9b); haste is NOT part of the copy (CR 707.9a) but an
// effect that ends with the turn (CR 611.2a), so a copy of the token made
// later has no haste. The sacrifice is a delayed trigger over that one
// token (CR 603.7). With the target gone (CR 608.2b) nothing is made.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "34a39862-c18b-4c1b-8920-207627c40934",
		Name:         "Gray and Adira, Forever Joined",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(mscM13CombatIfCastNoncreature(
				"Gray and Adira, Forever Joined — a nonlegendary token copy of target creature you control with haste; sacrifice it at the next end step",
				grayAndAdiraCopy),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}

// grayAndAdiraCopy makes the token, gives it haste until end of turn and
// schedules its sacrifice.
func grayAndAdiraCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var copyOf uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			copyOf = t.ID
			break
		}
	}
	if copyOf == uuid.Nil {
		return nil
	}
	cursor := b25LastEventSeq(g)
	if err := (CreateTokenCopy{
		Controller: item.Controller,
		Copy:       copyOf,
		N:          1,
		Except: func(t *game.Card) {
			t.TypeLine = removeLegendaryFromTypeLine(t.TypeLine)
		},
	}).Apply(ctx); err != nil {
		return err
	}
	tokens := b27TokensCreatedByAfter(g, item.Controller, cursor)
	if len(tokens) == 0 {
		return nil
	}
	for _, id := range tokens {
		if err := (GrantKeywordUntilEOT{
			Target:   id,
			Keywords: []string{"haste"},
			Label:    "Gray and Adira, Forever Joined — haste until end of turn",
		}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return ScheduleDelayedTrigger{
		Label: "Gray and Adira, Forever Joined — sacrifice the token",
		Cards: tokens,
		Body:  sacrificeListedCardsBody,
	}.Apply(ctx)
}
