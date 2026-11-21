package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vampire Gourmand — Creature — Vampire {1}{B}, 2/2 (Foundations, #4465):
//
//	"Whenever this creature attacks, you may sacrifice another creature.
//	 If you do, draw a card and this creature can't be blocked this
//	 turn."
//
// Beetle-Headed Merchants' attack trigger (maySacrificeOne, asked at
// resolution over the other creatures you control then), with Sword
// Coast Sailor's "can't be blocked this turn" restriction on itself in
// place of the counter. The restriction goes on the Gourmand only while
// it is still the object that attacked (CR 400.7); the card is drawn
// either way. Declining, or having no other creature, does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3386414c-32fa-40ae-9b30-dab976a158c1",
		Name:         "Vampire Gourmand",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Vampire Gourmand — you may sacrifice another creature to draw a card and be unblockable",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MaySacrificeOne(g, item, "Vampire Gourmand — sacrifice another creature to draw a card, and this creature can't be blocked this turn?",
						effects.Creature(), true, vampireGourmandSacrificed)
				}),
		},
	})
}

var vampireGourmandSacrificed = effects.RegisterSacrificedOne("prompt-run/vampire-gourmand-draw-unblockable", func(ctx *effects.Context, _ uuid.UUID) error {
	if err := (effects.DrawCards{Player: ctx.Controller(), N: 1}).Apply(ctx); err != nil {
		return err
	}
	if effects.SourceIsNewObject(ctx.Game, ctx.Item) || !effects.OnBattlefield(ctx.Game, ctx.Source()) {
		return nil
	}
	return effects.RestrictUntilEOT{
		Target:       ctx.Source(),
		Restrictions: game.CantBeBlocked,
		Label:        "Vampire Gourmand — can't be blocked",
	}.Apply(ctx)
})
