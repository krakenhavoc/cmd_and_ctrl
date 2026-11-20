package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dawnhart Mentor — Creature — Human Warlock {2}{G}, 0/4 (Midnight Hunt,
// #5497):
//
//	"When this creature enters, create a 1/1 white Human creature token.
//	 Coven — {5}{G}: Target creature you control gets +3/+3 and gains
//	 trample until end of turn. Activate only if you control three or more
//	 creatures with different powers."
//
// The coven gate is read when the ability is activated and is not
// rechecked on resolution (it is an activation restriction, not an
// intervening if).
//
// No simplification.
func init() {
	const label = "Dawnhart Mentor — target creature you control gets +3/+3 and gains trample until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "e12fb44a-524e-4603-a89c-fad8204fdb2b",
		Name:         "Dawnhart Mentor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Dawnhart Mentor — create a 1/1 white Human creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Human"), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:     "Coven — {5}{G}: Target creature you control gets +3/+3 and gains trample until end of turn. Activate only if you control three or more creatures with different powers.",
			Cost:      effects.ManaCost("{5}{G}"),
			Condition: effects.TrekB48Coven,
			Targets:   effects.TargetCreature("target creature you control", effects.YouControl()),
			Purpose:   effects.ForTargets(effects.BenefitUntilEOT(0, 3, 3, "trample")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				ts := ctx.LegalTargets()
				if len(ts) == 0 {
					return nil
				}
				if err := (effects.BoostUntilEOT{Target: ts[0].ID, Power: 3, Toughness: 3, Label: label}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Target: ts[0].ID, Keywords: []string{"trample"}, Label: label}.Apply(ctx)
			},
		}},
	})
}
