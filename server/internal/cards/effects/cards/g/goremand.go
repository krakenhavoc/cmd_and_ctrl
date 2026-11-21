package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goremand — Creature — Demon {4}{B}{B}, 5/5:
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Flying
//	 Trample
//	 When this creature enters, each opponent sacrifices a creature."
//
// The sacrifice is paid while casting (CR 601.2h), so a Goremand put
// onto the battlefield by another effect costs nothing and still
// triggers. The enters trigger is Grave Pact's "each other player" edict:
// each opponent chooses their own creature, an opponent with none is
// skipped, and nothing is targeted.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0f80f061-6304-4c7b-ab53-f406385e39df",
		Name:            "Goremand",
		Completeness:    effects.CompletenessFull,
		AdditionalCost:  effects.SacrificeCost("a creature", effects.Creature()),
		PrintedKeywords: []string{"flying", "trample"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerEdict(effects.WhenThisEnters("Goremand — each opponent sacrifices a creature", effects.Do(effects.EachPlayerSacrifices{
				ExceptController: true,
				Match:            effects.Creature(),
				Label:            "a creature",
			}))),
		},
	})
}
