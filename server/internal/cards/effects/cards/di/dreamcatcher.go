package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dreamcatcher — Creature — Spirit {U}, 1/1 (Saviors of Kamigawa, #5601):
//
//	"Whenever you cast a Spirit or Arcane spell, you may sacrifice this
//	 creature. If you do, draw a card."
//
// The "may" is asked as the trigger resolves (CR 603.5). "If you do"
// holds only when the Dreamcatcher is still on the battlefield as the same
// object (CR 400.7) and so is really sacrificed: Sanctum of Ugin's shape.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "61529c3a-9b7e-4e73-9f6e-58c992ef1ffc",
		Name:         "Dreamcatcher",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.WheneverYouCast(effects.ChkB1SpiritOrArcane(), "Dreamcatcher — sacrifice it to draw a card", dreamcatcherSacrificeToDraw),
				"Dreamcatcher — sacrifice this creature to draw a card?"),
		},
	})
}

// dreamcatcherSacrificeToDraw sacrifices the Dreamcatcher and, if it did,
// draws a card.
func dreamcatcherSacrificeToDraw(g *game.Game, item *game.StackItem) error {
	if !effects.OnBattlefield(g, item.SourceCardID) || effects.SourceIsNewObject(g, item) {
		return nil
	}
	ctx := effects.NewContext(g, item)
	if err := (effects.SacrificePermanent{Target: item.SourceCardID}).Apply(ctx); err != nil {
		return err
	}
	return effects.DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
}
