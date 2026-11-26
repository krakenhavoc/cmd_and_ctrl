package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Magma Phoenix — Creature — Phoenix {3}{R}{R}, 3/3 (Magic 2011, #5579):
//
//	"Flying
//	 When this creature dies, it deals 3 damage to each creature and each
//	 player.
//	 {3}{R}{R}: Return this card from your graveyard to your hand."
//
// The dies trigger deals the damage from the Phoenix's last-known information
// (CR 608.2h) to every creature on the battlefield and every player, as one
// instance (CR 120.1). The second ability works from the graveyard only
// (CR 113.6, 602.5): it returns the card to its owner's hand if it is still
// there as the ability resolves (Multani's shape).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9c6edb99-2625-4d80-a2e8-1b6039c0a960",
		Name:            "Magma Phoenix",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		// ADR 0126 §6: the dies trigger is a damage sweep of every creature.
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 3}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Magma Phoenix — 3 damage to each creature and each player", S82OldB476PhoenixBlast),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{R}{R}: Return this card from your graveyard to your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{3}{R}{R}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  effects.ReturnThisCardFromYourGraveyardToYourHand,
		}},
	})
}

// S82OldB476PhoenixBlast is "it deals 3 damage to each creature and each
// player".
func S82OldB476PhoenixBlast(g *game.Game, item *game.StackItem) error {
	return effects.DamageEachMatchingAndEachPlayerThen(effects.NewContext(g, item), effects.Creature(), 3, effects.RecipientThen{})
}
