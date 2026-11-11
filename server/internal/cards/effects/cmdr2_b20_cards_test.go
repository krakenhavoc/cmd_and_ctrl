package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_b20_cards_test.go — Commander wave 2, batch 20 (#5368): Aura
// payoffs and Auras entering without being cast.

const (
	cmdr2KnickknackOuphe  = "55adf2f0-f023-4f41-ae9d-3c002e9e6e8d"
	cmdr2TimberPaladin    = "0be44c76-9e3f-45c2-9b2c-6a5f00678780"
	cmdr2AuraGnarlid      = "586aebf2-4653-49fa-8425-13bd1f511bf0"
	cmdr2UmbraMystic      = "c0a4f47c-c7f9-4b95-aa1b-30580e4cf030"
	cmdr2Retether         = "b693c24f-07d9-40fa-aa4a-8f7ba61f3452"
	cmdr2LiberatedStock   = "d080cb03-b3e5-42e8-b289-2ce2151b4cb8"
	cmdr2Siona            = "89e17ccf-cde7-428d-bd64-47e2f79b40d7"
	cmdr2OxDrover         = "70f26726-0373-4e50-a1de-37392ca94890"
	cmdr2SilhanaLedgewalk = "5b15dcac-38dc-4957-bc01-40153ea07559"
)

func TestCmdr2B20EveryBuiltCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		cmdr2TimberPaladin:    "Timber Paladin",
		cmdr2AuraGnarlid:      "Aura Gnarlid",
		cmdr2OxDrover:         "Ox Drover",
		cmdr2SilhanaLedgewalk: "Silhana Ledgewalker",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestCmdr2B20TimberPaladinGrowsWithTheAurasOnIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	paladin := dskB1Creature(g, me.ID, "Timber Paladin", "Artifact Creature — Knight", cmdr2TimberPaladin, 1, 1)
	type want struct {
		p, tt            int
		vigilance, tramp bool
	}
	check := func(n int, w want) {
		t.Helper()
		if p, tt := ptOf(t, g, paladin); p != w.p || tt != w.tt {
			t.Errorf("%d Auras: %d/%d, want %d/%d", n, p, tt, w.p, w.tt)
		}
		if got := hasEffectiveKeyword(t, g, paladin, "vigilance"); got != w.vigilance {
			t.Errorf("%d Auras: vigilance = %v, want %v", n, got, w.vigilance)
		}
		if got := hasEffectiveKeyword(t, g, paladin, "trample"); got != w.tramp {
			t.Errorf("%d Auras: trample = %v, want %v", n, got, w.tramp)
		}
	}
	check(0, want{1, 1, false, false})
	cmdr2B19Attach(g, me.ID, paladin, "Aura One", "")
	check(1, want{3, 3, false, false})
	// An opponent's Aura counts as well.
	cmdr2B19Attach(g, opp.ID, paladin, "Aura Two", "")
	check(2, want{5, 5, true, false})
	cmdr2B19Attach(g, me.ID, paladin, "Aura Three", "")
	check(3, want{10, 10, true, true})
	cmdr2B19Attach(g, me.ID, paladin, "Aura Four", "")
	check(4, want{10, 10, true, true})
}

func TestCmdr2B20TimberPaladinCountersAndPumpsApplyOnTheBase(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	paladin := dskB1Creature(g, me.ID, "Timber Paladin", "Artifact Creature — Knight", cmdr2TimberPaladin, 1, 1)
	cmdr2B19Attach(g, me.ID, paladin, "Spectral Steel", cmdr2SpectralSteel)
	// Base 3/3 from one Aura, +2/+2 from Spectral Steel itself.
	if p, tt := ptOf(t, g, paladin); p != 5 || tt != 5 {
		t.Errorf("Paladin with Spectral Steel is %d/%d, want 5/5", p, tt)
	}
}

func TestCmdr2B20AuraGnarlidCountsEveryAuraAndOutranksSmallBlockers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gnarlid := dskB1Creature(g, me.ID, "Aura Gnarlid", "Creature — Beast", cmdr2AuraGnarlid, 2, 2)
	host := b12Creature(g, opp.ID, "Host", "Creature — Bear", 1, 1)
	if p, tt := ptOf(t, g, gnarlid); p != 2 || tt != 2 {
		t.Fatalf("Gnarlid with no Auras is %d/%d, want 2/2", p, tt)
	}
	cmdr2B19Attach(g, opp.ID, host, "Their Aura", "")
	cmdr2B19Attach(g, me.ID, gnarlid, "My Aura", "")
	if p, tt := ptOf(t, g, gnarlid); p != 4 || tt != 4 {
		t.Fatalf("Gnarlid with two Auras on the battlefield is %d/%d, want 4/4", p, tt)
	}
	small := b12Creature(g, opp.ID, "Small", "Creature — Bear", 3, 3)
	equal := b12Creature(g, opp.ID, "Equal", "Creature — Bear", 4, 4)
	declareAttack(t, g, opp.ID, gnarlid)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(small, gnarlid); err == nil {
		t.Error("a creature with less power than the Gnarlid blocked it")
	}
	if err := g.DeclareBlocker(equal, gnarlid); err != nil {
		t.Errorf("a creature with equal power could not block: %v", err)
	}
}

func TestCmdr2B20SilhanaLedgewalkerOnlyFlyersBlockAndItIsHexproof(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	silhana := dskB1Creature(g, me.ID, "Silhana Ledgewalker", "Creature — Elf Rogue", cmdr2SilhanaLedgewalk, 1, 1)
	if !hasKeywordOnBattlefield(t, g, silhana, "hexproof") {
		t.Error("Silhana Ledgewalker has no hexproof")
	}
	ground := b12Creature(g, opp.ID, "Ground", "Creature — Bear", 2, 2)
	flyer := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Flyer", TypeLine: "Creature — Bird",
		Power: 2, Toughness: 2, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID})
	declareAttack(t, g, opp.ID, silhana)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(ground, silhana); err == nil {
		t.Error("a creature without flying blocked Silhana")
	}
	if err := g.DeclareBlocker(flyer, silhana); err != nil {
		t.Errorf("a creature with flying could not block: %v", err)
	}
}

func TestCmdr2B20OxDroverGivesAnOpponentAnOxAndDrawsWhenItEntersAndAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "a", "b", "c", "d")
	advanceToMain(t, g)
	before := len(me.Hand.Cards)
	drover := ecl6Cast(t, g, me, "Ox Drover", "Creature — Human Peasant", cmdr2OxDrover, 4, 4, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if n := namedOnBattlefield(g, "Ox", opp.ID); n != 1 {
		t.Fatalf("%d Oxen for the target opponent after entering, want 1", n)
	}
	if got := len(me.Hand.Cards) - before; got != 1 {
		t.Errorf("drew %d cards on entering, want 1", got)
	}
	if !hasKeywordOnBattlefield(t, g, drover, "vigilance") {
		t.Error("no vigilance")
	}
	// Summoning sickness is over once it has been here a turn: attack on my next turn.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == drover {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	declareAttack(t, g, opp.ID, drover)
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if n := namedOnBattlefield(g, "Ox", opp.ID); n != 2 {
		t.Errorf("%d Oxen after attacking, want 2", n)
	}
	if got := len(me.Hand.Cards) - before; got != 2 {
		t.Errorf("drew %d cards after entering and attacking, want 2", got)
	}
	// The Oxen it hands out can't block it.
	advanceTo(t, g, game.StepDeclareBlockers)
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Ox" {
			if err := g.DeclareBlocker(c.InstanceID, drover); err == nil {
				t.Error("an Ox blocked Ox Drover")
			}
		}
	}
}
