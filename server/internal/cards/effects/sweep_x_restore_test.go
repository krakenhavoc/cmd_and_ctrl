package effects

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sweep_x_restore_test.go — #4253, the final seams (part X): each moved
// card paused on a prompt, restored STRICTLY (the boot path) and answered on
// the restored table:
//
//   - the damage continuations as data: a keyed Each walk
//     (Game.DealDamageEachKeyedForEffect, the walk riding each leg's damage
//     tail) under Ranger's Firebrand's tempt and Creeping Bloodsucker's
//     shared "gain life equal to the damage dealt", and the keyed creature
//     door (Game.DealDamageToCreatureKeyedForEffect) under Ballista
//     Wielder's "can't block";
//   - the keyed proliferate told what got a counter
//     (Game.ProliferateChoosingThenKeyedForEffect): Ripples of Potential;
//   - PutFromLibraryOntoBattlefield's pick with its set rule registered
//     (RegisterPutRule) and its counters as data (PutCounters): Ojer Kaslem
//     and Turntimber Symbiosis.

// damageOrderProbeOracles are two test cards, each "if a source would deal
// damage to you or a permanent you control, it deals that much instead":
// two declared effects, so such damage asks its controller for their order
// (CR 616.1). Catalog rows, so a restored table gathers them again.
var damageOrderProbeOracles = []string{"test-4253-x-damage-order-a", "test-4253-x-damage-order-b"}

func init() {
	for i, oracle := range damageOrderProbeOracles {
		Register(Spec{
			OracleID: oracle,
			Name:     "Damage Order Probe X" + string(rune('A'+i)),
			Replacements: []game.ReplacementEffect{{
				Watches: []game.EventKind{game.EventDealDamage},
				AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, self *game.Card) bool {
					if ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || ev.IsCombatDamage || self == nil {
						return false
					}
					if ev.DamageTarget == self.Controller {
						return true
					}
					c, ok := g.LookupCardForEffect(ev.DamageTarget)
					return ok && c.Controller == self.Controller && c.InstanceID != self.InstanceID
				},
				Replace: func(*game.ReplacementEvent, *game.Game, *game.Card) error { return nil },
				Label:   "Damage Order Probe X" + string(rune('A'+i)),
			}},
		})
	}
}

// seedDamageOrderProbes gives `who` the two probes.
func seedDamageOrderProbes(g *game.Game, who uuid.UUID) {
	seedOrderProbes(g, who, damageOrderProbeOracles, "Damage Order Probe X")
}

// --- Ranger's Firebrand ------------------------------------------------------

// Ranger's Firebrand at an opponent with the probes: the damage asks for the
// order, the walk's damage frame is written, and the restored answer deals
// the 2 and then the Ring tempts the caster (the item rode the keyed tail).
func TestARangersFirebrandDamageOrderIsARestorePointAndStillTempts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedDamageOrderProbes(g, opp.ID)
	life := opp.Life
	castCatalogSpell(t, g, "Ranger's Firebrand", "Sorcery", rangersFirebrandOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if opp.Life != life || ringCount(g, me.ID) != 0 {
		t.Fatalf("the damage or the tempt happened before the ordering prompt was answered")
	}

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeDamage)
	answerExileOrder(t, g)
	ringSettle(t, g)
	if got := lifeTotalOf(g, opp.ID); got != life-2 {
		t.Errorf("opponent at %d after the restored answer, want %d", got, life-2)
	}
	if ringCount(g, me.ID) != 1 {
		t.Error("the Ring did not tempt after the restored answer: the keyed tail did not run off the disk")
	}
}

// --- Creeping Bloodsucker ----------------------------------------------------

// Creeping Bloodsucker's upkeep drain with one opponent holding the probes:
// that opponent's damage asks for the order, the restored table still owes
// the rest of the walk, and the answer deals every opponent 1 and gains its
// controller the total, the shared tail off the disk.
func TestACreepingBloodsuckerDrainIsARestorePointMidWalk(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Creeping Bloodsucker", "Creature — Vampire", b14CreepingBloodsuckerOracle, 1, 2)
	seedDamageOrderProbes(g, g.Seats[1].ID)
	mine := me.Life
	theirs := lifeOfOpponents(g)
	b12ToMyNextUpkeep(t, g)
	passPriorityAroundTable(t, g)
	if lifeOfOpponents(g)[0] != theirs[0] || me.Life != mine {
		t.Fatal("the probed opponent's damage, or the gain, landed before the ordering prompt was answered")
	}

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeDamage)
	answerExileOrder(t, g)
	passPriorityAroundTable(t, g)
	for i, before := range theirs {
		if got := lifeOfOpponents(g)[i]; got != before-1 {
			t.Errorf("opponent %d: %d → %d, want -1", i, before, got)
		}
	}
	if got := lifeTotalOf(g, me.ID); got != mine+3 {
		t.Errorf("three opponents hit: life %d → %d, want +3 (the walk's total off the disk)", mine, got)
	}
}

// --- Ballista Wielder --------------------------------------------------------

// Ballista Wielder pings a creature whose controller holds the probes: the
// ping asks for the order, the restored answer deals the 1, and the rider
// (the keyed creature door's tail, told the creature and the damage) stops
// it blocking.
func TestABallistaWielderPingIsARestorePointAndStillStopsTheBlock(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	wielder := b12Push(g, me.ID, "Ballista Wielder", "Creature — Werewolf", dnaBallista+"#1", 5, 5)
	attacker := b12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	blocker := b12Creature(g, opp.ID, "Blocker", "Creature — Bear", 3, 3)
	seedDamageOrderProbes(g, opp.ID)
	floatMana(t, g, me, "{R}{R}{R}")
	if err := g.ActivateCatalogAbility(me.ID, wielder, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: blocker}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if damageMarkedOn(g, blocker) != 0 {
		t.Fatal("the ping landed before the ordering prompt was answered")
	}

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeDamage)
	answerExileOrder(t, g)
	passPriorityAroundTable(t, g)
	if got := damageMarkedOn(g, blocker); got != 1 {
		t.Fatalf("the blocker has %d damage after the restored answer, want 1", got)
	}
	declareAttack(t, g, opp.ID, attacker)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err == nil {
		t.Error("the creature the Wielder pinged blocked: the keyed rider did not run off the disk")
	}
}

// --- Ripples of Potential ----------------------------------------------------

// Ripples of Potential waiting on its proliferate: the restored answer
// proliferates, and the phase-out pick offers exactly what got a counter
// (the event the proliferate started after rode the keyed tail).
func TestARipplesOfPotentialProliferateIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := seedCreatureWithCounters(g, me.ID, map[string]int{game.CounterPlusOne: 1})
	b := seedCreatureWithCounters(g, me.ID, map[string]int{game.CounterPlusOne: 1})
	castRipples(t, g)

	g = restoreRoundTrip(t, g, true)
	answerProliferate(t, g, a)
	c := ripplesChoosePrompt(g)
	if c == nil {
		t.Fatal("no phase-out pick after the restored proliferate")
	}
	if !slices.Equal(c.ChooseCards, []uuid.UUID{a}) {
		t.Fatalf("offered %v, want only %v, the one that got a counter", c.ChooseCards, a)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{a}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !phasedOutInCatalogGame(g, a) || phasedOutInCatalogGame(g, b) {
		t.Errorf("phased out: a=%v b=%v, want only a", phasedOutInCatalogGame(g, a), phasedOutInCatalogGame(g, b))
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
			g, me, top := znrTurntimberBoard(t)
			g = restoreRoundTrip(t, g, true)
			prompt := latestChooseCards(g, me.ID)
			if prompt == nil {
				t.Fatal("no choose-cards prompt on the restored table")
			}
			if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{top[tc.pick]}); err != nil {
				t.Fatalf("ResolveChooseCards: %v", err)
			}
			c, ok := znrBattlefieldCard(g, top[tc.pick])
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	god := importToBattlefield(t, g, ojerKaslemRow(), me)
	var ids []uuid.UUID
	for _, tl := range []string{"Basic Land — Forest", "Creature — Ox", "Creature — Elf"} {
		id := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: id, Name: tl, TypeLine: tl, Owner: me.ID, Controller: me.ID})
		ids = append(ids, id)
	}
	forest, ox, elf := ids[0], ids[1], ids[2]
	dealCombatDamageToPlayer(g, god, opp.ID, 3)
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) == nil {
		t.Fatal("no pick prompt after the reveal")
	}

	g = restoreRoundTrip(t, g, true)
	c := chooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no pick prompt on the restored table")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{elf, ox}); !errors.Is(err, game.ErrChoiceSetRejected) {
		t.Fatalf("two creatures on the restored prompt: err %v, want ErrChoiceSetRejected", err)
	}
	answerChooseCards(t, g, me.ID, elf, forest)
	if !g.Battlefield.Contains(elf) || !g.Battlefield.Contains(forest) || g.Battlefield.Contains(ox) {
		t.Error("the restored answer did not put exactly the creature and the land onto the battlefield")
	}
}
