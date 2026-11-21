package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Garruk Wildspeaker — Legendary Planeswalker — Garruk {2}{G}{G}, starting
// loyalty 3 (Lorwyn, #5593):
//
//	"+1: Untap two target lands.
//	 −1: Create a 3/3 green Beast creature token.
//	 −4: Creatures you control get +3/+3 and gain trample until end of
//	     turn."
//
// The +1 is Argothian Elder's two-land untap: exactly two target lands,
// anyone's, and a land gone by resolution is skipped (CR 608.2b). The −4 is
// Overrun: the creatures it affects are the ones you control as it resolves
// (CR 611.2c), and both halves end in the cleanup step (CR 514.2). Each
// ability is activated only as a sorcery and once a turn (CR 606.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "186b4def-4fff-4a2b-bcbf-5318b0ac5fa9",
		Name:         "Garruk Wildspeaker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Untap two target lands.",
				Cost:    effects.LoyaltyCost(1),
				Targets: effects.TargetPermanent("two target lands", effects.Land()).WithCount(2, 2),
				Effect:  effects.UntapEachLegalTarget,
			},
			{
				Label:  "−1: Create a 3/3 green Beast creature token.",
				Cost:   effects.LoyaltyCost(-1),
				Effect: effects.Do(effects.CreateToken{Template: effects.TokenCard("3/3 green Beast"), N: 1}),
			},
			{
				Label: "−4: Creatures you control get +3/+3 and gain trample until end of turn.",
				Cost:  effects.LoyaltyCost(-4),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					match := effects.And(effects.Creature(), effects.YouControl())
					if err := (effects.BoostUntilEOT{Match: match, Power: 3, Toughness: 3, Label: "Garruk Wildspeaker — +3/+3"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Match: match, Keywords: []string{"trample"}, Label: "Garruk Wildspeaker — trample until end of turn"}.Apply(ctx)
				},
			},
		},
	})
}
