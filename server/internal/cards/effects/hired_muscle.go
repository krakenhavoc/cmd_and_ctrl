package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hired Muscle // Scarmaker — {1}{B}{B} Creature — Human Warrior 2/2 //
// Legendary Creature — Spirit 4/4, a flip card (#3216, ADR 0151):
//
//	Top:     "Whenever you cast a Spirit or Arcane spell, you may put a
//	          ki counter on this creature.
//	          At the beginning of the end step, if there are two or more
//	          ki counters on this creature, you may flip it."
//	Flipped: "Remove a ki counter from Scarmaker: Target creature gains
//	          fear until end of turn."
//
// Scarmaker is black and {1}{B}{B} (CR 710.1c). The fear grant is Shizo,
// Death's Storehouse's until-end-of-turn keyword grant.
//
// No simplification.
const hiredMuscleOracleID = "ea31c9a4-af4f-4cab-b53c-77ef4241d4a6"

func init() {
	Register(Spec{
		OracleID:     hiredMuscleOracleID,
		Name:         "Hired Muscle",
		Completeness: CompletenessFull,
		Triggered:    kiFlipTriggers("Hired Muscle"),
	})
	Register(Spec{
		OracleID:     hiredMuscleOracleID + "#1",
		Name:         "Scarmaker",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Remove a ki counter from Scarmaker: Target creature gains fear until end of turn.",
			Cost:    removeAKiCounter(),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if id, ok := b16FirstLegalTargetCard(ctx); ok {
					return GrantKeywordUntilEOT{
						Target:   id,
						Keywords: []string{"fear"},
						Label:    "Scarmaker — fear",
					}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
