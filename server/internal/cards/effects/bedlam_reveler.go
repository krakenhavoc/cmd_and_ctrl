package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bedlam Reveler — Creature — Devil Horror {6}{R}{R}, 3/4 (Innistrad
// Remastered, S82 GAP, #5948):
//
//	"This spell costs {1} less to cast for each instant and sorcery card
//	 in your graveyard.
//	 Prowess (Whenever you cast a noncreature spell, this creature gets
//	 +1/+1 until end of turn.)
//	 When this creature enters, discard your hand, then draw three cards."
//
// Cyan's self cost modifier over instant and sorcery cards, prowess as the
// shared keyword, and Case of the Crimson Pulse's hand dump: the whole hand
// goes first (a commander's command-zone offer can pause it), and the
// three cards are drawn only after it has finished (CR 608.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "01a7e0ce-c98f-4cf2-aba3-360aa0a9b9a7",
		Name:            "Bedlam Reveler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"prowess"},
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(instantAndSorceryCardsInYourGraveyard,
				"This spell costs {1} less to cast for each instant and sorcery card in your graveyard."),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Bedlam Reveler — discard your hand, then draw three cards", Do(
				Instruction(func(ctx *Context) error {
					_, err := discardWholeHand(ctx.Game, ctx.Controller())
					return err
				}),
				DrawCards{N: 3},
			)),
		},
	})
}
