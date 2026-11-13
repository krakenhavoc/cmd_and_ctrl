package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Damping Matrix — Artifact {3} (Mirrodin, #5606):
//
//	"Activated abilities of artifacts and creatures can't be activated
//	 unless they're mana abilities."
//
// Clarion Conqueror's board-wide activation restriction on artifacts
// and creatures, with mana abilities exempt (Null Rod's flag). It binds
// every player, the Matrix's controller too, and reads the source as it
// is now, so an animated land is covered while it is a creature. Equip,
// a creature's pump and an artifact's sacrifice ability are all stopped;
// a land's abilities and a planeswalker's loyalty abilities are not.
//
// "Artifacts and creatures" names no zone and no "card", so it means
// permanents on the battlefield (CR 109.2): an ability activated from a
// hand or a graveyard (cycling an artifact card, ninjutsu, a creature
// card's graveyard ability) is not stopped. That is the battlefield
// check in front of SourcesCantActivate's own body.
//
// No simplification.
func init() {
	matrix := SourcesCantActivate(
		"Damping Matrix — activated abilities of artifacts and creatures can't be activated unless they're mana abilities.",
		Or(Artifact(), Creature()), true)
	forbids := matrix.Forbids
	matrix.Forbids = func(q game.ActivationQuery) bool {
		return q.FromZone == game.ZoneBattlefield && forbids(q)
	}
	Register(Spec{
		OracleID:               "10403a8c-2534-421e-8809-5fc019623452",
		Name:                   "Damping Matrix",
		Completeness:           CompletenessFull,
		ActivationRestrictions: []game.ActivationRestriction{matrix},
	})
}
