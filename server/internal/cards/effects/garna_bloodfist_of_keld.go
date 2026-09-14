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
// the source excluded). "If it was attacking" is the part with no
// field to read: the battlefield-leave choke point clears the dead
// creature's attack before its dies event fires, and the harvester's
// last-known characteristic carries no combat state — so it is read
// off the event log in Build, at the event (b35WasAttackingWhenItLeft):
// an attack declaration naming the creature, found before the combat
// it attacked in ended or the turn changed, means it was attacking.
// A creature that attacked, left, and came back is a new object with
// no attack of its own, so it reads as not attacking, as printed.
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
				return b35AnotherCreatureYouControlDied(ev, source, g)
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				attacking := b35WasAttackingWhenItLeft(g, ev.CardID, ev.Seq)
				label := "Garna, Bloodfist of Keld — 1 damage to each opponent"
				if attacking {
					label = "Garna, Bloodfist of Keld — draw a card (it was attacking)"
				}
				return game.NewTriggeredItem(source, label, b35DrawIfAttackingElsePingOpponents(attacking))
			},
		}},
	})
}
