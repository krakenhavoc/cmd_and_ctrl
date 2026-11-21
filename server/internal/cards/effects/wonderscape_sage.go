package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wonderscape Sage — Creature — Moonfolk Wizard {1}{U}, 1/3 (Modern Horizons
// 3 Commander, #5948):
//
//	"Flying
//	 {T}, Return a land you control to its owner's hand: Draw a card. Then
//	 discard a card unless that land had a nonbasic land type."
//
// The return is a cost (CR 602.2b): the activator names the land as the
// ability is announced, and the engine records which permanent it was
// (PaidCost.ReturnedObjects). "That land had a nonbasic land type" is read
// from the land as it last existed on the battlefield (CR 400.7j, CR
// 608.2h), so a land made a Desert by an effect counts and a Desert that
// lost the type does not. Nonbasic land types are the CR 205.3i land types
// other than the five basic ones: Cave, Desert, Gate, Lair, Locus, Mine,
// Planet, Power-Plant, Sphere, Tower, Town and Urza's. The card is drawn
// first, so it is a legal discard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "99994c36-21ec-4a28-b850-d3f2ba8b7f19",
		Name:            "Wonderscape Sage",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Return a land you control to its owner's hand: Draw a card. Then discard a card unless that land had a nonbasic land type.",
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
			Cost:    Plus(TapCost(), ReturnAPermanentToHand("a land you control", Land())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				nonbasic := false
				for _, info := range ctx.ReturnedPermanents() {
					for _, sub := range info.Characteristic.Subtypes {
						if game.IsLandType(sub) && !game.IsBasicLandType(sub) {
							nonbasic = true
						}
					}
				}
				if nonbasic {
					return DrawCards{N: 1}.Apply(ctx)
				}
				return drawThenDiscard(g, item.Controller, item.SourceCardID, 1, 1, "")
			},
		}},
	})
}
