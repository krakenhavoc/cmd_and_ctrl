package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b187_test.go — Guilds of Ravnica, batch 187 (#5525): damage cards.
// Each test plays the real catalog entry and asserts what its text says, and
// what it refuses.

const (
	s82o187Uprising  = "53e540b5-8338-4b21-a61a-c5b7fe652a6e"
	s82o187Guildmage = "bb33abbe-b43c-4a74-b2db-cce949f0edf2"
	s82o187Blow      = "bc2bf46f-2c9b-4f85-b8d6-d13a3b64436d"
	s82o187Strike    = "bc63e7ca-e62a-4026-803d-eecdbbedf3c6"
	s82o187Clarion   = "c9bbebbd-9f2d-499c-adb1-8d3170c3fd0d"
	s82o187Staticast = "eae7dd41-3d5f-49a7-baaa-2c03bb341c3a"
	s82o187Giant     = "fc16a2ce-a570-454f-8498-f9c604c0d68b"
)

func TestS82Old187EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82o187Uprising:  "Undercity Uprising",
		s82o187Guildmage: "Legion Guildmage",
		s82o187Blow:      "Righteous Blow",
		s82o187Strike:    "Justice Strike",
		s82o187Clarion:   "Deafening Clarion",
		s82o187Staticast: "Ral's Staticaster",
		s82o187Giant:     "Swathcutter Giant",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func s82o187Walker(g *game.Game, owner uuid.UUID, name, subtype string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Legendary Planeswalker — " + subtype,
		Counters: map[string]int{game.CounterLoyalty: 5}, Owner: owner, Controller: owner,
	})
}

func TestS82Old187UndercityUprisingGivesDeathtouchThenFights(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mouse", "Creature — Mouse", 1, 1)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	effectstest.CastCatalogSpell(t, g, "Undercity Uprising", "Sorcery", s82o187Uprising, effectstest.CardRefs(mine, theirs))
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(theirs) {
		t.Error("the 9-toughness wall survived a 1-power deathtouch fight")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("the fighter died against a 0-power wall")
	}
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, other), "deathtouch") {
		t.Error("another creature you control did not gain deathtouch")
	}
}

func TestS82Old187UndercityUprisingRefusesYourOwnCreatureAsTheSecondTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Undercity Uprising", "Sorcery", s82o187Uprising, effectstest.CardRefs(a, b)); err == nil {
		t.Error("a creature you control was accepted for the second target")
	}
}

func TestS82Old187LegionGuildmageHitsEachOpponentAndTaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMainForCost(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	lg := effectstest.PushCatalogPermanent(g, me.ID, "Legion Guildmage", "Creature — Human Wizard", s82o187Guildmage, false)
	var lives []int
	for _, p := range g.Seats {
		lives = append(lives, p.Life)
	}
	effectstest.Acrb1AddMana(t, g, me.ID, "{R}{R}{R}{R}{R}{R}")
	if err := g.ActivateCatalogAbility(me.ID, lg, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		want := lives[i]
		if p.ID != me.ID {
			want -= 3
		}
		if p.Life != want {
			t.Errorf("seat %d life %d, want %d", i, p.Life, want)
		}
	}
}

func TestS82Old187LegionGuildmageTapsAnotherCreatureButNotItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMainForCost(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	lg := effectstest.PushCatalogPermanent(g, me.ID, "Legion Guildmage", "Creature — Human Wizard", s82o187Guildmage, false)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	effectstest.Acrb1AddMana(t, g, me.ID, "{W}{W}{W}")
	if err := g.ActivateCatalogAbility(me.ID, lg, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(lg)}); err == nil {
		t.Fatal("the Guildmage was allowed to target itself")
	}
	if err := g.ActivateCatalogAbility(me.ID, lg, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(victim)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCardByID(g, victim); !c.Tapped {
		t.Error("the target was not tapped")
	}
}

func TestS82Old187RighteousBlowHitsAnAttackerOrBlockerOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	idle := effectstest.B12Creature(g, opp.ID, "Idle", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Righteous Blow", "Instant", s82o187Blow, effectstest.CardRefs(idle)); err == nil {
		t.Fatal("a creature that is neither attacking nor blocking was accepted")
	}
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	blow := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: blow, Name: "Righteous Blow", TypeLine: "Instant", OracleID: s82o187Blow, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, blow, game.CastSpellParams{Targets: effectstest.CardRefs(attacker)}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(attacker) {
		t.Error("the attacking 2/2 survived 2 damage")
	}
}

func TestS82Old187JusticeStrikeDamagesItselfByItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Giant", 4, 4)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 3, 5)
	effectstest.CastCatalogSpell(t, g, "Justice Strike", "Instant", s82o187Strike, effectstest.CardRefs(big))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Justice Strike", "Instant", s82o187Strike, effectstest.CardRefs(small))
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(big) {
		t.Error("a 4/4 survived 4 damage from itself")
	}
	if !g.Battlefield.Contains(small) {
		t.Error("a 3/5 died to 3 damage from itself")
	}
}

func TestS82Old187DeafeningClarionModes(t *testing.T) {
	for _, tc := range []struct {
		name           string
		modes          []int
		mineDies, lifl bool
	}{
		{"damage only", []int{0}, true, false},
		{"lifelink only", []int{1}, false, true},
		{"both", []int{0, 1}, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
			tough := effectstest.B12Creature(g, me.ID, "Tough", "Creature — Wall", 0, 4)
			theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
			effectstest.CastCatalogSpellWithModes(t, g, "Deafening Clarion", "Sorcery", s82o187Clarion, tc.modes)
			effectstest.PassPriorityAroundTable(t, g)
			g.RunStateChecksForTest()
			if g.Battlefield.Contains(mine) == tc.mineDies || g.Battlefield.Contains(theirs) == tc.mineDies {
				t.Errorf("2/2s on the battlefield: mine=%v theirs=%v, dies=%v", g.Battlefield.Contains(mine), g.Battlefield.Contains(theirs), tc.mineDies)
			}
			if !g.Battlefield.Contains(tough) {
				t.Fatal("the 0/4 died to 3 damage")
			}
			if got := effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, tough), "lifelink"); got != tc.lifl {
				t.Errorf("surviving creature has lifelink = %v, want %v", got, tc.lifl)
			}
		})
	}
}

func TestS82Old187StaticasterGrowsByHandSizeWithARalOnly(t *testing.T) {
	for _, withRal := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		if withRal {
			s82o187Walker(g, me.ID, "Ral, Test", "Ral")
		}
		sc := effectstest.PushCatalogPermanent(g, me.ID, "Ral's Staticaster", "Creature — Lizard Wizard", s82o187Staticast, false)
		effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Fill", TypeLine: "Land", Owner: me.ID, Controller: me.ID})
		for i := 0; i < 3; i++ {
			me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
		}
		hand := me.Hand.Size()
		base := effectstest.EffectivePower(t, g, sc)
		effectstest.AlaB1Attack(t, g, opp.ID, sc)
		effectstest.PassPriorityAroundTable(t, g)
		want := base
		if withRal {
			want += hand
		}
		if p := effectstest.EffectivePower(t, g, sc); p != want {
			t.Errorf("with a Ral = %v: power %d, want %d (hand %d)", withRal, p, want, hand)
		}
	}
}

func TestS82Old187SwathcutterGiantPingsTheDefendersCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	gc := effectstest.PushCatalogPermanent(g, me.ID, "Swathcutter Giant", "Creature — Giant Soldier", s82o187Giant, false)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
	sturdy := effectstest.B12Creature(g, opp.ID, "Sturdy", "Creature — Wall", 0, 3)
	effectstest.AlaB1Attack(t, g, opp.ID, gc)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(theirs) {
		t.Error("the defender's 1/1 survived")
	}
	if !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(sturdy) {
		t.Error("your own creature or a sturdier one died")
	}
	if n := effectstest.Pr6Marked(g, sturdy); n != 1 {
		t.Errorf("sturdy creature has %d damage, want 1", n)
	}
}
