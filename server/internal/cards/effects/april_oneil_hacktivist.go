package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// April O'Neil, Hacktivist — Legendary Creature — Human Scientist
// {3}{U}, 1/5 (Teenage Mutant Ninja Turtles, #4457, #4677):
//
//	"At the beginning of your end step, draw a card for each card type
//	 among spells you've cast this turn."
//
// The count is CastTally.CardTypes, the card types (CR 205.2a) of each
// spell you cast this turn, recorded as the spell became cast (CR
// 601.2i), so a spell that has since left every zone still counts, and
// an MDFC cast as its back face counts that face's types. An artifact
// creature spell is two types, a kindred instant two; a copy is not cast
// (CR 707.10) and a land played is not a spell (CR 305.1). The number is
// read as the ability resolves, so a spell cast in response to it
// counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d379a919-02cd-46fd-a6ce-b66a77745ae7",
		Name:         "April O'Neil, Hacktivist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("April O'Neil, Hacktivist — draw a card for each card type among spells you've cast this turn",
				func(g *game.Game, item *game.StackItem) error {
					n := len(g.CastTallyFor(item.Controller).CardTypesCast())
					return DrawCards{Player: item.Controller, N: n}.Apply(NewContext(g, item))
				}),
		},
	})
}
