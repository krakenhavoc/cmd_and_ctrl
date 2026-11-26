package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Invasive Surgery — Instant {U} (Shadows over Innistrad, #5543):
//
//	"Counter target sorcery spell.
//	 Delirium — If there are four or more card types among cards in your
//	 graveyard, search the graveyard, hand, and library of that spell's
//	 controller for any number of cards with the same name as that spell,
//	 exile those cards, then that player shuffles."
//
// Quash's body on a sorcery, with "any number" in place of "all": the
// spell's name and controller are read while it is still on the stack,
// then it is countered, and with delirium (read as this resolves) its
// controller's graveyard, hand and library are searched. The caster takes
// what they like from the hidden zones, and the library is shuffled
// whatever was taken. A spell that can't be countered stays on the stack
// and the search still happens; a target that has left the stack makes
// the spell do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "545f56c6-b10f-4868-8aa8-ea3154980854",
		Name:         "Invasive Surgery",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target sorcery spell", effects.Sorcery()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			target, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			spell := ctx.Game.StackItemForEffect(target)
			c, found := ctx.Game.LookupCardForEffect(target)
			if spell == nil || !found {
				return nil
			}
			victim, names := spell.Controller, effects.ExtractionNames(c)
			if err := (effects.CounterTarget{StackID: target}).Apply(ctx); err != nil {
				return err
			}
			if !effects.DrcDeliriumMet(ctx.Game, ctx.Controller()) {
				return nil
			}
			return effects.ExtractionSearch(ctx, "Invasive Surgery", victim, names, false)
		},
	})
}
