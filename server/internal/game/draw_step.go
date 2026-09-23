package game

import "github.com/google/uuid"

// draw_step.go is CR 504.1's turn-based draw, widened — the draw
// step's answer to untap.go's Seedborn Muse family (#1315).
//
// "You draw a card during each opponent's draw step" (Teferi, Who
// Slows the Sunset's emblem) is printed as a continuous modification
// of a turn-based action, not as an "at the beginning of" trigger.
// The two are observably different, and the engine already has an
// example of each: Howling Mine and Dictate of Kruphix ARE ordinary
// triggers (their printed text says "At the beginning of… that player
// draws an additional card"), so their draw goes on the stack, can be
// responded to (tap the Mine first), and lands strictly after the
// turn-based draw. A DrawStepPermission draw has none of that — it IS
// the turn-based action, resolved with no stack and no priority
// window between it and the active player's own CR 504.1 draw.
//
// See untap.go's header for the fuller argument against modelling
// this as a trigger; this file is the same shape one turn-based
// action over, deliberately kept as small as untap.go's own
// UntapStepPermission: one predicate pair, no state of its own.

// DrawStepPermission declares one source's contribution to the extra
// draws that happen as part of a player's draw step (CR 504.1) —
// Teferi, Who Slows the Sunset's emblem today.
//
// Declared on effects.Spec.DrawStep (a permanent) or the mirror field
// on effects.EmblemSpec (an emblem — CR 114.3 runs its abilities in
// the command zone exactly like a permanent's), and bridged into this
// package by CatalogDrawStepPermissions.
//
// Both predicates run under g.mu held in write mode and MUST NOT call
// public locking mutators — the same contract UntapStepPermission's
// carry.
type DrawStepPermission struct {
	// AppliesTo reports whether this source contributes an extra draw
	// during `activePlayer`'s draw step. Every printed card of this
	// family so far reads "each OPPONENT's draw step", which is
	// `activePlayer != source.Controller`: the active player already
	// gets their own CR 504.1 draw, so a permission that also fired
	// there would double it. (No teams in this engine, so "opponent"
	// and "each other player" coincide — see
	// eachOtherPlayersUntapStep's identical note.)
	//
	// Nil means "never contributes" and the permission is skipped —
	// declaring one without both funcs is a programming error.
	AppliesTo func(g *Game, source *Card, activePlayer uuid.UUID) bool

	// Drawer names who draws. Every printed card of this family is
	// "YOU draw", i.e. the source's controller — a func rather than a
	// bare read of source.Controller so a future "target player
	// draws" wording does not need a second field.
	Drawer func(g *Game, source *Card) uuid.UUID

	// N is how many cards this source adds. Zero reads as one, which
	// is every printed "a card" so far.
	N int

	// Label names the clause for logs and debugging. Not player-
	// facing — an extra draw from a DrawStepPermission never goes on
	// the stack, so there is no overlay to title.
	Label string
}

// CatalogDrawStepPermissions returns the draw-step permissions
// declared by the given catalog key (one entry per
// effects.Spec.DrawStep / EmblemSpec.DrawStep element), or nil for
// every card and emblem but a handful. Populated at boot by
// carddef.go's init, alongside the other catalog hooks.
//
// Nil hook ⇒ no catalog wired ⇒ the draw step draws exactly what
// CR 504.1 says and nothing else, matching CatalogUntapStepPermissions.
var CatalogDrawStepPermissions func(key string) []DrawStepPermission

// boundDrawPermission is one declared permission paired with the
// source that declared it — a battlefield card or an emblem. Pointers
// into g.Battlefield.Cards or a seat's p.Emblems.Cards, valid only
// for the duration of the write-locked step that built them.
type boundDrawPermission struct {
	permission DrawStepPermission
	source     *Card
}

// activeDrawStepPermissionsLocked gathers the draw-step permissions
// live right now for `activePlayer`'s draw step: one walk of the
// battlefield, plus one walk of every seat's emblem zone (CR 114.3),
// mirroring activeUntapStepPermissionsLocked exactly — including
// which catalog-key accessor each walk uses and why (see that
// function's comment).
//
// Caller must hold g.mu in write mode.
func (g *Game) activeDrawStepPermissionsLocked(activePlayer uuid.UUID) []boundDrawPermission {
	if CatalogDrawStepPermissions == nil {
		return nil
	}
	var out []boundDrawPermission
	if g.Battlefield != nil {
		for i := range g.Battlefield.Cards {
			src := &g.Battlefield.Cards[i]
			// CatalogAbilityKey: a permanent that has lost all its
			// abilities grants nothing, exactly as the untap-step
			// gather reads it.
			oracle := CatalogAbilityKey(*src)
			if oracle == "" {
				continue
			}
			for _, p := range CatalogDrawStepPermissions(oracle) {
				if p.AppliesTo == nil || p.Drawer == nil {
					continue
				}
				if !p.AppliesTo(g, src, activePlayer) {
					continue
				}
				out = append(out, boundDrawPermission{permission: p, source: src})
			}
		}
	}
	// CR 114.3: an emblem's abilities function in the command zone,
	// exactly like a battlefield permanent's — Teferi, Who Slows the
	// Sunset's emblem grants this with no permanent on the
	// battlefield at all.
	for _, p := range g.Seats {
		if p == nil || p.Emblems == nil {
			continue
		}
		for i := range p.Emblems.Cards {
			src := &p.Emblems.Cards[i]
			// CatalogKey, not CatalogAbilityKey: nothing in the game
			// can name an emblem to remove its abilities (emblem.go),
			// so there is no removal state to read.
			oracle := CatalogKey(*src)
			if oracle == "" {
				continue
			}
			for _, perm := range CatalogDrawStepPermissions(oracle) {
				if perm.AppliesTo == nil || perm.Drawer == nil {
					continue
				}
				if !perm.AppliesTo(g, src, activePlayer) {
					continue
				}
				out = append(out, boundDrawPermission{permission: perm, source: src})
			}
		}
	}
	return out
}
