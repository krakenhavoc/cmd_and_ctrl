package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Knotvine Paladin — Creature — Human Knight {G}{W}, 2/2 (Alara Reborn):
//
//	"Whenever this creature attacks, it gets +1/+1 until end of turn for each
//	 untapped creature you control."
//
// The untapped creatures are counted as the ability resolves (CR 611.2c),
// so a creature tapped by the attack (Paladin itself, with no vigilance)
// is not counted. A Paladin that has left the battlefield gets nothing
// (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "75efff64-ea9b-4e3e-a729-fc003ee3db1e",
		Name:         "Knotvine Paladin",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Knotvine Paladin — it gets +1/+1 for each untapped creature you control",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.SourceIsStillThisPermanent(g, item) {
						return nil
					}
					n := effects.CountControlled(g, item.Controller, func(c game.Card) bool { return c.IsCreature() && !c.Tapped })
					return effects.BoostUntilEOT{Target: item.SourceCardID, Power: n, Toughness: n,
						Label: "Knotvine Paladin — +1/+1 per untapped creature"}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
