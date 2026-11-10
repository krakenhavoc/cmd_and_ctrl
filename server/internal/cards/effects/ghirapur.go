package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ghirapur — Plane — Avishkar (#4870):
//
//	"At the beginning of combat on your turn, until end of turn, each
//	 noncreature, non-Vehicle artifact you control becomes a 5/3 Vehicle
//	 in addition to its other types and gains trample, haste, and crew 2.
//	 Whenever chaos ensues, return target noncreature artifact card from
//	 your graveyard to your hand."
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6).
// See planesB2BecomeVehicles: the set is locked as the ability resolves (the
// ruling) and the affected artifacts keep their other types and abilities.
// The chaos ability (CR 311.7) targets a noncreature artifact card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a7c6b72c-f9ec-4048-8e4c-99210e4e7b85",
		Name:         "Ghirapur",
		Completeness: CompletenessFull,
		Grants:       []AbilityGrant{planesB2CrewGrant(ghirapurCrewGrant, 2)},
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat("Ghirapur — until end of turn, each noncreature, non-Vehicle artifact you control becomes a 5/3 Vehicle and gains trample, haste, and crew 2",
				planesB2BecomeVehicles("Ghirapur — 5/3 Vehicles with trample, haste and crew 2", ghirapurCrewGrant,
					And(Not(Creature()), Not(HasSubtype("Vehicle"))), 5, 3, "trample", "haste")),
			Targeting(WheneverChaosEnsues("Ghirapur — return target noncreature artifact card from your graveyard to your hand",
				returnFirstLegalGraveyardTargetToHand),
				TargetCardInGraveyard("target noncreature artifact card from your graveyard", Artifact(), Not(Creature()), YouOwn())),
		},
	})
}

// ghirapurCrewGrant is the granted bundle's key. An on-disk key.
const ghirapurCrewGrant = "ghirapur/crew-2"
