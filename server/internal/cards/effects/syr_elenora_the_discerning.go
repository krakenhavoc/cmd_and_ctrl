package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Syr Elenora, the Discerning — Legendary Creature — Human Knight
// {3}{U}{U}, */4:
//
//	"Syr Elenora's power is equal to the number of cards in your hand.
//	 When Syr Elenora enters, draw a card.
//	 Spells your opponents cast that target Syr Elenora cost {2} more to
//	 cast."
//
// The power is a characteristic-defining ability (CR 604.3, layer 7a).
// The tax is a cost modifier that reads the announced targets (CR
// 601.2c, 601.2f) and applies to spells only, as printed: abilities that
// target her pay nothing extra.
//
// No simplification.
func init() {
	tax := CostsMore(2, "Spells your opponents cast that target Syr Elenora cost {2} more to cast.",
		OpponentsSpell(), ThatTargetThisPermanent())
	tax.ReadsTargets = true
	Register(Spec{
		OracleID:     "8f6a3dcf-32e0-48c7-9265-0e298df4a5e4",
		Name:         "Syr Elenora, the Discerning",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			Labelled(powerIsCardsInYourHand(), "Syr Elenora's power is equal to the number of cards in your hand."),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Syr Elenora — draw a card", Do(DrawCards{N: 1})),
		},
		CostModifiers: []game.CostModifier{tax},
	})
}
