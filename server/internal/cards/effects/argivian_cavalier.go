package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Argivian Cavalier — Creature — Orc Knight {2}{W}, 2/2 (Dominaria United,
// #5487, seam #5658):
//
//	"Enlist (As this creature attacks, you may tap a nonattacking creature
//	 you control without summoning sickness. When you do, add its power to
//	 this creature's until end of turn.)
//	 When this creature enters, create a 1/1 white Soldier creature token."
//
// Enlist is the engine's keyword (CR 702.154, game/enlist.go): the creature
// to tap is chosen with the attack declaration, tapped as the declaration
// locks in, and the linked trigger adds its power as it resolves.
// The Soldier is S.H.I.E.L.D. Deployment Drone's token; one that has been
// under your control since the turn began can be the creature the Cavalier
// enlists.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c45090bd-a589-45a0-8388-025b8e0b3487",
		Name:            "Argivian Cavalier",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordEnlist},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Argivian Cavalier — create a 1/1 white Soldier", Do(CreateToken{Template: TokenCard("1/1 white Soldier"), N: 1})),
		},
	})
}
