package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Guild Globe — Artifact {2}:
//
//	"When this artifact enters, draw a card.
//	 {2}, {T}, Sacrifice this artifact: Add two mana of different
//	 colors."
//
// A cantrip that later cracks into a colour fix: two mana of any kind
// in, two mana of different colours out (#2558, DifferentColors(2)),
// never two of one colour. The sacrifice, the tap and the {2} are one
// cost paid before the mana is added (CR 605.3b), so the Globe is gone
// even when the colours are asked for afterwards.
//
// Waited on #2558 from the Sami Whammy deck request (#2190): the mana
// ability is the card's whole second half.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a7e7cacf-cc2f-45ce-abb1-7b66d1e2fabc",
		Name:         "Guild Globe",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{samiDrawOnETB("Guild Globe")},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Mana: "{2}", Tap: true, Sacrifice: true},
			Produced: DifferentColors(2),
			Label:    "{2}, {T}, Sacrifice this artifact: Add two mana of different colors",
		}},
	})
}
