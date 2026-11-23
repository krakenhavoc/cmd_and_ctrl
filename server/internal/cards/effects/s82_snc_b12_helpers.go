package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_snc_b12_helpers.go — Streets of New Capenna's "mana values among cards
// in your graveyard" family (S82, batch 12): the count of distinct mana
// values, its "five or more" threshold, and the self statics, condition and
// cost amounts the batch's cards read it through.

// sncMVThreshold is the printed "five or more mana values".
const sncMVThreshold = 5

// sncGraveyardMVs is the number of different mana values among the cards in
// `player`'s graveyard (CR 202.3): a land card or a card with no mana cost
// counts as 0. `skip` leaves out one card, the spell being cast, which a
// self cost reduction must not count (it is still in its source zone while
// the price is read). uuid.Nil skips nothing.
func SncGraveyardMVs(g *game.Game, player, skip uuid.UUID) int {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Graveyard == nil {
		return 0
	}
	seen := map[int]bool{}
	for _, c := range p.Graveyard.Cards {
		if c.InstanceID == skip {
			continue
		}
		seen[c.ManaValue()] = true
	}
	return len(seen)
}

// sncFiveMVs is "there are five or more mana values among cards in your
// graveyard".
func SncFiveMVs(g *game.Game, player uuid.UUID) bool {
	return SncGraveyardMVs(g, player, uuid.Nil) >= sncMVThreshold
}

// sncSelfPT is "As long as there are five or more mana values among cards in
// your graveyard, this creature gets +P/+T": a layer 7c self-modifier read
// live on every recompute (CR 613.4c).
func SncSelfPT(power, toughness int) game.StaticAbility {
	return game.StaticAbility{
		Label:    "Gets " + SignedPT(power, toughness) + " as long as there are five or more mana values among cards in your graveyard.",
		Layer:    game.Layer7PT,
		SubLayer: game.SubLayer7C_Modify,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			return target.InstanceID == source.InstanceID && SncFiveMVs(g, source.Controller)
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Power += power
			c.Toughness += toughness
		},
	}
}

// sncSelfKeywords is "As long as there are five or more mana values among
// cards in your graveyard, this creature has <keywords>": a layer 6 self-grant.
func SncSelfKeywords(keywords ...string) game.StaticAbility {
	s := B16GrantKeywords(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID && SncFiveMVs(g, source.Controller)
	}, keywords...)
	s.Label = "Has " + joinWords(keywords) + " as long as there are five or more mana values among cards in your graveyard."
	return s
}

// CondSNCFiveManaValues is the five-or-more threshold as an intervening-if
// Condition, read for the item's controller as it resolves. It declares no
// footprint (the graveyard is read, which other triggers can change).
type CondSNCFiveManaValues struct{}

func (CondSNCFiveManaValues) Holds(ctx *Context) bool {
	return SncFiveMVs(ctx.Game, ctx.Controller())
}

func (CondSNCFiveManaValues) Condition() (game.FootprintStep, bool) {
	return game.FootprintStep{}, false
}

// sncFiveMVsWhen is the same threshold as a trigger When, read as the
// trigger event fires (CR 603.4's first check).
func SncFiveMVsWhen(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return SncFiveMVs(g, source.Controller)
}

// sncCostLessPerDifferentMV is "This spell costs {1} less to cast for each
// different mana value among cards in your graveyard": the caster's graveyard,
// the spell itself left out.
func SncCostLessPerDifferentMV(q game.CostQuery) int {
	if q.Game == nil {
		return 0
	}
	return SncGraveyardMVs(q.Game, q.Controller, q.Card.InstanceID)
}

// sncCostLessThreeWithFive is "This ability costs {3} less to activate if
// there are five or more mana values among cards in your graveyard".
func SncCostLessThreeWithFive(q game.CostQuery) int {
	if q.Game == nil || !SncFiveMVs(q.Game, q.Controller) {
		return 0
	}
	return 3
}
