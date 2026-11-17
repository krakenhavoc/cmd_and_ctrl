package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sheoldred's Restoration — Sorcery {3}{B} (Dominaria United):
//
//	"Kicker {2}{W}
//	 Return target creature card from your graveyard to the battlefield. If
//	 this spell was kicked, you gain life equal to that card's mana value.
//	 Otherwise, you lose that much life.
//	 Exile Sheoldred's Restoration."
//
// Reanimate's single-target body, which hands back the card as it sat in the
// graveyard so the life is its mana value there (CR 608.2h, an X is 0). The
// life gain or loss is the caster's. "Exile Sheoldred's Restoration" is done
// first, as Bound's is: the spell is on the stack, not in the graveyard, so
// the order is not observable, and the engine does not route it to the
// graveyard while a prompt from the return is open. A target gone by
// resolution makes the spell do nothing (CR 608.2b), though it is still
// exiled.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "df587772-516f-4ae8-aefe-e6fc63f8b365",
		Name:          "Sheoldred's Restoration",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Kicker("{2}{W}")},
		Targets:       targetCreatureInYourGraveyard(),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (ExileTarget{Target: ctx.Source()}).Apply(ctx); err != nil {
				return err
			}
			card, ok := reanimateSingleTarget(ctx, ctx.Controller())
			if !ok {
				return nil
			}
			mv := card.ManaValue()
			if ctx.WasKicked() {
				return GainLife{Player: ctx.Controller(), Amount: mv}.Apply(ctx)
			}
			return GainLife{Player: ctx.Controller(), Amount: -mv}.Apply(ctx)
		},
	})
}
