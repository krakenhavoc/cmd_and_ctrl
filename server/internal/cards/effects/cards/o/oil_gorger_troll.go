package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oil-Gorger Troll — Creature — Phyrexian Troll Warrior {3}{G}{G}, 3/4
// (Phyrexia: All Will Be One):
//
//	"When this creature enters, you gain 3 life. Then if you control a
//	 permanent with an oil counter on it, draw a card."
//
// The life comes first and the "then if" is read afterwards, as the
// ability resolves, so a permanent that lost its last counter in the
// meantime does not count.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "238f2c86-4376-48f2-ae71-c10d3d785163",
		Name:         "Oil-Gorger Troll",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Oil-Gorger Troll — you gain 3 life; then if you control a permanent with an oil counter on it, draw a card",
				effects.Do(effects.GainLife{Amount: 3},
					effects.If{Cond: effects.S82YouControlAnOilPermanent, Then: []effects.Applier{effects.DrawCards{N: 1}}})),
		},
	})
}
