package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Prodigy's Prototype — {1}{W}{U} Artifact — Vehicle, 3/4:
//
//	"Whenever one or more Vehicles you control attack, create a 1/1
//	 colorless Pilot creature token with "This token crews Vehicles as
//	 though its power were 2 greater."
//	 Crew 2"
//
// "One or more ... attack" is one trigger per attack declaration
// (OncePerBatch, CR 603.2c), Aloy's shape: two Vehicles attacking together
// make one Pilot, and the Prototype itself need not be among them. The
// token is CrewPilotToken (ADR 0186 section 5). A creature that is a
// Vehicle only because it was crewed counts: the test reads the attacker's
// subtypes as it attacks.
//
// No simplification.
func init() {
	const label = "Prodigy's Prototype — create a Pilot token"
	Register(Spec{
		OracleID:     "405a54b3-fb9f-4503-a4fc-3ba6a5e58205",
		Name:         "Prodigy's Prototype",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Prodigy's Prototype"),
		}},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, vehicleYouControlAttacked, label,
				Do(CreateToken{Template: CrewPilotToken(), N: 1}))),
		},
	})
}

// vehicleYouControlAttacked is "a Vehicle you control attacks".
func vehicleYouControlAttacked(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if source == nil || !attackDeclaredByYou(ev, source.Controller) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.HasSubtype("Vehicle")
}
