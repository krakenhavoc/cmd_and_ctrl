package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b464_test.go — OLD batch 464 (#5576): Mirrodin Besieged cards
// Choking Fumes, Shriekhorn, Vedalken Infuser, Virulent Wound, Vedalken
// Anatomist, Sphere of the Suns and Gore Vassal.

const (
	oldB464Fumes     = "51bd7bc4-12cd-434e-8691-3123a018e634"
	oldB464Horn      = "571c1923-790f-48b7-86d2-70c3e79fdba3"
	oldB464Infuser   = "68cce1e6-1043-41b2-abd6-7e27ae0125e3"
	oldB464Wound     = "6a555c10-4dc2-455d-87b4-c7fdbf3d29a6"
	oldB464Anatomist = "7719fd5d-c9dc-49b8-87b4-24f7d42af62f"
	oldB464Sphere    = "7c709413-dacf-48af-8ac9-8f61b2699bf1"
	oldB464Vassal    = "e4f44a37-1e66-4f60-9af2-1e67ba58445b"
)

func TestOldB464EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB464Fumes: "Choking Fumes", oldB464Horn: "Shriekhorn", oldB464Infuser: "Vedalken Infuser",
		oldB464Wound: "Virulent Wound", oldB464Anatomist: "Vedalken Anatomist", oldB464Sphere: "Sphere of the Suns",
		oldB464Vassal: "Gore Vassal",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%s caveats=%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB464ChokingFumesPutsAMinusCounterOnEachAttackingCreatureOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 3, 3)
	second := effectstest.B12Creature(g, me.ID, "Second Attacker", "Creature — Bear", 3, 3)
	home := effectstest.B12Creature(g, me.ID, "Stay Home", "Creature — Bear", 3, 3)
	blocker := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	effectstest.DeclareAttack(t, g, opp.ID, attacker, second)
	effectstest.CastFromSeat(t, g, me, "Choking Fumes", "Instant", oldB464Fumes, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{attacker, second} {
		if got := effectstest.CountersOn(g, id, game.CounterMinusOne); got != 1 {
			t.Errorf("an attacker has %d -1/-1 counters, want 1", got)
		}
	}
	for _, id := range []uuid.UUID{home, blocker} {
		if got := effectstest.CountersOn(g, id, game.CounterMinusOne); got != 0 {
			t.Errorf("a creature that is not attacking has %d -1/-1 counters", got)
		}
	}
}

func TestOldB464ShriekhornEntersWithThreeChargeCountersAndMillsTwoPerCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	horn := effectstest.CastCatalogSpell(t, g, "Shriekhorn", "Artifact", oldB464Horn, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, horn, game.CounterCharge); got != 3 {
		t.Fatalf("%d charge counters on entering, want 3", got)
	}
	_ = me
	for want := 2; want >= 0; want-- {
		g.WithWriteLock(func() { effectstest.UntapForTest(g, horn) })
		lib := opp.Library.Size()
		effectstest.B16Activate(t, g, me.ID, horn, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.CountersOn(g, horn, game.CounterCharge); got != want {
			t.Errorf("%d charge counters left, want %d", got, want)
		}
		if milled := lib - opp.Library.Size(); milled != 2 {
			t.Errorf("milled %d, want 2", milled)
		}
	}
	g.WithWriteLock(func() { effectstest.UntapForTest(g, horn) })
	if err := g.ActivateCatalogAbility(me.ID, horn, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err == nil {
		t.Error("activated with no charge counters left")
	}
}

func TestOldB464SphereOfTheSunsEntersTappedAndSpendsACounterForAnyColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	sphere := effectstest.CastCatalogSpell(t, g, "Sphere of the Suns", "Artifact", oldB464Sphere, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, sphere); !c.Tapped {
		t.Error("it did not enter tapped")
	}
	if got := effectstest.CountersOn(g, sphere, game.CounterCharge); got != 3 {
		t.Fatalf("%d charge counters, want 3", got)
	}
	if err := g.ActivateManaAbility(me.ID, sphere, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("tapped for mana while tapped")
	}
	g.WithWriteLock(func() { effectstest.UntapForTest(g, sphere) })
	effectstest.TapForMana(t, g, me.ID, sphere)
	effectstest.ResolveManaPickAnyColor(t, g, me.ID)
	if got := me.ManaPool.Total(); got != 1 {
		t.Errorf("%d mana in the pool, want 1", got)
	}
	if got := effectstest.CountersOn(g, sphere, game.CounterCharge); got != 2 {
		t.Errorf("%d charge counters left, want 2", got)
	}
}

func TestOldB464VedalkenInfuserMayPutAChargeCounterOnAnArtifactAtUpkeep(t *testing.T) {
	for _, yes := range []bool{false, true} {
		g, me := effectstest.IwTable(t)
		effectstest.B12Push(g, me.ID, "Vedalken Infuser", "Creature — Vedalken Wizard", oldB464Infuser, 1, 4)
		rock := effectstest.SeedPermanent(g, me.ID, "Rock", "Artifact")
		seat := g.Turn.ActiveSeat
		effectstest.IwNextUpkeep(t, g, nil)
		for g.Turn.ActiveSeat != seat {
			effectstest.IwNextUpkeep(t, g, nil)
		}
		effectstest.PassPriorityUntilChoice(t, g)
		if effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePickTarget, me.ID) != nil {
			effectstest.AnswerFirstPickTarget(t, g)
			effectstest.PassPriorityUntilChoice(t, g)
		}
		effectstest.AnswerMayChoice(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if yes {
			want = 1
		}
		if got := effectstest.CountersOn(g, rock, game.CounterCharge); got != want {
			t.Errorf("yes=%v: %d charge counters, want %d", yes, got, want)
		}
	}
}

func TestOldB464VedalkenAnatomistPutsAMinusCounterAndMayTapOrUntap(t *testing.T) {
	for _, tc := range []struct {
		name   string
		pick   int
		tapped bool
	}{{"tap it", 0, true}, {"untap it", 1, false}, {"leave it", 2, true}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := effectstest.SeatsForCombat(g)
			effectstest.AdvanceToMain(t, g)
			ana := effectstest.B12Push(g, me.ID, "Vedalken Anatomist", "Creature — Phyrexian Vedalken Wizard", oldB464Anatomist, 1, 2)
			target := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Bear",
				TypeLine: "Creature — Bear", Power: 3, Toughness: 3, Tapped: tc.pick == 2, Owner: opp.ID, Controller: opp.ID})
			g.WithWriteLock(func() { effectstest.FillPool(me, 3) })
			effectstest.B16Activate(t, g, me.ID, ana, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(target)})
			effectstest.PassPriorityUntilChoice(t, g)
			effectstest.AnswerOptionPick(t, g, me.ID, tc.pick)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.CountersOn(g, target, game.CounterMinusOne); got != 1 {
				t.Errorf("%d -1/-1 counters, want 1", got)
			}
			if c, _ := effectstest.BattlefieldCard(g, target); c.Tapped != tc.tapped {
				t.Errorf("tapped=%v, want %v", c.Tapped, tc.tapped)
			}
		})
	}
}

func TestOldB464GoreVassalRegeneratesAThreeToughnessCreatureButNotAOneToughnessOne(t *testing.T) {
	for _, tc := range []struct {
		name      string
		toughness int
		dies      bool
	}{{"toughness 3", 3, false}, {"toughness 2 stays 1 or more", 2, false}, {"toughness 1 dies", 1, true}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := effectstest.SeatsForCombat(g)
			effectstest.AdvanceToMain(t, g)
			vassal := effectstest.B12Push(g, me.ID, "Gore Vassal", "Creature — Phyrexian Dog", oldB464Vassal, 2, 1)
			target := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, tc.toughness)
			effectstest.B16Activate(t, g, me.ID, vassal, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(target)})
			effectstest.PassPriorityAroundTable(t, g)
			g.RunStateChecksForTest()
			if effectstest.BattlefieldHas(g, vassal) {
				t.Error("the Vassal was not sacrificed")
			}
			if effectstest.BattlefieldHas(g, target) == tc.dies {
				t.Errorf("target alive=%v, want dies=%v", effectstest.BattlefieldHas(g, target), tc.dies)
			}
			if !tc.dies {
				c, _ := effectstest.BattlefieldCard(g, target)
				if c.RegenerationShields < 1 {
					t.Errorf("no regeneration shield on a toughness-%d creature", tc.toughness)
				}
			}
		})
	}
}

func TestOldB464VirulentWoundPoisonsTheControllerOfACreatureThatDiesThisTurn(t *testing.T) {
	for _, tc := range []struct {
		name      string
		toughness int
		killLater bool
		poison    int
	}{
		{"dies to the counter", 1, false, 1},
		{"survives and nothing kills it", 3, false, 0},
		{"survives then dies this turn", 3, true, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			effectstest.AdvanceToMain(t, g)
			target := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, tc.toughness)
			effectstest.CastCatalogSpell(t, g, "Virulent Wound", "Instant", oldB464Wound, effectstest.CardTarget(target))
			effectstest.PassPriorityAroundTable(t, g)
			g.RunStateChecksForTest()
			if tc.killLater {
				effectstest.KillCreature(t, g, opp.ID, target)
				g.RunStateChecksForTest()
				effectstest.PassPriorityAroundTable(t, g)
			}
			if opp.Poison != tc.poison || me.Poison != 0 {
				t.Errorf("their poison %d (want %d), mine %d (want 0)", opp.Poison, tc.poison, me.Poison)
			}
			if got := effectstest.CountersOn(g, target, game.CounterMinusOne); !tc.killLater && tc.toughness > 1 && got != 1 {
				t.Errorf("%d -1/-1 counters on the survivor, want 1", got)
			}
		})
	}
}
