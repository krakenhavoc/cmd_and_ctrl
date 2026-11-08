package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cactarantula — Creature — Plant Spider {4}{G}{G}, 6/5 (Outlaws of
// Thunder Junction, #4468):
//
//	"This spell costs {1} less to cast if you control a Desert.
//	 Reach
//	 Whenever this creature becomes the target of a spell or ability an
//	 opponent controls, you may draw a card."
//
// The discount is Prehistoric Turtlesaurus's self cost modifier, its
// condition any Desert you control (b02IsDesert, the post-layer land
// subtype). The draw is Leovold, Emissary of Trest's optional draw on
// Black Bolt's "becomes the target of a spell or ability an opponent
// controls" (the event's actor is the targeting object's controller).
// Your own spells and abilities never trigger it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "18a5fe34-0c97-486c-b09a-80535a99e982",
		Name:            "Cactarantula",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		SelfCostModifiers: []game.CostModifier{
			CostsLess(1, "This spell costs {1} less to cast if you control a Desert.", YouControlA(b02IsDesert)),
		},
		Triggered: []game.TriggeredAbility{
			OptionalDraw(On(game.EventBecomesTarget, blackBoltTargetedByAnOpponent,
				"Cactarantula — you may draw a card", Do(DrawCards{N: 1})),
				"Cactarantula — draw a card?", 1),
		},
	})
}
