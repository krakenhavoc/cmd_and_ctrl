package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b37_helpers.go — shared pieces for the Assassin's Creed cards of
// S82 GAP batch 37 (#5948). Append-only.

// GrantedAltCostFreerunning is the key of "<spells> you cast have
// freerunning {cost}" (Ezio Auditore da Firenze; CR 702.173a): a granted
// offer that, like the printed one, is claimable only once an Assassin or a
// commander you control has dealt combat damage to a player this turn. An
// on-disk identity (it lands on StackItem.AltCost): never renamed, never
// reused.
const GrantedAltCostFreerunning = "granted-freerunning"

// FreerunningForSpellsYouCast is "<spells> you cast have freerunning
// {cost}" for the Assassin creature type. The gate is Freerunning's own,
// asked when the cast is announced and in the view.
func FreerunningForSpellsYouCast(cost, creatureType string) game.GrantedAlternativeCost {
	printed := Freerunning(cost, creatureType)
	return game.GrantedAlternativeCost{
		Offer: game.AlternativeCost{
			Key:       GrantedAltCostFreerunning,
			Label:     "Freerunning " + cost,
			ManaCost:  cost,
			Condition: printed.Condition,
		},
		Spells: game.PermissionFilter{CreatureType: creatureType},
	}
}

// gapB37PlayerPayload is the player a reflexive trigger was handed.
func GapB37PlayerPayload(ctx *Context) (uuid.UUID, bool) {
	for _, t := range ctx.Payload() {
		if t.Kind == game.TargetPlayer {
			return t.ID, true
		}
	}
	return uuid.Nil, false
}

// gapB37VehicleYouControlDealtCombatDamageToAPlayer is "whenever a Vehicle
// you control deals combat damage to a player" (Edward Kenway, Adéwalé):
// one match per Vehicle that connects. Whether the dealer was a Vehicle is
// read as it dealt the damage, so a Vehicle that was crewed still counts.
func GapB37VehicleYouControlDealtCombatDamageToAPlayer(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !CombatDamageToPlayerBy(ev, source.Controller, g) {
		return false
	}
	src, ok := g.LookupCardForEffect(ev.Source)
	return ok && HasSubtype("Vehicle")(g, source.Controller, src)
}
