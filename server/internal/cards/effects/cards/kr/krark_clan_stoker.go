package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Krark-Clan Stoker — Creature — Goblin Shaman {2}{R} (Darksteel, #5605):
//
//	"{T}, Sacrifice an artifact: Add {R}{R}."
//
// Krark-Clan Ironworks's mana ability with a tap beside the sacrifice:
// it doesn't use the stack (CR 605.3a), and a summoning-sick Stoker can't
// pay {T} (CR 302.6). An artifact sacrifice, not a creature one, so it
// answers nothing (AnswerValue, ADR 0142).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "756b4bc9-6f3f-4eae-b6b0-63c94ed8482f",
		Name:         "Krark-Clan Stoker",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost: effects.ManaAbilityCost{
				Tap:            true,
				SacrificeOther: effects.B10SacrificeAnArtifact().SacrificeOther,
			},
			Produced: "{R}{R}",
			Label:    "{T}, Sacrifice an artifact: Add {R}{R}",
			Answers:  game.AnswerValue,
		}},
	})
}
