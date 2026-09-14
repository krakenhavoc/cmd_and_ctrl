package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scytheclaw Raptor — Creature — Dinosaur {2}{R}, 4/3 (EDHREC rank
// 3117):
//
//	"Whenever a player casts a spell, if it's not their turn, this
//	 creature deals 4 damage to them."
//
// The anti-flash Dinosaur. Any player's cast — the controller's
// own included, as printed — and the intervening "if" is whether
// the caster is the active player (b29SpellCastOffTurn); the turn
// cannot change while the trigger waits on the stack, so the check
// at trigger time holds at resolution. The damage is the Raptor's,
// to the caster.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8c653a0d-e35a-4596-bed6-b5156192955b",
		Name:         "Scytheclaw Raptor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b29SpellCastOffTurn(ev, g)
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				caster := ev.Actor
				return game.NewTriggeredItem(source, "Scytheclaw Raptor — 4 damage to the player who cast a spell on another player's turn",
					func(g *game.Game, item *game.StackItem) error {
						return DealDamage{Source: item.SourceCardID, Target: caster, Amount: 4}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
