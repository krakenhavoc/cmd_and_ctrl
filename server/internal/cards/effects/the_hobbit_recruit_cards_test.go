package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// the_hobbit_recruit_cards_test.go — the recruit cards of The Hobbit
// (#4475). Recruit itself (CR 701.70a) is pinned in recruit_test.go;
// these pin each card's trigger or clause reaching it.

const (
	lakeTownLookoutOracle        = "cf765efe-884c-48e2-9edb-9d45cf2756dd"
	longLakeNuisanceOracle       = "a833fdf1-db0c-4846-8452-d3b2059c2355"
	patientInstructorOracle      = "bddd7e99-ec74-4ca6-9137-155b85695a95"
	esgarothGarrisonOracle       = "84597236-2cfc-4654-b126-a8ee252f4d14"
	greatGildedBoatOracle        = "d4607026-80b2-4fa6-99d5-694cd3b43868"
	theQueenOfDaleOracle         = "a2ec1dd0-86c7-423d-b562-ed95b79bf8f7"
	bardsCompanyOracle           = "5c5bfbb2-0e63-4e43-b441-c4878983288f"
	soundTheTrumpetsOracle       = "dbcf2a79-b455-43ac-9cf0-9e63708800f4"
	celebrateTheMountainKingOrac = "d51136fa-3c13-48a5-83fd-51fe00010a4b"
	theMountainKingsReturnOracle = "32ad5b3e-92c0-45be-b2e4-6f1794552f36"
)

// hobEnter puts c into its owner's hand and moves it onto the
// battlefield, so its entry triggers fire as they would from a cast.
func hobEnter(t *testing.T, g *game.Game, owner uuid.UUID, c game.Card) uuid.UUID {
	t.Helper()
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = owner, owner
	p := playerByIDForTest(g, owner)
	p.Hand.PushTop(c)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: owner}, game.ZoneRef{Kind: game.ZoneBattlefield}, c.InstanceID); err != nil {
		t.Fatalf("enter %s: %v", c.Name, err)
	}
	return c.InstanceID
}

// hobRecruitAnswered resolves what is on the stack until the recruiter's
// discard opens, then discards a nonland card from the hand and reports
// how many Human Soldiers the recruiter has.
func hobRecruitAnswered(t *testing.T, g *game.Game, me *game.Player) int {
	t.Helper()
	for i := 0; i < 12 && discardChoiceFor(g, me.ID) == nil; i++ {
		hobAnswerOrder(t, g, me.ID)
		passPriorityAroundTable(t, g)
	}
	sweep01Discard(t, g, me.ID, recruitDiscarded.Key())
	var spell uuid.UUID
	for _, c := range me.Hand.Cards {
		if !c.IsLand() {
			spell = c.InstanceID
			break
		}
	}
	if spell == uuid.Nil {
		t.Fatal("no nonland card in hand to discard")
	}
	answerDiscard(t, g, me.ID, spell)
	return recruitSoldiers(g, me.ID)
}

func TestHobLakeTownLookoutRecruitsWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lookout := pushTestPermanent(g, me.ID, game.Card{Name: "Lake-town Lookout", OracleID: lakeTownLookoutOracle,
		TypeLine: "Creature — Human Scout", Power: 1, Toughness: 1})
	destroyNow(t, g, lookout)
	if got := hobRecruitAnswered(t, g, me); got != 1 {
		t.Errorf("%d Human Soldiers after the Lookout died and recruited, want 1", got)
	}
}

func TestHobLongLakeNuisanceRecruitsAsItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bird := hobEnter(t, g, me.ID, game.Card{Name: "Long Lake Nuisance", OracleID: longLakeNuisanceOracle,
		TypeLine: "Creature — Bird", Power: 3, Toughness: 1, Keywords: []string{"flying"}})
	if triggerOnStack(g, bird) == nil {
		t.Fatal("the Nuisance's entry trigger is not on the stack")
	}
	if got := hobRecruitAnswered(t, g, me); got != 1 {
		t.Errorf("%d Human Soldiers, want 1", got)
	}
}

func TestHobPatientInstructorRecruitsAsItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hobEnter(t, g, me.ID, game.Card{Name: "Patient Instructor", OracleID: patientInstructorOracle,
		TypeLine: "Creature — Human Citizen", Power: 2, Toughness: 2, Keywords: []string{"vigilance"}})
	if got := hobRecruitAnswered(t, g, me); got != 1 {
		t.Errorf("%d Human Soldiers, want 1", got)
	}
}

// The Garrison's power counts the creatures you control, itself and the
// Soldier its recruit makes included.
func TestHobEsgarothGarrisonCountsCreaturesAndRecruits(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	garrison := hobEnter(t, g, me.ID, game.Card{Name: "Esgaroth Garrison", OracleID: esgarothGarrisonOracle,
		TypeLine: "Creature — Human Soldier", Power: 0, Toughness: 5})
	if p, tough := effectivePTOf(t, g, garrison); p != 2 || tough != 5 {
		t.Errorf("Garrison is %d/%d with two creatures, want 2/5", p, tough)
	}
	if got := hobRecruitAnswered(t, g, me); got != 1 {
		t.Fatalf("%d Human Soldiers, want 1", got)
	}
	if p, _ := effectivePTOf(t, g, garrison); p != 3 {
		t.Errorf("Garrison's power after the Soldier = %d, want 3", p)
	}
}

// Whenever you attack: once per declaration, whatever attacks.
func TestHobGreatGildedBoatRecruitsWheneverYouAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	boat := pushTestPermanent(g, me.ID, game.Card{Name: "Great Gilded Boat", OracleID: greatGildedBoatOracle,
		TypeLine: "Artifact — Vehicle", Power: 4, Toughness: 4})
	a := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Other Bear", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{a, b} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	_ = boat
	if got := hobRecruitAnswered(t, g, me); got != 1 {
		t.Errorf("%d Human Soldiers, want 1", got)
	}
	passPriorityAroundTable(t, g)
	if discardChoiceFor(g, me.ID) != nil {
		t.Error("a second recruit for one attack declaration")
	}
}

// An opponent's first noncreature spell of the turn makes the Queen's
// controller recruit, on the opponent's turn; the second does not.
func TestHobTheQueenOfDaleRecruitsOnAnOpponentsFirstNoncreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	queenSide := g.Seats[1]
	queen := pushTestPermanent(g, queenSide.ID, game.Card{Name: "The Queen of Dale", OracleID: theQueenOfDaleOracle,
		TypeLine: "Legendary Creature — Human Noble", Power: 2, Toughness: 1})
	castCatalogSpell(t, g, "Colossal Dreadmaw", "Creature — Dinosaur", colossalDreadmawOracle, nil)
	if triggerOnStack(g, queen) != nil {
		t.Fatal("the Queen triggered on a creature spell")
	}
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: queenSide.ID}})
	if triggerOnStack(g, queen) == nil {
		t.Fatal("the Queen did not trigger on the first noncreature spell")
	}
	if got := hobRecruitAnswered(t, g, queenSide); got != 1 {
		t.Fatalf("%d Human Soldiers for the Queen's controller, want 1", got)
	}
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: queenSide.ID}})
	if triggerOnStack(g, queen) != nil {
		t.Error("the Queen triggered on the second noncreature spell of the turn")
	}
}

// Other creatures get +1/+1, the Company itself does not; it recruits as
// it enters; and it may be cast at instant speed only with a Human.
func TestHobBardsCompanyPumpsOthersRecruitsAndFlashesWithAHuman(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	company := hobEnter(t, g, me.ID, game.Card{Name: "Bard's Company", OracleID: bardsCompanyOracle,
		TypeLine: "Creature — Human Citizen", Power: 2, Toughness: 3})
	if p, tough := effectivePTOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the Bear is %d/%d, want 3/3", p, tough)
	}
	if p, tough := effectivePTOf(t, g, company); p != 2 || tough != 3 {
		t.Errorf("the Company is %d/%d, want its own 2/3", p, tough)
	}
	if got := hobRecruitAnswered(t, g, me); got != 1 {
		t.Fatalf("%d Human Soldiers, want 1", got)
	}

	spec, _ := Lookup(bardsCompanyOracle)
	if len(spec.AlternativeCosts) != 1 || !spec.AlternativeCosts[0].AsThoughFlash || spec.AlternativeCosts[0].ManaCost != "{2}{W}{U}" {
		t.Fatalf("the flash offer is %+v, want one as-though-flash offer at {2}{W}{U}", spec.AlternativeCosts)
	}
	cond := spec.AlternativeCosts[0].Condition
	fresh := newCatalogGame(t)
	var without, with bool
	fresh.WithWriteLock(func() { without = cond(fresh, fresh.Seats[0].ID) })
	pushVanillaCreature(fresh, fresh.Seats[1].ID, "Their Human", 1, 1) // not yours
	fresh.WithWriteLock(func() { without = without || cond(fresh, fresh.Seats[0].ID) })
	pushTestPermanent(fresh, fresh.Seats[0].ID, game.Card{Name: "Villager", TypeLine: "Creature — Human Peasant", Power: 1, Toughness: 1})
	fresh.WithWriteLock(func() { with = cond(fresh, fresh.Seats[0].ID) })
	if without || !with {
		t.Errorf("flash condition: no Human of yours %v, a Human of yours %v; want false, true", without, with)
	}
}

// Countering a spell of mana value 2 or less recruits; 3 or more does not.
func TestHobSoundTheTrumpetsRecruitsOnlyForACheapSpell(t *testing.T) {
	for _, tc := range []struct {
		name string
		cost string
		want bool
	}{{"two", "{1}{G}", true}, {"three", "{2}{G}", false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			advanceTo(t, g, game.StepPrecombatMain)
			bear := uuid.New()
			me.Hand.PushTop(game.Card{InstanceID: bear, Name: "Bear Spell", TypeLine: "Creature — Bear", ManaCost: tc.cost,
				Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
			if err := g.CastSpell(me.ID, bear, game.CastSpellParams{}); err != nil {
				t.Fatalf("cast the bear: %v", err)
			}
			castCatalogSpell(t, g, "Sound the Trumpets", "Instant", soundTheTrumpetsOracle,
				[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
			for i := 0; i < 4 && discardChoiceFor(g, me.ID) == nil && !me.Graveyard.Contains(bear); i++ {
				passPriorityAroundTable(t, g)
			}
			if !me.Graveyard.Contains(bear) {
				t.Fatalf("the %s-mana spell was not countered", tc.name)
			}
			if got := discardChoiceFor(g, me.ID) != nil; got != tc.want {
				t.Fatalf("recruit discard open = %v, want %v", got, tc.want)
			}
		})
	}
}

// hobAnswerTargets answers every open pick_target prompt of chooser's
// with the first offered card in want, or with nothing ("up to one").
func hobAnswerTargets(t *testing.T, g *game.Game, chooser uuid.UUID, want ...uuid.UUID) {
	t.Helper()
	for i := 0; i < 8; i++ {
		var ch *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoicePickTarget && c.Chooser == chooser {
				ch = c
				break
			}
		}
		if ch == nil {
			return
		}
		var refs []game.TargetRef
		for _, offered := range ch.PickTargetCards {
			for _, w := range want {
				if offered == w && len(refs) == 0 {
					refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: w})
				}
			}
		}
		if err := g.ResolvePickTargets(ch.ID, chooser, refs); err != nil {
			t.Fatalf("ResolvePickTargets: %v", err)
		}
	}
}

// One nonland permanent of each opponent is exiled until the enchantment
// leaves, and comes back when it does; the second trigger recruits.
func TestHobCelebrateTheMountainKingExilesUntilItLeavesAndRecruits(t *testing.T) {
	g := newCatalogGame(t)
	me, b, c := g.Seats[0], g.Seats[1], g.Seats[2]
	theirs := pushVanillaCreature(g, b.ID, "B's Bear", 2, 2)
	others := pushVanillaCreature(g, c.ID, "C's Bear", 2, 2)
	advanceTo(t, g, game.StepPrecombatMain)
	ench := castCatalogSpell(t, g, "Celebrate the Mountain-king", "Enchantment", celebrateTheMountainKingOrac, nil)
	passPriorityAroundTable(t, g)
	soldiers := 0
	for i := 0; i < 16; i++ {
		hobAnswerTargets(t, g, me.ID, theirs, others)
		hobAnswerOrder(t, g, me.ID)
		if discardChoiceFor(g, me.ID) != nil {
			soldiers = hobRecruitAnswered(t, g, me)
			continue
		}
		if len(g.StackMeta) == 0 {
			break
		}
		passPriorityAroundTable(t, g)
	}
	if g.Battlefield.Contains(theirs) || g.Battlefield.Contains(others) {
		hobDump(t, g)
		t.Fatal("the opponents' permanents were not exiled")
	}
	if soldiers != 1 {
		t.Fatalf("%d Human Soldiers, want 1", soldiers)
	}
	if !g.Exile.Contains(theirs) || !g.Exile.Contains(others) {
		t.Fatal("the picks are not in exile")
	}
	destroyNow(t, g, ench)
	g.RunStateChecksForTest()
	if battlefieldNamed(g, b.ID, "B's Bear") != 1 || battlefieldNamed(g, c.ID, "C's Bear") != 1 {
		t.Error("the exiled permanents did not return to their owners when the enchantment left")
	}
}

// I recruits, II returns a creature card with mana value 3 or less, III
// puts a +1/+1 counter on a creature.
func TestHobTheMountainKingsReturnChapters(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	advanceTo(t, g, game.StepPrecombatMain)
	saga := castCatalogSpell(t, g, "The Mountain-king's Return", "Enchantment — Saga", theMountainKingsReturnOracle, nil)
	passPriorityAroundTable(t, g)
	if got := loreCountersOn(g, saga); got != 1 {
		t.Fatalf("lore counters after entry = %d, want 1", got)
	}
	if got := hobRecruitAnswered(t, g, me); got != 1 {
		t.Fatalf("chapter I: %d Human Soldiers, want 1", got)
	}

	dead := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
			Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	})
	advanceToPrecombatMainOf(t, g, seat)
	hobAnswerTargets(t, g, me.ID, dead)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dead) {
		hobDump(t, g)
		t.Fatal("chapter II did not return the creature card")
	}

	advanceToPrecombatMainOf(t, g, seat)
	hobAnswerTargets(t, g, me.ID, dead)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, dead, game.CounterPlusOne); got != 1 {
		t.Errorf("chapter III: %d +1/+1 counters on the returned creature, want 1", got)
	}
}

func hobDump(t *testing.T, g *game.Game) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil {
			t.Logf("pending %s chooser %s reason %q cards %v", c.Kind, c.Chooser, c.Reason, c.PickTargetCards)
		}
	}
	for id, it := range g.StackMeta {
		t.Logf("stack %s %q", id, it.Label)
	}
	t.Logf("step %s active %d", g.Turn.Step, g.Turn.ActiveSeat)
	for _, c := range g.Battlefield.Cards {
		t.Logf("bf %s %v", c.Name, c.Counters)
	}
}

// hobAnswerOrder answers chooser's open trigger-order prompt, if any, in
// the order it offered, without passing priority.
func hobAnswerOrder(t *testing.T, g *game.Game, chooser uuid.UUID) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerOrder && c.Chooser == chooser {
			if err := g.ResolveTriggerOrder(c.ID, chooser, append([]uuid.UUID(nil), c.TriggerOrderIDs...)); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			return
		}
	}
}
