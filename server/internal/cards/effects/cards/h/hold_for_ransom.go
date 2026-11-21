package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const holdForRansomGrant = "hold-for-ransom/sacrifice-and-draw"

// Hold for Ransom — Enchantment — Aura {1}{W}:
//
//	"Enchant creature
//	 Enchanted creature can't attack or block and has '{7}: Hold for
//	 Ransom's controller sacrifices it and draws a card. Activate only as
//	 a sorcery.'"
//
// The creature has the ability (ADR 0093), so its controller pays {7},
// at sorcery speed. "It" is the Aura, the Grantor the activation names
// (StackItem.GrantedBy): its controller sacrifices it and draws. If the
// Aura has already gone the sacrifice does nothing, and its controller, read
// from last-known information (CR 608.2h), still draws as the ability
// resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0c9336c0-c0ec-4fb4-af7c-a7766265e927",
		Name:         "Hold for Ransom",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Grants: []effects.AbilityGrant{{
			Key: holdForRansomGrant,
			Activated: []effects.ActivatedAbility{{
				Label:        "{7}: Hold for Ransom's controller sacrifices it and draws a card. Activate only as a sorcery.",
				Cost:         effects.ManaCost("{7}"),
				SorcerySpeed: true,
				Effect:       holdForRansomPayOff,
			}},
			Text: "{7}: Hold for Ransom's controller sacrifices it and draws a card. Activate only as a sorcery.",
		}},
		Static: []game.StaticAbility{
			effects.RestrictAttached(game.CantAttackOrBlock),
			effects.GrantAbilitiesToAttached(holdForRansomGrant),
		},
	})
}

// holdForRansomPayOff sacrifices the Aura that granted the ability and
// draws for its controller, who is read first (the live Aura, or its
// last-known information once gone) and carried to the tail. The draw does
// not depend on the sacrifice, so it runs whatever happened to it.
func holdForRansomPayOff(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	aura := ctx.Grantor()
	if aura == uuid.Nil {
		return nil
	}
	var controller uuid.UUID
	if c, ok := g.LookupCardForEffect(aura); ok && effects.OnBattlefield(g, aura) {
		controller = c.Controller
	} else if info, ok := g.LastKnownPermanentForEffect(aura); ok {
		controller = info.Controller
	}
	if controller == uuid.Nil {
		return nil
	}
	return effects.SacrificePermanent{Target: aura, ThenTail: effects.WithCarry(ctx.ItemTail(holdForRansomDraw), controller)}.Apply(ctx)
}

// holdForRansomDraw is "and draws a card" for the Aura's controller, Carry[0].
// An on-disk key: never renamed, never reused.
var holdForRansomDraw = effects.CardTail("then/hold-for-ransom-draw", func(ctx *effects.Context, o game.TailOutcome) error {
	if p := effects.TailCarry0(o); p != uuid.Nil {
		return effects.DrawCards{Player: p, N: 1}.Apply(ctx)
	}
	return nil
})
