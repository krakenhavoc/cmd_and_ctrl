package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kappa Cannoneer — Artifact Creature — Turtle Warrior {5}{U}, 4/4:
//
//	"Improvise (Your artifacts can help cast this spell. Each artifact
//	 you tap after you're done activating mana abilities pays for {1}.)
//	 Ward {4}
//	 Whenever this creature or another artifact you control enters, put
//	 a +1/+1 counter on this creature. It can't be blocked this turn."
//
// Improvise is Spec.TapCost (Improvise(), CR 702.126), ward the shared
// Ward trigger, and the enters trigger watches the creature itself or any
// other artifact its controller controls. Both riders land on the
// Cannoneer: the counter, then RestrictUntilEOT's can't-be-blocked.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7fffbe7f-ad78-4bb6-a611-c37e495c6ee0",
		Name:         "Kappa Cannoneer",
		Completeness: effects.CompletenessFull,
		TapCost:      effects.Improvise(),
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{4}"), "Kappa Cannoneer — ward {4}"),
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.CardID == source.InstanceID {
					return true
				}
				c, ok := effects.EnteredUnderYourControl(ev, source, g, true)
				return ok && c.IsArtifact()
			}, "Kappa Cannoneer — +1/+1 counter; it can't be blocked this turn", effects.Do(
				// The restriction waits for a prompt the counter asks (CR 608.2c, #4155).
				effects.Instruction(func(ctx *effects.Context) error {
					return effects.AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
				}),
				effects.Instruction(func(ctx *effects.Context) error {
					return effects.RestrictUntilEOT{
						Target:       ctx.Source(),
						Restrictions: game.CantBeBlocked,
						Label:        "Kappa Cannoneer — can't be blocked",
					}.Apply(ctx)
				}),
			)),
		},
	})
}
