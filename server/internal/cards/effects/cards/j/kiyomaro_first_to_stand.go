package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kiyomaro, First to Stand — Legendary Creature — Spirit {3}{W}{W}, */*
// (Saviors of Kamigawa, S82 OLD b583):
//
//	"Kiyomaro's power and toughness are each equal to the number of cards in
//	 your hand.
//	 As long as you have four or more cards in hand, Kiyomaro has vigilance.
//	 Whenever Kiyomaro deals damage, if you have seven or more cards in hand,
//	 you gain 7 life."
//
// Maro's characteristic-defining ability over the hand, and Hellbent's
// self-keyword shape held to four cards instead of none; both are read live.
// The trigger is Putrid Warrior's "deals damage" (combat or not, to anything)
// once for each batch of damage the creature deals at once (a creature it
// damages together with a player is one event of dealing damage), with an
// intervening "if" (CR 603.4): read as the damage is dealt and again as the
// ability resolves.
//
// No simplification.
func init() {
	const label = "Kiyomaro, First to Stand — you gain 7 life"
	effects.Register(effects.Spec{
		OracleID:     "340ec369-1761-4134-a96d-5537b4e08703",
		Name:         "Kiyomaro, First to Stand",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.S82OldB021CardsInHandPT("Kiyomaro's power and toughness are each equal to the number of cards in your hand."),
			effects.B583SelfKeywordWhileYou(effects.B583YouHaveCardsInHand(4), "vigilance", "As long as you have four or more cards in hand, Kiyomaro has vigilance."),
		},
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.InterveningIfYou("you have seven or more cards in hand", effects.B583YouHaveCardsInHand(7),
				effects.On(game.EventDealDamage, effects.ThisDealtDamage, label, effects.Do(effects.GainLife{Amount: 7})))),
		},
	})
}
