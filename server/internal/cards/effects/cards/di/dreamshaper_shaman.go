package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dreamshaper Shaman — Enchantment Creature — Minotaur Shaman {5}{R}, 5/4
// (Theros Beyond Death):
//
//	"At the beginning of your end step, you may pay {2}{R} and sacrifice a
//	 nonland permanent. If you do, reveal cards from the top of your library
//	 until you reveal a nonland permanent card. Put that card onto the
//	 battlefield and the rest on the bottom of your library in a random
//	 order."
//
// The "may" is one decision over two payments, and a cost nobody can pay is
// not offered (CR 118.3): with no nonland permanent to sacrifice nobody is
// asked about the mana. The mana is the pay prompt (CR 118.12); once it is
// paid, the sacrifice is mandatory (a floor of one, Dreadfeast Demon's) over
// your nonland permanents, the Shaman itself included. Only a permanent that
// really left starts the reveal (Kethek's keyed "if you do"): the run stops
// on the first nonland permanent card (an instant or sorcery does not stop
// it), puts it onto the battlefield, and puts every other revealed card on the
// bottom in a random order; a library with no such card puts everything
// revealed there.
//
// No simplification.
func init() {
	const label = "Dreamshaper Shaman — you may pay {2}{R} and sacrifice a nonland permanent; if you do, reveal until a nonland permanent card and put it onto the battlefield"
	effects.Register(effects.Spec{
		OracleID:     "d06ef178-f02a-45a4-8193-140dee52c99d",
		Name:         "Dreamshaper Shaman",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep(label, func(g *game.Game, item *game.StackItem) error {
				if len(effects.PermanentsControlledByMatching(g, item.Controller, effects.Nonland())) == 0 {
					return nil
				}
				return effects.MayPay{
					Chooser:  item.Controller,
					Cost:     "{2}{R}",
					Question: "Dreamshaper Shaman — pay {2}{R} and sacrifice a nonland permanent to reveal until a nonland permanent card?",
					Then:     dreamshaperPaid,
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}

// dreamshaperPaid is the sacrifice half of the payment, asked once the mana
// has been paid. An on-disk key: never renamed, never reused.
var dreamshaperPaid = effects.PayUnlessBranches("pay-unless/dreamshaper-shaman-sacrifice",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.MustSacrificeOne(ctx.Game, ctx.Item, "Dreamshaper Shaman — sacrifice a nonland permanent",
			effects.Nonland(), dreamshaperSacrificed)
	}, nil)

// dreamshaperSacrificed is the "if you do" of the sacrifice. An on-disk key:
// never renamed, never reused.
var dreamshaperSacrificed = effects.RegisterSacrificedOne("prompt-run/dreamshaper-shaman-sacrifice", func(ctx *effects.Context, _ uuid.UUID) error {
	return effects.RevealUntilThenPutOntoBattlefield{
		Match:  func(c game.Card) bool { return c.IsPermanent() && !c.IsLand() },
		Reason: "Dreamshaper Shaman — a nonland permanent card",
	}.Apply(ctx)
})
