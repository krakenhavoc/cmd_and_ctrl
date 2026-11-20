package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Elvish Harbinger — Creature — Elf Druid {2}{G}, 1/2 (Lorwyn, #5593):
//
//	"When this creature enters, you may search your library for an Elf card,
//	 reveal it, then shuffle and put that card on top.
//	 {T}: Add one mana of any color."
//
// The Harbinger enters tutor (lrwB1Harbinger): a "you may" asked as the
// trigger resolves (CR 603.5), a search for an Elf card that may find
// nothing (CR 701.23b), shown to the table, and placed on top after the
// shuffle. A Kindred card of the type counts, and so does a changeling
// card (CR 702.73a).
//
// The mana ability is the plain any-colour tap of Birds of Paradise.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "28ddc23c-5b0b-4747-9a33-6845acf27bec",
		Name:         "Elvish Harbinger",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			lrwB1Harbinger("Elvish Harbinger", "an Elf card", lrwB1CardOfType("Elf")),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
	})
}
