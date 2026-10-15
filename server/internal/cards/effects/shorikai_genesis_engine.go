package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shorikai, Genesis Engine — {2}{W}{U} Legendary Artifact — Vehicle, 8/8:
//
//	"{1}, {T}: Draw two cards, then discard a card. Create a 1/1
//	 colorless Pilot creature token with "This token crews Vehicles as
//	 though its power were 2 greater."
//	 Crew 8"
//
// The token is CrewPilotToken (ADR 0186 section 5): it counts 3 toward a
// crew cost and only its plain 1 toward a saddle cost. The draw and the
// discard are one go (the discard is a prompt built from the post-draw
// hand, so a card just drawn is a legal pitch); the token is a separate
// sentence and does not wait on the discard. The {T} is the Vehicle's own
// tap, which a Vehicle that is not a creature can always pay.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e2962883-6daf-4a84-af73-917afc73092c",
		Name:         "Shorikai, Genesis Engine",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{1}, {T}: Draw two cards, then discard a card. Create a 1/1 colorless Pilot creature token with \"This token crews Vehicles as though its power were 2 greater.\"",
				Cost:    Plus(ManaCost("{1}"), TapCost()),
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Effect:  shorikaiDrawAndMakePilot,
			},
			{
				Label:   "Crew 8",
				Cost:    CrewCost(8),
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect:  CrewEffect("Shorikai, Genesis Engine"),
			},
		},
	})
}

func shorikaiDrawAndMakePilot(g *game.Game, item *game.StackItem) error {
	if err := drawThenDiscard(g, item.Controller, item.SourceCardID, 2, 1, ""); err != nil {
		return err
	}
	return CreateToken{Template: CrewPilotToken(), N: 1}.Apply(NewContext(g, item))
}
