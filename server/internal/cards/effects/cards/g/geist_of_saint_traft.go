package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Geist of Saint Traft — Legendary Creature — Spirit Cleric {1}{W}{U}, 2/2:
//
//	"Hexproof
//	 Whenever Geist of Saint Traft attacks, create a 4/4 white Angel
//	 creature token with flying that's tapped and attacking. Exile that token
//	 at end of combat."
//
// Kari Zev's shape: the token enters tapped and attacking the player (or
// planeswalker) the Geist attacks when the attacker choice is not the
// controller's to make, and a delayed trigger at the end of combat exiles
// exactly the token made, by identity.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4304035c-c3e4-4a19-aa1f-92a83d8aed1f",
		Name:            "Geist of Saint Traft",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"hexproof"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Geist of Saint Traft — create a 4/4 white Angel creature token with flying that's tapped and attacking; exile it at end of combat",
				func(g *game.Game, item *game.StackItem) error {
					return effects.CreateTokensAttackingYourChoice{
						Controller: item.Controller,
						Template:   effects.TokenCard("4/4 white Angel with flying"),
						N:          1,
						Tapped:     true,
						Prefer:     effects.TriggerAttackTarget(item),
						Label:      "Geist of Saint Traft",
						Tail:       game.KeyedTail{Ref: geistOfSaintTraftExileAngel},
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

var geistOfSaintTraftExileAngel = effects.AttackingTokensThen("attacking-tokens/geist-of-saint-traft-exile-angel",
	func(ctx *effects.Context, created []uuid.UUID, _ game.KeyedTail) error {
		if len(created) == 0 {
			return nil
		}
		return effects.ScheduleDelayedTrigger{At: game.StepEndCombat, Controller: ctx.Item.Controller,
			Label: "Geist of Saint Traft — exile the Angel", Cards: created, Body: effects.ExileListedCardsBody}.Apply(ctx)
	})
