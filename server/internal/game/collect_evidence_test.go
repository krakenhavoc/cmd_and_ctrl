package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// collect_evidence_test.go — #5193: collect evidence N (CR 701.59) as an
// optional additional cost to cast, the engine half. The catalog is
// stubbed, so the plumbing is tested without a card: the gate
// (ExileCost.FloorMet), the walk, the validator, the payer and the record
// on the paid cost. The card files' tests are cards/effects/mkm_collect_
// evidence_cards_test.go.

const testCollectEvidenceOracle = "test-collect-evidence"

// withCollectEvidence stubs a spell whose only optional cost is "you may
// collect evidence n".
func withCollectEvidence(t *testing.T, n int) {
	t.Helper()
	stubOptionalCosts(t, testCollectEvidenceOracle, []AdditionalCost{{
		Optional:   true,
		Key:        CollectEvidenceKey,
		ExileCards: CollectEvidenceCost(n, "cards with total mana value N or greater"),
		Label:      "Collect evidence",
	}})
}

// evidenceInGraveyard puts one card per mana cost into p's graveyard and
// returns their IDs in the same order.
func evidenceInGraveyard(p *Player, costs ...string) []uuid.UUID {
	ids := make([]uuid.UUID, len(costs))
	for i, cost := range costs {
		c := NewCard("Evidence", p.ID)
		c.TypeLine = "Sorcery"
		c.ManaCost = cost
		ids[i] = c.InstanceID
		p.Graveyard.PushTop(c)
	}
	return ids
}

// castCollectingEvidence puts the stub spell in hand and casts it, taking
// the collect evidence offer with `ids` when `collect` is set.
func castCollectingEvidence(t *testing.T, g *Game, me *Player, collect bool, ids []uuid.UUID) (uuid.UUID, error) {
	t.Helper()
	advanceTo(t, g, StepPrecombatMain)
	c := NewCard("Test Detective", me.ID)
	c.TypeLine = "Sorcery"
	c.ManaCost = "{1}"
	c.OracleID = testCollectEvidenceOracle
	me.Hand.PushTop(c)
	params := CastSpellParams{ExileIDs: ids}
	if collect {
		params.OptionalCosts = []int{0}
	}
	return c.InstanceID, g.CastSpell(me.ID, c.InstanceID, params)
}

// CR 701.59a: cards whose mana values add up to exactly N pay, and so do
// cards that add up to more. They are exiled with the spell on the stack
// and recorded on the paid cost beside the announced optional cost.
func TestCollectEvidencePaysWithExactlyNAndWithMore(t *testing.T) {
	for _, tc := range []struct {
		name  string
		costs []string
		pick  []int
	}{
		{"exactly 4", []string{"{2}{U}", "{1}", "{5}"}, []int{0, 1}},
		{"more than 4", []string{"{2}{U}", "{1}", "{5}"}, []int{2}},
		{"more than 4 from several", []string{"{2}{U}", "{1}", "{5}"}, []int{0, 1, 2}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			withCollectEvidence(t, 4)
			g := newActiveGame(t)
			me := g.Seats[0]
			gy := evidenceInGraveyard(me, tc.costs...)
			var named []uuid.UUID
			for _, i := range tc.pick {
				named = append(named, gy[i])
			}
			spell, err := castCollectingEvidence(t, g, me, true, named)
			if err != nil {
				t.Fatalf("cast: %v", err)
			}
			for _, id := range named {
				if !g.Exile.Contains(id) || me.Graveyard.Contains(id) {
					t.Errorf("evidence %s was not exiled", id)
				}
			}
			if got := me.Graveyard.Size(); got != len(tc.costs)-len(named) {
				t.Errorf("graveyard holds %d, want the %d cards not named", got, len(tc.costs)-len(named))
			}
			paid := g.StackMeta[spell].Paid
			if len(paid.Exiled) != len(named) {
				t.Fatalf("Paid.Exiled = %v, want the named cards", paid.Exiled)
			}
			if spellCard, ok := g.LookupCardForEffect(spell); !ok || OptionalCostTimesPaid(spellCard, paid.OptionalCosts, CollectEvidenceKey) != 1 {
				t.Errorf("Paid.OptionalCosts = %v, want the collect evidence offer recorded", paid.OptionalCosts)
			}
		})
	}
}

// CR 701.59b: a graveyard whose whole total is below N can't collect
// evidence, and a pick that falls short is refused even when the graveyard
// could pay. Nothing moves either way.
func TestCollectEvidenceIsRefusedBelowN(t *testing.T) {
	t.Run("the whole graveyard is short", func(t *testing.T) {
		withCollectEvidence(t, 6)
		g := newActiveGame(t)
		me := g.Seats[0]
		gy := evidenceInGraveyard(me, "{2}", "{3}")
		var payable bool
		g.ReadSnapshot(func() {
			payable = g.ExileCostPayableForEffect(me.ID, uuid.Nil, CollectEvidenceCost(6, ""))
		})
		if payable {
			t.Error("ExileCostPayableForEffect: a graveyard totalling 5 pays collect evidence 6")
		}
		spell, err := castCollectingEvidence(t, g, me, true, gy)
		if !errors.Is(err, ErrInvalidParam) {
			t.Fatalf("cast = %v, want ErrInvalidParam", err)
		}
		if g.Stack.Contains(spell) || me.Graveyard.Size() != 2 || g.Exile.Size() != 0 {
			t.Error("a refused collect evidence moved something")
		}
	})
	t.Run("the pick is short", func(t *testing.T) {
		withCollectEvidence(t, 6)
		g := newActiveGame(t)
		me := g.Seats[0]
		gy := evidenceInGraveyard(me, "{2}", "{3}", "{4}")
		for name, ids := range map[string][]uuid.UUID{
			"five of six": gy[:2],
			"nothing":     nil,
		} {
			spell, err := castCollectingEvidence(t, g, me, true, ids)
			if !errors.Is(err, ErrInvalidParam) {
				t.Fatalf("%s: cast = %v, want ErrInvalidParam", name, err)
			}
			if g.Stack.Contains(spell) || me.Graveyard.Size() != 3 {
				t.Errorf("%s: a refused collect evidence moved something", name)
			}
		}
	})
	t.Run("cards named without the offer", func(t *testing.T) {
		withCollectEvidence(t, 2)
		g := newActiveGame(t)
		me := g.Seats[0]
		gy := evidenceInGraveyard(me, "{2}")
		if _, err := castCollectingEvidence(t, g, me, false, gy); !errors.Is(err, ErrInvalidParam) {
			t.Fatalf("cast = %v, want ErrInvalidParam for exile_ids with no cost", err)
		}
	})
}

// Declining the offer is always a legal cast, whatever the graveyard holds,
// and it records nothing.
func TestCollectEvidenceDeclined(t *testing.T) {
	withCollectEvidence(t, 6)
	g := newActiveGame(t)
	me := g.Seats[0]
	evidenceInGraveyard(me, "{6}")
	spell, err := castCollectingEvidence(t, g, me, false, nil)
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	paid := g.StackMeta[spell].Paid
	if len(paid.Exiled) != 0 || len(paid.OptionalCosts) != 0 || me.Graveyard.Size() != 1 {
		t.Errorf("a declined offer paid something: %+v", paid)
	}
}

// A card's mana value in the graveyard is its printed one: a modal
// double-faced card has only its front face there (CR 712.8a) and an
// adventurer card only its normal characteristics (CR 715.4), whatever the
// other face costs.
func TestCollectEvidenceReadsTheFrontFaceInTheGraveyard(t *testing.T) {
	owner := uuid.New()
	mdfc := cheapBackMDFCFixture(owner)           // front {5}{B}{B}, back {1}{B}
	adventure := bigCreatureCheapAdventure(owner) // {6}{G}, Adventure {2}{G}
	for name, c := range map[string]Card{"modal DFC": mdfc, "adventurer": adventure} {
		if got := TotalManaValue([]Card{c}); got != 7 {
			t.Errorf("%s: TotalManaValue = %d, want 7 (the front face)", name, got)
		}
		if !CollectEvidenceMet([]Card{c}, 7) {
			t.Errorf("%s: does not collect evidence 7", name)
		}
		if CollectEvidenceMet([]Card{c}, 8) {
			t.Errorf("%s: collects evidence 8", name)
		}
	}

	withCollectEvidence(t, 7)
	g := newActiveGame(t)
	me := g.Seats[0]
	c := cheapBackMDFCFixture(me.ID)
	me.Graveyard.PushTop(c)
	if _, err := castCollectingEvidence(t, g, me, true, []uuid.UUID{c.InstanceID}); err != nil {
		t.Fatalf("collect evidence 7 with a {5}{B}{B} // {1}{B} card: %v", err)
	}
}

// The effect form's walk offers nothing when the graveyard can't reach N,
// and the whole graveyard when it can.
func TestCollectEvidenceOptionsForEffect(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	gy := evidenceInGraveyard(me, "{1}", "{2}")
	g.ReadSnapshot(func() {
		if got := g.CollectEvidenceOptionsForEffect(me.ID, uuid.Nil, 4); got != nil {
			t.Errorf("options at 4 over a total of 3 = %v, want none", got)
		}
		if got := g.CollectEvidenceOptionsForEffect(me.ID, uuid.Nil, 3); len(got) != len(gy) {
			t.Errorf("options at 3 = %v, want the whole graveyard", got)
		}
	})
}
