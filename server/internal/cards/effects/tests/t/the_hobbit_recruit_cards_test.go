package testst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hobRecruitAnswered resolves what is on the stack until the recruiter's
// discard opens, then discards a nonland card from the hand and reports
// how many Human Soldiers the recruiter has.
func hobRecruitAnswered(t *testing.T, g *game.Game, me *game.Player) int {
	t.Helper()
	for i := 0; i < 12 && effectstest.DiscardChoiceFor(g, me.ID) == nil; i++ {
		effectstest.HobAnswerOrder(t, g, me.ID)
		effectstest.PassPriorityAroundTable(t, g)
	}
	effectstest.Sweep01Discard(t, g, me.ID, effects.RecruitDiscarded.Key())
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
	effectstest.AnswerDiscard(t, g, me.ID, spell)
	return effectstest.RecruitSoldiers(g, me.ID)
}

func TestHobLakeTownLookoutRecruitsWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lookout := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Lake-town Lookout", OracleID: effectstest.LakeTownLookoutOracle,
		TypeLine: "Creature — Human Scout", Power: 1, Toughness: 1})
	effectstest.DestroyNow(t, g, lookout)
	if got := hobRecruitAnswered(t, g, me); got != 1 {
		t.Errorf("%d Human Soldiers after the Lookout died and recruited, want 1", got)
	}
}

func TestHobLongLakeNuisanceRecruitsAsItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bird := effectstest.HobEnter(t, g, me.ID, game.Card{Name: "Long Lake Nuisance", OracleID: effectstest.LongLakeNuisanceOracle,
		TypeLine: "Creature — Bird", Power: 3, Toughness: 1, Keywords: []string{"flying"}})
	if effectstest.TriggerOnStack(g, bird) == nil {
		t.Fatal("the Nuisance's entry trigger is not on the stack")
	}
	if got := hobRecruitAnswered(t, g, me); got != 1 {
		t.Errorf("%d Human Soldiers, want 1", got)
	}
}

func TestHobPatientInstructorRecruitsAsItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.HobEnter(t, g, me.ID, game.Card{Name: "Patient Instructor", OracleID: effectstest.PatientInstructorOracle,
		TypeLine: "Creature — Human Citizen", Power: 2, Toughness: 2, Keywords: []string{"vigilance"}})
	if got := hobRecruitAnswered(t, g, me); got != 1 {
		t.Errorf("%d Human Soldiers, want 1", got)
	}
}

// The Garrison's power counts the creatures you control, itself and the
// Soldier its recruit makes included.
func TestHobEsgarothGarrisonCountsCreaturesAndRecruits(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	garrison := effectstest.HobEnter(t, g, me.ID, game.Card{Name: "Esgaroth Garrison", OracleID: effectstest.EsgarothGarrisonOracle,
		TypeLine: "Creature — Human Soldier", Power: 0, Toughness: 5})
	if p, tough := effectstest.EffectivePTOf(t, g, garrison); p != 2 || tough != 5 {
		t.Errorf("Garrison is %d/%d with two creatures, want 2/5", p, tough)
	}
	if got := hobRecruitAnswered(t, g, me); got != 1 {
		t.Fatalf("%d Human Soldiers, want 1", got)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, garrison); p != 3 {
		t.Errorf("Garrison's power after the Soldier = %d, want 3", p)
	}
}

// Whenever you attack: once per declaration, whatever attacks.
func TestHobGreatGildedBoatRecruitsWheneverYouAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	boat := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Great Gilded Boat", OracleID: effectstest.GreatGildedBoatOracle,
		TypeLine: "Artifact — Vehicle", Power: 4, Toughness: 4})
	a := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "Other Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
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
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) != nil {
		t.Error("a second recruit for one attack declaration")
	}
}

// An opponent's first noncreature spell of the turn makes the Queen's
// controller recruit, on the opponent's turn; the second does not.
func TestHobTheQueenOfDaleRecruitsOnAnOpponentsFirstNoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	queenSide := g.Seats[1]
	queen := effectstest.PushTestPermanent(g, queenSide.ID, game.Card{Name: "The Queen of Dale", OracleID: effectstest.TheQueenOfDaleOracle,
		TypeLine: "Legendary Creature — Human Noble", Power: 2, Toughness: 1})
	effectstest.CastCatalogSpell(t, g, "Colossal Dreadmaw", "Creature — Dinosaur", effectstest.ColossalDreadmawOracle, nil)
	if effectstest.TriggerOnStack(g, queen) != nil {
		t.Fatal("the Queen triggered on a creature spell")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: queenSide.ID}})
	if effectstest.TriggerOnStack(g, queen) == nil {
		t.Fatal("the Queen did not trigger on the first noncreature spell")
	}
	if got := hobRecruitAnswered(t, g, queenSide); got != 1 {
		t.Fatalf("%d Human Soldiers for the Queen's controller, want 1", got)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: queenSide.ID}})
	if effectstest.TriggerOnStack(g, queen) != nil {
		t.Error("the Queen triggered on the second noncreature spell of the turn")
	}
}

// Other creatures get +1/+1, the Company itself does not; it recruits as
// it enters; and it may be cast at instant speed only with a Human.
func TestHobBardsCompanyPumpsOthersRecruitsAndFlashesWithAHuman(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	company := effectstest.HobEnter(t, g, me.ID, game.Card{Name: "Bard's Company", OracleID: effectstest.BardsCompanyOracle,
		TypeLine: "Creature — Human Citizen", Power: 2, Toughness: 3})
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the Bear is %d/%d, want 3/3", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, company); p != 2 || tough != 3 {
		t.Errorf("the Company is %d/%d, want its own 2/3", p, tough)
	}
	if got := hobRecruitAnswered(t, g, me); got != 1 {
		t.Fatalf("%d Human Soldiers, want 1", got)
	}

	spec, _ := effects.Lookup(effectstest.BardsCompanyOracle)
	// #5207: a timing permission, not an alternative cost.
	if len(spec.AlternativeCosts) != 0 || spec.FlashIf == nil || spec.FlashIf.Condition == nil {
		t.Fatalf("offers %+v, FlashIf %+v; want no offer and a conditional FlashIf", spec.AlternativeCosts, spec.FlashIf)
	}
	cond := spec.FlashIf.Condition
	fresh := effectstest.NewCatalogGame(t)
	var without, with bool
	fresh.WithWriteLock(func() { without = cond(fresh, fresh.Seats[0].ID) })
	effectstest.PushVanillaCreature(fresh, fresh.Seats[1].ID, "Their Human", 1, 1) // not yours
	fresh.WithWriteLock(func() { without = without || cond(fresh, fresh.Seats[0].ID) })
	effectstest.PushTestPermanent(fresh, fresh.Seats[0].ID, game.Card{Name: "Villager", TypeLine: "Creature — Human Peasant", Power: 1, Toughness: 1})
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
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
			bear := uuid.New()
			me.Hand.PushTop(game.Card{InstanceID: bear, Name: "Bear Spell", TypeLine: "Creature — Bear", ManaCost: tc.cost,
				Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
			if err := g.CastSpell(me.ID, bear, game.CastSpellParams{}); err != nil {
				t.Fatalf("cast the bear: %v", err)
			}
			effectstest.CastCatalogSpell(t, g, "Sound the Trumpets", "Instant", effectstest.SoundTheTrumpetsOracle,
				[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
			for i := 0; i < 4 && effectstest.DiscardChoiceFor(g, me.ID) == nil && !me.Graveyard.Contains(bear); i++ {
				effectstest.PassPriorityAroundTable(t, g)
			}
			if !me.Graveyard.Contains(bear) {
				t.Fatalf("the %s-mana spell was not countered", tc.name)
			}
			if got := effectstest.DiscardChoiceFor(g, me.ID) != nil; got != tc.want {
				t.Fatalf("recruit discard open = %v, want %v", got, tc.want)
			}
		})
	}
}

// One nonland permanent of each opponent is exiled until the enchantment
// leaves, and comes back when it does; the second trigger recruits.
func TestHobCelebrateTheMountainKingExilesUntilItLeavesAndRecruits(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, b, c := g.Seats[0], g.Seats[1], g.Seats[2]
	theirs := effectstest.PushVanillaCreature(g, b.ID, "B's Bear", 2, 2)
	others := effectstest.PushVanillaCreature(g, c.ID, "C's Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	ench := effectstest.CastCatalogSpell(t, g, "Celebrate the Mountain-king", "Enchantment", effectstest.CelebrateTheMountainKingOrac, nil)
	effectstest.PassPriorityAroundTable(t, g)
	soldiers := 0
	for i := 0; i < 16; i++ {
		effectstest.HobAnswerTargets(t, g, me.ID, theirs, others)
		effectstest.HobAnswerOrder(t, g, me.ID)
		if effectstest.DiscardChoiceFor(g, me.ID) != nil {
			soldiers = hobRecruitAnswered(t, g, me)
			continue
		}
		if len(g.StackMeta) == 0 {
			break
		}
		effectstest.PassPriorityAroundTable(t, g)
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
	effectstest.DestroyNow(t, g, ench)
	g.RunStateChecksForTest()
	if effectstest.BattlefieldNamed(g, b.ID, "B's Bear") != 1 || effectstest.BattlefieldNamed(g, c.ID, "C's Bear") != 1 {
		t.Error("the exiled permanents did not return to their owners when the enchantment left")
	}
}

// I recruits, II returns a creature card with mana value 3 or less, III
// puts a +1/+1 counter on a creature.
func TestHobTheMountainKingsReturnChapters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	saga := effectstest.CastCatalogSpell(t, g, "The Mountain-king's Return", "Enchantment — Saga", effectstest.TheMountainKingsReturnOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoreCountersOn(g, saga); got != 1 {
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
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.HobAnswerTargets(t, g, me.ID, dead)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dead) {
		hobDump(t, g)
		t.Fatal("chapter II did not return the creature card")
	}

	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.HobAnswerTargets(t, g, me.ID, dead)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, dead, game.CounterPlusOne); got != 1 {
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
