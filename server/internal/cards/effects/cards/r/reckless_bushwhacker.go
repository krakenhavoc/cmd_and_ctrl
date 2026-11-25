package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reckless Bushwhacker — Creature — Goblin Warrior Ally {2}{R}, 2/1 (Oath of
// the Gatewatch, #5545):
//
//	"Surge {1}{R} (You may cast this spell for its surge cost if you or a
//	 teammate has cast another spell this turn.)
//	 Haste
//	 When this creature enters, if its surge cost was paid, other creatures
//	 you control get +1/+0 and gain haste until end of turn."
//
// The trigger is an intervening "if" over the permanent's cast provenance
// (CR 603.4, CR 400.7d), so a hard-cast Bushwhacker does nothing. The set of
// creatures is locked as the trigger resolves (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "4fb83127-00cf-4ce7-aedf-cb0b2740c66f",
		Name:             "Reckless Bushwhacker",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"haste"},
		AlternativeCosts: []game.AlternativeCost{effects.Surge("{1}{R}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AllOf(effects.Self, effects.CastWithSurge),
				"Reckless Bushwhacker — other creatures you control get +1/+0 and gain haste until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					match := effects.And(effects.Creature(), effects.YouControl(), effects.OtherThan(item.SourceCardID))
					if err := (effects.BoostUntilEOT{Match: match, Power: 1,
						Label: "Reckless Bushwhacker — other creatures you control get +1/+0"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Match: match, Keywords: []string{"haste"},
						Label: "Reckless Bushwhacker — other creatures you control gain haste"}.Apply(ctx)
				}),
		},
	})
}
