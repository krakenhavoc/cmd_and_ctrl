package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Éomer, King of Rohan — Legendary Creature — Human Noble {3}{R}{W}, 2/2
// (Tales of Middle-earth Commander):
//
//	"Double strike
//	 Éomer enters with a +1/+1 counter on it for each other Human you
//	 control.
//	 When Éomer enters, target player becomes the monarch. Éomer deals
//	 damage equal to its power to any target."
//
// Squad Captain's entry counters, counted over the other Humans the caster
// controls as Éomer enters (CR 614.1c), so they are on him before the enter
// trigger reads his power. The trigger is Survey Mechan's two-slot clause,
// "target player" (slot 0) then "any target" (slot 1), each judged on its
// own at resolution (CR 608.2b); the damage is his power as the ability
// resolves, or as he last existed on the battlefield (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4f8191e6-a6aa-4426-9014-eff4e0efee9e",
		Name:            "Éomer, King of Rohan",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Replacements: []game.ReplacementEffect{
			effects.B19EntersWithCountersCounted(game.CounterPlusOne, s82GapB118OtherHumansYouControl, "Éomer enters with a +1/+1 counter on it for each other Human you control"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Éomer, King of Rohan — target player becomes the monarch; Éomer deals damage equal to its power to any target",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if t, ok := ctx.ClauseTarget(0); ok {
						if err := (effects.BecomeTheMonarch{Player: t.ID}).Apply(ctx); err != nil {
							return err
						}
					}
					if t, ok := ctx.ClauseTarget(1); ok {
						if power := effects.SourcePowerNow(ctx); power > 0 {
							return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: power}.Apply(ctx)
						}
					}
					return nil
				}),
				effects.TargetPlayer("target player").Then(effects.TargetAny())),
		},
	})
}

// s82GapB118OtherHumansYouControl counts the other Humans src's controller
// controls, read through the layered view.
func s82GapB118OtherHumansYouControl(g *game.Game, src *game.Card) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == src.Controller && c.InstanceID != src.InstanceID && c.HasSubtype(src.TextSubtype("Human")) {
			n++
		}
	}
	return n
}
