package testsu

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestWhenPaidSetsTheCostsClauseAndRefusesNil(t *testing.T) {
	clause := effects.TargetCreature("target creature")
	got := effects.WhenPaid(effects.Teamwork(2), clause)
	if got.Targets != clause || got.Teamwork != 2 || got.Key != game.TeamworkKey || !got.Optional {
		t.Errorf("WhenPaid(Teamwork(2), …) = %+v, want the teamwork cost carrying the clause", got)
	}
	effectstest.MustPanic(t, "nil clause", func() { effects.WhenPaid(effects.Kicker("{1}"), nil) })
}

// TestRegisterAcceptsATeamworkClauseRewrite is the refusal #1703's
// builder hit on Cruel Alliance, lifted: a clause rewrite is not a
// payment component, so teamwork and blight may carry one.
func TestRegisterAcceptsATeamworkClauseRewrite(t *testing.T) {
	for _, cost := range []game.AdditionalCost{
		effects.WhenPaid(effects.Teamwork(2), effects.TargetCreature("target creature")),
		effects.WhenPaid(effects.OptionalBlight(1), effects.TargetCreature("target creature")),
	} {
		effectstest.RegisterForTest(t, effects.Spec{OracleID: "wp-test-" + cost.Key, Name: "Rewrite " + cost.Key,
			Completeness: effects.CompletenessFull, OptionalCosts: []game.AdditionalCost{cost},
			Targets: effects.TargetCreature("target creature with mana value 3 or less", effects.ManaValueLE(3))})
	}
}

func TestCruelAllianceExilesSmallOrAnyWithTeamworkAndGainsLife(t *testing.T) {
	t.Run("plain", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
		four := effectstest.WpCreature(g, opp.ID, "Four Drop", "{3}{G}")
		three := effectstest.WpCreature(g, opp.ID, "Three Drop", "{2}{G}")
		if _, err := effectstest.CastPaying(t, g, "Cruel Alliance", "Sorcery", effectstest.CruelAllianceOracle, effectstest.TwTarget(four),
			nil, nil, nil); !errors.Is(err, game.ErrIllegalTarget) {
			t.Fatalf("err = %v, want ErrIllegalTarget for a mana-value-4 creature without teamwork", err)
		}
		life := me.Life
		if _, err := effectstest.CastPaying(t, g, "Cruel Alliance", "Sorcery", effectstest.CruelAllianceOracle, effectstest.TwTarget(three),
			nil, nil, nil); err != nil {
			t.Fatalf("CastSpell: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(three) {
			t.Error("the three-drop should be exiled")
		}
		if me.Life != life {
			t.Errorf("a plain cast gains no life: %d → %d", life, me.Life)
		}
	})
	t.Run("teamwork", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
		team := effectstest.PushVanillaCreature(g, me.ID, "Ally", 2, 2)
		five := effectstest.WpCreature(g, opp.ID, "Five Drop", "{4}{G}")
		life := me.Life
		if _, err := effectstest.CastPaying(t, g, "Cruel Alliance", "Sorcery", effectstest.CruelAllianceOracle, effectstest.TwTarget(five),
			[]int{0}, []uuid.UUID{team}, nil); err != nil {
			t.Fatalf("a teamwork cast may target any creature: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(five) {
			t.Error("the five-drop should be exiled")
		}
		if me.Life != life+3 {
			t.Errorf("a teamwork cast gains 3 life: %d → %d", life, me.Life)
		}
	})
}

func TestBloodchiefsThirstKickedDestroysAnyCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	big := effectstest.WpCreature(g, opp.ID, "Three Drop", "{2}{B}")
	if _, err := effectstest.CastWithOptionalCosts(t, g, "Bloodchief's Thirst", "Sorcery", effectstest.BloodchiefsThirstOracle,
		effectstest.TwTarget(big), nil, nil); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("err = %v, want ErrIllegalTarget for a three-drop unkicked", err)
	}
	if _, err := effectstest.CastWithOptionalCosts(t, g, "Bloodchief's Thirst", "Sorcery", effectstest.BloodchiefsThirstOracle,
		effectstest.TwTarget(big), []int{0}, nil); err != nil {
		t.Fatalf("kicked: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("the kicked Thirst destroys the three-drop")
	}
}

func TestTearAsunderKickedExilesACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	bear := effectstest.WpCreature(g, opp.ID, "Bear", "{1}{G}")
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	if _, err := effectstest.CastWithOptionalCosts(t, g, "Tear Asunder", "Instant", effectstest.TearAsunderOracle,
		effectstest.TwTarget(bear), nil, nil); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("err = %v, want ErrIllegalTarget for a creature unkicked", err)
	}
	if _, err := effectstest.CastWithOptionalCosts(t, g, "Tear Asunder", "Instant", effectstest.TearAsunderOracle,
		effectstest.TwTarget(rock), nil, nil); err != nil {
		t.Fatalf("unkicked at an artifact: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, err := effectstest.CastWithOptionalCosts(t, g, "Tear Asunder", "Instant", effectstest.TearAsunderOracle,
		effectstest.TwTarget(bear), []int{0}, nil); err != nil {
		t.Fatalf("kicked: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || g.Battlefield.Contains(bear) {
		t.Error("both the artifact (plain) and the creature (kicked) should be exiled")
	}
	if !g.Exile.Contains(bear) {
		t.Error("the creature goes to exile, not the graveyard")
	}
}

func TestExpelTheUnworthyKickedExilesAndPaysItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	five := effectstest.WpCreature(g, opp.ID, "Five Drop", "{4}{W}")
	if _, err := effectstest.CastWithOptionalCosts(t, g, "Expel the Unworthy", "Sorcery", effectstest.ExpelTheUnworthyOracle,
		effectstest.TwTarget(five), nil, nil); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("err = %v, want ErrIllegalTarget for a five-drop unkicked", err)
	}
	life := opp.Life
	if _, err := effectstest.CastWithOptionalCosts(t, g, "Expel the Unworthy", "Sorcery", effectstest.ExpelTheUnworthyOracle,
		effectstest.TwTarget(five), []int{0}, nil); err != nil {
		t.Fatalf("kicked: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(five) {
		t.Error("the kicked Expel exiles the five-drop")
	}
	if opp.Life != life+5 {
		t.Errorf("its controller gains life equal to its mana value: %d → %d, want +5", life, opp.Life)
	}
}

func TestBloodBeckoningKickedReturnsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := effectstest.B17GraveyardCard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	b := effectstest.B17GraveyardCard(me, "Dead Elf", "Creature — Elf", "{G}")
	two := []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}
	if _, err := effectstest.CastWithOptionalCosts(t, g, "Blood Beckoning", "Sorcery", effectstest.BloodBeckoningOracle,
		two, nil, nil); err == nil {
		t.Fatal("unkicked, Blood Beckoning takes one target")
	}
	if _, err := effectstest.CastWithOptionalCosts(t, g, "Blood Beckoning", "Sorcery", effectstest.BloodBeckoningOracle,
		effectstest.TwTarget(a), []int{0}, nil); err == nil {
		t.Fatal("kicked, Blood Beckoning must name two targets")
	}
	if _, err := effectstest.CastWithOptionalCosts(t, g, "Blood Beckoning", "Sorcery", effectstest.BloodBeckoningOracle,
		two, []int{0}, nil); err != nil {
		t.Fatalf("kicked: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) || !me.Hand.Contains(b) {
		t.Error("the kicked Beckoning returns both creature cards")
	}
}

func TestDivineResilienceKickedProtectsEveryTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	a := effectstest.WpCreature(g, me.ID, "Bear A", "{1}{G}")
	b := effectstest.WpCreature(g, me.ID, "Bear B", "{1}{G}")
	c := effectstest.WpCreature(g, me.ID, "Bear C", "{1}{G}")
	theirs := effectstest.WpCreature(g, opp.ID, "Their Bear", "{1}{G}")
	three := []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}, {Kind: game.TargetCard, ID: c}}
	if _, err := effectstest.CastWithOptionalCosts(t, g, "Divine Resilience", "Instant", effectstest.DivineResilienceOracle,
		three, nil, nil); err == nil {
		t.Fatal("unkicked, Divine Resilience takes one target")
	}
	if _, err := effectstest.CastWithOptionalCosts(t, g, "Divine Resilience", "Instant", effectstest.DivineResilienceOracle,
		effectstest.TwTarget(theirs), []int{0}, nil); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("err = %v, want ErrIllegalTarget — the kicked clause is still creatures you control", err)
	}
	if _, err := effectstest.CastWithOptionalCosts(t, g, "Divine Resilience", "Instant", effectstest.DivineResilienceOracle,
		three, []int{0}, nil); err != nil {
		t.Fatalf("kicked: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b, c} {
		if !effectstest.EffectiveAbilitiesContain(t, g, id, "indestructible") {
			t.Errorf("%s should be indestructible", id)
		}
	}
}
