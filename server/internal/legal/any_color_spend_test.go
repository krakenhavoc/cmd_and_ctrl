package legal_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// any_color_spend_test.go — #1600, the #544 half of a player's "you may
// spend mana as though it were mana of any color" (Chromatic Orrery,
// Mycosynth Lattice, Oath of Nissa). The enumerator's affordability
// probe reads the grant through game.CostAsPaidByForEffect, the function
// every engine payment reads it with, so a move only the grant pays for
// is offered, every offered move is accepted (each test dispatches what
// it is offered), and a move the grant no longer pays for is not.

const (
	oracleChromaticOrrery  = "95c3976c-33f3-490b-bfd3-7f1af2fe0416"
	oracleMycosynthLattice = "ae1f2ab5-c6a5-4d49-a746-3cb4668bf805"
	oracleOathOfNissa      = "c4efdbab-711d-4269-9b24-b05d36f7e5c7"
	oraclePestilence       = "dafe63ef-f3d6-45e7-877a-573da92ba85e"
	oracleCascadeBluffs    = "f1603384-4361-49c9-98aa-7785fc3504c4"
)

// spendTable is a main phase for the active seat with an empty hand and
// two Mountains — no blue source anywhere.
func spendTable(t *testing.T) (*game.Game, *game.Player) {
	t.Helper()
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	battlefieldCard(g, seat, basic("Mountain", "Mountain"))
	battlefieldCard(g, seat, basic("Mountain", "Mountain"))
	return g, seat
}

func grantPermanent(g *game.Game, p *game.Player, name, typeLine, oracle string) uuid.UUID {
	return battlefieldCard(g, p, game.Card{Name: name, TypeLine: typeLine, OracleID: oracle})
}

func blueSorcery(p *game.Player, typeLine string) uuid.UUID {
	return handCard(p, game.Card{Name: "Blue Spell", TypeLine: typeLine, ManaCost: "{U}{U}", Layout: "normal"})
}

// A {U}{U} spell off two Mountains is offered under the Orrery and
// dispatches; the same board without it, or with an opponent's, is not.
func TestAnyColorSpendCastIsOfferedOnlyUnderTheGrant(t *testing.T) {
	t.Run("own Orrery", func(t *testing.T) {
		g, seat := spendTable(t)
		grantPermanent(g, seat, "Chromatic Orrery", "Legendary Artifact", oracleChromaticOrrery)
		card := blueSorcery(seat, "Sorcery")
		moves := legal.EnumerateFor(g, seat.ID)
		casts := castMovesFor(moves, card)
		if len(casts) != 1 {
			t.Fatalf("offered %d casts of {U}{U} off two Mountains under the Orrery, want 1", len(casts))
		}
		dispatchAll(t, g, seat.ID, casts)
		dispatchOne(t, g, seat.ID, casts[0])
		if !g.Stack.Contains(card) {
			t.Error("the spell is not on the stack")
		}
	})
	t.Run("no Orrery", func(t *testing.T) {
		g, seat := spendTable(t)
		card := blueSorcery(seat, "Sorcery")
		if casts := castMovesFor(legal.EnumerateFor(g, seat.ID), card); len(casts) != 0 {
			t.Errorf("offered %v with no grant", labels(casts))
		}
	})
	t.Run("an opponent's Orrery", func(t *testing.T) {
		g, seat := spendTable(t)
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		grantPermanent(g, opp, "Chromatic Orrery", "Legendary Artifact", oracleChromaticOrrery)
		card := blueSorcery(seat, "Sorcery")
		if casts := castMovesFor(legal.EnumerateFor(g, seat.ID), card); len(casts) != 0 {
			t.Errorf("offered %v under an opponent's Orrery", labels(casts))
		}
	})
	t.Run("an opponent's Lattice", func(t *testing.T) {
		g, seat := spendTable(t)
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		grantPermanent(g, opp, "Mycosynth Lattice", "Artifact", oracleMycosynthLattice)
		card := blueSorcery(seat, "Sorcery")
		casts := castMovesFor(legal.EnumerateFor(g, seat.ID), card)
		if len(casts) != 1 {
			t.Fatalf("offered %d casts under an opponent's Lattice, want 1", len(casts))
		}
		dispatchAll(t, g, seat.ID, casts)
	})
}

// The grant is read live: once the Orrery leaves, the move goes with it.
func TestAnyColorSpendMoveGoesWhenTheOrreryLeaves(t *testing.T) {
	g, seat := spendTable(t)
	orrery := grantPermanent(g, seat, "Chromatic Orrery", "Legendary Artifact", oracleChromaticOrrery)
	card := blueSorcery(seat, "Sorcery")
	if casts := castMovesFor(legal.EnumerateFor(g, seat.ID), card); len(casts) != 1 {
		t.Fatalf("offered %d casts with the Orrery, want 1", len(casts))
	}
	g.WithWriteLock(func() {
		if _, err := g.Battlefield.Remove(orrery); err != nil {
			t.Fatalf("remove the Orrery: %v", err)
		}
	})
	if casts := castMovesFor(legal.EnumerateFor(g, seat.ID), card); len(casts) != 0 {
		t.Errorf("still offered %v after the Orrery left", labels(casts))
	}
}

// Oath of Nissa's grant is narrowed to planeswalker spells: the walker
// is offered off two Mountains, the sorcery is not.
func TestAnyColorSpendOathOfNissaOffersOnlyPlaneswalkers(t *testing.T) {
	g, seat := spendTable(t)
	grantPermanent(g, seat, "Oath of Nissa", "Legendary Enchantment", oracleOathOfNissa)
	walker := blueSorcery(seat, "Legendary Planeswalker — Jace")
	sorcery := blueSorcery(seat, "Sorcery")
	moves := legal.EnumerateFor(g, seat.ID)
	if casts := castMovesFor(moves, sorcery); len(casts) != 0 {
		t.Errorf("offered %v — the Oath does not cover a sorcery", labels(casts))
	}
	casts := castMovesFor(moves, walker)
	if len(casts) != 1 {
		t.Fatalf("offered %d casts of the planeswalker, want 1", len(casts))
	}
	dispatchAll(t, g, seat.ID, casts)
}

// An activated ability's coloured cost: Pestilence's {B} off a Mountain.
func TestAnyColorSpendActivationIsOfferedUnderTheGrant(t *testing.T) {
	for _, withOrrery := range []bool{true, false} {
		g, seat := spendTable(t)
		if withOrrery {
			grantPermanent(g, seat, "Chromatic Orrery", "Legendary Artifact", oracleChromaticOrrery)
		}
		pest := grantPermanent(g, seat, "Pestilence", "Enchantment", oraclePestilence)
		acts := movesFrom(legal.EnumerateFor(g, seat.ID), pest, legal.KindActivate)
		if withOrrery != (len(acts) > 0) {
			t.Fatalf("Orrery %v: offered %v", withOrrery, labels(acts))
		}
		dispatchAll(t, g, seat.ID, acts)
	}
}

// A mana ability's own coloured cost, paid from floating mana: Cascade
// Bluffs' {U/R} with a colorless in the pool.
func TestAnyColorSpendManaAbilityCostIsOfferedUnderTheGrant(t *testing.T) {
	for _, withOrrery := range []bool{true, false} {
		g, seat := spendTable(t)
		if withOrrery {
			grantPermanent(g, seat, "Chromatic Orrery", "Legendary Artifact", oracleChromaticOrrery)
		}
		bluffs := grantPermanent(g, seat, "Cascade Bluffs", "Land", oracleCascadeBluffs)
		// #2215: the move tops its cost up from untapped sources, and
		// a Mountain pays {U/R} on its own. Tap them, so the floating
		// {C} is the only mana and the grant is what decides.
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].Name == "Mountain" {
					g.Battlefield.Cards[i].Tapped = true
				}
			}
		})
		seat.ManaPool.AddMana(game.ManaToken{Color: "C"})
		var filter []legal.Move
		for _, m := range movesFrom(legal.EnumerateFor(g, seat.ID), bluffs, legal.KindMana) {
			if strings.Contains(m.Label, "{U/R}") {
				filter = append(filter, m)
			}
		}
		if withOrrery != (len(filter) > 0) {
			t.Fatalf("Orrery %v: filter moves %v", withOrrery, labels(filter))
		}
		dispatchAll(t, g, seat.ID, filter)
	}
}

// "Unless that player pays {U}": the pay answer is offered to an
// Orrery's controller with only red, and the engine takes the payment.
func TestAnyColorSpendPayUnlessIsOfferedUnderTheGrant(t *testing.T) {
	for _, withOrrery := range []bool{true, false} {
		g := newTable(t)
		payer := g.Seats[1]
		if withOrrery {
			grantPermanent(g, payer, "Chromatic Orrery", "Legendary Artifact", oracleChromaticOrrery)
		}
		payer.ManaPool.AddMana(game.ManaToken{Color: "R"})
		g.WithWriteLock(func() {
			if err := g.QueuePayUnlessForEffect(payer.ID, uuid.New(), "{U}",
				"Pay {U}?", func(*game.Game) error { return nil }); err != nil {
				t.Fatalf("QueuePayUnlessForEffect: %v", err)
			}
		})
		var choices []legal.Move
		for _, m := range legal.EnumerateFor(g, payer.ID) {
			if m.Kind == legal.KindChoice {
				choices = append(choices, m)
			}
		}
		want := 1 // decline only
		if withOrrery {
			want = 2
		}
		if len(choices) != want {
			t.Fatalf("Orrery %v: %d answers %v, want %d", withOrrery, len(choices), labels(choices), want)
		}
		dispatchAll(t, g, payer.ID, choices)
	}
}
