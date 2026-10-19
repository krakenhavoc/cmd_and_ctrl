package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Castle Locthwain — Land (EDHREC rank 790):
//
//	"This land enters tapped unless you control a Swamp.
//	 {T}: Add {B}.
//	 {1}{B}{B}, {T}: Draw a card, then you lose life equal to the
//	 number of cards in your hand."
//
// Black's card-draw land: cheap when the hand is empty, expensive when
// it isn't, which is the whole design. The life is lost AFTER the draw
// and counts the drawn card — "then" is the printed order, and it is
// the order here.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "be811e70-aaaa-41f3-bf9e-5d3f9f719b49",
		Name:         "Castle Locthwain",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{b06EntersTappedUnlessLandType("swamp")},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "Add {B}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}{B}, {T}: Draw a card, then you lose life equal to the number of cards in your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{1}{B}{B}"), TapCost()),
			Effect: Do(
				// The loss counts the hand after the draw, and the draw's prompt
				// (a dredge offer) comes first (CR 608.2c, #4155).
				DrawCards{N: 1},
				Instruction(func(ctx *Context) error {
					p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
					if p == nil {
						return nil
					}
					return GainLife{Amount: -p.Hand.Size()}.Apply(ctx)
				}),
			),
		}},
	})
}
