package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chromatic Star — Artifact {1} (EDHREC rank 2832):
//
//	"{1}, {T}, Sacrifice this artifact: Add one mana of any color.
//	 When this artifact is put into a graveyard from the battlefield,
//	 draw a card."
//
// The colour-fixing cantrip egg. The mana ability carries all three
// cost components — {1} paid from the pool, the tap, and the
// sacrifice — validated together so a failed activation leaves the
// Star untouched; the any-colour pick keeps its printed width. The
// draw is a leaves-for-the-graveyard trigger on the Star itself, not
// a rider on the ability: it fires whether the Star was cracked,
// destroyed or sacrificed to something else, as printed, and it uses
// the stack (the mana ability does not), so the mana is in the pool
// before the card is drawn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fedbd40b-e3a5-449c-a8a3-b42e9da191a9",
		Name:         "Chromatic Star",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true, Sacrifice: true, Mana: "{1}"},
			Produced: "{W|U|B|R|G}",
			Label:    "{1}, {T}, Sacrifice: Add one mana of any color",
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Chromatic Star — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
