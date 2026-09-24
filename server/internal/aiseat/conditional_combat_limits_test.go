package aiseat_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// conditional_combat_limits_test.go — the bot half of #1534: Mirri,
// Weatherlight Duelist's per-opponent block limit and The Eternal
// Wanderer's per-planeswalker attack limit must not wedge a bot table
// (ADR 0045 amendment of 2026-09-24, Decision 46). Same drive as
// combat_limits_test.go — one combat, move by move, through the
// enumerator and actions.Dispatch — and deliberately not behind
// AISEAT_GAME_TESTS.

const (
	botMirriOracle    = "60bc9cd2-2e03-4889-b1ef-fdab9a9a6d08"
	botWandererOracle = "20a1671d-e8a4-4cf1-87a7-f2f6319f4b9e"
)

// TestBotsFinishACombatUnderMirri — four seats, Mirri attacking beside
// three Ogres: her trigger resolves, every opponent blocks with at most
// one creature, and the combat ends.
func TestBotsFinishACombatUnderMirri(t *testing.T) {
	for name, pol := range map[string]aiseat.Policy{"heuristic": heuristic.New(), "aggressive": aggressive()} {
		t.Run(name, func(t *testing.T) {
			g := limitTable(t, 4)
			active := g.Seats[g.Turn.ActiveSeat]
			limitPush(g, active, "Mirri, Weatherlight Duelist", "Legendary Creature — Cat Warrior", botMirriOracle, 3, 2)
			tally := driveOneCombat(t, g, pol)
			for seat, n := range tally.blockersBy {
				if n > 1 {
					t.Errorf("seat %s blocked with %d creatures under Mirri", seat, n)
				}
			}
			if name == "aggressive" && tally.blockers == 0 {
				t.Errorf("the aggressive policy should still have blocked with one creature somewhere: %+v", tally)
			}
		})
	}
}

// TestBotsFinishACombatUnderTheEternalWanderer — four seats, the
// Wanderer on the next seat: at most one creature attacks her, the
// aggressive policies still attack with all three Ogres, and the
// combat ends. "wanderer-first" is the aggressive policy aimed at her,
// the one that walks straight into the limit: exactly one Ogre goes at
// the Wanderer and the other two go elsewhere.
func TestBotsFinishACombatUnderTheEternalWanderer(t *testing.T) {
	for name, pol := range map[string]aiseat.Policy{
		"heuristic":      heuristic.New(),
		"aggressive":     aggressive(),
		"wanderer-first": &scripted{prefer: []string{"Attack The Eternal Wanderer", "Attack", "Block"}},
	} {
		t.Run(name, func(t *testing.T) {
			g := limitTable(t, 4)
			owner := g.Seats[(g.Turn.ActiveSeat+1)%4]
			wanderer := game.Card{
				InstanceID: uuid.New(),
				Name:       "The Eternal Wanderer", TypeLine: "Legendary Planeswalker", OracleID: botWandererOracle,
				Owner: owner.ID, Controller: owner.ID,
				Counters: map[string]int{game.CounterLoyalty: 5},
			}
			wanderer.AddKnowersAll(seatIDs(g))
			g.Battlefield.PushTop(wanderer)
			tally := driveOneCombat(t, g, pol)
			if n := tally.perDefender[wanderer.InstanceID]; n > 1 {
				t.Errorf("%d creatures attacked The Eternal Wanderer", n)
			}
			if name != "heuristic" && tally.attackers != 3 {
				t.Errorf("the %s policy should still attack with all three: %+v", name, tally)
			}
			if name == "wanderer-first" && tally.perDefender[wanderer.InstanceID] != 1 {
				t.Errorf("the wanderer-first policy should have sent exactly one Ogre at her: %+v", tally)
			}
		})
	}
}
