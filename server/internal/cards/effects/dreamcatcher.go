package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "61529c3a-9b7e-4e73-9f6e-58c992ef1ffc",
		Name:         "Dreamcatcher",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(WheneverYouCast(chkB1SpiritOrArcane(), "Dreamcatcher — sacrifice it to draw a card", dreamcatcherSacrificeToDraw),
				"Dreamcatcher — sacrifice this creature to draw a card?"),
		},
	})
}

// dreamcatcherSacrificeToDraw sacrifices the Dreamcatcher and, if it did,
// draws a card.
func dreamcatcherSacrificeToDraw(g *game.Game, item *game.StackItem) error {
	if !onBattlefield(g, item.SourceCardID) || sourceIsNewObject(g, item) {
		return nil
	}
	ctx := NewContext(g, item)
	if err := (SacrificePermanent{Target: item.SourceCardID}).Apply(ctx); err != nil {
		return err
	}
	return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
}
