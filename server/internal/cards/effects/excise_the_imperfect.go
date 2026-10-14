package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Excise the Imperfect — Instant {1}{W}{W} (March of the Machine):
//
//	"Exile target nonland permanent. Its controller incubates X, where
//	 X is its mana value."
//
// "Its controller" is the permanent's controller as it was on the
// battlefield, and X is its mana value there (an X in a cost is 0 on
// the battlefield, which is what Card.ManaValue reads, and a token has
// none). Both are read BEFORE the exile, as last-known information,
// because the card is in another zone afterwards. The incubate
// (CR 701.53a, ADR 0185) is that player's, not the caster's, and it
// happens even if a replacement sent the permanent somewhere other than
// exile.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d3264482-2a5a-45e8-987f-593b8aceac12",
		Name:         "Excise the Imperfect",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target nonland permanent", Not(Land())),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				card, ok := ctx.Game.LookupCardForEffect(t.ID)
				if !ok {
					continue
				}
				controller, x := card.Controller, card.ManaValue()
				if err := (ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
				return Incubate{Controller: controller, N: x}.Apply(ctx)
			}
			return nil
		},
	})
}
