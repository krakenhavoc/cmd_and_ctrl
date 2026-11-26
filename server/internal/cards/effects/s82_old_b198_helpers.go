package effects

import (
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b198_helpers.go — shared vocabulary of Core Set 2019 batch 198
// (S82 older sets, #5527): Suncleanser's counter clauses and Fraying
// Omnipotence's three halvings.

// S82OldB198PlayerLosesAllCounters is "<player> loses all counters": every
// kind a player has (poison, energy, experience, rad, tickets) comes off.
func S82OldB198PlayerLosesAllCounters(ctx *Context, id uuid.UUID) error {
	p := ctx.PlayerByID(id)
	if p == nil {
		return nil
	}
	kinds := make([]string, 0, len(p.Counters))
	for kind, n := range p.Counters {
		if n > 0 {
			kinds = append(kinds, kind)
		}
	}
	sort.Strings(kinds)
	for _, kind := range kinds {
		if n := p.Counters[kind]; n > 0 {
			if err := ctx.Game.AddPlayerCounterForEffect(id, kind, -n); err != nil {
				return err
			}
		}
	}
	return nil
}

// S82OldB198CantHaveCountersWhileSourceRemains is "It can't have counters put
// on it for as long as <this creature> remains on the battlefield"
// (Suncleanser): a layer-6 restriction record on the one permanent, as it is
// now, that ends when the source leaves (CR 611.2b) and never begins if the
// source already has.
func S82OldB198CantHaveCountersWhileSourceRemains(ctx *Context, id uuid.UUID, label string) error {
	d, ok := DurationWhileSourceRemains(ctx, ctx.Source())
	if !ok {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(id)
	if !ok || !OnBattlefield(ctx.Game, id) {
		return nil
	}
	ctx.Game.RegisterScopedEffectForEffect(ctx.Source(),
		[]game.AffectedObject{game.PinObject(id, c.EnteredBattlefieldAt)},
		[]game.Mod{{Kind: game.ModAddRestrictions, ExtraRestrictions: game.ObjectCantHaveCounters}},
		d, label)
	return nil
}

// S82OldB198EachPlayerLosesHalfTheirLifeRoundedUp is "each player loses half
// their life, rounded up": the caster first, then each opponent in seat order.
// A player at zero or less loses nothing.
func S82OldB198EachPlayerLosesHalfTheirLifeRoundedUp(ctx *Context) error {
	for _, id := range TablePlayers(ctx) {
		p := ctx.PlayerByID(id)
		if p == nil || p.Eliminated || p.Life <= 0 {
			continue
		}
		if half := (p.Life + 1) / 2; half > 0 {
			if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), id, -half); err != nil {
				return err
			}
		}
	}
	return nil
}
