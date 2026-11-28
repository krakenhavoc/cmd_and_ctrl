package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b772_cards_test.go — Weatherlight cumulative upkeep cards, S82 OLD batch
// 772 (#5629), played through the real catalog entries. Every permanent belongs
// to seat 1, whose upkeep is the next one the game reaches.

const (
	o772Unicorn      = "18a8b50b-2328-459b-979d-04c4c769e569"
	o772Heart        = "1e1d8c64-4cea-4c2c-8933-be19d13bf07c"
	o772WaveOfTerror = "2346be19-7344-4f11-be2c-a22750cb7e1d"
	o772AncestralKn  = "2786f210-ec64-46ea-9a7f-b03759fb1dae"
	o772Aboroth      = "28f70f86-19a9-4811-bc10-423a05842d39"
	o772InnerSanct   = "2ae36d31-08b7-481a-bfa1-a5c65921dee0"
	o772Ooze         = "42b2c6d0-e0da-41b0-8876-f17dbe5db79a"
	o772PsychicVort  = "451e35dd-e0b5-4157-8511-bacb315e19bd"
	o772Reserves     = "a2dabd52-0b75-42d7-a07d-0df7b38646e2"
)

func TestS82OldB772CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		o772Unicorn: "Revered Unicorn", o772Heart: "Heart of Bogardan", o772WaveOfTerror: "Wave of Terror",
		o772AncestralKn: "Ancestral Knowledge", o772Aboroth: "Aboroth", o772InnerSanct: "Inner Sanctum",
		o772Ooze: "Mwonvuli Ooze", o772PsychicVort: "Psychic Vortex", o772Reserves: "Volunteer Reserves",
	} {
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

// o772Upkeep brings seat 1's next upkeep to its cumulative upkeep prompt and
// answers it: paying with `mana` floated in the pool, or declining.
func o772Upkeep(t *testing.T, g *game.Game, first bool, pay bool, mana ...game.ManaToken) {
	t.Helper()
	if first {
		effectstest.AdvanceToUpkeepOf(t, g, 1)
	} else {
		effectstest.LeaveUpkeepThenAdvanceTo(t, g, 1)
	}
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.RemoraUpkeepPrompt(t, g, g.Seats[1].ID)
	if pay {
		g.Seats[1].ManaPool.AddMana(mana...)
	}
	if err := g.ResolvePayUnless(prompt.ID, g.Seats[1].ID, pay); err != nil {
		t.Fatalf("answer the upkeep prompt (pay=%v): %v", pay, err)
	}
}

func o772C(n int) []game.ManaToken {
	out := make([]game.ManaToken, n)
	for i := range out {
		out[i] = game.ManaToken{Color: "C"}
	}
	return out
}

func TestS82OldB772RevereUnicornGainsLifeForItsAgeCountersWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	uni := effectstest.B12Push(g, me.ID, "Revered Unicorn", "Creature — Unicorn", o772Unicorn, 2, 3)
	o772Upkeep(t, g, true, true, o772C(1)...)
	if !g.Battlefield.Contains(uni) {
		t.Fatal("a paid upkeep sacrificed the Unicorn")
	}
	life := me.Life
	// The second upkeep costs {1}{1}: decline it, and the Unicorn goes with two counters.
	o772Upkeep(t, g, false, false)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(uni) {
		t.Fatal("a declined upkeep did not sacrifice the Unicorn")
	}
	if me.Life != life+2 {
		t.Errorf("life is %d, want %d (two age counters)", me.Life, life+2)
	}
}

func TestS82OldB772HeartOfBogardanDamagesAPlayerAndTheirCreaturesForTwiceTheCountersLessTwo(t *testing.T) {
	for _, tc := range []struct{ prior, damage int }{{0, 0}, {2, 4}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[1], g.Seats[0]
		heart := effectstest.PushCatalogPermanent(g, me.ID, "Heart of Bogardan", "Enchantment", o772Heart, false)
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == heart && tc.prior > 0 {
					if g.Battlefield.Cards[i].Counters == nil {
						g.Battlefield.Cards[i].Counters = map[string]int{}
					}
					g.Battlefield.Cards[i].Counters[game.CounterAge] = tc.prior
				}
			}
		})
		mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
		small := effectstest.B12Creature(g, opp.ID, "Their Rat", "Creature — Rat", 1, 1)
		big := effectstest.B12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 9)
		life := opp.Life
		o772Upkeep(t, g, true, false)
		if tc.damage > 0 {
			effectstest.AnswerTriggerTargets2(t, g, me.ID, opp.ID)
		} else if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
			effectstest.AnswerTriggerTargets2(t, g, me.ID, opp.ID)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(heart) {
			t.Errorf("prior %d: the Heart was not sacrificed", tc.prior)
		}
		if opp.Life != life-tc.damage {
			t.Errorf("prior %d: the target is at %d, want %d", tc.prior, opp.Life, life-tc.damage)
		}
		if got := effectstest.DamageMarkedOn(g, big); got != tc.damage {
			t.Errorf("prior %d: their 0/9 has %d damage, want %d", tc.prior, got, tc.damage)
		}
		if tc.damage > 0 && g.Battlefield.Contains(small) {
			t.Errorf("prior %d: their 1/1 survived", tc.prior)
		}
		if got := effectstest.DamageMarkedOn(g, mine); got != 0 {
			t.Errorf("prior %d: the Heart's controller's creature took %d", tc.prior, got)
		}
	}
}

func TestS82OldB772WaveOfTerrorDestroysCreaturesWithManaValueEqualToItsCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Wave of Terror", "Enchantment", o772WaveOfTerror, false)
	one := effectstest.B12Creature(g, opp.ID, "One Drop", "Creature — Bear", 1, 1)
	two := effectstest.B12Creature(g, opp.ID, "Two Drop", "Creature — Bear", 2, 2)
	mineOne := effectstest.B12Creature(g, me.ID, "My One Drop", "Creature — Bear", 1, 1)
	token := effectstest.B12Creature(g, opp.ID, "Zero Drop", "Creature — Bear", 1, 1)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			switch g.Battlefield.Cards[i].InstanceID {
			case one, mineOne:
				g.Battlefield.Cards[i].ManaCost = "{1}"
			case two:
				g.Battlefield.Cards[i].ManaCost = "{1}{1}"
			}
		}
	})
	o772Upkeep(t, g, true, true, o772C(1)...)
	effectstest.AdvanceToDrawStepOfSeat(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(one) || g.Battlefield.Contains(mineOne) {
		t.Error("a creature with mana value 1 survived one age counter")
	}
	if !g.Battlefield.Contains(two) || !g.Battlefield.Contains(token) {
		t.Error("a creature of another mana value was destroyed")
	}
}

func TestS82OldB772AncestralKnowledgeExilesAnyNumberOfTheTopTenAndShufflesWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	top := make([]uuid.UUID, 0, 10)
	g.WithWriteLock(func() {
		me.Library.Cards = nil
		for i := 0; i < 30; i++ {
			c := game.Card{InstanceID: uuid.New(), Name: "Lib " + string(rune('A'+i%26)) + string(rune('a'+i/26)), TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID}
			me.Library.Cards = append(me.Library.Cards, c)
		}
		// Index 0 is the bottom; the top ten are the last ten cards.
		for i := len(me.Library.Cards) - 1; i >= len(me.Library.Cards)-10; i-- {
			top = append(top, me.Library.Cards[i].InstanceID)
		}
	})
	aKn := effectstest.CastCatalogSpell(t, g, "Ancestral Knowledge", "Enchantment", o772AncestralKn, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("the controller was not asked which cards to exile")
	}
	if len(pick.ChooseCards) != 10 {
		t.Fatalf("the pick offers %d cards, want the top ten", len(pick.ChooseCards))
	}
	effectstest.AnswerChooseCards(t, g, me.ID, top[0], top[1], top[2])
	for _, id := range top[:3] {
		if !effectstest.InExile(g, id) {
			t.Errorf("card %s was not exiled", id)
		}
	}
	order := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if order == nil {
		t.Fatal("the rest were not offered back in any order")
	}
	if len(me.Library.Cards) != 27 {
		t.Errorf("the library holds %d cards, want 27", len(me.Library.Cards))
	}
	_ = aKn
}

func TestS82OldB772AncestralKnowledgeShufflesWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	aKn := effectstest.PushCatalogPermanent(g, me.ID, "Ancestral Knowledge", "Enchantment", o772AncestralKn, false)
	var before []uuid.UUID
	g.WithWriteLock(func() {
		for i := range me.Library.Cards {
			before = append(before, me.Library.Cards[i].InstanceID)
		}
	})
	if len(before) < 10 {
		t.Fatalf("test library too small: %d", len(before))
	}
	effectstest.B14Kill(g, aKn)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	same := true
	g.WithWriteLock(func() {
		for i := range me.Library.Cards {
			if me.Library.Cards[i].InstanceID != before[i] {
				same = false
			}
		}
	})
	if same {
		t.Error("the library order is unchanged after the enchantment left")
	}
}

func TestS82OldB772AborothPutsACounterOnItselfPerAgeCounterOrIsSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	abo := effectstest.B12Push(g, me.ID, "Aboroth", "Creature — Elemental", o772Aboroth, 9, 9)
	o772Upkeep(t, g, true, true)
	if p, tough := effectstest.PtOf(t, g, abo); p != 8 || tough != 8 {
		t.Errorf("after one payment the Aboroth is %d/%d, want 8/8", p, tough)
	}
	o772Upkeep(t, g, false, true)
	if p, tough := effectstest.PtOf(t, g, abo); p != 6 || tough != 6 {
		t.Errorf("after the second payment (two counters at once) the Aboroth is %d/%d, want 6/6", p, tough)
	}
	o772Upkeep(t, g, false, false)
	if g.Battlefield.Contains(abo) {
		t.Error("a declined upkeep did not sacrifice the Aboroth")
	}
}

func TestS82OldB772InnerSanctumPaysLifeAndPreventsDamageToYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Inner Sanctum", "Enchantment", o772InnerSanct, false)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	life := me.Life
	o772Upkeep(t, g, true, true)
	if me.Life != life-2 {
		t.Errorf("life is %d, want %d (2 life for one age counter)", me.Life, life-2)
	}
	g.WithWriteLock(func() {
		_ = g.DealDamageToCreatureForEffect(opp.ID, mine, 3)
		_ = g.DealDamageToCreatureForEffect(me.ID, theirs, 3)
	})
	g.RunStateChecksForTest()
	if !g.Battlefield.Contains(mine) || effectstest.DamageMarkedOn(g, mine) != 0 {
		t.Error("damage to the controller's creature was not prevented")
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("damage to an opponent's creature was prevented")
	}
}

func TestS82OldB772MwonvuliOozeGrowsWithItsAgeCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	ooze := effectstest.B12Push(g, me.ID, "Mwonvuli Ooze", "Creature — Ooze", o772Ooze, 0, 0)
	if p, tough := effectstest.PtOf(t, g, ooze); p != 1 || tough != 1 {
		t.Errorf("with no counters the Ooze is %d/%d, want 1/1", p, tough)
	}
	o772Upkeep(t, g, true, true, o772C(2)...)
	if p, tough := effectstest.PtOf(t, g, ooze); p != 3 || tough != 3 {
		t.Errorf("with one counter the Ooze is %d/%d, want 3/3", p, tough)
	}
	o772Upkeep(t, g, false, true, o772C(4)...)
	if p, tough := effectstest.PtOf(t, g, ooze); p != 5 || tough != 5 {
		t.Errorf("with two counters the Ooze is %d/%d, want 6/6", p, tough)
	}
}

func TestS82OldB772PsychicVortexDrawsToPayAndSacrificesALandAndDiscardsAtEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	vortex := effectstest.PushCatalogPermanent(g, me.ID, "Psychic Vortex", "Enchantment", o772PsychicVort, false)
	land := effectstest.PushCatalogPermanent(g, me.ID, "Island", "Basic Land — Island", "", false)
	hand := me.Hand.Size()
	o772Upkeep(t, g, true, true)
	if !g.Battlefield.Contains(vortex) {
		t.Fatal("a paid upkeep sacrificed the Vortex")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("the hand holds %d cards, want %d (drew to pay)", me.Hand.Size(), hand+1)
	}
	effectstest.AdvanceToEndStepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, me.ID, land)
	if g.Battlefield.Contains(land) {
		t.Error("the only land was not sacrificed")
	}
	if me.Hand.Size() != 0 {
		t.Errorf("the hand holds %d cards, want 0", me.Hand.Size())
	}
}

func TestS82OldB772VolunteerReservesHasBandingAndPaysOrIsSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	res := effectstest.B12Push(g, me.ID, "Volunteer Reserves", "Creature — Human Soldier", o772Reserves, 2, 4)
	if !effectstest.HasEffectiveKeyword(t, g, res, game.KeywordBanding) {
		t.Error("the Reserves have no banding")
	}
	o772Upkeep(t, g, true, true, o772C(1)...)
	if !g.Battlefield.Contains(res) {
		t.Fatal("a paid upkeep sacrificed the Reserves")
	}
	o772Upkeep(t, g, false, false)
	if g.Battlefield.Contains(res) {
		t.Error("a declined upkeep did not sacrifice the Reserves")
	}
}
