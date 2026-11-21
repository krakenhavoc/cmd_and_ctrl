package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gwenna, Eyes of Gaea — Legendary Creature — Elf Druid Scout {2}{G},
// 2/3:
//
//	"{T}: Add two mana in any combination of colors. Spend this mana
//	 only to cast creature spells or activate abilities of creature
//	 sources.
//	 Whenever you cast a creature spell with power 5 or greater, put a
//	 +1/+1 counter on Gwenna and untap it."
//
// The mana is two independent any-colour slots (AnyCombinationOfColors)
// carrying one restriction tag that admits either alternative:
// casting a creature spell, or activating an ability whose SOURCE is a
// creature (CR 106.6 reads the source's types). Mana abilities with a
// mana cost are activations too. An unknown purpose is a refusal, so
// the mana can't pay anything else.
//
// The trigger reads the spell's power on the stack. It untaps Gwenna
// whether or not the counter could be placed, and does nothing to a
// Gwenna that has left the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7d95e72a-461e-4993-a190-847466a4b17c",
		Name:         "Gwenna, Eyes of Gaea",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: effects.AnyCombinationOfColors(2),
			Label:    "Add two mana in any combination of colors. Spend this mana only to cast creature spells or activate abilities of creature sources",
			Restrictions: []string{game.ManaRestrictAnyOf(
				[]string{game.ManaRestrictCast, game.ManaRestrictType("Creature")},
				[]string{game.ManaRestrictActivate, game.ManaRestrictType("Creature")},
			)},
		}},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.And(effects.Creature(), effects.PowerGE(5)),
				"Gwenna, Eyes of Gaea — put a +1/+1 counter on Gwenna and untap it",
				effects.Do(
					// The untap waits for a prompt the counter asks (CR 608.2c, #4155).
					effects.Instruction(effects.PlusOneCounterOnThisIfStillHere),
					effects.Instruction(func(ctx *effects.Context) error {
						if !effects.SourceIsStillThisPermanent(ctx.Game, ctx.Item) {
							return nil
						}
						return effects.UntapTarget{Target: ctx.Item.SourceCardID}.Apply(ctx)
					}),
				)),
		},
	})
}
