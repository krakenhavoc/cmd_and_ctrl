package protocol

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const viewFlierWardOracle = "test-view-flier-ward"

// filteredWardTable is a table where seat 1 controls a "Test Ward" that
// writes a players-only "can't attack you" restriction filtered by `filter`
// onto every creature, and the active seat has a flying 1/1 Hawk and a
// ground Bear of power bearPower, both ready to attack, at declare
// attackers.
func filteredWardTable(t *testing.T, filter game.PermanentQuery, bearPower int) (g *game.Game, active, warded *game.Player, flier, bear uuid.UUID) {
	t.Helper()
	prev := game.CatalogStaticAbilities
	t.Cleanup(func() { game.CatalogStaticAbilities = prev })
	game.CatalogStaticAbilities = func(key string) []game.StaticAbility {
		if key != viewFlierWardOracle {
			if prev == nil {
				return nil
			}
			return prev(key)
		}
		return []game.StaticAbility{{
			Layer:     game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool { return target.IsCreature() },
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
				c.AttackTargetRestrictions = append(c.AttackTargetRestrictions, game.AttackTargetRestriction{
					Source: source.InstanceID, SourceName: source.Name, NotPlayer: source.Controller, NotPlayerOnly: true,
					OnlyAttackersMatching: []game.PermanentQuery{filter},
				})
			},
		}}
	}
	g = busyTable(t, 1)
	active = g.Seats[g.Turn.ActiveSeat]
	warded = g.Seats[(g.Turn.ActiveSeat+1)%4]
	ward := uuid.New()
	flier, bear = uuid.New(), uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: ward, Name: "Test Ward", TypeLine: "Enchantment",
		OracleID: viewFlierWardOracle, Owner: warded.ID, Controller: warded.ID})
	g.Battlefield.PushTop(game.Card{InstanceID: flier, Name: "Hawk", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Owner: active.ID, Controller: active.ID, Keywords: []string{"flying"}})
	g.Battlefield.PushTop(game.Card{InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear",
		Power: bearPower, Toughness: 2, Owner: active.ID, Controller: active.ID})
	knowTheTable(g)
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{ward, flier, bear} {
			g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: id, OldZone: game.ZoneHand, NewZone: game.ZoneBattlefield})
		}
		for i := range g.Battlefield.Cards {
			g.Battlefield.Cards[i].SummonedThisTurn = false
		}
	})
	advanceTo(t, g, game.StepDeclareAttackers)

	return g, active, warded, flier, bear
}

// #4601: a restriction with an attacker filter ("creatures with flying
// can't attack you") is a chip only on the creatures it binds right now,
// it names the player alone (no planeswalkers), and the digest leaves that
// player out of the flier's attack_targets and not the ground creature's.
func TestAttackerFilteredRestrictionIsAChipOnlyOnMatchingCreatures(t *testing.T) {
	g, active, warded, flier, bear := filteredWardTable(t, game.PermanentQuery{Types: []string{"creature"}, Keyword: "flying"}, 2)
	v := ViewOfGameFor(g, active.ID.String())
	fv, bv := cardInFrame(v, flier.String()), cardInFrame(v, bear.String())
	if fv == nil || bv == nil {
		t.Fatal("the creatures are not in the frame")
	}
	want := []AttackTargetRestrictionView{{
		Player: warded.ID.String(), Source: "Test Ward", Reason: "it's a creature with flying",
	}}
	if !slices.Equal(fv.AttackTargetRestrictions, want) {
		t.Errorf("flier chip = %+v, want %+v", fv.AttackTargetRestrictions, want)
	}
	if len(bv.AttackTargetRestrictions) != 0 {
		t.Errorf("a creature without flying carries a chip: %+v", bv.AttackTargetRestrictions)
	}
	if v.LegalActions == nil || v.LegalActions.Sources[flier.String()] == nil || v.LegalActions.Sources[bear.String()] == nil {
		t.Fatal("the creatures have no digest entries")
	}
	if slices.Contains(v.LegalActions.Sources[flier.String()].AttackTargets, warded.ID.String()) {
		t.Error("the flier's digest still offers the protected player")
	}
	if !slices.Contains(v.LegalActions.Sources[bear.String()].AttackTargets, warded.ID.String()) {
		t.Error("the ground creature's digest lost the protected player")
	}
}

// #4659: a power bound is read live like flying is: a power 2 or less
// filter chips the Hawk (power 1) with "it's a creature with power 2 or
// less" and leaves a power 3 Bear alone.
func TestPowerBoundedRestrictionIsAChipOnlyOnCreaturesWithinTheBound(t *testing.T) {
	g, active, warded, flier, bear := filteredWardTable(t,
		game.PermanentQuery{Types: []string{"creature"}}.WithPowerAtMost(2), 3)
	v := ViewOfGameFor(g, active.ID.String())
	fv, bv := cardInFrame(v, flier.String()), cardInFrame(v, bear.String())
	if fv == nil || bv == nil {
		t.Fatal("the creatures are not in the frame")
	}
	want := []AttackTargetRestrictionView{{
		Player: warded.ID.String(), Source: "Test Ward", Reason: "it's a creature with power 2 or less",
	}}
	if !slices.Equal(fv.AttackTargetRestrictions, want) {
		t.Errorf("small creature chip = %+v, want %+v", fv.AttackTargetRestrictions, want)
	}
	if len(bv.AttackTargetRestrictions) != 0 {
		t.Errorf("a creature outside the bound carries a chip: %+v", bv.AttackTargetRestrictions)
	}
	if slices.Contains(v.LegalActions.Sources[flier.String()].AttackTargets, warded.ID.String()) {
		t.Error("the small creature's digest still offers the protected player")
	}
	if !slices.Contains(v.LegalActions.Sources[bear.String()].AttackTargets, warded.ID.String()) {
		t.Error("the big creature's digest lost the protected player")
	}
}
