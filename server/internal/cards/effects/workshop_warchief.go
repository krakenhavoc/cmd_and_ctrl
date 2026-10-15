package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Workshop Warchief — Creature — Rhino Warrior {3}{G}{G}, 5/3:
//
//	"Trample
//	 When this creature enters, you gain 3 life.
//	 When this creature dies, create a 4/4 green Rhino Warrior creature
//	 token.
//	 Blitz {4}{G}{G}"
//
// Blitz (CR 702.152, #3772): blitzed, the Warchief gains 3 life, swings
// with haste, and at the end step leaves a 4/4 and a card behind.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "0695af85-7405-4512-bb97-42517a532ada",
		Name:             "Workshop Warchief",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"trample"},
		AlternativeCosts: []game.AlternativeCost{Blitz("{4}{G}{G}")},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Workshop Warchief — you gain 3 life", Do(GainLife{Amount: 3})),
			WhenThisDies("Workshop Warchief — create a 4/4 green Rhino Warrior",
				Do(CreateToken{Template: TokenCard("4/4 green Rhino Warrior"), N: 1})),
		},
	})
}
