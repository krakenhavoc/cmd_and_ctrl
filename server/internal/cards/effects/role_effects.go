package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// role_effects.go — the effect bodies and predicates the Role cards
// share (#1945). Append-only, per docs/adding-cards.md's shared-
// vocabulary rule.

// createRoleOnFirstTarget is "create a <Role> token attached to
// target creature" as a triggered or activated ability's whole body,
// read off the announce-time target. A target that left makes the
// instruction do nothing (CR 608.2b), and so does an "up to one" clause
// that chose no target.
func createRoleOnFirstTarget(kind RoleKind) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := b16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		return CreateRoleToken{Role: kind, Host: id}.Apply(ctx)
	}
}

// createRoleOnClauseTarget is "create a <Role> token attached to
// <the creature chosen for clause slot>" as the last sentence of a
// spell with several target clauses (Eriette's Whisper, Shatter the
// Oath). Nothing happens when the slot chose nothing ("up to one") or
// its target is no longer legal (CR 608.2b).
func createRoleOnClauseTarget(ctx *Context, slot int, kind RoleKind) error {
	t, ok := ctx.ClauseTarget(slot)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	return CreateRoleToken{Role: kind, Host: t.ID}.Apply(ctx)
}

// createRoleOnThis is "create a <Role> token attached to it" /
// "attached to this creature". CreateRoleToken creates nothing for a
// host that is not a creature on the battlefield.
func createRoleOnThis(kind RoleKind) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return CreateRoleToken{Role: kind, Host: item.SourceCardID}.Apply(NewContext(g, item))
	}
}

// isEnchantedByAura reports whether an Aura is attached to the
// permanent: "enchanted creature" (Ellivere, Syr Armont). Walks the
// battlefield slice directly, so a layer-7 static can ask it for every
// candidate without a per-call copy of the whole battlefield
// (b15IsEnchanted copies it; PumpAttachedPer's note says why not).
func isEnchantedByAura(g *game.Game, id uuid.UUID) bool {
	if g == nil || g.Battlefield == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		a := &g.Battlefield.Cards[i]
		if a.IsAttachedTo(id) && a.IsAura() {
			return true
		}
	}
	return false
}
