package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b148_helpers.go — shared bodies for S82 OLD batch 148 (#5516).

// OldB148AttachedCreatureDealtCombatDamageToACreature is "whenever equipped
// creature deals combat damage to a creature" (Giant's Skewer): a positive
// combat damage event whose source is the creature this Equipment is attached
// to and whose target is a creature still on the battlefield. A planeswalker
// or battle that took the damage does not match.
func OldB148AttachedCreatureDealtCombatDamageToACreature(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 || !source.IsAttachedTo(ev.Source) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Target)
	if !ok {
		return false
	}
	if z := g.FindCardZoneForEffect(ev.Target); z == nil || z.Kind != game.ZoneBattlefield {
		return false
	}
	return c.IsCreature()
}

// OldB148YouCantCastMoreThan is "you can cast no more than N spells each
// turn" on a permanent (Fires of Invention): only its controller is held to it,
// counted off the engine's per-turn cast tally, which is bumped after a cast
// succeeds, so a player who has cast N already is announcing the (N+1)th.
func OldB148YouCantCastMoreThan(n int, label string) game.CastRestriction {
	return game.CastRestriction{
		Label: label,
		Forbids: func(q game.CastQuery) bool {
			return q.Source.Controller == q.Controller && q.Game.CastTallyFor(q.Controller).Total >= n
		},
	}
}

// OldB148GrantedAltCostFreeUpToLandCount is the key of the free offer below.
const OldB148GrantedAltCostFreeUpToLandCount = "granted-free-up-to-land-count"

// OldB148CastFreeUpToLandCount is Fires of Invention's static: "You may cast
// spells with mana value less than or equal to the number of lands you control
// without paying their mana costs." Omnipresence's offer with the cap following
// the lands its caster controls right now (MaxManaValueFunc), reaching every
// zone a spell is cast from. An {X} in the cost counts as 0 (CR 107.3b).
func OldB148CastFreeUpToLandCount() game.GrantedAlternativeCost {
	return game.GrantedAlternativeCost{
		Offer: game.AlternativeCost{
			Key:   OldB148GrantedAltCostFreeUpToLandCount,
			Label: "Cast it without paying its mana cost",
		},
		MaxManaValueFunc: oldB148LandsControlledCount,
		MaxManaValueText: "the number of lands you control",
	}
}

// oldB148LandsControlledCount is the number of lands `controller` controls, off
// effective characteristics. Caller must hold g.mu.
func oldB148LandsControlledCount(g *game.Game, controller uuid.UUID) int {
	if g == nil || g.Battlefield == nil || controller == uuid.Nil {
		return 0
	}
	n := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == controller && c.IsLand() {
			n++
		}
	}
	return n
}

// OldB148YouControlAnOkoPlaneswalker is "activate only if you control an Oko
// planeswalker" as an activation condition (Bramblefort Fink): an Oko is a
// planeswalker whose subtype is Oko, read through the layers.
func OldB148YouControlAnOkoPlaneswalker(g *game.Game, controller, _ uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsPlaneswalker() && c.HasSubtype("Oko") {
			return true
		}
	}
	return false
}
