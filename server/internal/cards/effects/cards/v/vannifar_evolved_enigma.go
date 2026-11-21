package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vannifar, Evolved Enigma — Legendary Creature — Elf Ooze Wizard
// {2}{G}{U} (Murders at Karlov Manor, #4469):
//
//	"At the beginning of combat on your turn, choose one —
//	 • Cloak a card from your hand. (Put it onto the battlefield face down
//	   as a 2/2 creature with ward {2}. Turn it face up any time for its
//	   mana cost if it's a creature card.)
//	 • Put a +1/+1 counter on each colorless creature you control."
//
// The first mode is the cloak seam (CloakACardFromYourHand, game/cloak.go).
// The second is read through the layered view: a face-down creature is
// colorless (CR 708.2a) and so are Eldrazi and artifact creatures, and a
// creature a color-changing effect has colored is not. The set is fixed as
// the mode resolves (CR 611.2c).
//
// No simplification.
func init() {
	trigger := effects.AtBeginningOfYourCombat("Vannifar, Evolved Enigma — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	trigger.Modes = effects.ChooseOne(
		effects.ModeDoing("Cloak a card from your hand.", nil,
			func(_ *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.CloakACardFromYourHand{}.Apply(ctx)
			}),
		effects.ModeDoing("Put a +1/+1 counter on each colorless creature you control.", nil,
			func(_ *game.StackItem, ctx *effects.Context, _ int) error {
				for _, c := range ctx.Game.BattlefieldCardsForEffect() {
					if c.Controller != ctx.Controller() || !c.IsCreature() || len(c.EffectiveColors()) != 0 {
						continue
					}
					if err := (effects.AddCounter{Target: c.InstanceID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.AsGroupMember()); err != nil {
						return err
					}
				}
				return nil
			}),
	)
	effects.Register(effects.Spec{
		OracleID:     "e8aa0d7d-e5f6-4d66-bdc1-315483e1b256",
		Name:         "Vannifar, Evolved Enigma",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{trigger},
	})
}
