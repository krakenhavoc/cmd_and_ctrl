package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Finale of Eternity — Sorcery {X}{B}{B}:
//
//	"Destroy up to three target creatures with toughness X or less. If
//	 X is 10 or more, return all creature cards from your graveyard to
//	 the battlefield."
//
// ADR 0109 §9 (#1842): the toughness bound reads the announced X,
// chosen before the targets (CR 601.2b–c) and re-checked as the spell
// resolves (CR 608.2b). The destruction comes first, in the printed
// order, so with X of 10 or more a creature of yours it destroyed is
// in your graveyard when the creature cards come back, and returns
// with them — the printed interaction. The return is one simultaneous
// entry (ReturnFromGraveyardTogether).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3113afec-d4ae-46b5-9952-89e2e2c9ae7b",
		Name:         "Finale of Eternity",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Targets: effects.TargetCreature("up to three target creatures with toughness X or less").
			WithCount(0, 3).WithToughnessAtMostX(),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			if ctx.X() < 10 {
				return nil
			}
			return effects.ReturnFromGraveyardTogether{
				Targets: effects.GraveyardCardIDs(ctx, item.Controller, func(c game.Card) bool { return c.IsCreature() }),
			}.Apply(ctx)
		},
	})
}
