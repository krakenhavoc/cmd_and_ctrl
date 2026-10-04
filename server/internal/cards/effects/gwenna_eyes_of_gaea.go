package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "7d95e72a-461e-4993-a190-847466a4b17c",
		Name:         "Gwenna, Eyes of Gaea",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: AnyCombinationOfColors(2),
			Label:    "Add two mana in any combination of colors. Spend this mana only to cast creature spells or activate abilities of creature sources",
			Restrictions: []string{game.ManaRestrictAnyOf(
				[]string{game.ManaRestrictCast, game.ManaRestrictType("Creature")},
				[]string{game.ManaRestrictActivate, game.ManaRestrictType("Creature")},
			)},
		}},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(And(Creature(), PowerGE(5)),
				"Gwenna, Eyes of Gaea — put a +1/+1 counter on Gwenna and untap it",
				func(g *game.Game, item *game.StackItem) error {
					if !sourceIsStillThisPermanent(g, item) {
						return nil
					}
					ctx := NewContext(g, item)
					if err := (AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
						return err
					}
					return UntapTarget{Target: item.SourceCardID}.Apply(ctx)
				}),
		},
	})
}
