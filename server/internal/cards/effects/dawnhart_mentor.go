package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "e12fb44a-524e-4603-a89c-fad8204fdb2b",
		Name:         "Dawnhart Mentor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Dawnhart Mentor — create a 1/1 white Human creature token",
				Do(CreateToken{Template: TokenCard("1/1 white Human"), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:     "Coven — {5}{G}: Target creature you control gets +3/+3 and gains trample until end of turn. Activate only if you control three or more creatures with different powers.",
			Cost:      ManaCost("{5}{G}"),
			Condition: trekB48Coven,
			Targets:   TargetCreature("target creature you control", YouControl()),
			Purpose:   withAnswers(ForTargets(BenefitUntilEOT(0, 3, 3, "trample")), game.AnswerPump|game.AnswerCombatGrant),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				ts := ctx.LegalTargets()
				if len(ts) == 0 {
					return nil
				}
				if err := (BoostUntilEOT{Target: ts[0].ID, Power: 3, Toughness: 3, Label: label}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{Target: ts[0].ID, Keywords: []string{"trample"}, Label: label}.Apply(ctx)
			},
		}},
	})
}
