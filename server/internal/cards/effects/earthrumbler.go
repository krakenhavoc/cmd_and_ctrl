package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Earthrumbler — Artifact — Vehicle {4}{G}, 7/6 (Aetherdrift, #4464):
//
//	"Vigilance, trample
//	 Exile an artifact or creature card from your graveyard: This Vehicle
//	 becomes an artifact creature until end of turn.
//	 Crew 3 (Tap any number of creatures you control with total power 3
//	 or more: This Vehicle becomes an artifact creature until end of
//	 turn.)"
//
// The second way to animate is crew's effect (BecomeCreatureUntilEOT)
// behind Moorland Haunt's graveyard cost (ExileFromGraveyard): only an
// artifact or creature card in your own graveyard can pay, and with none
// there the ability is not offered. It is not crewing, so nothing that
// cares about the Vehicle becoming crewed sees it (CR 702.122e). The
// keywords are the Vehicle's whether or not it is a creature; crew is the
// engine's (CR 702.122a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ba5982a9-e8c1-4e7c-9a73-c4caad0f7336",
		Name:            "Earthrumbler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "trample"},
		Activated: []ActivatedAbility{
			{
				Label:   "Exile an artifact or creature card from your graveyard: This Vehicle becomes an artifact creature until end of turn.",
				Cost:    ExileFromGraveyard(1, "an artifact or creature card", func(c game.Card) bool { return c.IsArtifact() || c.IsCreature() }),
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return BecomeCreatureUntilEOT{Label: "Earthrumbler — becomes an artifact creature"}.Apply(NewContext(g, item))
				},
			},
			{
				Label:   "Crew 3",
				Cost:    CrewCost(3),
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect:  CrewEffect("Earthrumbler"),
			},
		},
	})
}
