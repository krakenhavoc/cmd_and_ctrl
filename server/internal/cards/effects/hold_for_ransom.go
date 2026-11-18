package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "0c9336c0-c0ec-4fb4-af7c-a7766265e927",
		Name:         "Hold for Ransom",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Grants: []AbilityGrant{{
			Key: holdForRansomGrant,
			Activated: []ActivatedAbility{{
				Label:        "{7}: Hold for Ransom's controller sacrifices it and draws a card. Activate only as a sorcery.",
				Cost:         ManaCost("{7}"),
				SorcerySpeed: true,
				Effect:       holdForRansomPayOff,
			}},
			Text: "{7}: Hold for Ransom's controller sacrifices it and draws a card. Activate only as a sorcery.",
		}},
		Static: []game.StaticAbility{
			RestrictAttached(game.CantAttackOrBlock),
			GrantAbilitiesToAttached(holdForRansomGrant),
		},
	})
}

// holdForRansomPayOff sacrifices the Aura that granted the ability and
// draws for its controller, who is read first (the live Aura, or its
// last-known information once gone) and carried to the tail. The draw does
// not depend on the sacrifice, so it runs whatever happened to it.
func holdForRansomPayOff(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	aura := ctx.Grantor()
	if aura == uuid.Nil {
		return nil
	}
	var controller uuid.UUID
	if c, ok := g.LookupCardForEffect(aura); ok && onBattlefield(g, aura) {
		controller = c.Controller
	} else if info, ok := g.LastKnownPermanentForEffect(aura); ok {
		controller = info.Controller
	}
	if controller == uuid.Nil {
		return nil
	}
	return SacrificePermanent{Target: aura, ThenTail: withCarry(ctx.ItemTail(holdForRansomDraw), controller)}.Apply(ctx)
}

// holdForRansomDraw is "and draws a card" for the Aura's controller, Carry[0].
// An on-disk key: never renamed, never reused.
var holdForRansomDraw = CardTail("then/hold-for-ransom-draw", func(ctx *Context, o game.TailOutcome) error {
	if p := tailCarry0(o); p != uuid.Nil {
		return DrawCards{Player: p, N: 1}.Apply(ctx)
	}
	return nil
})
