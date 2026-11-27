package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Descendant of Kiyomaro — Creature — Human Soldier {1}{W}{W}, 2/3 (Saviors of
// Kamigawa, S82 OLD b583):
//
//	"As long as you have more cards in hand than each opponent, this creature
//	 gets +1/+2 and has "Whenever this creature deals combat damage, you gain
//	 3 life.""
//
// A layer 7c pump and a layer 6 grant of the card's own trigger bundle
// (Brimstone Mage's self grant), both held to the same comparison of hand
// sizes and read live (Death of a Thousand Stings' test). The granted trigger
// is Quintessential Katana's: it is the creature's own ability, once for each
// batch of combat damage it deals.
//
// No simplification.
func init() {
	const grant = "descendant-of-kiyomaro/combat-damage-gain-3"
	effects.Register(effects.Spec{
		OracleID:     "bbf07bec-9c02-455b-885d-7cb5fc77b5de",
		Name:         "Descendant of Kiyomaro",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Triggered: []game.TriggeredAbility{
				effects.OncePerBatch(effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.Combat && ev.Amount > 0 && ev.Source == source.InstanceID
				}, "Descendant of Kiyomaro — you gain 3 life", effects.Do(effects.GainLife{Amount: 3}))),
			},
			Text: "Whenever this creature deals combat damage, you gain 3 life.",
		}},
		Static: []game.StaticAbility{
			effects.B583SelfPumpWhileYou(effects.B583YouHaveMoreCardsInHandThanEachOpponent, 1, 2, "As long as you have more cards in hand than each opponent, this creature gets +1/+2."),
			effects.B583SelfGrantWhileYou(effects.B583YouHaveMoreCardsInHandThanEachOpponent, grant, `As long as you have more cards in hand than each opponent, this creature has "Whenever this creature deals combat damage, you gain 3 life."`),
		},
	})
}
