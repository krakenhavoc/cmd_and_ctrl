package effects

// Welding Jar — Artifact {0}:
//
//	"Sacrifice this artifact: Regenerate target artifact."
//
// A free artifact whose whole job is to sit there until somebody
// points a Shatterstorm at the rock you actually care about. One
// shield (CR 701.19a), one use, and the Jar is gone.
//
// Note what the shield does to an ARTIFACT, which is the part players
// misremember: tap it, remove all damage from it, remove it from
// combat. A tapped mana rock is a mana rock that has not produced
// this turn, so saving a Sol Ring with this costs you the Sol Ring's
// mana for the turn.
//
// Shatterstorm says the artifacts can't be regenerated (CR 701.19c),
// so the Jar does not save anything from that one — and the shield it
// made is not even spent trying (CR 701.19c), which matters because
// the Jar itself is already in the graveyard by then.
//
// The sacrifice is a COST, paid at announce, so responding to the
// ability by destroying the Jar does not stop the regeneration.
func init() {
	Register(Spec{
		OracleID:     "e4a6c421-bef1-4360-a2dc-ae9c61d86b2f",
		Name:         "Welding Jar",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice this artifact: Regenerate target artifact.",
			Cost:    SacrificeThis(),
			Targets: TargetPermanent("target artifact", Artifact()),
			Effect:  regenerateTheTargetPermanent,
		}},
	})
}
