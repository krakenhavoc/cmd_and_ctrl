package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// the_hobbit_storied_cards_test.go — the storied cards of The Hobbit
// (#4475, CR 702.195). The designation itself is pinned in
// game/enduring_story_test.go; these pin each card's reader.

const (
	oriKeeperOfSongsOracle  = "69cc089a-abd3-4edb-bc79-f02daac70837"
	oinTheBraveOracle       = "a25ba5cc-6ff8-4215-b70d-7e86e15e3c1a"
	bomburGentleDreamerOrac = "925ce68e-9323-4985-8fef-0d3f36e0d74c"
	dainLordOfTheIronHills  = "5b79b68a-db71-4638-824e-ab89b1716c86"
	thorinOakenshieldOracle = "bdd41af0-bbd1-4ecd-a699-99f006f5e5ce"
	filiThePathfinderOracle = "c91c7d6b-a502-4b66-9509-3c1eda00e763"
	balinLoremasterOracle   = "0d420e41-43e9-41d6-832c-5a9f410c994e"
	bifurMelodicRiderOracle = "b8d563e4-e2bc-4e8b-8841-6655beff9138"
)

// giveStory gives a player an enduring story the way the engine does
// (the flag and its event, which invalidates the layer pass), without a
// board of three artifacts.
func giveStory(g *game.Game, p *game.Player) {
	g.WithWriteLock(func() {
		p.EnduringStory = true
		g.EmitEvent(game.Event{Kind: game.EventEnduringStory, Actor: p.ID})
	})
}

func TestHobEveryStoriedCardDeclaresTheKeywordAndIsFull(t *testing.T) {
	for name, oracle := range map[string]string{
		"Ori": oriKeeperOfSongsOracle, "Óin": oinTheBraveOracle, "Bombur": bomburGentleDreamerOrac,
		"Dáin": dainLordOfTheIronHills, "Thorin": thorinOakenshieldOracle, "Fíli": filiThePathfinderOracle,
		"Balin": balinLoremasterOracle, "Bifur": bifurMelodicRiderOracle,
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", name)
			continue
		}
		if !hasKeyword(spec.PrintedKeywords, game.KeywordStoried) {
			t.Errorf("%s does not declare storied", name)
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s completeness = %v caveats %v, want full with none", name, spec.Completeness, spec.Caveats)
		}
	}
}

// Ori earns the story for real: he is legendary, so two artifacts make
// three; and then he is 4/3 with vigilance.
func TestHobOriKeeperOfSongsIsBiggerAndVigilantWithAStory(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ori := pushTestPermanent(g, me.ID, game.Card{Name: "Ori, Keeper of Songs", OracleID: oriKeeperOfSongsOracle,
		TypeLine: "Legendary Creature — Dwarf Bard", Power: 3, Toughness: 3})
	pushTypedCard(g, me.ID, "Relic", "Artifact", "")
	g.RunStateChecksForTest()
	if me.EnduringStory {
		t.Fatal("Ori and one artifact gave an enduring story")
	}
	if p, tough := effectivePTOf(t, g, ori); p != 3 || tough != 3 || hasEffectiveKeyword(t, g, ori, "vigilance") {
		t.Errorf("without a story Ori is %d/%d vigilance=%v, want 3/3 without", p, tough, hasEffectiveKeyword(t, g, ori, "vigilance"))
	}
	pushTypedCard(g, me.ID, "Relic", "Artifact", "")
	g.RunStateChecksForTest()
	if !me.EnduringStory {
		t.Fatal("Ori (legendary) and two artifacts gave no enduring story")
	}
	if p, tough := effectivePTOf(t, g, ori); p != 4 || tough != 3 || !hasEffectiveKeyword(t, g, ori, "vigilance") {
		t.Errorf("with a story Ori is %d/%d vigilance=%v, want 4/3 with", p, tough, hasEffectiveKeyword(t, g, ori, "vigilance"))
	}
}

// Óin: +1/+0 and haste with a story; his loot ability pays {1}, taps and
// discards, then draws.
func TestHobOinTheBraveIsHastyWithAStoryAndLoots(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oin := pushTestPermanent(g, me.ID, game.Card{Name: "Óin the Brave", OracleID: oinTheBraveOracle,
		TypeLine: "Legendary Creature — Dwarf Warrior", Power: 1, Toughness: 3})
	if p, _ := effectivePTOf(t, g, oin); p != 1 || hasEffectiveKeyword(t, g, oin, "haste") {
		t.Errorf("without a story Óin has power %d, haste %v; want 1, false", p, hasEffectiveKeyword(t, g, oin, "haste"))
	}
	giveStory(g, me)
	if p, tough := effectivePTOf(t, g, oin); p != 2 || tough != 3 || !hasEffectiveKeyword(t, g, oin, "haste") {
		t.Errorf("with a story Óin is %d/%d haste=%v, want 2/3 with", p, tough, hasEffectiveKeyword(t, g, oin, "haste"))
	}

	advanceTo(t, g, game.StepPrecombatMain)
	var pitched uuid.UUID
	g.WithWriteLock(func() { pitched = handCardFull(me, "Pitched", "Sorcery", "{1}", "", nil) })
	handBefore := me.Hand.Size()
	fillPool(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, oin, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{pitched}}); err != nil {
		t.Fatalf("activate Óin: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(pitched) || me.Hand.Size() != handBefore || !isTapped(g, oin) {
		t.Errorf("after the loot: discarded %v, hand %d (want %d), tapped %v", me.Graveyard.Contains(pitched), me.Hand.Size(), handBefore, isTapped(g, oin))
	}
}

// Bombur stays tapped through your untap step until you have a story.
func TestHobBomburUntapsOnlyWithAStory(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bombur := pushTestPermanent(g, me.ID, game.Card{Name: "Bombur, Gentle Dreamer", OracleID: bomburGentleDreamerOrac,
		TypeLine: "Legendary Creature — Dwarf Bard", Power: 5, Toughness: 3, Tapped: true})
	advanceToMainOf(t, g, 0)
	aroundTheTableToMain(t, g)
	if !isTapped(g, bombur) {
		t.Fatal("Bombur untapped without an enduring story")
	}
	giveStory(g, me)
	aroundTheTableToMain(t, g)
	if isTapped(g, bombur) {
		t.Error("Bombur did not untap in your untap step with an enduring story")
	}
}

// Dáin: creatures attacking you cost {1} each, only with the story.
func TestHobDainTaxesAttackersOnlyWithAStory(t *testing.T) {
	g, bear, opp := taxTable(t)
	other := pushVanillaCreature(g, g.Seats[0].ID, "Second Bear", 2, 2)
	pushTestPermanent(g, opp.ID, game.Card{Name: "Dáin, Lord of the Iron Hills", OracleID: dainLordOfTheIronHills,
		TypeLine: "Legendary Creature — Dwarf Noble", Power: 2, Toughness: 2})
	attack := []game.AttackDeclaration{{Attacker: bear, Target: opp.ID}, {Attacker: other, Target: opp.ID}}
	if price := g.PriceAttackDeclaration(attack); price.Total.Generic != 0 {
		t.Errorf("without a story the attack costs %d, want 0", price.Total.Generic)
	}
	giveStory(g, opp)
	if price := g.PriceAttackDeclaration(attack); price.Total.Generic != 2 {
		t.Errorf("with a story two attackers cost %d, want 2", price.Total.Generic)
	}
	// Another player's story does not tax attacks on Dáin's controller's
	// opponents, and Dáin taxes only attacks on his controller.
	third := g.Seats[2]
	if price := g.PriceAttackDeclaration([]game.AttackDeclaration{{Attacker: bear, Target: third.ID}}); price.Total.Generic != 0 {
		t.Errorf("an attack on another player costs %d, want 0", price.Total.Generic)
	}
}

// Thorin: with a story, an opponent's spell targeting your artifact or
// creature triggers the ward; without one it does not.
func TestHobThorinGivesWardOnlyWithAStory(t *testing.T) {
	for _, story := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		thorin := pushTestPermanent(g, me.ID, game.Card{Name: "Thorin Oakenshield", OracleID: thorinOakenshieldOracle,
			TypeLine: "Legendary Creature — Dwarf Noble", Power: 3, Toughness: 2})
		bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
		if story {
			giveStory(g, me)
		}
		for g.Turn.Step != game.StepPrecombatMain {
			if _, err := g.AdvanceStep(); err != nil {
				t.Fatalf("AdvanceStep: %v", err)
			}
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
		castAtWardedCreature(t, g, opp, bear)
		if got := triggerOnStack(g, thorin) != nil; got != story {
			t.Errorf("story %v: ward trigger on the stack = %v, want %v", story, got, story)
		}
	}
}

// Fíli: he and each nontoken Dwarf that enters under you make a 2/2 red
// Dwarf; the tokens make none; the story pumps every creature you
// control.
func TestHobFiliMakesDwarvesAndPumpsWithAStory(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fili := hobEnter(t, g, me.ID, game.Card{Name: "Fíli the Pathfinder", OracleID: filiThePathfinderOracle,
		TypeLine: "Legendary Creature — Dwarf Scout", Power: 2, Toughness: 2})
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Dwarf"); n != 1 {
		t.Fatalf("%d Dwarf tokens after Fíli entered, want 1", n)
	}
	hobEnter(t, g, me.ID, game.Card{Name: "Some Dwarf", TypeLine: "Creature — Dwarf Warrior", Power: 1, Toughness: 1})
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Dwarf"); n != 2 {
		t.Fatalf("%d Dwarf tokens after a nontoken Dwarf entered, want 2 (and none for the tokens)", n)
	}
	hobEnter(t, g, me.ID, game.Card{Name: "Some Elf", TypeLine: "Creature — Elf", Power: 1, Toughness: 1})
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Dwarf"); n != 2 {
		t.Errorf("%d Dwarf tokens after an Elf entered, want 2", n)
	}
	if p, _ := effectivePTOf(t, g, fili); p != 2 {
		t.Errorf("Fíli's power without a story = %d, want 2", p)
	}
	giveStory(g, me)
	if p, tough := effectivePTOf(t, g, fili); p != 3 || tough != 3 {
		t.Errorf("Fíli with a story is %d/%d, want 3/3", p, tough)
	}
}

// Balin: on a yes, discard the hand and draw that many; with a story he
// deals that much to each opponent.
func TestHobBalinDiscardsDrawsAndBurnsWithAStory(t *testing.T) {
	for _, story := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		if story {
			giveStory(g, me)
		}
		g.WithWriteLock(func() {
			me.Hand.Cards = nil
			for range 3 {
				handCardFull(me, "Old Card", "Sorcery", "{1}", "", nil)
			}
		})
		before := map[uuid.UUID]int{}
		for _, p := range g.Seats[1:] {
			before[p.ID] = p.Life
		}
		graveBefore := me.Graveyard.Size()
		hobEnter(t, g, me.ID, game.Card{Name: "Balin, Loremaster", OracleID: balinLoremasterOracle,
			TypeLine: "Legendary Creature — Dwarf Bard", Power: 4, Toughness: 4})
		passPriorityAroundTable(t, g)
		answerLatestTriggerPrompt(t, g, me.ID, true)
		passPriorityAroundTable(t, g)
		if got := me.Graveyard.Size() - graveBefore; got != 3 {
			t.Errorf("story %v: %d cards discarded, want 3", story, got)
		}
		if got := me.Hand.Size(); got != 3 {
			t.Errorf("story %v: hand %d after drawing X, want 3", story, got)
		}
		want := 0
		if story {
			want = 3
		}
		for _, p := range g.Seats[1:] {
			if got := before[p.ID] - p.Life; got != want {
				t.Errorf("story %v: %s lost %d, want %d", story, p.Name, got, want)
			}
		}
	}
}

// Bifur: with a story, a Dwarf's triggered ability (his own entry
// trigger here) triggers an additional time; without one, once.
func TestHobBifurDoublesDwarfTriggersWithAStory(t *testing.T) {
	for _, story := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
		if story {
			giveStory(g, me)
		}
		bifur := hobEnter(t, g, me.ID, game.Card{Name: "Bifur, Melodic Rider", OracleID: bifurMelodicRiderOracle,
			TypeLine: "Legendary Creature — Dwarf Bard", Power: 4, Toughness: 5})
		for i := 0; i < 6; i++ {
			hobAnswerTargets(t, g, me.ID, bear)
			passPriorityAroundTable(t, g)
		}
		want := 1
		if story {
			want = 2
		}
		if got := counterCount(g, bear, game.CounterPlusOne); got != want {
			t.Errorf("story %v: %d +1/+1 counters from Bifur's entry, want %d", story, got, want)
		}
		_ = bifur
	}
}
