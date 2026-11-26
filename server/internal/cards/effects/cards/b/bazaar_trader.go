package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bazaar Trader — Creature — Goblin {1}{R}, 1/1 (Worldwake, S82 OLD b487):
//
//	"{T}: Target player gains control of target artifact, creature, or land you
//	 control."
//
// Two clauses: a player, and one of your artifacts, creatures or lands (the
// Trader itself is a legal pick). The control change has no stated duration, so
// it lasts indefinitely (CR 611.2a), a layer-2 effect pinned to the object the
// permanent was as the ability resolved (CR 400.7). Both clauses are re-checked
// as it resolves (CR 608.2b): with the player gone it does nothing, and with the
// permanent gone or no longer yours (it was taken in response) it does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "21ad7067-3210-42d6-9009-0784b2834227",
		Name:         "Bazaar Trader",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{T}: Target player gains control of target artifact, creature, or land you control.",
			Cost:  effects.TapCost(),
			Targets: effects.Clauses(
				effects.TargetPlayer("target player"),
				effects.TargetPermanent("target artifact, creature, or land you control", effects.Or(effects.Artifact(), effects.Creature(), effects.Land()), effects.YouControl()),
			),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				player, ok := ctx.ClauseTarget(0)
				if !ok || player.Kind != game.TargetPlayer {
					return nil
				}
				perm, ok := ctx.ClauseTarget(1)
				if !ok || perm.Kind != game.TargetCard {
					return nil
				}
				return effects.GainControl{Target: perm.ID, Controller: player.ID, Duration: game.IndefiniteDuration(), Label: "Bazaar Trader — gains control"}.Apply(ctx)
			},
		}},
	})
}
