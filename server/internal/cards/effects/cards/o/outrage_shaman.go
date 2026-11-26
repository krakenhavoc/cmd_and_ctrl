package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Outrage Shaman — {3}{R}{R} Creature — Goblin Shaman, 2/2:
//
//	"Chroma — When this creature enters, it deals damage to target creature
//	 equal to the number of red mana symbols in the mana costs of permanents
//	 you control."
//
// The count is read as the trigger resolves; the Shaman's own {R}{R} count.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "08d77e5b-7de1-4b0a-a5a7-75e60d452d27",
		Name:         "Outrage Shaman",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Outrage Shaman — damage to target creature for each red mana symbol",
				effects.SourceDealsComputedDamageToEachLegalTarget(func(ctx *effects.Context) int {
					return effects.S82OldB521ChromaOfPermanents(ctx.Game, ctx.Controller(), ctx.TextColor("R"))
				})),
				effects.TargetCreature("target creature")),
		},
	})
}
