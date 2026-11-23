package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b228_cards_test.go — Commander Legends: Battle for Baldur's Gate
// instants, enchantments and Backgrounds, GAP batch 228 (#5948). Each test
// plays the real catalog entry through the engine.

const (
	gap228StrikeOracle     = "139574fd-c61b-4615-9200-59f2c65f5703"
	gap228BanishmentOracle = "9c9b3bf7-afb7-4e9b-bdbc-e2b4d5ba477f"
	gap228TravelerOracle   = "b9432430-df01-497d-b025-7576dce830e4"
	gap228SweepOracle      = "2a747cdb-afbe-451d-8b1b-9458401a5419"
	gap228SoldierOracle    = "6bbf1410-800d-4d05-93d2-ba3c1656728e"
	gap228ThirstOracle     = "0236ad13-d8e4-49b4-b397-fe2f6136665d"
	gap228LanternOracle    = "4af03564-1d1c-4582-9248-cfe9551bad30"
	gap228RespiteOracle    = "1c19b04d-b307-4f70-b89e-af306374b9db"
	gap228ReclamationOracl = "98266582-12fd-4a7b-a9bb-e9cc4370ed28"
	gap228PoisonOracle     = "8af22295-0d6b-4cdc-b815-eb81ade241ec"
	gap228Background       = "Legendary Enchantment — Background"
)

func TestGapB228EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		gap228StrikeOracle:         "Stunning Strike",
		gap228BanishmentOracle:     "Banishment",
		gap228TravelerOracle:       "Far Traveler",
		gap228SweepOracle:          "Undercellar Sweep",
		gap228SoldierOracle:        "Veteran Soldier",
		gap228ThirstOracle:         "Astarion's Thirst",
		gap228LanternOracle:        "Ghost Lantern",
		gap228LanternOracle + "#1": "Bind Spirit",
		gap228RespiteOracle:        "Jaheira's Respite",
		gap228ReclamationOracl:     "Natural Reclamation",
		gap228PoisonOracle:         "Poison the Blade",
	}
	for oracle, name := range want {
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

// --- Poison the Blade ---

func TestGapB228PoisonTheBladeGivesDeathtouchAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedCheapLibrary(g, me, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Poison the Blade", "Instant", gap228PoisonOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "deathtouch") {
		t.Error("the Bear has no deathtouch")
	}
	// The spell went from hand to the stack and graveyard; one card drawn.
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d (the spell was put in hand then cast, a card was drawn)", got, hand+1)
	}
}

// --- Natural Reclamation ---

func TestGapB228NaturalReclamationDestroysAnArtifactOrEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	effectstest.CastCatalogSpell(t, g, "Natural Reclamation", "Instant", gap228ReclamationOracl, effectstest.CardTarget(rock))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the Rock survived")
	}
	spec, _ := effects.Lookup(gap228ReclamationOracl)
	cascades := false
	for _, tr := range spec.Triggered {
		if tr.Keyword == effects.KeywordCascade {
			cascades = true
		}
	}
	if !cascades {
		t.Error("Natural Reclamation has no cascade trigger")
	}
}

// --- Astarion's Thirst ---

func TestGapB228AstarionsThirstExilesAndPutsItsPowerInCountersOnACommander(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	commander := effectstest.B21Commander(g, me.ID, "My Commander", "Legendary Creature — Human", 3, 3)
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Victim", 4, 4)
	effectstest.CastCatalogSpell(t, g, "Astarion's Thirst", "Instant", gap228ThirstOracle, effectstest.CardTarget(victim))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, victim) != game.ZoneExile {
		t.Errorf("the victim is in %v, want exile", effectstest.B12ZoneOf(g, victim))
	}
	if n := effectstest.CounterOn(g, commander, game.CounterPlusOne); n != 4 {
		t.Errorf("the commander has %d counters, want 4", n)
	}
}

// --- Jaheira's Respite ---

func TestGapB228JaheirasRespiteFetchesALandPerAttackerAndPreventsCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 3; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	opp := g.Seats[g.Turn.ActiveSeat]
	a := effectstest.B12Creature(g, opp.ID, "Attacker A", "Creature — Bear", 3, 3)
	b := effectstest.B12Creature(g, opp.ID, "Attacker B", "Creature — Bear", 3, 3)
	effectstest.AttackWith(t, g, me.ID, a, b)
	spell := effectstest.HandSpell(me, "Jaheira's Respite", "Instant", "{0}")
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == spell {
				me.Hand.Cards[i].OracleID = gap228RespiteOracle
			}
		}
	})
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Forest" {
			n++
			if !c.Tapped {
				t.Error("a fetched Forest entered untapped")
			}
		}
	}
	if n != 2 {
		t.Errorf("%d Forests fetched, want 2 (two creatures attack me)", n)
	}
	before := me.Life
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before {
		t.Errorf("I lost %d life: combat damage was not prevented", before-me.Life)
	}
}

// --- Banishment ---

func TestGapB228BanishmentExilesEveryOpposingPermanentWithTheSameNameUntilItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Permanent(g, opp.ID, "Twin Rock", "Artifact")
	b := effectstest.B12Permanent(g, g.Seats[2].ID, "Twin Rock", "Artifact")
	mine := effectstest.B12Permanent(g, me.ID, "Twin Rock", "Artifact")
	other := effectstest.B12Permanent(g, opp.ID, "Other Rock", "Artifact")
	effectstest.B12PlayFromHand(t, g, "Banishment", "Enchantment", gap228BanishmentOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.HobAnswerTargets(t, g, me.ID, a)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if effectstest.B12ZoneOf(g, id) != game.ZoneExile {
			t.Errorf("an opposing Twin Rock is in %v, want exile", effectstest.B12ZoneOf(g, id))
		}
	}
	if !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(other) {
		t.Error("my own Twin Rock or an opponent's Other Rock was exiled")
	}
	ban := effectstest.BattlefieldIDNamed(g, me.ID, "Banishment")
	effectstest.KillCreature(t, g, me.ID, ban)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountNamed(g, "Twin Rock"); n != 3 {
		t.Errorf("%d Twin Rocks after Banishment left, want 3 (both came back)", n)
	}
}

// --- Stunning Strike ---

func TestGapB228StunningStrikeTapsAndHoldsANonlegendaryCreature(t *testing.T) {
	for _, legendary := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		tl := "Creature — Bear"
		if legendary {
			tl = "Legendary Creature — Bear"
		}
		bear := effectstest.B12Creature(g, opp.ID, "Their Bear", tl, 2, 2)
		effectstest.AdvanceToMain(t, g)
		aura := effectstest.HandSpell(me, "Stunning Strike", "Enchantment — Aura", "{0}")
		g.WithWriteLock(func() {
			for i := range me.Hand.Cards {
				if me.Hand.Cards[i].InstanceID == aura {
					me.Hand.Cards[i].OracleID = gap228StrikeOracle
				}
			}
		})
		if err := g.CastSpell(me.ID, aura, game.CastSpellParams{Targets: effectstest.CardTarget(bear)}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.IsTapped(g, bear) {
			t.Errorf("legendary=%v: the creature was not tapped", legendary)
		}
		// Pass the turn around to the opponent's next untap step.
		for i := 0; i < 6 && g.Turn.ActiveSeat != 1; i++ {
			effectstest.AdvanceToNextSeatsTurn(t, g)
		}
		effectstest.AdvanceTo(t, g, game.StepUpkeep)
		stillTapped := effectstest.IsTapped(g, bear)
		if stillTapped == legendary {
			t.Errorf("legendary=%v: tapped after the untap step = %v", legendary, stillTapped)
		}
	}
}

// --- Undercellar Sweep ---

func TestGapB228UndercellarSweepMakesAttackingSoldiersOnlyWithTheInitiative(t *testing.T) {
	for _, initiative := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.PushCatalogPermanent(g, me.ID, "Undercellar Sweep", "Enchantment", gap228SweepOracle, false)
		attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
		if initiative {
			if err := g.SetInitiative(me.ID); err != nil {
				t.Fatalf("SetInitiative: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.TakeSearchPrompts(t, g)
		}
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		effectstest.LockInAttacks(t, g)
		effectstest.PassAndSettleAttackSplit(t, g, me.ID)
		n := effectstest.CountOnBattlefield(g, "Soldier", me.ID)
		want := 0
		if initiative {
			want = 2
		}
		if n != want {
			t.Errorf("initiative=%v: %d Soldiers, want %d", initiative, n, want)
		}
	}
}

// --- Ghost Lantern // Bind Spirit ---

func TestGapB228GhostLanternGrowsTheEquippedCreatureWhenYoursDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	lantern := effectstest.PushCatalogPermanent(g, me.ID, "Ghost Lantern", "Artifact — Equipment", gap228LanternOracle, false)
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	victim := effectstest.B12Creature(g, me.ID, "Victim", "Creature — Bear", 1, 1)
	theirs := effectstest.PushVanillaCreature(g, g.Seats[1].ID, "Theirs", 1, 1)
	effectstest.B06AddMana(me, "B")
	effectstest.EquipTo(t, g, me.ID, lantern, host)
	effectstest.KillCreature(t, g, g.Seats[1].ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterOn(g, host, game.CounterPlusOne); n != 0 {
		t.Errorf("an opponent's creature dying gave %d counters", n)
	}
	effectstest.KillCreature(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterOn(g, host, game.CounterPlusOne); n != 1 {
		t.Errorf("the host has %d counters, want 1", n)
	}
}

func TestGapB228BindSpiritReturnsACreatureCardToYourHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dead := effectstest.PushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	effectstest.CastCatalogSpell(t, g, "Bind Spirit", "Instant — Adventure", gap228LanternOracle+"#1", effectstest.CardTarget(dead))
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(dead) {
		t.Error("the creature card did not return to hand")
	}
}

// --- Backgrounds ---

func TestGapB228FarTravelerBlinksATappedCreatureAtYourEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B21Push(g, me.ID, "Far Traveler", gap228Background, gap228TravelerOracle, 0, 0)
	effectstest.B21Commander(g, me.ID, "My Commander", "Legendary Creature — Human", 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Tapped Bear", "Creature — Bear", 2, 2)
	untapped := effectstest.B12Creature(g, me.ID, "Fresh Bear", "Creature — Bear", 2, 2)
	effectstest.AvB4SetTapped(g, bear)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt at the end step")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, effectstest.CardRefs(untapped)); err == nil {
		t.Fatal("an untapped creature was a legal target")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, effectstest.CardRefs(bear)); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear kept its identity: it was not blinked")
	}
	if n := effectstest.CountOnBattlefield(g, "Tapped Bear", me.ID); n != 1 {
		t.Errorf("%d Tapped Bears, want 1", n)
	}
}

func TestGapB228VeteranSoldierMakesASoldierForEachOpponentTappedAndAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B21Push(g, me.ID, "Veteran Soldier", gap228Background, gap228SoldierOracle, 0, 0)
	commander := effectstest.B21Commander(g, me.ID, "My Commander", "Legendary Creature — Human", 3, 3)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(commander, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Soldier" {
			n++
			if !c.Tapped || c.AttackingTarget == uuid.Nil {
				t.Errorf("a Soldier is tapped=%v attacking=%v", c.Tapped, c.AttackingTarget != uuid.Nil)
			}
		}
	}
	if n != 3 {
		t.Errorf("%d Soldiers, want 3 (one for each of three opponents)", n)
	}
}
