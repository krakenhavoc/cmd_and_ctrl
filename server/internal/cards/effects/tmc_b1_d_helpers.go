package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tmc_b1_d_helpers.go — shared pieces for TMNT Commander batch 1,
// slice d (#5011): squad, and "for each opponent, create a token that's
// tapped and attacking that player".

// squadKey is the optional-cost key of CR 702.157's squad. The engine
// reads no optional-cost key but its own few (game.KickerKey and
// friends); squad is read only by SquadTrigger, off the paid record the
// resolution carries onto the permanent (CR 400.7d, ADR 0073 §5).
const squadKey = "squad"

// squadMaxTimes is the engine's cap on one announcement, multikicker's
// reason (Multikicker): paper squad is unbounded, but an announcement
// has to be finite, the client's stepper has to stop and the bot's
// expansion has to terminate. Far above what any board pays for.
const squadMaxTimes = 20

// Squad is the first of CR 702.157a's two linked abilities: "As an
// additional cost to cast this spell, you may pay [cost] any number of
// times." It is an optional mana cost paid any number of times, exactly
// multikicker's shape (CR 702.157a: "Paying a spell's squad cost follows
// the rules for paying additional costs in rules 601.2b and 601.2f–h").
// Declare SquadTrigger beside it for the second ability.
//
//	OptionalCosts: []game.AdditionalCost{Squad("{1}{W}")},
//	Triggered:     []game.TriggeredAbility{SquadTrigger("Endless Foot Assault — squad")},
func Squad(mana string) game.AdditionalCost {
	return game.AdditionalCost{
		Optional: true,
		Key:      squadKey,
		ManaCost: mana,
		Repeat:   squadMaxTimes,
		Label:    "Squad " + mana,
	}
}

// squadTimesPaid is how many times the permanent's squad cost was paid
// as it was cast: zero for one that was not cast (a token copy, a
// reanimated or flickered permanent), whose record is empty.
func squadTimesPaid(c game.Card) int {
	return game.OptionalCostTimesPaid(c, c.Provenance.OptionalCosts, squadKey)
}

// SquadTrigger is the second of CR 702.157a's linked abilities: "When
// this [permanent] enters, if its squad cost was paid, create a token
// that's a copy of it for each time its squad cost was paid."
//
// The intervening "if" (CR 603.4) is checked as it triggers: an unpaid
// squad puts nothing on the stack. The count is read then too, off the
// paid record on the permanent, and carried on the item (Params.Amount),
// because the record is cleared as the permanent leaves (CR 400.7) and
// the ruling is that the copies are still made if it has left by the
// time the ability resolves. The payments cannot change once the spell
// was cast, so the re-check as it resolves is the count on the item.
//
// The copies are of the permanent as it is now, or of its last-known
// copiable values once it has left (CR 707.2, CR 608.2h,
// game.TokenCopySourceForEffect), all created by one event, so they
// enter at the same time. A token copy was not cast, so its own squad
// ability never triggers.
func SquadTrigger(label string) game.TriggeredAbility {
	t := On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return Self(ev, source, lki, g) && squadTimesPaid(*source) > 0
	}, label, func(g *game.Game, item *game.StackItem) error {
		n := item.Params.Amount
		if n <= 0 {
			return nil
		}
		return CreateTokenCopy{Controller: item.Controller, Copy: item.SourceCardID, N: n}.Apply(NewContext(g, item))
	})
	t.Build = func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Amount = squadTimesPaid(*source)
		return item
	}
	return t
}

// tokenTappedAndAttackingEachOpponent is "for each opponent, create
// <tmpl> that's tapped and attacking that player": one token per
// opponent still in the game as the ability resolves, whether or not
// that player was attacked. The tokens were never declared as attackers,
// so nothing that triggers on a creature attacking sees them (CR 508.4).
func tokenTappedAndAttackingEachOpponent(ctx *Context, tmpl func() game.Card) error {
	for _, opp := range ctx.Opponents() {
		t := tmpl()
		t.Tapped = true
		t.AttackingTarget = opp
		if err := (CreateToken{Controller: ctx.Controller(), Template: t, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
