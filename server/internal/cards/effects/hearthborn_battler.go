package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hearthborn Battler — Creature — Lizard Warlock {2}{R}, 2/3
// (Bloomburrow, #4467):
//
//	"Haste
//	 Whenever a player casts their second spell each turn, this creature
//	 deals 2 damage to target opponent."
//
// Lotho, Corrupt Shirriff's condition (any seat's second spell, its
// controller's included: the cast tally is bumped before EventCast, so
// the second cast reads exactly two) with Deadeye Duelist's damage to
// target opponent. The opponent is chosen as the ability goes on the
// stack (CR 603.3d), and the Battler deals the damage, last known
// information if it has left.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "beb02f49-95a6-49bf-b176-4e4b2ab5b651",
		Name:            "Hearthborn Battler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventCast, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return g.CastTallyFor(ev.Actor).Total == 2
			}, "Hearthborn Battler — 2 damage to target opponent", sourceDealsDamageToEachLegalTarget(2)),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}
