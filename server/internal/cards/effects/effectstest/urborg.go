package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// urborg_test.go is the end-to-end proof that Layer 4 is load-bearing
// rather than wire-only. Urborg, Tomb of Yawgmoth is the card #258
// wrote up and then refused to ship: its one static "would apply
// cleanly and do nothing," because everything downstream of the
// layer engine read the printed type line.
//
// The assertions deliberately go all the way to real mana in a real
// pool via the real activation path. An assertion on
// Effective().Subtypes alone would have passed before this branch
// and is exactly the check that let the gap survive four sprints.

const (
	UrborgOracle        = "db6174d7-211d-4817-b8e4-8384594c83f9"
	SunpetalGroveOracle = "402ec768-76fb-474e-ae74-babc90d833c4"
)

// seedLand puts a land on the battlefield through the layer-aware
// push so the listener stamps EnteredBattlefieldAt and invalidates
// the cached resolution.
func SeedLand(g *game.Game, owner uuid.UUID, name, typeLine, oracleID string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       name,
		TypeLine:   typeLine,
		OracleID:   oracleID,
		Owner:      owner,
		Controller: owner,
	})
}

// poolColors lists the colour letters currently in a player's mana
// pool, in pool order.
func PoolColors(p *game.Player) []string {
	out := make([]string, 0, p.ManaPool.Total())
	for _, tok := range p.ManaPool.Units() {
		out = append(out, tok.Color)
	}
	return out
}

// manaAbilityProduced returns the Produced string of each mana
// ability the engine currently offers on a battlefield card,
// through a snapshot so the layer engine has resolved.
func ManaAbilityProduced(t *testing.T, g *game.Game, id uuid.UUID) []string {
	t.Helper()
	var out []string
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID != id {
				continue
			}
			found = true
			for _, ab := range game.ManaAbilitiesForCard(c) {
				out = append(out, ab.Produced)
			}
			return
		}
	})
	if !found {
		t.Fatalf("card %s not on battlefield", id)
	}
	return out
}

// effectiveSubtypes returns Effective().Subtypes for a battlefield
// card, forcing a snapshot.
func EffectiveSubtypes(t *testing.T, g *game.Game, id uuid.UUID) []string {
	t.Helper()
	var out []string
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				out, found = c.Effective().Subtypes, true
				return
			}
		}
	})
	if !found {
		t.Fatalf("card %s not on battlefield", id)
	}
	return out
}
