package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// when_paid_test.go — #1716: a target clause widened by an announced
// optional cost (WhenPaid). One test per proof card, each proving the
// printed clause on a plain cast and the wider one on a paid cast.

const (
	cruelAllianceOracle     = "6897f9e0-f654-4c0a-9fda-2ad4e264bf9a"
	bloodchiefsThirstOracle = "4236851b-5366-43a1-bde4-f525b4fbcbce"
	tearAsunderOracle       = "610af0f7-b5e3-43fb-9d02-7c59bd99034c"
	expelTheUnworthyOracle  = "4af2e62f-150e-4fd0-98b0-c6e72f5f9a51"
	bloodBeckoningOracle    = "67a48e3f-2388-42a9-a8b1-97c08761f807"
	divineResilienceOracle  = "4f6e2e47-34df-4bf3-a546-e06b42840167"
)

// wpCreature seeds a creature with a printed mana cost on the
// battlefield, so its mana value is what the clause reads.
func wpCreature(g *game.Game, owner uuid.UUID, name, manaCost string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Test", ManaCost: manaCost,
		Power: 3, Toughness: 3, Owner: owner, Controller: owner,
	})
}

func TestWhenPaidSetsTheCostsClauseAndRefusesNil(t *testing.T) {
	clause := TargetCreature("target creature")
	got := WhenPaid(Teamwork(2), clause)
	if got.Targets != clause || got.Teamwork != 2 || got.Key != game.TeamworkKey || !got.Optional {
		t.Errorf("WhenPaid(Teamwork(2), …) = %+v, want the teamwork cost carrying the clause", got)
	}
	mustPanic(t, "nil clause", func() { WhenPaid(Kicker("{1}"), nil) })
}

// TestRegisterAcceptsATeamworkClauseRewrite is the refusal #1703's
// builder hit on Cruel Alliance, lifted: a clause rewrite is not a
// payment component, so teamwork and blight may carry one.
func TestRegisterAcceptsATeamworkClauseRewrite(t *testing.T) {
	for _, cost := range []game.AdditionalCost{
		WhenPaid(Teamwork(2), TargetCreature("target creature")),
		WhenPaid(OptionalBlight(1), TargetCreature("target creature")),
	} {
		registerForTest(t, Spec{OracleID: "wp-test-" + cost.Key, Name: "Rewrite " + cost.Key,
			Completeness: CompletenessFull, OptionalCosts: []game.AdditionalCost{cost},
			Targets: TargetCreature("target creature with mana value 3 or less", ManaValueLE(3))})
	}
}

func TestCruelAllianceExilesSmallOrAnyWithTeamworkAndGainsLife(t *testing.T) {
	t.Run("plain", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
		four := wpCreature(g, opp.ID, "Four Drop", "{3}{G}")
		three := wpCreature(g, opp.ID, "Three Drop", "{2}{G}")
		if _, err := castPaying(t, g, "Cruel Alliance", "Sorcery", cruelAllianceOracle, twTarget(four),
			nil, nil, nil); !errors.Is(err, game.ErrIllegalTarget) {
			t.Fatalf("err = %v, want ErrIllegalTarget for a mana-value-4 creature without teamwork", err)
		}
		life := me.Life
		if _, err := castPaying(t, g, "Cruel Alliance", "Sorcery", cruelAllianceOracle, twTarget(three),
			nil, nil, nil); err != nil {
			t.Fatalf("CastSpell: %v", err)
		}
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(three) {
			t.Error("the three-drop should be exiled")
		}
		if me.Life != life {
			t.Errorf("a plain cast gains no life: %d → %d", life, me.Life)
		}
	})
	t.Run("teamwork", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
		team := pushVanillaCreature(g, me.ID, "Ally", 2, 2)
		five := wpCreature(g, opp.ID, "Five Drop", "{4}{G}")
		life := me.Life
		if _, err := castPaying(t, g, "Cruel Alliance", "Sorcery", cruelAllianceOracle, twTarget(five),
			[]int{0}, []uuid.UUID{team}, nil); err != nil {
			t.Fatalf("a teamwork cast may target any creature: %v", err)
		}
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(five) {
			t.Error("the five-drop should be exiled")
		}
		if me.Life != life+3 {
			t.Errorf("a teamwork cast gains 3 life: %d → %d", life, me.Life)
		}
	})
}

func TestBloodchiefsThirstKickedDestroysAnyCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	big := wpCreature(g, opp.ID, "Three Drop", "{2}{B}")
	if _, err := castWithOptionalCosts(t, g, "Bloodchief's Thirst", "Sorcery", bloodchiefsThirstOracle,
		twTarget(big), nil, nil); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("err = %v, want ErrIllegalTarget for a three-drop unkicked", err)
	}
	if _, err := castWithOptionalCosts(t, g, "Bloodchief's Thirst", "Sorcery", bloodchiefsThirstOracle,
		twTarget(big), []int{0}, nil); err != nil {
		t.Fatalf("kicked: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("the kicked Thirst destroys the three-drop")
	}
}

func TestTearAsunderKickedExilesACreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	bear := wpCreature(g, opp.ID, "Bear", "{1}{G}")
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	if _, err := castWithOptionalCosts(t, g, "Tear Asunder", "Instant", tearAsunderOracle,
		twTarget(bear), nil, nil); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("err = %v, want ErrIllegalTarget for a creature unkicked", err)
	}
	if _, err := castWithOptionalCosts(t, g, "Tear Asunder", "Instant", tearAsunderOracle,
		twTarget(rock), nil, nil); err != nil {
		t.Fatalf("unkicked at an artifact: %v", err)
	}
	passPriorityAroundTable(t, g)
	if _, err := castWithOptionalCosts(t, g, "Tear Asunder", "Instant", tearAsunderOracle,
		twTarget(bear), []int{0}, nil); err != nil {
		t.Fatalf("kicked: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || g.Battlefield.Contains(bear) {
		t.Error("both the artifact (plain) and the creature (kicked) should be exiled")
	}
	if !g.Exile.Contains(bear) {
		t.Error("the creature goes to exile, not the graveyard")
	}
}

func TestExpelTheUnworthyKickedExilesAndPaysItsController(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	five := wpCreature(g, opp.ID, "Five Drop", "{4}{W}")
	if _, err := castWithOptionalCosts(t, g, "Expel the Unworthy", "Sorcery", expelTheUnworthyOracle,
		twTarget(five), nil, nil); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("err = %v, want ErrIllegalTarget for a five-drop unkicked", err)
	}
	life := opp.Life
	if _, err := castWithOptionalCosts(t, g, "Expel the Unworthy", "Sorcery", expelTheUnworthyOracle,
		twTarget(five), []int{0}, nil); err != nil {
		t.Fatalf("kicked: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(five) {
		t.Error("the kicked Expel exiles the five-drop")
	}
	if opp.Life != life+5 {
		t.Errorf("its controller gains life equal to its mana value: %d → %d, want +5", life, opp.Life)
	}
}

func TestBloodBeckoningKickedReturnsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := b17GraveyardCard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	b := b17GraveyardCard(me, "Dead Elf", "Creature — Elf", "{G}")
	two := []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}
	if _, err := castWithOptionalCosts(t, g, "Blood Beckoning", "Sorcery", bloodBeckoningOracle,
		two, nil, nil); err == nil {
		t.Fatal("unkicked, Blood Beckoning takes one target")
	}
	if _, err := castWithOptionalCosts(t, g, "Blood Beckoning", "Sorcery", bloodBeckoningOracle,
		twTarget(a), []int{0}, nil); err == nil {
		t.Fatal("kicked, Blood Beckoning must name two targets")
	}
	if _, err := castWithOptionalCosts(t, g, "Blood Beckoning", "Sorcery", bloodBeckoningOracle,
		two, []int{0}, nil); err != nil {
		t.Fatalf("kicked: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) || !me.Hand.Contains(b) {
		t.Error("the kicked Beckoning returns both creature cards")
	}
}

func TestDivineResilienceKickedProtectsEveryTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	a := wpCreature(g, me.ID, "Bear A", "{1}{G}")
	b := wpCreature(g, me.ID, "Bear B", "{1}{G}")
	c := wpCreature(g, me.ID, "Bear C", "{1}{G}")
	theirs := wpCreature(g, opp.ID, "Their Bear", "{1}{G}")
	three := []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}, {Kind: game.TargetCard, ID: c}}
	if _, err := castWithOptionalCosts(t, g, "Divine Resilience", "Instant", divineResilienceOracle,
		three, nil, nil); err == nil {
		t.Fatal("unkicked, Divine Resilience takes one target")
	}
	if _, err := castWithOptionalCosts(t, g, "Divine Resilience", "Instant", divineResilienceOracle,
		twTarget(theirs), []int{0}, nil); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("err = %v, want ErrIllegalTarget — the kicked clause is still creatures you control", err)
	}
	if _, err := castWithOptionalCosts(t, g, "Divine Resilience", "Instant", divineResilienceOracle,
		three, []int{0}, nil); err != nil {
		t.Fatalf("kicked: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b, c} {
		if !effectiveAbilitiesContain(t, g, id, "indestructible") {
			t.Errorf("%s should be indestructible", id)
		}
	}
}
