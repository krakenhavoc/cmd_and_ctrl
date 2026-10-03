package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// adr0108_pr8_shields_test.go — ADR 0108 PR 8, owner decision 2 (#1906):
// the scoped "prevent the next N damage" and "prevent all combat damage"
// shields whose CR 615.5 additional effect does something with what they
// prevented. Each test proves the shield, the follow-up's amount, and
// CR 615.12: damage that can't be prevented gets through, makes the
// follow-up do nothing, and leaves a charge whole.

const (
	pr8Temper         = "9cefff18-96b1-4090-93cc-1ca981a58fbd"
	pr8SacredBoon     = "32a8d49b-cdfb-4944-bfb5-afd704cc3658"
	pr8ScarsOfVeteran = "5f0ad09f-a9ee-42df-b32d-9d842ce5dd00"
	pr8AcolytesReward = "944689b2-f9f7-45c3-952c-54287b9824d5"
	pr8VengefulArchon = "c3ef4311-c7f1-45ae-8bd5-9c05bdf2ae88"
	pr8Inkshield      = "da93264e-4e04-401a-9809-e4e1056cf604"
)

// pr8CastX casts a catalogued spell from the active player's hand with X.
func pr8CastX(t *testing.T, g *game.Game, name, typeLine, oracle string, x int, targets []game.TargetRef) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	c := game.NewCard(name, active.ID)
	c.TypeLine, c.OracleID = typeLine, oracle
	active.Hand.PushTop(c)
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.CastSpell(active.ID, c.InstanceID, game.CastSpellParams{Targets: targets, XValue: x}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	passPriorityAroundTable(t, g)
}

func pr8Toughness(g *game.Game, id uuid.UUID) int {
	return pr8Counters(g, id, "+0/+1")
}

func pr8Card(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

// Temper with X = 5 facing 3: only 3 counters (the ruling), and the rest
// of the charge prevents the next 2.
func TestADR0108PR8ShieldsTemper(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := apaPush(g, me.ID, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	src := pr7Creature(g, opp.ID, "Src", 6, "R")
	pr8CastX(t, g, "Temper", "Instant", pr8Temper, 5, pr8Card(bear))
	pr8Unpreventable(t, g, src, bear, 1)
	if pr6Marked(g, bear) != 1 || pr8Counters(g, bear, game.CounterPlusOne) != 0 {
		t.Fatalf("unpreventable: damage %d (want 1), counters %d (want 0)", pr6Marked(g, bear), pr8Counters(g, bear, game.CounterPlusOne))
	}
	pr6Damage(t, g, src, bear, 3)
	if pr6Marked(g, bear) != 1 || pr8Counters(g, bear, game.CounterPlusOne) != 3 {
		t.Fatalf("3 of a 5 charge: damage %d (want 1), counters %d (want 3)", pr6Marked(g, bear), pr8Counters(g, bear, game.CounterPlusOne))
	}
	pr6Damage(t, g, src, bear, 4)
	if pr6Marked(g, bear) != 3 || pr8Counters(g, bear, game.CounterPlusOne) != 5 {
		t.Fatalf("the last 2 of the charge: damage %d (want 3), counters %d (want 5)", pr6Marked(g, bear), pr8Counters(g, bear, game.CounterPlusOne))
	}
}

// Sacred Boon: two preventions before the end step join ONE delayed
// trigger, which puts on the total at the next end step; unpreventable
// damage adds nothing.
func TestADR0108PR8ShieldsSacredBoon(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := apaPush(g, me.ID, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 5})
	src := pr7Creature(g, opp.ID, "Src", 6, "R")
	castCatalogSpell(t, g, "Sacred Boon", "Instant", pr8SacredBoon, pr8Card(bear))
	passPriorityAroundTable(t, g)
	pr8Unpreventable(t, g, src, bear, 1)
	pr6Damage(t, g, src, bear, 1)
	pr6Damage(t, g, src, bear, 1)
	if pr6Marked(g, bear) != 1 || len(g.DelayedTriggers) != 1 || g.DelayedTriggers[0].Params.Amount != 2 {
		var amounts []int
		for _, d := range g.DelayedTriggers {
			amounts = append(amounts, d.Params.Amount)
		}
		t.Fatalf("damage %d (want 1), delayed triggers %v (want one, of 2)", pr6Marked(g, bear), amounts)
	}
	if pr8Toughness(g, bear) != 0 {
		t.Fatal("the counters went on before the end step")
	}
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if pr8Toughness(g, bear) != 2 {
		t.Fatalf("end step: %d +0/+1 counters, want 2", pr8Toughness(g, bear))
	}
}

// Scars of the Veteran: on a player, the shield prevents and no trigger is
// made ("If it's a creature"); on a creature, 7 of 9 are prevented and 7
// counters arrive at the end step.
func TestADR0108PR8ShieldsScarsOfTheVeteran(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pr7Creature(g, opp.ID, "Src", 9, "R")
	life := me.Life
	castCatalogSpell(t, g, "Scars of the Veteran", "Instant", pr8ScarsOfVeteran, []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	passPriorityAroundTable(t, g)
	pr7Hit(t, g, src, me.ID, 5)
	if me.Life != life || len(g.DelayedTriggers) != 0 {
		t.Fatalf("player: life %d (want %d), %d delayed triggers (want 0)", me.Life, life, len(g.DelayedTriggers))
	}

	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	wall := apaPush(g, me.ID, me.ID, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 5})
	src = pr7Creature(g, opp.ID, "Src", 9, "R")
	castCatalogSpell(t, g, "Scars of the Veteran", "Instant", pr8ScarsOfVeteran, pr8Card(wall))
	passPriorityAroundTable(t, g)
	pr6Damage(t, g, src, wall, 9)
	if pr6Marked(g, wall) != 2 {
		t.Fatalf("wall damage %d, want 2", pr6Marked(g, wall))
	}
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if pr8Toughness(g, wall) != 7 {
		t.Fatalf("end step: %d +0/+1 counters, want 7", pr8Toughness(g, wall))
	}
}

// Acolyte's Reward: X is your devotion to white; the prevented amount is
// dealt to the second target; an excess gets through at once; damage that
// can't be prevented deals none.
func TestADR0108PR8ShieldsAcolytesReward(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	apaPush(g, me.ID, me.ID, game.Card{Name: "Devout", TypeLine: "Creature — Cleric", Power: 1, Toughness: 1, ManaCost: "{1}{W}{W}"})
	apaPush(g, me.ID, me.ID, game.Card{Name: "Pious", TypeLine: "Enchantment", ManaCost: "{W}"})
	bear := apaPush(g, me.ID, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 6})
	src := pr7Creature(g, opp.ID, "Src", 6, "R")
	life := opp.Life
	castCatalogSpell(t, g, "Acolyte's Reward", "Instant", pr8AcolytesReward, []game.TargetRef{
		{Kind: game.TargetCard, ID: bear, Slot: 0}, {Kind: game.TargetPlayer, ID: opp.ID, Slot: 1},
	})
	passPriorityAroundTable(t, g)
	pr8Unpreventable(t, g, src, bear, 1)
	if pr6Marked(g, bear) != 1 || opp.Life != life {
		t.Fatalf("unpreventable: bear damage %d (want 1), opponent at %d (want %d)", pr6Marked(g, bear), opp.Life, life)
	}
	pr6Damage(t, g, src, bear, 2)
	pr6Damage(t, g, src, bear, 2)
	if pr6Marked(g, bear) != 2 || opp.Life != life-3 {
		t.Fatalf("devotion 3: bear damage %d (want 2), opponent at %d (want %d)", pr6Marked(g, bear), opp.Life, life-3)
	}
}

// Acolyte's Reward with its second target gone by the time damage is
// prevented: the damage is still prevented, and nothing is dealt.
func TestADR0108PR8ShieldsAcolytesRewardSecondTargetGone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	apaPush(g, me.ID, me.ID, game.Card{Name: "Pious", TypeLine: "Enchantment", ManaCost: "{W}{W}"})
	bear := apaPush(g, me.ID, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 6})
	victim := pr7Creature(g, opp.ID, "Victim", 1, "B")
	src := pr7Creature(g, opp.ID, "Src", 6, "R")
	castCatalogSpell(t, g, "Acolyte's Reward", "Instant", pr8AcolytesReward, []game.TargetRef{
		{Kind: game.TargetCard, ID: bear, Slot: 0}, {Kind: game.TargetCard, ID: victim, Slot: 1},
	})
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(victim); err != nil {
			t.Fatal(err)
		}
	})
	pr6Damage(t, g, src, bear, 2)
	if pr6Marked(g, bear) != 0 {
		t.Fatalf("bear damage %d, want 0: still prevented", pr6Marked(g, bear))
	}
}

// Vengeful Archon: {X} shields you for X; the Archon deals what it
// prevented to the target player; damage that can't be prevented deals
// none and leaves the charge.
func TestADR0108PR8ShieldsVengefulArchon(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	archon := pushCatalogPermanent(g, me.ID, "Vengeful Archon", "Creature — Archon", pr8VengefulArchon, false)
	src := pr7Creature(g, opp.ID, "Src", 6, "R")
	pr7Activate(t, g, me.ID, archon, 0, game.ActivateAbilityParams{XValue: 4, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	life, oppLife := me.Life, opp.Life
	pr8Unpreventable(t, g, src, me.ID, 2)
	if me.Life != life-2 || opp.Life != oppLife {
		t.Fatalf("unpreventable: me %d (want %d), opponent %d (want %d)", me.Life, life-2, opp.Life, oppLife)
	}
	pr7Hit(t, g, src, me.ID, 3)
	pr7Hit(t, g, src, me.ID, 3)
	if me.Life != life-4 || opp.Life != oppLife-4 {
		t.Fatalf("charge of 4: me %d (want %d), opponent %d (want %d)", me.Life, life-4, opp.Life, oppLife-4)
	}
}

// Inkshield: two attackers' combat damage in one step is one instance:
// one batch of Inklings, one per point prevented; noncombat damage is not
// prevented.
func TestADR0108PR8ShieldsInkshield(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pr7Creature(g, opp.ID, "A", 2, "R")
	b := pr7Creature(g, opp.ID, "B", 3, "G")
	advanceToDeclareAttackersOf(t, g, 1)
	ink := game.NewCard("Inkshield", me.ID)
	ink.TypeLine, ink.OracleID = "Instant", pr8Inkshield
	me.Hand.PushTop(ink)
	if err := g.CastSpell(me.ID, ink.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Inkshield: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, attacker := range []uuid.UUID{a, b} {
		if err := g.DeclareAttacker(attacker, me.ID); err != nil {
			t.Fatal(err)
		}
	}
	life := me.Life
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	inklings := len(battlefieldIDsNamed(g, "Inkling"))
	if me.Life != life || inklings != 5 {
		t.Fatalf("combat: life %d (want %d), %d Inklings (want 5)", me.Life, life, inklings)
	}
	pr7Hit(t, g, a, me.ID, 1)
	if me.Life != life-1 || len(battlefieldIDsNamed(g, "Inkling")) != 5 {
		t.Fatalf("noncombat: life %d (want %d), %d Inklings (want 5)", me.Life, life-1, len(battlefieldIDsNamed(g, "Inkling")))
	}
}
