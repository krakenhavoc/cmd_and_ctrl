package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Garna, Bloodfist of Keld — Legendary Creature — Human Berserker
// {1}{B}{R}{R}, 4/3 (EDHREC rank 3702):
//
//	"Whenever another creature you control dies, draw a card if it
//	 was attacking. Otherwise, Garna deals 1 damage to each opponent."
//
// The aristocrats commander that pays either way. The condition is
// "another creature you control died" (diedCreature read post-move,
// the source excluded). "If it was attacking" is diedWhileAttacking
// (#1661): the attack as it last existed, carried on the dies event
// itself, because the exit has already cleared it from the card
// (CR 603.10a).
//
// It used to be read off the event log instead — an EventAttack
// naming the creature earlier in the same combat — and that was wrong
// twice: a creature put onto the battlefield attacking (a ninja, a
// token created tapped and attacking) was never declared and read as
// not attacking, and one removed from combat by a control change
// (CR 506.4) still read as attacking.
//
// The damage is dealt by Garna, so a Fog-class shield stops it; the
// draw is the controller's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3fb83218-e18d-405d-91c8-46b5e4e672a9",
		Name:         "Garna, Bloodfist of Keld",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return anotherCreatureYouControlDied(ev, source, g)
			},
			Key: "Garna, Bloodfist of Keld — draw a card or deal 1 damage to each opponent",
			// "Was it attacking" reads the event log at trigger time
			// (ADR 0041 P9's fill-in Build): a fact about the moment the
			// creature died, not something the resolving item can
			// re-derive from the board later.
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				_, attacking := diedWhileAttacking(ev, g)
				label := "Garna, Bloodfist of Keld — 1 damage to each opponent"
				if attacking {
					label = "Garna, Bloodfist of Keld — draw a card (it was attacking)"
				}
				item := game.NewTriggeredItem(source, label)
				if attacking {
					item.Params.Amount = 1
				}
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return b35DrawIfAttackingElsePingOpponents(item.Params.Amount != 0)(g, item)
			},
		}},
	})
}
