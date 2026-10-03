package legal_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// emblem_targeting_test.go — ADR 0109 Delivery PR 6: the enumerator asks
// the engine's own gates, so a bot is never offered a cast an emblem
// refuses (§5, #1899) or a target a zone restriction refuses (§6,
// #1885), and every move it IS offered is one the engine accepts (ADR
// 0033 §1). Real catalog cards.

const (
	oracleNarset1899     = "e1de0c94-0ecc-425a-9b92-27c2745d07e7"
	oracleGroundSeal1885 = "13f6f960-ef79-4f2c-8874-90fe9e77099e"
	oracleTomik1885      = "9895a33f-9bbd-4440-8c1a-0d401431b77f"
	oracleRegrowth1885   = "e6e4a8bd-5c40-4654-8de1-0da9afed90fd"
)

// castTargetsOffered is every card target the cast moves for `source`
// name.
func castTargetsOffered(t *testing.T, moves []legal.Move, source uuid.UUID) map[uuid.UUID]bool {
	t.Helper()
	out := map[uuid.UUID]bool{}
	for _, m := range castMovesFor(moves, source) {
		var p struct {
			Targets []struct {
				ID uuid.UUID `json:"id"`
			} `json:"targets"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("params %s: %v", string(m.Params), err)
		}
		for _, tg := range p.Targets {
			out[tg.ID] = true
		}
	}
	return out
}

// An opponent's Narset emblem takes the active player's noncreature
// casts out of the enumeration and leaves the creature cast in it, and
// the engine refuses exactly what the enumerator dropped.
func TestNoNoncreatureCastIsOfferedUnderNarsetsEmblem(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(active)
	bolt := handCard(active, game.Card{Name: "Burst Lightning", TypeLine: "Instant", OracleID: oracleBurstLightning, ManaCost: "{R}"})
	bear := handCard(active, creature("Bear", "{G}", 2, 2))
	lands(g, active, "Mountain", "Mountain", 2)
	lands(g, active, "Forest", "Forest", 2)
	narset := battlefieldCard(g, opp, game.Card{Name: "Narset Transcendent", TypeLine: "Legendary Planeswalker — Narset", OracleID: oracleNarset1899})
	advanceTo(t, g, game.StepPrecombatMain)
	if len(castMovesFor(legal.EnumerateFor(g, active.ID), bolt)) == 0 {
		t.Fatal("setup: Burst Lightning is not offered before the emblem")
	}

	var err error
	g.WithWriteLock(func() { err = g.CreateEmblemForEffect(opp.ID, narset) })
	if err != nil {
		t.Fatalf("CreateEmblemForEffect: %v", err)
	}
	moves := legal.EnumerateFor(g, active.ID)
	if got := castMovesFor(moves, bolt); len(got) != 0 {
		t.Errorf("a noncreature cast was offered under Narset's emblem: %v", labels(got))
	}
	bearMoves := castMovesFor(moves, bear)
	if len(bearMoves) == 0 {
		t.Fatal("the creature cast was not offered under Narset's emblem")
	}
	dispatchAll(t, g, active.ID, bearMoves)
	if err := g.CastSpell(active.ID, bolt, game.CastSpellParams{}); !errors.Is(err, game.ErrCantCast) {
		t.Errorf("CastSpell of the instant under the emblem: %v, want ErrCantCast", err)
	}
}

// Under Ground Seal no graveyard card is a target, so Regrowth, whose
// only targets are graveyard cards, is not offered at all.
func TestNoGraveyardTargetIsOfferedUnderGroundSeal(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(active)
	regrowth := handCard(active, game.Card{Name: "Regrowth", TypeLine: "Sorcery", OracleID: oracleRegrowth1885, ManaCost: "{1}{G}"})
	dead := graveyardCard(active, creature("Dead Bear", "{1}{G}", 2, 2))
	lands(g, active, "Forest", "Forest", 2)
	advanceTo(t, g, game.StepPrecombatMain)
	if !castTargetsOffered(t, legal.EnumerateFor(g, active.ID), regrowth)[dead] {
		t.Fatal("setup: Regrowth is not offered at the graveyard card")
	}
	battlefieldCard(g, opp, game.Card{Name: "Ground Seal", TypeLine: "Enchantment", OracleID: oracleGroundSeal1885})
	if got := castMovesFor(legal.EnumerateFor(g, active.ID), regrowth); len(got) != 0 {
		t.Errorf("Regrowth was offered under Ground Seal: %v", labels(got))
	}
}

// Tomik refuses his opponents a land card in a graveyard and nothing
// else: the active player's Regrowth is offered the creature card and
// not the land card, and the offers it makes are accepted.
func TestTomikRefusesHisOpponentsGraveyardLandTargets(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(active)
	regrowth := handCard(active, game.Card{Name: "Regrowth", TypeLine: "Sorcery", OracleID: oracleRegrowth1885, ManaCost: "{1}{G}"})
	deadBear := graveyardCard(active, creature("Dead Bear", "{1}{G}", 2, 2))
	deadLand := graveyardCard(active, basic("Dead Forest", "Forest"))
	lands(g, active, "Forest", "Forest", 2)
	battlefieldCard(g, opp, game.Card{Name: "Tomik, Distinguished Advokist", TypeLine: "Legendary Creature — Human Advisor", OracleID: oracleTomik1885})
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	offered := castTargetsOffered(t, moves, regrowth)
	if !offered[deadBear] {
		t.Error("the creature card in the graveyard was not offered under Tomik")
	}
	if offered[deadLand] {
		t.Error("the land card in the graveyard was offered to Tomik's opponent")
	}
	dispatchAll(t, g, active.ID, castMovesFor(moves, regrowth))
}
