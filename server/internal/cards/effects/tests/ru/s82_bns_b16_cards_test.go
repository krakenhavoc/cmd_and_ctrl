package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bns_b16_cards_test.go — S82 BNS batch 16 (#5929), Ravnica: Clue Edition:
// twelve creatures. Every test plays the real catalog card through the engine.

const (
	bnsB16StampedeSurfer = "22640b75-8db6-4264-a4aa-9174c159a228"
	bnsB16SteepleCreeper = "f53f33ed-5c8c-4835-bc5f-65a8acfa3f5b"
	bnsB16Stormblood     = "84e56125-3fbb-4229-8aa0-8f8acff27038"
	bnsB16SumalaRumblers = "09189f67-b33f-4775-bfd5-862154da7526"
	bnsB16SunhomeStalwrt = "340d6218-d351-4b2e-80da-7d4afa77190a"
	bnsB16SyndicOfTithes = "8f5c52d2-0583-4c94-a043-15b2118877fa"
	bnsB16SyndEnforcer   = "3a067b98-e775-4b44-addf-9d31716c2191"
	bnsB16SyndHeavy      = "01becf76-aa53-4fd5-94c1-665c64674219"
	bnsB16SyndMessenger  = "f5f4d78b-ed2b-471b-9ece-237fb55adc13"
	bnsB16TerritorialBr  = "848fc2b8-7208-4415-8be1-add84c780f6e"
	bnsB16ThrashMossdog  = "8778ddbe-9aba-4a45-a22a-5453fefef952"
	bnsB16ThunderDrake   = "a0956c3d-d382-486a-81f1-9aa5cf49b2ab"
)

func TestBnsB16EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		bnsB16StampedeSurfer: "Stampede Surfer", bnsB16SteepleCreeper: "Steeple Creeper",
		bnsB16Stormblood: "Stormblood Berserker", bnsB16SumalaRumblers: "Sumala Rumblers",
		bnsB16SunhomeStalwrt: "Sunhome Stalwart", bnsB16SyndicOfTithes: "Syndic of Tithes",
		bnsB16SyndEnforcer: "Syndicate Enforcer", bnsB16SyndHeavy: "Syndicate Heavy",
		bnsB16SyndMessenger: "Syndicate Messenger", bnsB16TerritorialBr: "Territorial Boar",
		bnsB16ThrashMossdog: "Thrashing Mossdog", bnsB16ThunderDrake: "Thunder Drake",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%s caveats=%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestBnsB16StampedeSurferMakesABoarAttackingEachOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	surfer := effectstest.B12Push(g, me.ID, "Stampede Surfer", "Creature — Human Warrior", bnsB16StampedeSurfer, 4, 4)
	effectstest.FinABAttack(t, g, effectstest.FinABOpp(g).ID, surfer)
	effectstest.PassPriorityAroundTable(t, g)
	boars := dmuB17Tokens(g, "Boar")
	if len(boars) != len(g.Seats)-1 {
		t.Fatalf("%d Boars, want one per opponent (%d)", len(boars), len(g.Seats)-1)
	}
	targets := map[uuid.UUID]bool{}
	for _, b := range boars {
		if !b.Tapped || b.AttackingTarget == uuid.Nil || b.Power != 2 || b.Toughness != 2 || !b.HasColor("G") {
			t.Errorf("Boar %+v: want a tapped, attacking green 2/2", b)
		}
		targets[b.AttackingTarget] = true
	}
	if len(targets) != len(g.Seats)-1 || targets[me.ID] {
		t.Errorf("Boars attack %v, want each opponent exactly once", targets)
	}
}

func TestBnsB16SumalaRumblersPowerCountsYourCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rum := effectstest.B12Push(g, me.ID, "Sumala Rumblers", "Creature — Wurm", bnsB16SumalaRumblers, 0, 4)
	if p, tt := effectstest.PtOf(t, g, rum); p != 1 || tt != 4 {
		t.Errorf("alone: %d/%d, want 1/4", p, tt)
	}
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if p, _ := effectstest.PtOf(t, g, rum); p != 3 {
		t.Errorf("with two more of mine and one of theirs: power %d, want 3", p)
	}
}

func TestBnsB16ThunderDrakeGrowsOnTheSecondSpellOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	drake := effectstest.B12Push(g, me.ID, "Thunder Drake", "Creature — Elemental Drake", bnsB16ThunderDrake, 2, 3)
	for i, want := range []int{0, 1, 1} {
		effectstest.CastCatalogSpell(t, g, "Sol Ring", "Artifact", effectstest.SolRingOracle, nil)
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.CounterCount(g, drake, game.CounterPlusOne); got != want {
			t.Errorf("after spell %d: %d counters, want %d", i+1, got, want)
		}
	}
}

func TestBnsB16TerritorialBoarGrowsForABigCreatureOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	boar := effectstest.B12Push(g, me.ID, "Territorial Boar", "Creature — Boar", bnsB16TerritorialBr, 2, 2)
	effectstest.RfpbCastCreature(t, g, "Small", 3)
	if p, tt := effectstest.PtOf(t, g, boar); p != 2 || tt != 2 {
		t.Errorf("after a power-3 creature: %d/%d, want 2/2", p, tt)
	}
	effectstest.RfpbCastCreature(t, g, "Big", 4)
	if p, tt := effectstest.PtOf(t, g, boar); p != 3 || tt != 3 {
		t.Errorf("after a power-4 creature: %d/%d, want 3/3", p, tt)
	}
	if !effectstest.HasEffectiveKeyword(t, g, boar, "vigilance") {
		t.Error("the Boar did not gain vigilance")
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if p, _ := effectstest.PtOf(t, g, boar); p != 2 || effectstest.HasEffectiveKeyword(t, g, boar, "vigilance") {
		t.Error("the bonus outlived the turn")
	}
}

func TestBnsB16SyndicateHeavyInvestigatesAfterFourLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Syndicate Heavy", "Creature — Giant Rogue", bnsB16SyndHeavy, 4, 4)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3) })
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHasName(g, "Clue") {
		t.Fatal("3 life gained made a Clue")
	}
	effectstest.AdvanceToMainOf(t, g, 0)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 4) })
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.BattlefieldHasName(g, "Clue") {
		t.Error("4 life gained did not investigate")
	}
}

func TestBnsB16ExtortCardsDrainOnACast(t *testing.T) {
	for _, tc := range []struct{ name, line, oracle string }{
		{"Syndic of Tithes", "Creature — Human Cleric", bnsB16SyndicOfTithes},
		{"Syndicate Enforcer", "Creature — Human Rogue", bnsB16SyndEnforcer},
		{"Syndicate Heavy", "Creature — Giant Rogue", bnsB16SyndHeavy},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.B12Push(g, me.ID, tc.name, tc.line, tc.oracle, 2, 2)
			before, mine := effectstest.LifeOf(g, g.Seats[1].ID), effectstest.LifeOf(g, me.ID)
			effectstest.E2CastSpellOfColor(t, g, "Some Spell", "Sorcery", "U")
			for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, me.ID); i++ {
				if err := g.PassPriority(); err != nil {
					t.Fatal(err)
				}
			}
			if !effectstest.HasPayUnlessFor(g, me.ID) {
				t.Fatal("no extort payment offered")
			}
			me.ManaPool.AddMana(game.ManaToken{Color: "B"})
			effectstest.AnswerPayUnless(t, g, me.ID, true)
			if got := effectstest.LifeOf(g, g.Seats[1].ID); got != before-1 {
				t.Errorf("opponent life %d, want %d", got, before-1)
			}
			if got := effectstest.LifeOf(g, me.ID); got < mine+1 {
				t.Errorf("own life %d, want at least %d", got, mine+1)
			}
		})
	}
}

func TestBnsB16SyndicateMessengerLeavesASpiritWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	msg := effectstest.B12Push(g, me.ID, "Syndicate Messenger", "Creature — Bird", bnsB16SyndMessenger, 2, 3)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(msg) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	spirits := dmuB17Tokens(g, "Spirit")
	if len(spirits) != 1 || !spirits[0].HasColor("W") || !spirits[0].HasColor("B") || spirits[0].Power != 1 {
		t.Fatalf("spirits = %+v, want one white and black 1/1", spirits)
	}
	if !effectstest.HasEffectiveKeyword(t, g, spirits[0].InstanceID, "flying") {
		t.Error("the Spirit lacks flying")
	}
}

func TestBnsB16StormbloodBerserkerEntersBigOnlyAfterDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cast := func() uuid.UUID {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Stormblood Berserker", TypeLine: "Creature — Human Berserker",
			OracleID: bnsB16Stormblood, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
		effectstest.ToMainForCost(t, g)
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		return id
	}
	if p, _ := effectstest.PtOf(t, g, cast()); p != 1 {
		t.Errorf("no damage dealt: power %d, want 1", p)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(uuid.Nil, g.Seats[1].ID, 1) })
	if p, _ := effectstest.PtOf(t, g, cast()); p != 3 {
		t.Errorf("after damage: power %d, want 3", p)
	}
}

func TestBnsB16SteepleCreeperGainsFlyingForFourBlueMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	c := effectstest.B12Push(g, me.ID, "Steeple Creeper", "Creature — Frog Snake", bnsB16SteepleCreeper, 4, 2)
	if effectstest.HasEffectiveKeyword(t, g, c, "flying") {
		t.Fatal("flying before activating")
	}
	effectstest.RavB2MustActivate(t, g, me.ID, c, 0, game.ActivateAbilityParams{}, "C", "C", "C", "U")
	if !effectstest.HasEffectiveKeyword(t, g, c, "flying") {
		t.Error("no flying after activating")
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if effectstest.HasEffectiveKeyword(t, g, c, "flying") {
		t.Error("flying outlived the turn")
	}
}

func TestBnsB16SunhomeStalwartMentorsASmallerAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	st := effectstest.B12Push(g, me.ID, "Sunhome Stalwart", "Creature — Human Soldier", bnsB16SunhomeStalwrt, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	if !effectstest.HasEffectiveKeyword(t, g, st, "first strike") {
		t.Error("no first strike")
	}
	effectstest.FinABAttack(t, g, effectstest.FinABOpp(g).ID, st, bear)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickCard(t, g, me.ID, bear)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the smaller attacker has %d counters, want 1", got)
	}
	if got := effectstest.CounterCount(g, st, game.CounterPlusOne); got != 0 {
		t.Errorf("the mentor has %d counters, want 0", got)
	}
}

func TestBnsB16ThrashingMossdogScavengesForItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	id, _ := effectstest.ActivateFromGraveyard(t, g, "Thrashing Mossdog", "Creature — Plant Dog", bnsB16ThrashMossdog, 3, 3,
		"{C}{C}{C}{C}{G}{G}", game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 3 {
		t.Errorf("the target got %d counters, want 3", got)
	}
	if effectstest.BattlefieldHas(g, id) {
		t.Error("the Mossdog is on the battlefield")
	}
}
