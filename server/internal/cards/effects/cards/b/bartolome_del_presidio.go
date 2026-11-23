package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bartolomé del Presidio — Legendary Creature — Vampire Knight
// {W}{B}, 2/1 (EDHREC rank 1780):
//
//	"Sacrifice another creature or artifact: Put a +1/+1 counter on
//	 Bartolomé del Presidio."
//
// The free sac outlet. One activated ability whose whole cost is the
// sacrifice — Warren Soultrader's shape without the life — over "a
// creature or artifact" that isn't Bartolomé himself. "Another" is
// object identity (effects.Another, CR 109.1), so a token copy of him
// can be fed to his own ability, as printed.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f47e4c56-0a0b-422d-bf3d-7a20ee289f15",
		Name:         "Bartolomé del Presidio",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice another creature or artifact: Put a +1/+1 counter on Bartolomé del Presidio.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerSacOutlet},
			Cost: game.AbilityCost{
				SacrificeOther: effects.Another(effects.SacrificeSpec("another creature or artifact",
					effects.Or(effects.Creature(), effects.Artifact()))),
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.B15OnBattlefield(g, item.SourceCardID) {
					return nil
				}
				return effects.AddCounter{Target: item.SourceCardID, Kind: "+1/+1", N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
