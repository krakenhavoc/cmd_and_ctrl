package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Guidelight Pathmaker — Artifact — Vehicle {4}{W}{U}, 6/5 (Aetherdrift,
// #4464):
//
//	"Vigilance
//	 When this Vehicle enters, you may search your library for an artifact
//	 card and reveal it. Put it onto the battlefield if its mana value is
//	 2 or less. Otherwise, put it into your hand. If you search your
//	 library this way, shuffle.
//	 Crew 2"
//
// The search is optional, revealed, and shuffles only if it was made (the
// 2024 Oracle update); a failure to find is allowed (CR 701.23b). The card
// goes to the hand and a mana value of 2 or less is then put onto the
// battlefield from it (Archdruid's Charm's tail), read as it is found; an
// {X} in a card's mana cost is 0 there (CR 202.3e). Crew is the engine's
// (CR 702.122a).
//
// No simplification.
func init() {
	trigger := WhenThisEnters("Guidelight Pathmaker — you may search your library for an artifact card; put it onto the battlefield if its mana value is 2 or less, otherwise into your hand",
		func(g *game.Game, item *game.StackItem) error {
			return SearchLibrary{
				Player:    item.Controller,
				Predicate: func(c game.Card) bool { return c.IsArtifact() },
				Dest:      game.ZoneHand,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
				Optional:  true,
				Reason:    "Guidelight Pathmaker — an artifact card",
				ThenTail:  game.KeyedTail{Ref: thenPutFoundCheapArtifactsOntoBattlefield, Player: item.Controller, Source: item.SourceCardID},
			}.Apply(NewContext(g, item))
		})
	trigger.Purpose = game.Purpose{Tutors: 1}
	Register(Spec{
		OracleID:        "f8103865-5086-4f16-a35a-a56acdab1547",
		Name:            "Guidelight Pathmaker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered:       []game.TriggeredAbility{trigger},
		Activated: []ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Guidelight Pathmaker"),
		}},
	})
}
