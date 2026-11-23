package testsst

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Ranger's Firebrand ------------------------------------------------------

// Ranger's Firebrand at an opponent with the probes: the damage asks for the
// order, the walk's damage frame is written, and the restored answer deals
// the 2 and then the Ring tempts the caster (the item rode the keyed tail).
func TestARangersFirebrandDamageOrderIsARestorePointAndStillTempts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedDamageOrderProbes(g, opp.ID)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Ranger's Firebrand", "Sorcery", effectstest.RangersFirebrandOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life || effectstest.RingCount(g, me.ID) != 0 {
		t.Fatalf("the damage or the tempt happened before the ordering prompt was answered")
	}

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeDamage)
	effectstest.AnswerExileOrder(t, g)
	effectstest.RingSettle(t, g)
	if got := effectstest.LifeTotalOf(g, opp.ID); got != life-2 {
		t.Errorf("opponent at %d after the restored answer, want %d", got, life-2)
	}
	if effectstest.RingCount(g, me.ID) != 1 {
		t.Error("the Ring did not tempt after the restored answer: the keyed tail did not run off the disk")
	}
}

// --- Creeping Bloodsucker ----------------------------------------------------

// Creeping Bloodsucker's upkeep drain with one opponent holding the probes:
// that opponent's damage asks for the order, the restored table still owes
// the rest of the walk, and the answer deals every opponent 1 and gains its
// controller the total, the shared tail off the disk.
func TestACreepingBloodsuckerDrainIsARestorePointMidWalk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Creeping Bloodsucker", "Creature — Vampire", effectstest.B14CreepingBloodsuckerOracle, 1, 2)
	effectstest.SeedDamageOrderProbes(g, g.Seats[1].ID)
	mine := me.Life
	theirs := effectstest.LifeOfOpponents(g)
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LifeOfOpponents(g)[0] != theirs[0] || me.Life != mine {
		t.Fatal("the probed opponent's damage, or the gain, landed before the ordering prompt was answered")
	}

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeDamage)
	effectstest.AnswerExileOrder(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	for i, before := range theirs {
		if got := effectstest.LifeOfOpponents(g)[i]; got != before-1 {
			t.Errorf("opponent %d: %d → %d, want -1", i, before, got)
		}
	}
	if got := effectstest.LifeTotalOf(g, me.ID); got != mine+3 {
		t.Errorf("three opponents hit: life %d → %d, want +3 (the walk's total off the disk)", mine, got)
	}
}

// --- Ballista Wielder --------------------------------------------------------

// Ballista Wielder pings a creature whose controller holds the probes: the
// ping asks for the order, the restored answer deals the 1, and the rider
// (the keyed creature door's tail, told the creature and the damage) stops
// it blocking.
func TestABallistaWielderPingIsARestorePointAndStillStopsTheBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	wielder := effectstest.B12Push(g, me.ID, "Ballista Wielder", "Creature — Werewolf", effectstest.DnaBallista+"#1", 5, 5)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 3, 3)
	effectstest.SeedDamageOrderProbes(g, opp.ID)
	effectstest.FloatMana(t, g, me, "{R}{R}{R}")
	if err := g.ActivateCatalogAbility(me.ID, wielder, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: blocker}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DamageMarkedOn(g, blocker) != 0 {
		t.Fatal("the ping landed before the ordering prompt was answered")
	}

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeDamage)
	effectstest.AnswerExileOrder(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, blocker); got != 1 {
		t.Fatalf("the blocker has %d damage after the restored answer, want 1", got)
	}
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err == nil {
		t.Error("the creature the Wielder pinged blocked: the keyed rider did not run off the disk")
	}
}

// --- Ripples of Potential ----------------------------------------------------

// Ripples of Potential waiting on its proliferate: the restored answer
// proliferates, and the phase-out pick offers exactly what got a counter
// (the event the proliferate started after rode the keyed tail).
func TestARipplesOfPotentialProliferateIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := effectstest.SeedCreatureWithCounters(g, me.ID, map[string]int{game.CounterPlusOne: 1})
	b := effectstest.SeedCreatureWithCounters(g, me.ID, map[string]int{game.CounterPlusOne: 1})
	effectstest.CastRipples(t, g)

	g = effectstest.RestoreRoundTrip(t, g, true)
	effectstest.AnswerProliferate(t, g, a)
	c := effectstest.RipplesChoosePrompt(g)
	if c == nil {
		t.Fatal("no phase-out pick after the restored proliferate")
	}
	if !slices.Equal(c.ChooseCards, []uuid.UUID{a}) {
		t.Fatalf("offered %v, want only %v, the one that got a counter", c.ChooseCards, a)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{a}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !effectstest.PhasedOutInCatalogGame(g, a) || effectstest.PhasedOutInCatalogGame(g, b) {
		t.Errorf("phased out: a=%v b=%v, want only a", effectstest.PhasedOutInCatalogGame(g, a), effectstest.PhasedOutInCatalogGame(g, b))
	}
}

// --- Turntimber Symbiosis ----------------------------------------------------

// Turntimber Symbiosis waiting on its pick: the gated counters ride the
// prompt as data, so the restored answer still puts a mana value 3 creature
// in with three +1/+1 counters, and a mana value 5 one with none.
func TestATurntimberSymbiosisPickIsARestorePointAndKeepsItsCounters(t *testing.T) {
	for _, tc := range []struct {
		name string
		pick int
		want int
	}{{"mana value 3", 4, 3}, {"mana value 5", 2, 0}} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, top := effectstest.ZnrTurntimberBoard(t)
			g = effectstest.RestoreRoundTrip(t, g, true)
			prompt := effectstest.LatestChooseCards(g, me.ID)
			if prompt == nil {
				t.Fatal("no choose-cards prompt on the restored table")
			}
			if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{top[tc.pick]}); err != nil {
				t.Fatalf("ResolveChooseCards: %v", err)
			}
			c, ok := effectstest.ZnrBattlefieldCard(g, top[tc.pick])
			if !ok {
				t.Fatal("the chosen creature is not on the battlefield")
			}
			if got := c.Counters[game.CounterPlusOne]; got != tc.want {
				t.Errorf("entered with %d +1/+1 counters, want %d", got, tc.want)
			}
		})
	}
}

// --- Ojer Kaslem -------------------------------------------------------------

// Ojer Kaslem's reveal waiting on its pick: the set rule is the registered
// rule, so the restored prompt still refuses two creatures and takes a
// creature and a land.
func TestAnOjerKaslemPickIsARestorePointAndKeepsItsRule(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	god := effectstest.ImportToBattlefield(t, g, effectstest.OjerKaslemRow(), me)
	var ids []uuid.UUID
	for _, tl := range []string{"Basic Land — Forest", "Creature — Ox", "Creature — Elf"} {
		id := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: id, Name: tl, TypeLine: tl, Owner: me.ID, Controller: me.ID})
		ids = append(ids, id)
	}
	forest, ox, elf := ids[0], ids[1], ids[2]
	effectstest.DealCombatDamageToPlayer(g, god, opp.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) == nil {
		t.Fatal("no pick prompt after the reveal")
	}

	g = effectstest.RestoreRoundTrip(t, g, true)
	c := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no pick prompt on the restored table")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{elf, ox}); !errors.Is(err, game.ErrChoiceSetRejected) {
		t.Fatalf("two creatures on the restored prompt: err %v, want ErrChoiceSetRejected", err)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, elf, forest)
	if !g.Battlefield.Contains(elf) || !g.Battlefield.Contains(forest) || g.Battlefield.Contains(ox) {
		t.Error("the restored answer did not put exactly the creature and the land onto the battlefield")
	}
}
