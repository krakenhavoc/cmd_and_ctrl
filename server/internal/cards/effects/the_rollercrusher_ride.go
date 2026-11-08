package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Rollercrusher Ride — Legendary Enchantment {X}{2}{R} (Duskmourn:
// House of Horror, #4466):
//
//	"Delirium — If a source you control would deal noncombat damage to a
//	 permanent or player while there are four or more card types among
//	 cards in your graveyard, it deals double that damage instead.
//	 When The Rollercrusher Ride enters, it deals X damage to each of up
//	 to X target creatures."
//
// The doubler is Solphim's replacement (CR 614.1a) without its opponent
// clause: noncombat damage from a source you control, to anything, while
// your graveyard has delirium (CR 207.2c, read as the damage would be
// dealt). Its own enters damage is noncombat damage from a source you
// control, so it doubles too.
//
// The enters trigger is Extraordinary Journey's "up to X target
// creatures" (UpToCastXCreatureTargets) with Kinetic Ooze's X carried on
// the item (Card.CastX, CR 107.3m), so the amount is exact even if the
// Ride has left by the time the trigger resolves. The damage is one
// instance to every target still legal (CR 608.2b), Batroc the Leaper's
// sourceDealsComputedDamageToEachLegalTarget. A Ride that was not cast
// has X of zero and targets nothing; X of zero still leaves the doubler,
// so the card does not declare XMatters.
//
// No simplification.
func init() {
	const label = "The Rollercrusher Ride — X damage to each of up to X target creatures"
	enters := WhenThisEnters(label, sourceDealsComputedDamageToEachLegalTarget(func(ctx *Context) int {
		return ctx.Item.Params.Amount
	}))
	enters.TargetsFrom = UpToCastXCreatureTargets
	enters.Build = func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Amount = source.CastX()
		return item
	}
	Register(Spec{
		OracleID:     "aed1f0cd-8df8-415f-afd6-ccfd12234334",
		Name:         "The Rollercrusher Ride",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters},
		Replacements: []game.ReplacementEffect{{
			Watches:   []game.EventKind{game.EventDealDamage},
			RowLabel:  "Delirium — Noncombat damage your sources would deal is doubled while there are four or more card types among cards in your graveyard.",
			AppliesTo: rollercrusherRideDoubles,
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.DamageAmount *= 2
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
		}},
	})
}

// rollercrusherRideDoubles is the AppliesTo of "if a source you control
// would deal noncombat damage to a permanent or player while there are
// four or more card types among cards in your graveyard".
func rollercrusherRideDoubles(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
	if ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || ev.IsCombatDamage || ev.DamageTarget == uuid.Nil {
		return false
	}
	return damageSourceControlledBy(ev, g, src.Controller) && b16CardTypesInGraveyard(g, src.Controller) >= 4
}
