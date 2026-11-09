package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Clarion Conqueror — Creature — Dragon {2}{W}, 3/3 (Tarkir: Dragonstorm,
// #4463):
//
//	"Flying
//	 Activated abilities of artifacts, creatures, and planeswalkers can't
//	 be activated."
//
// Cursed Totem's gate (activation_restriction.go, CR 602.5) over three
// types: symmetrical, derived from the Conqueror being on the
// battlefield, with no mana exemption, so a Birds of Paradise, a Sol Ring
// and a loyalty ability are all shut off, the Conqueror's controller's
// own included. The predicate reads the layered types, so a land
// creature or an artifact land counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f1266a1b-bae6-4ded-b4d5-7586a5b6c5d0",
		Name:            "Clarion Conqueror",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		ActivationRestrictions: []game.ActivationRestriction{
			SourcesCantActivate(
				"Clarion Conqueror — activated abilities of artifacts, creatures, and planeswalkers can't be activated.",
				Or(Artifact(), Creature(), Planeswalker()), false),
		},
	})
}
