package testss82m

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	oldB550Sprite    = "4c5b2be9-5d48-4e62-ae67-d61f3708bb4f"
	oldB550Harasser  = "5c0c3f08-4f65-4b04-8bfc-1532160f3e28"
	oldB550Goatnaper = "6875839a-9815-4ed2-adcf-b4d097c9bafd"
	oldB550Boggart   = "8c07878f-348c-4fd9-955f-87e975ed61fa"
	oldB550Oak       = "98d3d1f1-ac72-4b1e-87da-b4c1e4f03d08"
	oldB550Ashling   = "b68bc097-482a-4fe5-9fca-df5941f3dc5c"
	oldB550Divers    = "dbdf3c82-b0fc-479c-acce-f7e60be3a2b1"
	oldB550Spiderwig = "dd0bbf91-1826-42c7-82af-6820646fa003"
)

func TestOldB550CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB550Sprite: "Thieving Sprite", oldB550Harasser: "Hornet Harasser", oldB550Goatnaper: "Goatnapper",
		oldB550Boggart: "Exiled Boggart", oldB550Oak: "Battlewand Oak", oldB550Ashling: "Ashling's Prerogative",
		oldB550Divers: "Inkfathom Divers", oldB550Spiderwig: "Spiderwig Boggart",
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

// oldB550Enter puts a card with a printed mana cost onto the battlefield from
// its owner's hand through the entry pipeline, so replacements see it.
func oldB550Enter(t *testing.T, g *game.Game, owner *game.Player, name, typeLine, cost string) uuid.UUID {
	t.Helper()
	c := game.NewCard(name, owner.ID)
	c.TypeLine, c.ManaCost, c.Power, c.Toughness = typeLine, cost, 1, 1
	owner.Hand.PushTop(c)
	var id uuid.UUID
	g.WithWriteLock(func() {
		var err error
		id, err = g.PutFromHandOntoBattlefieldForEffect(c.InstanceID, game.HandEntryOptions{Controller: owner.ID})
		if err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	return id
}

func TestOldB550ThievingSpriteRevealsAsManyCardsAsYouHaveFaeries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Other Faerie", "Kindred Enchantment — Faerie", "", 0, 0) // a Faerie permanent that is no creature counts
	effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() { opp.Hand.Cards = nil })
	var hand []uuid.UUID
	for i := 1; i <= 4; i++ {
		id := uuid.New()
		opp.Hand.PushTop(game.Card{InstanceID: id, Name: fmt.Sprintf("Their Card %d", i), TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
		hand = append(hand, id)
	}
	effectstest.TdCastFromHand(t, g, game.Card{Name: "Thieving Sprite", TypeLine: "Creature — Faerie Rogue", OracleID: oldB550Sprite, Power: 1, Toughness: 1}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	ask := effectstest.ChooseCardsChoiceFor(g, opp.ID)
	if ask == nil || ask.ChooseMin != 2 || ask.ChooseMax != 2 {
		t.Fatalf("the target should reveal exactly two cards (the Sprite and the other Faerie), got %+v", ask)
	}
	effectstest.AnswerChooseCards(t, g, opp.ID, hand[0], hand[1])
	rp := effectstest.RevealPickFor(g, me.ID)
	if rp == nil || len(rp.ChooseCards) != 2 {
		t.Fatalf("want two revealed cards to pick from, got %+v", rp)
	}
	if err := g.ResolveRevealPick(rp.ID, me.ID, []uuid.UUID{hand[1]}); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Contains(hand[1]) || !opp.Graveyard.Contains(hand[1]) {
		t.Error("the chosen card was not discarded")
	}
	if !opp.Hand.Contains(hand[0]) || len(opp.Hand.Cards) != 3 {
		t.Errorf("the other cards were touched (hand %d)", len(opp.Hand.Cards))
	}
}

func TestOldB550HornetHarasserShrinksATargetWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	harasser := effectstest.B12Push(g, me.ID, "Hornet Harasser", "Creature — Goblin Shaman", oldB550Harasser, 2, 2)
	victim := effectstest.B12Push(g, opp.ID, "Ogre", "Creature — Ogre", "", 3, 3)
	effectstest.B25Destroy(g, harasser)
	effectstest.AnswerTriggerTargets(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, victim); got != 1 {
		t.Errorf("the Ogre's power is %d, want 1", got)
	}
	if !g.Battlefield.Contains(victim) {
		t.Error("a 3/3 under -2/-2 should survive as a 1/1")
	}
}

func TestOldB550ExiledBoggartMakesYouDiscardWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	boggart := effectstest.B12Push(g, me.ID, "Exiled Boggart", "Creature — Goblin Rogue", oldB550Boggart, 2, 2)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	for i := range 2 {
		me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: fmt.Sprintf("Card %d", i), TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	}
	effectstest.B25Destroy(g, boggart)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCardsOrDiscard(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != 1 {
		t.Errorf("hand %d, want 1", len(me.Hand.Cards))
	}
}

func TestOldB550InkfathomDiversLooksAtTheTopFourOnEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	for i := range 6 {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: fmt.Sprintf("Lib %d", i), TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	divers := effectstest.MrdEnter(t, g, me, "Inkfathom Divers", "Creature — Merfolk Soldier", oldB550Divers)
	if !effectstest.HasKeywordOf(g, divers, "islandwalk") {
		t.Error("the Divers lack islandwalk")
	}
	effectstest.PassPriorityAroundTable(t, g)
	open := false
	for _, c := range g.PendingChoices {
		if c != nil && c.Chooser == me.ID {
			open = true
		}
	}
	if !open {
		t.Error("no prompt to order the top four cards")
	}
}

func TestOldB550GoatnapperTakesAGoatUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	goat := effectstest.B12Push(g, opp.ID, "Goat", "Creature — Goat", "", 0, 1)
	bear := effectstest.B12Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.TapForTest(t, g, goat)
	effectstest.MrdEnter(t, g, me, "Goatnapper", "Creature — Goblin Rogue", oldB550Goatnaper)
	opts := effectstest.LatestPickTarget(g, me.ID)
	if opts == nil {
		t.Fatal("the Goatnapper did not ask for a target")
	}
	for _, id := range opts.PickTargetCards {
		if id == bear {
			t.Error("a Bear was offered as a Goat")
		}
	}
	effectstest.AnswerTriggerTargets(t, g, me.ID, goat)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ControllerOf(t, g, goat); got != me.ID {
		t.Errorf("the Goat is controlled by %v, want the Goatnapper's controller", got)
	}
	if effectstest.FindBattlefieldCardForTest(g, goat).Tapped {
		t.Error("the Goat was not untapped")
	}
	if !effectstest.HasKeywordOf(g, goat, "haste") {
		t.Error("the Goat did not gain haste")
	}
}

func TestOldB550SpiderwigBoggartGivesFearUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.MrdEnter(t, g, me, "Spiderwig Boggart", "Creature — Goblin Shaman", oldB550Spiderwig)
	effectstest.AnswerTriggerTargets(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOf(g, bear, "fear") {
		t.Error("the Bear did not gain fear")
	}
}

func TestOldB550BattlewandOakGrowsForForestsAndTreefolkSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	oak := effectstest.B12Push(g, me.ID, "Battlewand Oak", "Creature — Treefolk Warrior", oldB550Oak, 1, 3)
	if got := effectstest.EffectivePower(t, g, oak); got != 1 {
		t.Fatalf("power %d, want 1", got)
	}
	oldB550Enter(t, g, me, "Mountain", "Basic Land — Mountain", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, oak); got != 1 {
		t.Errorf("a Mountain grew the Oak to %d", got)
	}
	oldB550Enter(t, g, opp, "Their Forest", "Basic Land — Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, oak); got != 1 {
		t.Errorf("an opponent's Forest grew the Oak to %d", got)
	}
	oldB550Enter(t, g, me, "Forest", "Basic Land — Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, oak); got != 3 {
		t.Errorf("a Forest left the Oak at %d, want 3", got)
	}
	effectstest.TdCastFromHand(t, g, game.Card{Name: "Tree Spell", TypeLine: "Instant — Treefolk", ManaCost: "", Colors: []string{"G"}}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, oak); got != 5 {
		t.Errorf("a Treefolk spell left the Oak at %d, want 5", got)
	}
	effectstest.TdCastFromHand(t, g, game.Card{Name: "Bear Spell", TypeLine: "Instant — Bear", Colors: []string{"G"}}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, oak); got != 5 {
		t.Errorf("a non-Treefolk spell grew the Oak to %d", got)
	}
}

func TestOldB550AshlingsPrerogativeGivesHasteAndTapsByParity(t *testing.T) {
	for _, odd := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		word := "Even"
		if odd {
			word = "Odd"
		}
		effectstest.CastSiege(t, g, "Ashling's Prerogative", oldB550Ashling, word)
		oddCost, evenCost := "{1}", "{2}"
		hasteCost, tapCost := evenCost, oddCost
		if odd {
			hasteCost, tapCost = oddCost, evenCost
		}
		for _, owner := range []*game.Player{me, opp} {
			fast := oldB550Enter(t, g, owner, "Fast", "Creature — Bear", hasteCost)
			slow := oldB550Enter(t, g, owner, "Slow", "Creature — Bear", tapCost)
			if effectstest.FindBattlefieldCardForTest(g, fast).Tapped {
				t.Errorf("%s: a creature of the chosen quality entered tapped", word)
			}
			if !effectstest.FindBattlefieldCardForTest(g, slow).Tapped {
				t.Errorf("%s: a creature without the chosen quality entered untapped", word)
			}
			if !effectstest.HasKeywordOf(g, fast, "haste") {
				t.Errorf("%s: a creature of the chosen quality lacks haste", word)
			}
			if effectstest.HasKeywordOf(g, slow, "haste") {
				t.Errorf("%s: a creature without the chosen quality has haste", word)
			}
		}
		land := oldB550Enter(t, g, me, "Plains", "Basic Land — Plains", "")
		if effectstest.FindBattlefieldCardForTest(g, land).Tapped {
			t.Errorf("%s: a land entered tapped", word)
		}
		// A creature token has mana value 0: even.
		zero := oldB550Enter(t, g, me, "Zero", "Creature — Bear", "")
		if got := effectstest.FindBattlefieldCardForTest(g, zero).Tapped; got != odd {
			t.Errorf("%s: a mana value 0 creature tapped = %v, want %v", word, got, odd)
		}
	}
}
