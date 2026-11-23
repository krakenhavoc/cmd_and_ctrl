package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Deathleaper, Terror Weapon — Legendary Creature — Tyranid {2}{R}{G}, 3/3:
//
//	"Flash
//	 Haste
//	 Flesh Hooks — Creatures you control that entered this turn have double
//	 strike."
//
// "Entered this turn" is the engine's per-turn entry record
// (Game.EnteredThisTurn), judged on the layered view each recompute.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8e22495d-a921-4c91-ada3-79098851e5d4",
		Name:            "Deathleaper, Terror Weapon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "haste"},
		Static: []game.StaticAbility{
			effects.Labelled(effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller && g.EnteredThisTurn(target.InstanceID)
			}, "double strike"), "Creatures you control that entered this turn have double strike."),
		},
	})
}
