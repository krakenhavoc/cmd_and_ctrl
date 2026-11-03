package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "0f80f061-6304-4c7b-ab53-f406385e39df",
		Name:            "Goremand",
		Completeness:    CompletenessFull,
		AdditionalCost:  SacrificeCost("a creature", Creature()),
		PrintedKeywords: []string{"flying", "trample"},
		Triggered: []game.TriggeredAbility{
			TriggerEdict(WhenThisEnters("Goremand — each opponent sacrifices a creature", Do(EachPlayerSacrifices{
				ExceptController: true,
				Match:            Creature(),
				Label:            "a creature",
			}))),
		},
	})
}
