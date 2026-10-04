package aiseat_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// annihilator_pick_test.go — the bot half of #2073 (ADR 0113 §2). An
// annihilator trigger asks the defending player to choose N permanents
// they control in one own_permanents prompt. The enumerator answers it
// like Lotus Field's: whole sets of N, the seat's cheapest fuel first
// (ADR 0098), capped rather than every combination. And the real
// heuristic must answer it, or a bot defender wedges the table in the
// declare attackers step.

// queueAnnihilator puts the engine's annihilator prompt up for
// `defender`, as an annihilator `n` trigger from `attacker` resolving
// would.
func queueAnnihilator(t *testing.T, g *game.Game, attacker, defender uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		var src *game.Card
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == attacker {
				src = &g.Battlefield.Cards[i]
			}
		}
		if src == nil {
			t.Fatal("setup: the attacker is not on the battlefield")
		}
		item := game.NewTriggeredItem(src, "Annihilator 2 — defending player sacrifices 2 permanents")
		item.Params.Player = defender
		if err := g.AnnihilatorSacrificeForEffect(item, n); err != nil {
			t.Fatalf("AnnihilatorSacrificeForEffect: %v", err)
		}
	})
	if !owesChoice(g, defender) {
		t.Fatal("setup: the defender was not asked anything")
	}
}

// annihilatorBoard gives `defender` two 1/1 tokens and six bigger
// creatures, and `attacker` an Eldrazi.
func annihilatorBoard(g *game.Game, attacker, defender uuid.UUID) (eldrazi uuid.UUID, tokens, creatures []uuid.UUID) {
	g.WithWriteLock(func() {
		push := func(owner uuid.UUID, name, typeLine string, p int) uuid.UUID {
			id := uuid.New()
			g.Battlefield.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine,
				ManaCost: "{4}", Power: p, Toughness: p, Owner: owner, Controller: owner})
			return id
		}
		eldrazi = push(attacker, "Eldrazi", "Creature — Eldrazi", 5)
		for i := 0; i < 2; i++ {
			tokens = append(tokens, push(defender, "Eldrazi Spawn", "Token Creature — Eldrazi Spawn", 1))
		}
		for i := 0; i < 6; i++ {
			creatures = append(creatures, push(defender, "Big Beast", "Creature — Beast", 6))
		}
	})
	return eldrazi, tokens, creatures
}

func TestEnumeratorOffersAnnihilatorSetsCheapestFirst(t *testing.T) {
	g := newSettledTable(t, 73)
	attacker, defender := g.Seats[0].ID, g.Seats[1].ID
	eldrazi, tokens, _ := annihilatorBoard(g, attacker, defender)
	queueAnnihilator(t, g, eldrazi, defender, 2)

	isToken := map[uuid.UUID]bool{tokens[0]: true, tokens[1]: true}
	moves := legal.EnumerateForWithOptions(g, defender, legal.Options{
		OrderCostFuel: func(c legal.TargetCandidate) float64 {
			if isToken[c.ID] {
				return 0
			}
			return 6
		},
	})
	var sets [][]string
	for _, m := range moves {
		var p struct {
			CardIDs []string `json:"card_ids"`
		}
		if err := json.Unmarshal(m.Params, &p); err == nil && len(p.CardIDs) > 0 {
			sets = append(sets, p.CardIDs)
		}
	}
	if len(sets) == 0 {
		t.Fatal("the defender was offered no answer")
	}
	for _, s := range sets {
		if len(s) != 2 {
			t.Errorf("an answer names %d permanents, want exactly 2: %v", len(s), s)
		}
	}
	// C(8,2) = 28 sets exist; the enumerator offers whole sets under its
	// cap, not every combination.
	if len(sets) >= 28 {
		t.Errorf("%d sets offered — every combination, not a capped list", len(sets))
	}
	first := map[string]bool{sets[0][0]: true, sets[0][1]: true}
	if !first[tokens[0].String()] || !first[tokens[1].String()] {
		t.Errorf("the first set offered is %v, want the two tokens (cheapest fuel first)", sets[0])
	}
}

func TestHeuristicAnswersTheAnnihilatorSacrifice(t *testing.T) {
	g := newSettledTable(t, 74)
	attacker, defender := g.Seats[0].ID, g.Seats[1].ID
	eldrazi, tokens, creatures := annihilatorBoard(g, attacker, defender)
	queueAnnihilator(t, g, eldrazi, defender, 2)

	taken := driveChoices(t, g, heuristic.New(), defender, 10)
	if len(taken) != 1 {
		t.Fatalf("the heuristic answered %d times, want one choice: %v", len(taken), taken)
	}
	left := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == defender {
				left++
			}
		}
	})
	if want := len(tokens) + len(creatures) - 2; left != want {
		t.Errorf("the defender has %d permanents left, want %d", left, want)
	}
	t.Logf("the heuristic chose: %v", taken)
}
