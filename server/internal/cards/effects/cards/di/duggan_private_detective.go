package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Duggan, Private Detective — Legendary Creature — Human Detective {2}{G}{U},
// */* (Doctor Who):
//
//	"Duggan's power and toughness are each equal to the number of cards in your
//	 hand.
//	 Whenever Duggan enters or attacks, investigate.
//	 The Most Important Punch in History — {1}{G}, {T}: Duggan deals damage equal
//	 to twice its power to another target creature. Activate only once."
//
// The power and toughness are layer-7a characteristic-defining statics over the
// controller's hand size (Body of Knowledge's pair). The investigate trigger is
// one ability with two conditions. The punch is the engine's once-only marker
// (CR 602.5b: it is spent when activated, even if it was countered), and deals
// twice the power as the ability resolves, read as it last was if Duggan has
// left.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0fbe3805-b59b-47fc-ab21-f05a4aa4cfe0",
		Name:         "Duggan, Private Detective",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Static: []game.StaticAbility{
			effects.PowerIsCardsInYourHand(),
			effects.Cmdr2B08ToughnessIsCardsInYourHand(),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrAttacks("Duggan — investigate", effects.Do(effects.Investigate{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:    "The Most Important Punch in History — {1}{G}, {T}: Duggan deals damage equal to twice its power to another target creature. Activate only once.",
			Purpose:  game.Purpose{Answers: game.AnswerRemove},
			Cost:     effects.Plus(effects.ManaCost("{1}{G}"), effects.TapCost()),
			OnlyOnce: true,
			Targets:  effects.Another(effects.TargetCreature("another target creature")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				power := effects.B43PowerNowOrLastKnown(g, item.SourceCardID)
				targets := ctx.LegalTargets()
				if power <= 0 || len(targets) == 0 {
					return nil
				}
				return effects.DealDamage{Source: item.SourceCardID, Target: targets[0].ID, Amount: 2 * power}.Apply(ctx)
			},
		}},
	})
}
