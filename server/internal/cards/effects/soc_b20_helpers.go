package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b20_helpers.go — shared readers for Secrets of Strixhaven batch 20
// (#4456).

// socB20TomekeeperLabel is the stack label (and effect key) of Biblioplex
// Tomekeeper's entry trigger.
const socB20TomekeeperLabel = "Biblioplex Tomekeeper — a creature becomes prepared or unprepared"

// socB20UntappedCreaturesYouControl is every untapped creature `player`
// controls, read through the layered view (Aziza, Mage Tower Captain).
func socB20UntappedCreaturesYouControl(g *game.Game, player uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && !c.Tapped && c.IsCreature() {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// socB20DealDamageWithExcess is b27DealDamageWithExcess with CR 120.4a's
// deathtouch clause: damage from a source with deathtouch needs only 1 to be
// lethal, so every point beyond the first is excess (the ruling on
// Archaic's Agony, "in some unusual cases ... may have deathtouch"). The
// source's deathtouch is read off the resolving spell the way the damage
// itself reads it (CR 702.2). It returns the excess that landed.
func socB20DealDamageWithExcess(ctx *Context, target uuid.UUID, amount int) (int, error) {
	c, ok := ctx.Game.LookupCardForEffect(target)
	if !ok || amount <= 0 {
		return 0, nil
	}
	lethal := max(c.CurrentToughness()-c.DamageMarked, 0)
	if src, found := ctx.Game.LookupCardForEffect(ctx.Source()); found && game.HasKeyword(&src, "deathtouch") && lethal > 1 {
		lethal = 1
	}
	cursor := b25LastEventSeq(ctx.Game)
	if err := (DealDamage{Source: ctx.Source(), Target: target, Amount: amount}).Apply(ctx); err != nil {
		return 0, err
	}
	return max(b27DamageDealtToAfter(ctx.Game, ctx.Source(), target, cursor)-lethal, 0), nil
}

// socB20SmallCreatureYouControlDied is "a creature you control with power
// or toughness 1 or less dies" (Arnyn, Deathbloom Botanist): a creature
// went to the graveyard from the battlefield under your control, and its
// power or toughness as it last existed there (CR 603.10a, the departure
// record, counters and effects included, not clamped) was 1 or less.
func socB20SmallCreatureYouControlDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	dead, ok := diedCreature(ev, g)
	if !ok || leftUnderControlOf(ev, dead) != source.Controller {
		return false
	}
	info, ok := g.LastKnownPermanentForEffect(ev.CardID)
	if !ok {
		return dead.CurrentPower() <= 1 || dead.CurrentToughness() <= 1
	}
	return info.Power <= 1 || info.Toughness <= 1
}

// socB20DamageToTheModeTarget is "<this card> deals n damage to target
// creature" as a mode's body (Artistic Process, Konstrari Charm): the
// mode's own target, if it was chosen, takes n from the resolving spell.
func socB20DamageToTheModeTarget(n int) func(*game.StackItem, *Context, int) error {
	return func(_ *game.StackItem, ctx *Context, occ int) error {
		t, ok := ModeTarget(ctx, occ)
		if !ok {
			return nil
		}
		return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: n}.Apply(ctx)
	}
}
