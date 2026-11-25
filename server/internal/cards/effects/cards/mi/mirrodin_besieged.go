package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mirrodin Besieged — Enchantment {2}{U} (Modern Horizons):
//
//	"As this enchantment enters, choose Mirran or Phyrexian.
//	 • Mirran — Whenever you cast an artifact spell, create a 1/1 colorless Myr
//	   artifact creature token.
//	 • Phyrexian — At the beginning of your end step, draw a card, then discard a
//	   card. Then if there are fifteen or more artifact cards in your graveyard,
//	   target opponent loses the game."
//
// The choice is made as it enters (CR 614.12), Monastery Siege's way, and each
// bullet is a trigger that works only for its own choice. The Phyrexian one
// targets an opponent as it goes on the stack; the draw, the discard, and then
// the test of the graveyard follow in order, the test read after the discard has
// really been made (CR 608.2c), and an opponent who is no longer a legal target
// does not lose.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "674e2683-31c0-4fee-95fb-98b1201e41e7",
		Name:         "Mirrodin Besieged",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseOptionAsEnters("Mirrodin Besieged", "Mirran", "Phyrexian"),
		Triggered: []game.TriggeredAbility{
			effects.WhenChosen("Mirran", effects.WheneverYouCast(effects.Artifact(), "Mirrodin Besieged — create a 1/1 colorless Myr artifact creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Myr artifact"), N: 1}))),
			effects.Targeting(effects.WhenChosen("Phyrexian", effects.AtYourEndStep("Mirrodin Besieged — draw a card, then discard a card; then if there are fifteen or more artifact cards in your graveyard, target opponent loses the game",
				effects.Do(effects.DrawCards{N: 1},
					effects.Instruction(func(ctx *effects.Context) error { return effects.DiscardOne(ctx.Game, ctx.Item, ctx.Controller()) }),
					effects.If{Cond: effects.CondBoard{Name: "there are fifteen or more artifact cards in your graveyard", Test: func(ctx *effects.Context) bool {
						return effects.OldB166ArtifactCardsInGraveyard(ctx.Game, ctx.Controller()) >= 15
					}}, Then: []effects.Applier{effects.Instruction(func(ctx *effects.Context) error {
						for _, t := range ctx.LegalTargets() {
							if t.Kind == game.TargetPlayer && t.ID != uuid.Nil {
								return effects.LoseTheGame{Player: t.ID}.Apply(ctx)
							}
						}
						return nil
					})}}))),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
