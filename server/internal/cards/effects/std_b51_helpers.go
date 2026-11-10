package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b51_helpers.go — the seams batch 51 of the Standard-set leftovers
// (#4452) needs.

// ExploresTimes has the permanent `Explorer` explore N times in a row under
// the effect's controller (CR 701.44a): each explore finishes before the
// next reveals, so a nonland card left on top is revealed again. N of 0 or
// less explores not at all (Jadelight Spelunker cast for X=0).
type ExploresTimes struct {
	Explorer uuid.UUID
	N        int
	// Then is a registered tail run once after the last explore, with the
	// explorer as Carry[0] ("explores. Then you may ...").
	Then game.TailRef
}

// Apply reads the explorer's object identity now (CR 400.7), as Explores does.
func (e ExploresTimes) Apply(ctx *Context) error {
	ref := game.ObjectRef{ID: e.Explorer}
	if c, ok := ctx.Game.LookupCardForEffect(e.Explorer); ok {
		ref.Epoch = c.ObjectEpoch
	}
	return ctx.Game.ExploreTimesThenForEffect(ctx.Source(), ref, ctx.Controller(), e.N, e.Then)
}

// CantBeBlockedThisTurnByPowerAtLeast registers "<target creature> can't be
// blocked by creatures with power N or greater this turn" (Squeak By, CR 509.1b):
// CantBeBlockedThisTurnByPower's other direction, a cantBeBlockedByPowerAtLeast
// record pinned to Target at resolution (CR 611.2c) and swept at the turn's end.
type CantBeBlockedThisTurnByPowerAtLeast struct {
	Target  uuid.UUID
	AtLeast int
	Label   string
}

// Apply registers the record.
func (r CantBeBlockedThisTurnByPowerAtLeast) Apply(ctx *Context) error {
	text := "creatures with power " + strconv.Itoa(r.AtLeast) + " or greater"
	return ScopedEffectFor{
		Target:   r.Target,
		Mods:     []game.Mod{game.CantBeBlockedByPowerAtLeastMod(r.AtLeast, text)},
		Duration: DurationUntilEndOfTurn(ctx),
		Label:    eotLabel(r.Label, "can't be blocked this turn by "+text),
	}.Apply(ctx)
}

// permanentTypesInGraveyard counts the distinct permanent types (CR 110.4:
// artifact, battle, creature, enchantment, land, planeswalker) among the
// cards in `player`'s graveyard: "four or more permanent types among cards
// in your graveyard" (Matzalantli, the Great Door). A card with two of them
// counts for both. Printed characteristics, as every graveyard read is.
func permanentTypesInGraveyard(g *game.Game, player uuid.UUID) int {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Graveyard == nil {
		return 0
	}
	var artifact, battle, creature, enchantment, land, walker bool
	for _, c := range p.Graveyard.Cards {
		artifact = artifact || c.IsArtifact()
		battle = battle || c.IsBattle()
		creature = creature || c.IsCreature()
		enchantment = enchantment || c.IsEnchantment()
		land = land || c.IsLand()
		walker = walker || c.IsPlaneswalker()
	}
	n := 0
	for _, has := range []bool{artifact, battle, creature, enchantment, land, walker} {
		if has {
			n++
		}
	}
	return n
}
