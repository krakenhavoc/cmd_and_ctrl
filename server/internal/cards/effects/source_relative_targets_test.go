package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// source_relative_targets_test.go — #2146. A target described relative
// to its source ("with lesser power") is judged against the source's
// effective power at announce (the offered set and CR 601.2c) and again
// at resolution (CR 608.2b), and against its last-known information
// once the source has left (CR 608.2h).

const (
	rangersOfIthilienOracle  = "4d5fd9f0-ecb3-4b91-ae55-f02336d7cf37"
	hammerDropperOracle      = "088c9fa7-65ec-48c6-90cc-9087bc9df43e"
	bladeInstructorOracle    = "77cff683-013c-4568-8ae0-f86ef0590fe1"
	bargingSergeantOracle    = "af422a4b-087c-4753-9288-c31fb3457740"
	unlivingPsychopathOracle = "038463c8-d2e1-4e7b-829d-67a744ae2660"
)

// ithilienPrompt casts a 3/3 Rangers of Ithilien, resolves the spell,
// and stops at the trigger's target prompt.
func ithilienPrompt(t *testing.T, g *game.Game) (rangers uuid.UUID, prompt *game.PendingChoice) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	rangers = lftCastSizedCreature(t, g, "Rangers of Ithilien", "Creature — Human Ranger", rangersOfIthilienOracle, 3, nil)
	passPriorityAroundTable(t, g)
	prompt = latestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("the Rangers' trigger asked for no target")
	}
	return rangers, prompt
}

func TestRangersOfIthilienOffersOnlyLesserPower(t *testing.T) {
	g := newCatalogGame(t)
	victim := g.Seats[1]
	p1 := lftSeed(g, victim.ID, "One", "Creature — Bear", 1, 1)
	p2 := lftSeed(g, victim.ID, "Two", "Creature — Bear", 2, 2)
	p3 := lftSeed(g, victim.ID, "Three", "Creature — Bear", 3, 3)
	p4 := lftSeed(g, victim.ID, "Four", "Creature — Bear", 4, 4)
	_, prompt := ithilienPrompt(t, g)

	offered := prompt.PickTargetCards
	if !hasID(offered, p1) || !hasID(offered, p2) {
		t.Errorf("offered %v, want the 1- and 2-power creatures", offered)
	}
	if hasID(offered, p3) || hasID(offered, p4) {
		t.Errorf("offered %v, which includes a creature of power 3 or more", offered)
	}
}

func TestRangersOfIthilienRefusesAnEqualPowerPickAtAnnounce(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	p3 := lftSeed(g, victim.ID, "Three", "Creature — Bear", 3, 3)
	lftSeed(g, victim.ID, "One", "Creature — Bear", 1, 1) // so the trigger has a target to ask for
	_, prompt := ithilienPrompt(t, g)

	err := g.ResolvePickTarget(prompt.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: p3})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("picking an equal-power creature = %v, want ErrIllegalTarget (CR 601.2c)", err)
	}
}

func TestRangersOfIthilienStealsALesserCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	p2 := lftSeed(g, victim.ID, "Two", "Creature — Bear", 2, 2)
	_, _ = ithilienPrompt(t, g)
	pickCard(t, g, me.ID, p2)
	passPriorityAroundTable(t, g)
	if got := findBattlefieldCardByID(g, p2).Controller; got != me.ID {
		t.Errorf("controller = %s, want the Rangers' controller", got)
	}
}

func TestRangersOfIthilienShrunkInResponseFizzlesTheSteal(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	p2 := lftSeed(g, victim.ID, "Two", "Creature — Bear", 2, 2)
	rangers, _ := ithilienPrompt(t, g)
	pickCard(t, g, me.ID, p2)

	// In response: a -1/-1 counter, so the Rangers are 2/2 and the
	// target no longer has LESSER power.
	findBattlefieldCardByID(g, rangers).Counters = map[string]int{game.CounterMinusOne: 1}
	passPriorityAroundTable(t, g)

	if got := findBattlefieldCardByID(g, p2).Controller; got != victim.ID {
		t.Errorf("controller = %s, want the victim — the target was no longer legal at resolution (CR 608.2b)", got)
	}
}

func TestRangersOfIthilienGrownInResponseStillSteals(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	p2 := lftSeed(g, victim.ID, "Two", "Creature — Bear", 2, 2)
	rangers, _ := ithilienPrompt(t, g)
	pickCard(t, g, me.ID, p2)

	findBattlefieldCardByID(g, rangers).Counters = map[string]int{game.CounterPlusOne: 2}
	passPriorityAroundTable(t, g)

	if got := findBattlefieldCardByID(g, p2).Controller; got != me.ID {
		t.Errorf("controller = %s, want the Rangers' controller", got)
	}
}

func TestRangersOfIthilienTheTargetGrowingPastItFizzles(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	p2 := lftSeed(g, victim.ID, "Two", "Creature — Bear", 2, 2)
	_, _ = ithilienPrompt(t, g)
	pickCard(t, g, me.ID, p2)

	findBattlefieldCardByID(g, p2).Counters = map[string]int{game.CounterPlusOne: 1}
	passPriorityAroundTable(t, g)

	if got := findBattlefieldCardByID(g, p2).Controller; got != victim.ID {
		t.Errorf("controller = %s, want the victim — its power reached the Rangers'", got)
	}
}

func TestRangersOfIthilienOffersNothingWithNoLesserCreature(t *testing.T) {
	g := newCatalogGame(t)
	victim := g.Seats[1]
	big := lftSeed(g, victim.ID, "Big", "Creature — Giant", 5, 5)
	lftCastSizedCreature(t, g, "Rangers of Ithilien", "Creature — Human Ranger", rangersOfIthilienOracle, 3, nil)
	passPriorityAroundTable(t, g)
	if got := findBattlefieldCardByID(g, big).Controller; got != victim.ID {
		t.Errorf("the 5-power creature changed hands")
	}
	if !stackFullyEmpty(g) {
		t.Error("the table is still waiting on the up-to-one trigger")
	}
}

// --- mentor: the source dead, last-known information ------------------

// mentorAttack puts the mentor and its allies on the battlefield, declares
// them all as attackers and returns at the mentor's target prompt.
func mentorAttack(t *testing.T, g *game.Game, mentor uuid.UUID, allies ...uuid.UUID) *game.PendingChoice {
	t.Helper()
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	advanceToDeclareAttackersOf(t, g, seat)
	for _, id := range append([]uuid.UUID{mentor}, allies...) {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	prompt := latestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("mentor asked for no target")
	}
	return prompt
}

func mentorGame(t *testing.T, name, oracle string, power int) (*game.Game, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Soldier", OracleID: oracle,
		Power: power, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	return g, id
}

func ally(g *game.Game, name string, power int) uuid.UUID {
	me := g.Seats[g.Turn.ActiveSeat]
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Soldier", Power: power, Toughness: 2,
		Owner: me.ID, Controller: me.ID,
	})
}

var mentorCards = []struct {
	name, oracle string
	power        int
}{
	{"Hammer Dropper", hammerDropperOracle, 5},
	{"Blade Instructor", bladeInstructorOracle, 3},
	{"Barging Sergeant", bargingSergeantOracle, 4},
}

func TestMentorCardsOfferOnlyAttackersWithLesserPower(t *testing.T) {
	for _, tc := range mentorCards {
		t.Run(tc.name, func(t *testing.T) {
			g, mentor := mentorGame(t, tc.name, tc.oracle, tc.power)
			me := g.Seats[g.Turn.ActiveSeat]
			small := ally(g, "Small", tc.power-1)
			equal := ally(g, "Equal", tc.power)
			home := ally(g, "Home", 0) // not attacking
			prompt := mentorAttack(t, g, mentor, small, equal)

			if offered := prompt.PickTargetCards; len(offered) != 1 || offered[0] != small {
				t.Fatalf("offered %v, want only the attacker with lesser power", offered)
			}
			if err := g.ResolvePickTarget(prompt.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: equal}); !errors.Is(err, game.ErrIllegalTarget) {
				t.Errorf("an equal-power pick = %v, want ErrIllegalTarget", err)
			}
			if err := g.ResolvePickTarget(prompt.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: home}); !errors.Is(err, game.ErrIllegalTarget) {
				t.Errorf("a non-attacking pick = %v, want ErrIllegalTarget", err)
			}
			pickCard(t, g, me.ID, small)
			passPriorityAroundTable(t, g)
			if got := allCountersOn(t, g, small)["+1/+1"]; got != 1 {
				t.Errorf("+1/+1 counters = %d, want 1", got)
			}
		})
	}
}

func TestMentorDroppedWhenNoAttackerHasLesserPower(t *testing.T) {
	g, mentor := mentorGame(t, "Hammer Dropper", hammerDropperOracle, 5)
	equal := ally(g, "Equal", 5)
	advanceToDeclareAttackersOf(t, g, g.Turn.ActiveSeat)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	for _, id := range []uuid.UUID{mentor, equal} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatal(err)
		}
	}
	lockInAttacks(t, g)
	if latestPickTarget(g, g.Seats[g.Turn.ActiveSeat].ID) != nil {
		t.Error("mentor asked for a target with no lesser attacker (CR 603.3d)")
	}
}

func TestMentorShrunkInResponseDoesNotPutTheCounter(t *testing.T) {
	g, mentor := mentorGame(t, "Hammer Dropper", hammerDropperOracle, 5)
	me := g.Seats[g.Turn.ActiveSeat]
	small := ally(g, "Small", 4)
	mentorAttack(t, g, mentor, small)
	pickCard(t, g, me.ID, small)

	// In response the mentor is shrunk below the target: the counter
	// must not be put (CR 608.2b).
	findBattlefieldCardByID(g, mentor).Counters = map[string]int{game.CounterMinusOne: 2}
	passPriorityAroundTable(t, g)
	if got := allCountersOn(t, g, small)["+1/+1"]; got != 0 {
		t.Errorf("+1/+1 counters = %d, want 0 — the mentor's power fell to 3, below the target's 4", got)
	}
}

// TestMentorSourceDeadUsesLastKnownPower: the mentor, carrying two +1/+1
// counters (power 7), dies in response. Its graveyard card prints
// power 5; the target's power 6 is lesser than the LAST-KNOWN power
// only, so the counter landing proves the check read the record
// (CR 608.2h) and not the card in the graveyard.
func TestMentorSourceDeadUsesLastKnownPower(t *testing.T) {
	g, mentor := mentorGame(t, "Hammer Dropper", hammerDropperOracle, 5)
	me := g.Seats[g.Turn.ActiveSeat]
	findBattlefieldCardByID(g, mentor).Counters = map[string]int{game.CounterPlusOne: 2}
	target := ally(g, "Six", 6)
	mentorAttack(t, g, mentor, target)
	pickCard(t, g, me.ID, target)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(mentor, game.DestroyOptions{}) })
	if findBattlefieldCardByID(g, mentor) != nil {
		t.Fatal("the mentor did not die")
	}
	passPriorityAroundTable(t, g)
	if got := allCountersOn(t, g, target)["+1/+1"]; got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1 — power 6 is lesser than the mentor's last-known 7 (CR 608.2h)", got)
	}
}

// --- Unliving Psychopath: an activated ability ------------------------

func psychopathGame(t *testing.T) (*game.Game, uuid.UUID, *game.Player) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Unliving Psychopath", TypeLine: "Creature — Zombie Assassin",
		OracleID: unlivingPsychopathOracle, Power: 0, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	return g, id, me
}

func TestUnlivingPsychopathDestroysOnlyLesserPower(t *testing.T) {
	g, psycho, me := psychopathGame(t)
	victim := g.Seats[1]
	zero := lftSeed(g, victim.ID, "Wall", "Creature — Wall", 0, 5)
	one := lftSeed(g, victim.ID, "One", "Creature — Bear", 1, 1)
	two := lftSeed(g, victim.ID, "Two", "Creature — Bear", 2, 2)

	// At its printed power 0 nothing is lesser.
	fillPoolColored(me, "B", 1)
	err := g.ActivateCatalogAbility(me.ID, psycho, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: zero}},
	})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("destroy at power 0 = %v, want ErrIllegalTarget", err)
	}

	// Pump once: power 1, so the 0-power wall is legal and the 1-power bear is not.
	fillPoolColored(me, "B", 1)
	if err := g.ActivateCatalogAbility(me.ID, psycho, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("pump: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{one, two} {
		fillPoolColored(me, "B", 1)
		err := g.ActivateCatalogAbility(me.ID, psycho, 1, game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: id}},
		})
		if !errors.Is(err, game.ErrIllegalTarget) {
			t.Fatalf("destroy a power >= 1 creature = %v, want ErrIllegalTarget", err)
		}
	}
	fillPoolColored(me, "B", 1)
	if err := g.ActivateCatalogAbility(me.ID, psycho, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: zero}},
	}); err != nil {
		t.Fatalf("destroy the power-0 wall: %v", err)
	}
	passPriorityAroundTable(t, g)
	if findBattlefieldCardByID(g, zero) != nil {
		t.Error("the wall survived")
	}
}

func TestUnlivingPsychopathShrunkInResponseFizzles(t *testing.T) {
	g, psycho, me := psychopathGame(t)
	victim := g.Seats[1]
	bear := lftSeed(g, victim.ID, "Bear", "Creature — Bear", 1, 1)
	findBattlefieldCardByID(g, psycho).Counters = map[string]int{game.CounterPlusOne: 2}

	fillPoolColored(me, "B", 1)
	if err := g.ActivateCatalogAbility(me.ID, psycho, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("destroy: %v", err)
	}
	findBattlefieldCardByID(g, psycho).Counters = nil // the counters are removed in response
	passPriorityAroundTable(t, g)
	if findBattlefieldCardByID(g, bear) == nil {
		t.Error("the bear was destroyed although the source's power fell to 0 (CR 608.2b)")
	}
}

// TestSourceRelativeSpecWithNoSourceAdmitsNothing: a walk that cannot
// say what it is relative to must not admit everything.
func TestSourceRelativeSpecWithNoSourceAdmitsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lftSeed(g, g.Seats[1].ID, "Bear", "Creature — Bear", 1, 1)
	spec := RelativeToSource(TargetCreature("target creature with lesser power"), LesserPower())
	var lt game.LegalTargets
	g.WithWriteLock(func() { lt = g.LegalTargetsForEffect(game.SourceChooser(me.ID), spec) })
	if len(lt.Cards) != 0 {
		t.Errorf("a source-less walk admitted %v", lt.Cards)
	}
}
