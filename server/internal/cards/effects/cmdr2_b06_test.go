package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pirates batch (#5368, cmdr2 batch 6).

const (
	cmdr2B06BeckettOracle   = "43b5e462-d860-473d-828f-6c513fc7768a"
	cmdr2B06RavagerOracle   = "b83080af-9616-4dc3-bc47-6ef84b9ccc1b"
	cmdr2B06RamirezOracle   = "a3b8690e-ac47-444f-924e-ab2a12a85431"
	cmdr2B06BrassOracle     = "4df179b4-7c6b-4e63-9d7d-62f4df377989"
	cmdr2B06FranciscoOracle = "1dbe4eb4-8313-459a-89ce-982442cd0cd0"
	cmdr2B06SaboteurOracle  = "d36f64a0-53f7-4c56-9f40-a7927a5c2608"
	cmdr2B06CaptainOracle   = "7bcf3090-95c6-4384-a857-9fe571e6da85"
	cmdr2B06WarkiteOracle   = "8b566757-e5e4-4d47-b5ef-80ed7d853679"
	cmdr2B06SkeletonOracle  = "2536750c-3c4f-4e00-b517-680ad4fd2ecc"
	cmdr2B06RaidersOracle   = "8b388331-eaa7-41aa-99de-4451e51cb221"
)

func TestCmdr2B06Registered(t *testing.T) {
	for oracle, name := range map[string]string{
		cmdr2B06RavagerOracle:  "Dire Fleet Ravager",
		cmdr2B06SaboteurOracle: "Daring Saboteur",
		cmdr2B06WarkiteOracle:  "Warkite Marauder",
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 || spec.Name != name {
			t.Errorf("%s: %+v %v", name, spec, ok)
		}
	}
}

// --- Daring Saboteur ------------------------------------------------

func TestCmdr2B06SaboteurPaysToBecomeUnblockable(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sab := cmdr2B05Card(g, me.ID, "Daring Saboteur", "Creature — Human Pirate", cmdr2B06SaboteurOracle, 2, 1)
	advanceToMain(t, g)
	assertRestrictions(t, g, sab, 0)
	if err := g.ActivateCatalogAbility(me.ID, sab, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no mana")
	}
	floatForTest(g, me, "CCU")
	b16Activate(t, g, me.ID, sab, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	assertRestrictions(t, g, sab, game.CantBeBlocked)
}

func TestCmdr2B06SaboteurLootsOnCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	msc18Library(me, 8)
	sab := cmdr2B05Card(g, me.ID, "Daring Saboteur", "Creature — Human Pirate", cmdr2B06SaboteurOracle, 2, 1)
	dealCombatDamageToPlayer(g, sab, opp.ID, 2)
	passPriorityAroundTable(t, g)
	ask := latestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil || ask.Chooser != me.ID {
		t.Fatal("no loot prompt")
	}
	hand := me.Hand.Size()
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d, want %d after the draw", me.Hand.Size(), hand+1)
	}
	discardFromHand(t, g, me.ID)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d after the discard, want %d", me.Hand.Size(), hand)
	}
}

// --- Dire Fleet Ravager -----------------------------------------------

func TestCmdr2B06RavagerEveryPlayerLosesAThirdRoundedUp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lives := []int{20, 10, 1, 7, 40}
	for i, p := range g.Seats {
		if i < len(lives) {
			p.Life = lives[i]
		}
	}
	want := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		want[i] = p.Life - (p.Life+2)/3
	}
	castCatalogSpell(t, g, "Dire Fleet Ravager", "Creature — Orc Pirate Wizard", cmdr2B06RavagerOracle, nil)
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if p.Life != want[i] {
			t.Errorf("seat %d (%s) has %d life, want %d", i, map[bool]string{true: "controller", false: "opponent"}[p.ID == me.ID], p.Life, want[i])
		}
	}
	if g.Seats[0].Life != 13 || g.Seats[1].Life != 6 || g.Seats[2].Life != 0 {
		t.Errorf("20→13, 10→6 and 1→0 expected, got %d %d %d", g.Seats[0].Life, g.Seats[1].Life, g.Seats[2].Life)
	}
}

// --- Warkite Marauder -------------------------------------------------

func TestCmdr2B06WarkiteStripsAbilitiesAndSetsBasePT(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wm := cmdr2B05Card(g, me.ID, "Warkite Marauder", "Creature — Human Pirate", cmdr2B06WarkiteOracle, 2, 1)
	victim := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Big Flyer", TypeLine: "Creature — Dragon", Keywords: []string{"flying"},
		Power: 5, Toughness: 5, Owner: opp.ID, Controller: opp.ID,
	})
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(victim, game.CounterPlusOne, 1) })
	declareAttack(t, g, opp.ID, wm)
	if p := latestPickTarget(g, me.ID); p != nil {
		b40PickCard(t, g, me.ID, victim)
	}
	passPriorityAroundTable(t, g)
	if got := b39CurrentPower(t, g, victim); got != 1 {
		t.Errorf("victim power %d, want 1 (base 0 plus its counter)", got)
	}
	if got := effectiveToughness(t, g, victim); got != 1 {
		t.Errorf("victim toughness %d, want base 1", got)
	}
	if hasKeywordOf(g, victim, "flying") {
		t.Error("the victim kept flying")
	}
	if got := effectivePower(t, g, mine); got != 2 {
		t.Errorf("an unrelated creature changed: %d", got)
	}
	// Until end of turn.
	b39NextTurnOf(t, g, 1)
	if got := effectivePower(t, g, victim); got != 5 {
		t.Errorf("the effect outlived the turn: power %d", got)
	}
}
