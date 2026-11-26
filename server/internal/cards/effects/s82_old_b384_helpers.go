package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b384_helpers.go — shared bodies for Commander 2013 batch 384
// (#5560).

// OldB384DefendingPlayerControlsTheMostCreatures is "as long as defending
// player controls the most creatures or is tied for the most" (Hooded
// Horror), read of the attacker as blockers are declared (a CantBeBlockedWhile
// condition): the player it attacks controls at least as many creatures as
// every other player still in the game.
func OldB384DefendingPlayerControlsTheMostCreatures(g *game.Game, _ uuid.UUID, attacker game.Card) bool {
	defender := DefendingPlayerOfAttacker(g, attacker)
	if defender == uuid.Nil {
		return false
	}
	counts := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		if p != nil && !p.Eliminated {
			counts[p.ID] = 0
		}
	}
	if _, ok := counts[defender]; !ok {
		return false
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if _, in := counts[c.Controller]; in && c.IsCreature() {
			counts[c.Controller]++
		}
	}
	for id, n := range counts {
		if id != defender && n > counts[defender] {
			return false
		}
	}
	return true
}

// OldB384ReturnOwnedNontokenPermanents is "each player gains control of all
// nontoken permanents they own" (Brooding Saurian): every nontoken permanent
// on the battlefield that its owner doesn't control goes back to its owner,
// with no end (CR 611.2a).
func OldB384ReturnOwnedNontokenPermanents(g *game.Game, item *game.StackItem) error {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsToken() || c.Owner == uuid.Nil || c.Owner == c.Controller {
			continue
		}
		g.GainControlForEffect(item.SourceCardID, c.InstanceID, c.Owner, game.IndefiniteDuration(),
			"Brooding Saurian — its owner gains control of it")
	}
	return nil
}

// OldB384ExchangePowerWithTarget is "exchange its power and the power of
// target creature" until end of combat (Serene Master): both are read now,
// and each gets the other's as its base power until the combat ends.
func OldB384ExchangePowerWithTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target := FirstLegalBattlefieldTarget(ctx)
	if target == uuid.Nil || !OnBattlefield(g, item.SourceCardID) || SourceIsNewObject(g, item) {
		return nil
	}
	g.RecomputeLayersIfStaleLocked()
	self, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok {
		return nil
	}
	other, ok := g.LookupCardForEffect(target)
	if !ok {
		return nil
	}
	d, ok := g.UntilEndOfCombatDuration()
	if !ok {
		return nil
	}
	mine, theirs := self.CurrentPower(), other.CurrentPower()
	for _, set := range []struct {
		id    uuid.UUID
		power int
	}{{item.SourceCardID, theirs}, {target, mine}} {
		if err := (ScopedEffectFor{
			Target:   set.id,
			Mods:     []game.Mod{game.SetBasePowerMod(set.power)},
			Duration: g.PinnedTo(d, set.id),
			Label:    "Serene Master — power exchanged until end of combat",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
