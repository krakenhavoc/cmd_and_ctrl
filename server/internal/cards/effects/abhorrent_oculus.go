package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Abhorrent Oculus — Creature — Eye {2}{U} 5/5 (#3536, ADR 0189 slice 2):
//
//	"As an additional cost to cast this spell, exile six cards from your
//	 graveyard.
//	 Flying
//	 At the beginning of each opponent's upkeep, manifest dread."
//
// The additional cost is the activated ability's exile-N-cards component on
// a cast (ExileFromGraveyardCost, AdditionalCost.ExileCards): the caster
// names six graveyard cards on cast_spell as exile_ids, they are exiled with
// the spell on the stack (a countered Oculus does not give them back, and a
// graveyard-leaves watcher triggers above it), and a graveyard of fewer than
// six cards makes the spell uncastable (CR 118.3). Because it is an
// ADDITIONAL cost it is charged under every alternative cost and every "cast
// without paying its mana cost", as the 2024-09-20 ruling says.
//
// The trigger is controlled by Oculus's controller and fires on each
// OPPONENT's upkeep (AtEachOpponentsUpkeep), never on its own controller's.
// The manifest dread is the controller's: they look at the top two of their
// own library.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b1a0e129-27f0-4a09-b00d-497e18777306",
		Name:            "Abhorrent Oculus",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		AdditionalCost:  ExileFromGraveyardCost(6, "six cards", nil),
		Triggered: []game.TriggeredAbility{
			AtEachOpponentsUpkeep("Abhorrent Oculus — manifest dread", Do(ManifestDread{})),
		},
	})
}
