package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Command Chair — Artifact {6} (Star Trek Commander, #4974):
//
//	"This artifact enters tapped.
//	 {T}: Add three mana of any one color.
//	 {T}: Add X mana of any one color, where X is the greatest mana value
//	 among commanders you control."
//
// The first ability is Gilded Lotus's. The second reads X as it is
// activated: the greatest mana value among commander permanents on the
// battlefield that you control, whoever owns them (a commander in the
// command zone is not one you control). With none, X is zero and the
// ability adds nothing (CR 107.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7132b0ad-7bab-4379-ad82-a9af289f835e",
		Name:         "Command Chair",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: OneColorOfAmount(3),
				Label:    "Add three mana of any one color",
			},
			{
				Cost:         ManaAbilityCost{Tap: true},
				ProducedFunc: commandChairProduced,
				Label:        "Add X mana of any one color, where X is the greatest mana value among commanders you control",
			},
		},
	})
}

// commandChairProduced is the second ability's output: X of any one
// color, X the greatest mana value among commanders you control.
func commandChairProduced(g *game.Game, controller, _ uuid.UUID) string {
	best := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCommander && c.Controller == controller && c.ManaValue() > best {
			best = c.ManaValue()
		}
	}
	return OneColorOfAmount(best)
}
