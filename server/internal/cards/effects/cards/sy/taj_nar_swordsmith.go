package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Taj-Nar Swordsmith — Creature — Cat Soldier {3}{W}, 2/3 (Commander 2017, S82
// older sets):
//
//	"When this creature enters, you may pay {X}. If you do, search your
//	 library for an Equipment card with mana value X or less, put that card
//	 onto the battlefield, then shuffle."
//
// Decree of Justice's MayPayX: X is chosen up to what the controller can pay,
// the payment is offered, and only a paid X runs the search (CR 608.2d). The
// search is for one Equipment card whose mana value is at most X.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dcbf8d72-3367-43e1-9a71-e674e0ed52d9",
		Name:         "Taj-Nar Swordsmith",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Taj-Nar Swordsmith — you may pay {X}; if you do, search for an Equipment card with mana value X or less and put it onto the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPayX{
						Cost:  "{X}",
						Label: "Taj-Nar Swordsmith",
						Buys:  "search for an Equipment card with mana value X or less",
						Goal:  effects.XAsHighAsYouCan,
						Unit:  game.PayAmountOther,
						Then:  tajNarPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// tajNarPaid is the paid half, keyed: the X paid rides the prompt.
var tajNarPaid game.PayUnlessElse

func init() {
	tajNarPaid = effects.PayUnlessBranches("pay-unless/taj-nar-swordsmith-paid",
		func(ctx *effects.Context, r game.PayUnlessAnswer) error {
			x := max(r.Number, 0)
			return effects.SearchLibrary{
				Player: ctx.Controller(),
				Predicate: func(c game.Card) bool {
					return c.HasSubtype("Equipment") && c.ManaValue() <= x
				},
				Dest:    game.ZoneBattlefield,
				Limit:   1,
				Reveal:  true,
				Shuffle: true,
				Reason:  "Taj-Nar Swordsmith — an Equipment card with mana value X or less",
			}.Apply(ctx)
		}, nil)
}
