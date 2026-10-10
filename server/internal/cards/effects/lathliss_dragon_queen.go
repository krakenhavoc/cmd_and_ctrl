package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lathliss, Dragon Queen — Legendary Creature — Dragon {4}{R}{R}:
//
//	"Flying
//	 Whenever another nontoken Dragon you control enters, create a 5/5
//	 red Dragon creature token with flying.
//	 {1}{R}: Dragons you control get +1/+0 until end of turn."
//
// The token is a Dragon, but it is a token, so it never triggers a
// second Lathliss — and a Dragon that arrives as a token copy
// (Miirym's) does not either. The pump locks its set as it resolves
// (CR 611.2c): a Dragon that enters afterwards is not pumped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "15fef45d-f1a9-49b2-abaa-fe77bb9d1afd",
		Name:            "Lathliss, Dragon Queen",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AnotherNontokenCreatureOfTypeEnteredUnderYourControl("Dragon"),
				"Lathliss, Dragon Queen — create a 5/5 red Dragon creature token with flying",
				Do(CreateToken{Template: TokenCard("5/5 red Dragon with flying"), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}: Dragons you control get +1/+0 until end of turn",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{1}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return BoostUntilEOT{
					Match: And(OfCreatureType("Dragon"), YouControl()),
					Power: 1,
					Label: "Lathliss, Dragon Queen — Dragons you control get +1/+0",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
