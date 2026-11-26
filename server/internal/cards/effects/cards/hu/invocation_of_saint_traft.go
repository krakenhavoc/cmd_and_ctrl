package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Invocation of Saint Traft — Enchantment — Aura {1}{W}{U} (Shadows over
// Innistrad, S82 OLD):
//
//	"Enchant creature
//	 Enchanted creature has "Whenever this creature attacks, create a 4/4
//	 white Angel creature token with flying that's tapped and attacking. Exile
//	 that token at end of combat.""
//
// Geist of Saint Traft's attack trigger as a granted bundle (ADR 0093), so the
// enchanted creature, not the Aura, is the source and its controller makes the
// Angel: the token enters tapped and attacking the player or planeswalker the
// creature attacks, and a delayed trigger at the end of combat exiles exactly
// the token made, by identity.
//
// No simplification.
func init() {
	const grantKey = "invocation-of-saint-traft/angel"
	effects.Register(effects.Spec{
		OracleID:     "cd6e9465-f4eb-482d-9c49-3a16b52c997c",
		Name:         "Invocation of Saint Traft",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit, Tokens: 1},
		Grants: []effects.AbilityGrant{{
			Key: grantKey,
			Triggered: []game.TriggeredAbility{
				effects.WheneverThisAttacks("Invocation of Saint Traft — create a 4/4 white Angel creature token with flying that's tapped and attacking; exile it at end of combat",
					func(g *game.Game, item *game.StackItem) error {
						return effects.CreateTokensAttackingYourChoice{
							Controller: item.Controller,
							Template:   effects.TokenCard("4/4 white Angel with flying"),
							N:          1,
							Tapped:     true,
							Prefer:     effects.TriggerAttackTarget(item),
							Label:      "Invocation of Saint Traft",
							Tail:       game.KeyedTail{Ref: invocationOfSaintTraftExileAngel},
						}.Apply(effects.NewContext(g, item))
					}),
			},
			Text: "Whenever this creature attacks, create a 4/4 white Angel creature token with flying that's tapped and attacking. Exile that token at end of combat.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(grantKey)},
	})
}

// invocationOfSaintTraftExileAngel schedules the end-of-combat exile of the
// token just made. An on-disk key: never renamed, never reused.
var invocationOfSaintTraftExileAngel = effects.AttackingTokensThen("attacking-tokens/invocation-of-saint-traft-exile-angel",
	func(ctx *effects.Context, created []uuid.UUID, _ game.KeyedTail) error {
		if len(created) == 0 {
			return nil
		}
		return effects.ScheduleDelayedTrigger{At: game.StepEndCombat, Controller: ctx.Item.Controller,
			Label: "Invocation of Saint Traft — exile the Angel", Cards: created, Body: effects.ExileListedCardsBody}.Apply(ctx)
	})
