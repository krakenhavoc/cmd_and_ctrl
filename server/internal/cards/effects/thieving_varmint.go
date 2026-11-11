package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thieving Varmint — Creature — Varmint {1}{B}, 2/1 (Outlaws of Thunder
// Junction Commander, #5457):
//
//	"Deathtouch, lifelink
//	 {T}, Pay 1 life: Add two mana of any one color. Spend this mana only
//	 to cast spells you don't own."
//
// The mana is two of ONE colour (picked once, ProducedOneColor), each
// carrying game.ManaRestrictSpellNotOwned: it pays only for a spell whose
// owner is not the player spending it, so never for an activation and
// never for a spell its controller owns. Deathtouch and lifelink are the
// printed keywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b9f3cf9f-24c7-43ed-9ba1-f4570df3d09e",
		Name:         "Thieving Varmint",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:           ManaAbilityCost{Tap: true, Life: 1},
			ProducedAmount: ProducedOneColor(func(*game.Game, uuid.UUID, uuid.UUID) int { return 2 }),
			Label:          "{T}, Pay 1 life: Add two mana of any one color. Spend this mana only to cast spells you don't own",
			Restrictions:   []string{ManaRestrictSpellNotOwned},
		}},
	})
}
