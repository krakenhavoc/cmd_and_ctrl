package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// teamwork_blight_test.go — the enumerator half of #1703: a card with
// teamwork or an optional blight is two casts, and the paid one carries
// a payment the engine accepts — or is not offered at all.

const (
	oracleRepulsorBlast = "150f920c-c942-4feb-824f-79dd9b531687"
	oracleCinderStrike  = "54421e69-d79e-4c2e-8ce6-96994d168835"
)

type teamworkBlightWire struct {
	OptionalCosts []int    `json:"optional_costs"`
	TeamworkIDs   []string `json:"teamwork_ids"`
	BlightIDs     []string `json:"blight_ids"`
}

func teamworkBlightParams(t *testing.T, m legal.Move) teamworkBlightWire {
	t.Helper()
	var p teamworkBlightWire
	if err := json.Unmarshal(m.Params, &p); err != nil {
		t.Fatalf("decode %q: %v", m.Label, err)
	}
	return p
}

// TestTeamworkIsOfferedWithAPaymentOnlyWhenThePowerIsThere: Repulsor
// Blast with a 1/1 and a 1/1 on the board is offered paid (teamwork 2,
// both tapped) and unpaid; with only one 1/1 the paid cast is never
// offered. Every offered move dispatches.
func TestTeamworkIsOfferedWithAPaymentOnlyWhenThePowerIsThere(t *testing.T) {
	for _, tc := range []struct {
		name     string
		elves    int
		wantPaid bool
	}{
		{"enough power", 2, true},
		{"too little power", 1, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newTable(t)
			active := g.Seats[g.Turn.ActiveSeat]
			opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			clearHand(active)
			blast := handCard(active, game.Card{
				Name: "Repulsor Blast", TypeLine: "Sorcery", OracleID: oracleRepulsorBlast, ManaCost: "{3}{R}",
			})
			lands(g, active, "Mountain", "Mountain", 4)
			for i := 0; i < tc.elves; i++ {
				battlefieldCard(g, active, creature("Elf", "{G}", 1, 1))
			}
			battlefieldCard(g, opp, creature("Ogre", "{2}{R}", 3, 3))
			advanceTo(t, g, game.StepPrecombatMain)

			moves := castMovesFor(legal.EnumerateFor(g, active.ID), blast)
			paid, unpaid := 0, 0
			for _, m := range moves {
				p := teamworkBlightParams(t, m)
				if len(p.OptionalCosts) == 0 {
					if len(p.TeamworkIDs) != 0 {
						t.Errorf("%q names teamwork creatures without announcing teamwork", m.Label)
					}
					unpaid++
					continue
				}
				if len(p.TeamworkIDs) != 2 {
					t.Errorf("%q pays teamwork 2 with %d creatures, want both 1/1s", m.Label, len(p.TeamworkIDs))
				}
				paid++
			}
			if unpaid == 0 {
				t.Errorf("the unpaid cast was not offered; labels: %v", labels(moves))
			}
			if (paid > 0) != tc.wantPaid {
				t.Errorf("paid casts offered: %d, want any = %v; labels: %v", paid, tc.wantPaid, labels(moves))
			}
			dispatchAll(t, g, active.ID, moves)
		})
	}
}

// TestBlightIsPaidOntoASurvivorFirst: with a 1/1 and a 3/3 the bot's
// blight 1 goes on the 3/3, which lives through it.
func TestBlightIsPaidOntoASurvivorFirst(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(active)
	strike := handCard(active, game.Card{
		Name: "Cinder Strike", TypeLine: "Sorcery", OracleID: oracleCinderStrike, ManaCost: "{R}",
	})
	lands(g, active, "Mountain", "Mountain", 1)
	battlefieldCard(g, active, creature("Goblin", "{R}", 1, 1))
	ogre := battlefieldCard(g, active, creature("Ogre", "{2}{R}", 3, 3))
	battlefieldCard(g, opp, creature("Wall", "{1}", 0, 10))
	advanceTo(t, g, game.StepPrecombatMain)

	moves := castMovesFor(legal.EnumerateFor(g, active.ID), strike)
	paid := 0
	for _, m := range moves {
		p := teamworkBlightParams(t, m)
		if len(p.OptionalCosts) == 0 {
			if len(p.BlightIDs) != 0 {
				t.Errorf("%q names a blighted creature without announcing the blight", m.Label)
			}
			continue
		}
		paid++
		if len(p.BlightIDs) != 1 || p.BlightIDs[0] != ogre.String() {
			t.Errorf("%q blights %v, want the Ogre (it survives a -1/-1 counter)", m.Label, p.BlightIDs)
		}
	}
	if paid == 0 {
		t.Errorf("the blighted cast was not offered; labels: %v", labels(moves))
	}
	dispatchAll(t, g, active.ID, moves)
}

// TestTeamworkCreatureIsNotAlsoTappedForMana: the fourth mana comes
// from a Llanowar Elves, and the only team that reaches power 2 needs
// the Elves too. One creature cannot pay both (CR 118.3, the auto-tap
// exclusion), so the paid cast is unaffordable and must not be offered.
func TestTeamworkCreatureIsNotAlsoTappedForMana(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(active)
	strike := handCard(active, game.Card{
		Name: "Repulsor Blast", TypeLine: "Sorcery", OracleID: oracleRepulsorBlast, ManaCost: "{3}{R}",
	})
	// Three Mountains and an Elves: four mana only with the Elves. The
	// Elves (1 power) plus a 1/1 would pay teamwork 2 — but then the
	// cast is one mana short, so only the unpaid cast is offered.
	lands(g, active, "Mountain", "Mountain", 3)
	elves := creature("Llanowar Elves", "{G}", 1, 1)
	elves.OracleID = oracleLlanowarElves
	battlefieldCard(g, active, elves)
	battlefieldCard(g, active, creature("Goblin", "{R}", 1, 1))
	battlefieldCard(g, opp, creature("Ogre", "{2}{R}", 3, 3))
	advanceTo(t, g, game.StepPrecombatMain)

	moves := castMovesFor(legal.EnumerateFor(g, active.ID), strike)
	for _, m := range moves {
		if p := teamworkBlightParams(t, m); len(p.OptionalCosts) > 0 {
			t.Errorf("%q pays teamwork with the Elves the {3}{R} needs", m.Label)
		}
	}
	if len(moves) == 0 {
		t.Errorf("the unpaid cast (Elves for mana) was not offered")
	}
	dispatchAll(t, g, active.ID, moves)
}
