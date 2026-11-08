package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Predator Ooze — Creature — Ooze {G}{G}{G}, 1/1 (Foundations, #4465):
//
//	"Indestructible
//	 Whenever this creature attacks, put a +1/+1 counter on it.
//	 Whenever a creature dealt damage by this creature this turn dies,
//	 put a +1/+1 counter on this creature."
//
// The attack trigger is the plain ThisAttacked shape. The death trigger is
// Zurgo Helmsmasher's: the turn's record of damaged objects
// (dealtDamageByThisTurn), so the Ooze and its victim are each a new
// object once they leave the battlefield and come back, and damage
// prevented to nothing was never dealt. The creature that died may be any
// player's. Each counter goes on only while the Ooze is still the object
// that triggered (CounterOnThis). Indestructible rides PrintedKeywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a51ec70d-1356-4403-a858-b99445bac54a",
		Name:            "Predator Ooze",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, ThisAttacked, "Predator Ooze — put a +1/+1 counter on it",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, died := diedCreature(ev, g)
				return died && ev.CardID != source.InstanceID && dealtDamageByThisTurn(g, source.InstanceID, ev.CardID)
			}, "Predator Ooze — a creature it damaged died: put a +1/+1 counter on it",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
