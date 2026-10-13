package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cremate — Instant {B}:
//
//	"Exile target card from a graveyard.
//	 Draw a card."
//
// Graveyard hate that replaces itself. The draw is a separate
// sentence, so it happens whether or not the exile did: a CR 614
// replacement that keeps the card in its graveyard does not stop the
// cantrip, and a spell whose only target left is countered on
// resolution (CR 608.2b) and draws nothing. (#3213, the Kuja deck.)
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c6a2e410-b182-48d1-aeb2-bc8de27e9cd2",
		Name:         "Cremate",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target card from a graveyard"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			controller := item.Controller
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return DrawCards{Player: controller, N: 1}.Apply(ctx)
		},
	})
}
