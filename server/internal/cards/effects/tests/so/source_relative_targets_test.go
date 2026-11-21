package testsso

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ithilienPrompt casts a 3/3 Rangers of Ithilien, resolves the spell,
// and stops at the trigger's target prompt.
func ithilienPrompt(t *testing.T, g *game.Game) (rangers uuid.UUID, prompt *game.PendingChoice) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	rangers = effectstest.LftCastSizedCreature(t, g, "Rangers of Ithilien", "Creature — Human Ranger", effectstest.RangersOfIthilienOracle, 3, nil)
	effectstest.PassPriorityAroundTable(t, g)
	prompt = effectstest.LatestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("the Rangers' trigger asked for no target")
	}
	return rangers, prompt
}

func TestRangersOfIthilienOffersOnlyLesserPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	victim := g.Seats[1]
	p1 := effectstest.LftSeed(g, victim.ID, "One", "Creature — Bear", 1, 1)
	p2 := effectstest.LftSeed(g, victim.ID, "Two", "Creature — Bear", 2, 2)
	p3 := effectstest.LftSeed(g, victim.ID, "Three", "Creature — Bear", 3, 3)
	p4 := effectstest.LftSeed(g, victim.ID, "Four", "Creature — Bear", 4, 4)
	_, prompt := ithilienPrompt(t, g)

	offered := prompt.PickTargetCards
	if !effectstest.HasID(offered, p1) || !effectstest.HasID(offered, p2) {
		t.Errorf("offered %v, want the 1- and 2-power creatures", offered)
	}
	if effectstest.HasID(offered, p3) || effectstest.HasID(offered, p4) {
		t.Errorf("offered %v, which includes a creature of power 3 or more", offered)
	}
}

func TestRangersOfIthilienRefusesAnEqualPowerPickAtAnnounce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	p3 := effectstest.LftSeed(g, victim.ID, "Three", "Creature — Bear", 3, 3)
	effectstest.LftSeed(g, victim.ID, "One", "Creature — Bear", 1, 1) // so the trigger has a target to ask for
	_, prompt := ithilienPrompt(t, g)

	err := g.ResolvePickTarget(prompt.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: p3})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("picking an equal-power creature = %v, want ErrIllegalTarget (CR 601.2c)", err)
	}
}

func TestRangersOfIthilienStealsALesserCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	p2 := effectstest.LftSeed(g, victim.ID, "Two", "Creature — Bear", 2, 2)
	_, _ = ithilienPrompt(t, g)
	effectstest.PickCard(t, g, me.ID, p2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindBattlefieldCardByID(g, p2).Controller; got != me.ID {
		t.Errorf("controller = %s, want the Rangers' controller", got)
	}
}

func TestRangersOfIthilienShrunkInResponseFizzlesTheSteal(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	p2 := effectstest.LftSeed(g, victim.ID, "Two", "Creature — Bear", 2, 2)
	rangers, _ := ithilienPrompt(t, g)
	effectstest.PickCard(t, g, me.ID, p2)

	// In response: a -1/-1 counter, so the Rangers are 2/2 and the
	// target no longer has LESSER power.
	effectstest.FindBattlefieldCardByID(g, rangers).Counters = map[string]int{game.CounterMinusOne: 1}
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.FindBattlefieldCardByID(g, p2).Controller; got != victim.ID {
		t.Errorf("controller = %s, want the victim — the target was no longer legal at resolution (CR 608.2b)", got)
	}
}

func TestRangersOfIthilienGrownInResponseStillSteals(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	p2 := effectstest.LftSeed(g, victim.ID, "Two", "Creature — Bear", 2, 2)
	rangers, _ := ithilienPrompt(t, g)
	effectstest.PickCard(t, g, me.ID, p2)

	effectstest.FindBattlefieldCardByID(g, rangers).Counters = map[string]int{game.CounterPlusOne: 2}
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.FindBattlefieldCardByID(g, p2).Controller; got != me.ID {
		t.Errorf("controller = %s, want the Rangers' controller", got)
	}
}

func TestRangersOfIthilienTheTargetGrowingPastItFizzles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	p2 := effectstest.LftSeed(g, victim.ID, "Two", "Creature — Bear", 2, 2)
	_, _ = ithilienPrompt(t, g)
	effectstest.PickCard(t, g, me.ID, p2)

	effectstest.FindBattlefieldCardByID(g, p2).Counters = map[string]int{game.CounterPlusOne: 1}
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.FindBattlefieldCardByID(g, p2).Controller; got != victim.ID {
		t.Errorf("controller = %s, want the victim — its power reached the Rangers'", got)
	}
}

func TestRangersOfIthilienOffersNothingWithNoLesserCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	victim := g.Seats[1]
	big := effectstest.LftSeed(g, victim.ID, "Big", "Creature — Giant", 5, 5)
	effectstest.LftCastSizedCreature(t, g, "Rangers of Ithilien", "Creature — Human Ranger", effectstest.RangersOfIthilienOracle, 3, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindBattlefieldCardByID(g, big).Controller; got != victim.ID {
		t.Errorf("the 5-power creature changed hands")
	}
	if !effectstest.StackFullyEmpty(g) {
		t.Error("the table is still waiting on the up-to-one trigger")
	}
}

var mentorCards = []struct {
	name, oracle string
	power        int
}{
	{"Hammer Dropper", effectstest.HammerDropperOracle, 5},
	{"Blade Instructor", effectstest.BladeInstructorOracle, 3},
	{"Barging Sergeant", effectstest.BargingSergeantOracle, 4},
}

func TestMentorCardsOfferOnlyAttackersWithLesserPower(t *testing.T) {
	for _, tc := range mentorCards {
		t.Run(tc.name, func(t *testing.T) {
			g, mentor := effectstest.MentorGame(t, tc.name, tc.oracle, tc.power)
			me := g.Seats[g.Turn.ActiveSeat]
			small := effectstest.Ally(g, "Small", tc.power-1)
			equal := effectstest.Ally(g, "Equal", tc.power)
			home := effectstest.Ally(g, "Home", 0) // not attacking
			prompt := effectstest.MentorAttack(t, g, mentor, small, equal)

			if offered := prompt.PickTargetCards; len(offered) != 1 || offered[0] != small {
				t.Fatalf("offered %v, want only the attacker with lesser power", offered)
			}
			if err := g.ResolvePickTarget(prompt.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: equal}); !errors.Is(err, game.ErrIllegalTarget) {
				t.Errorf("an equal-power pick = %v, want ErrIllegalTarget", err)
			}
			if err := g.ResolvePickTarget(prompt.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: home}); !errors.Is(err, game.ErrIllegalTarget) {
				t.Errorf("a non-attacking pick = %v, want ErrIllegalTarget", err)
			}
			effectstest.PickCard(t, g, me.ID, small)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.AllCountersOn(t, g, small)["+1/+1"]; got != 1 {
				t.Errorf("+1/+1 counters = %d, want 1", got)
			}
		})
	}
}

func TestMentorDroppedWhenNoAttackerHasLesserPower(t *testing.T) {
	g, mentor := effectstest.MentorGame(t, "Hammer Dropper", effectstest.HammerDropperOracle, 5)
	equal := effectstest.Ally(g, "Equal", 5)
	effectstest.AdvanceToDeclareAttackersOf(t, g, g.Turn.ActiveSeat)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	for _, id := range []uuid.UUID{mentor, equal} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.LockInAttacks(t, g)
	if effectstest.LatestPickTarget(g, g.Seats[g.Turn.ActiveSeat].ID) != nil {
		t.Error("mentor asked for a target with no lesser attacker (CR 603.3d)")
	}
}

func TestMentorShrunkInResponseDoesNotPutTheCounter(t *testing.T) {
	g, mentor := effectstest.MentorGame(t, "Hammer Dropper", effectstest.HammerDropperOracle, 5)
	me := g.Seats[g.Turn.ActiveSeat]
	small := effectstest.Ally(g, "Small", 4)
	effectstest.MentorAttack(t, g, mentor, small)
	effectstest.PickCard(t, g, me.ID, small)

	// In response the mentor is shrunk below the target: the counter
	// must not be put (CR 608.2b).
	effectstest.FindBattlefieldCardByID(g, mentor).Counters = map[string]int{game.CounterMinusOne: 2}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.AllCountersOn(t, g, small)["+1/+1"]; got != 0 {
		t.Errorf("+1/+1 counters = %d, want 0 — the mentor's power fell to 3, below the target's 4", got)
	}
}

// TestMentorSourceDeadUsesLastKnownPower: the mentor, carrying two +1/+1
// counters (power 7), dies in response. Its graveyard card prints
// power 5; the target's power 6 is lesser than the LAST-KNOWN power
// only, so the counter landing proves the check read the record
// (CR 608.2h) and not the card in the graveyard.
func TestMentorSourceDeadUsesLastKnownPower(t *testing.T) {
	g, mentor := effectstest.MentorGame(t, "Hammer Dropper", effectstest.HammerDropperOracle, 5)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.FindBattlefieldCardByID(g, mentor).Counters = map[string]int{game.CounterPlusOne: 2}
	target := effectstest.Ally(g, "Six", 6)
	effectstest.MentorAttack(t, g, mentor, target)
	effectstest.PickCard(t, g, me.ID, target)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(mentor, game.DestroyOptions{}) })
	if effectstest.FindBattlefieldCardByID(g, mentor) != nil {
		t.Fatal("the mentor did not die")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.AllCountersOn(t, g, target)["+1/+1"]; got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1 — power 6 is lesser than the mentor's last-known 7 (CR 608.2h)", got)
	}
}

// --- Unliving Psychopath: an activated ability ------------------------

func psychopathGame(t *testing.T) (*game.Game, uuid.UUID, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Unliving Psychopath", TypeLine: "Creature — Zombie Assassin",
		OracleID: effectstest.UnlivingPsychopathOracle, Power: 0, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	return g, id, me
}

func TestUnlivingPsychopathDestroysOnlyLesserPower(t *testing.T) {
	g, psycho, me := psychopathGame(t)
	victim := g.Seats[1]
	zero := effectstest.LftSeed(g, victim.ID, "Wall", "Creature — Wall", 0, 5)
	one := effectstest.LftSeed(g, victim.ID, "One", "Creature — Bear", 1, 1)
	two := effectstest.LftSeed(g, victim.ID, "Two", "Creature — Bear", 2, 2)

	// At its printed power 0 nothing is lesser.
	effectstest.FillPoolColored(me, "B", 1)
	err := g.ActivateCatalogAbility(me.ID, psycho, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: zero}},
	})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("destroy at power 0 = %v, want ErrIllegalTarget", err)
	}

	// Pump once: power 1, so the 0-power wall is legal and the 1-power bear is not.
	effectstest.FillPoolColored(me, "B", 1)
	if err := g.ActivateCatalogAbility(me.ID, psycho, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("pump: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{one, two} {
		effectstest.FillPoolColored(me, "B", 1)
		err := g.ActivateCatalogAbility(me.ID, psycho, 1, game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: id}},
		})
		if !errors.Is(err, game.ErrIllegalTarget) {
			t.Fatalf("destroy a power >= 1 creature = %v, want ErrIllegalTarget", err)
		}
	}
	effectstest.FillPoolColored(me, "B", 1)
	if err := g.ActivateCatalogAbility(me.ID, psycho, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: zero}},
	}); err != nil {
		t.Fatalf("destroy the power-0 wall: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardByID(g, zero) != nil {
		t.Error("the wall survived")
	}
}

func TestUnlivingPsychopathShrunkInResponseFizzles(t *testing.T) {
	g, psycho, me := psychopathGame(t)
	victim := g.Seats[1]
	bear := effectstest.LftSeed(g, victim.ID, "Bear", "Creature — Bear", 1, 1)
	effectstest.FindBattlefieldCardByID(g, psycho).Counters = map[string]int{game.CounterPlusOne: 2}

	effectstest.FillPoolColored(me, "B", 1)
	if err := g.ActivateCatalogAbility(me.ID, psycho, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("destroy: %v", err)
	}
	effectstest.FindBattlefieldCardByID(g, psycho).Counters = nil // the counters are removed in response
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardByID(g, bear) == nil {
		t.Error("the bear was destroyed although the source's power fell to 0 (CR 608.2b)")
	}
}

// TestSourceRelativeSpecWithNoSourceAdmitsNothing: a walk that cannot
// say what it is relative to must not admit everything.
func TestSourceRelativeSpecWithNoSourceAdmitsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.LftSeed(g, g.Seats[1].ID, "Bear", "Creature — Bear", 1, 1)
	spec := effects.RelativeToSource(effects.TargetCreature("target creature with lesser power"), effects.LesserPower())
	var lt game.LegalTargets
	g.WithWriteLock(func() { lt = g.LegalTargetsForEffect(game.SourceChooser(me.ID), spec) })
	if len(lt.Cards) != 0 {
		t.Errorf("a source-less walk admitted %v", lt.Cards)
	}
}
