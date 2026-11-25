package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b288_test.go — Shadows over Innistrad instants and sorceries, S82
// OLD batch 288 (#5543), played through the real catalog entries.

const (
	oldB288Epiphany   = "33befbd6-56ac-45a7-a4d0-448167162707"
	oldB288Dissension = "717266b6-2bbc-4a04-8f30-e3eaf0283acd"
	oldB288Pieces     = "7b599b4d-95bc-4ed2-a167-84a697944747"
	oldB288Welcome    = "b8647f54-f50b-4bc6-bd44-fda25ef189fa"
	oldB288ClipWings  = "cefa28e2-6a95-4aca-86f4-b0b741f61939"
	oldB288Nagging    = "e74c70be-2b70-4b04-9909-161d6f41a156"
)

func TestOldB288EveryCardIsRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{
		oldB288Epiphany: "Epiphany at the Drownyard", oldB288Dissension: "Dissension in the Ranks",
		oldB288Pieces: "Pieces of the Puzzle", oldB288Welcome: "Welcome to the Fold",
		oldB288ClipWings: "Clip Wings", oldB288Nagging: "Nagging Thoughts",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s registration: %+v", name, spec)
		}
	}
}

func TestOldB288EpiphanyRevealsXPlusOneAndTheOpponentPicksThePile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	revealed := effectstest.SeedFactOrFictionLibrary(me, "Top", "Second", "Third", "Fourth", "Fifth")
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "U", "C", "C")
	hand, graves := me.Hand.Size(), me.Graveyard.Size()
	effectstest.CastFromHandForTest(t, g, me, "Epiphany at the Drownyard", "Instant", "{X}{U}", oldB288Epiphany, game.CastSpellParams{Strict: true, XValue: 2})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	split := effectstest.LatestRevealPickFor(g, me.ID)
	if split == nil {
		t.Fatalf("the caster separates the piles: %+v", g.PendingChoices)
	}
	if got := len(split.ChooseCards); got != 3 {
		t.Errorf("%d cards revealed for X = 2, want 3", got)
	}
	if err := g.ResolveRevealPick(split.ID, me.ID, revealed[:1]); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	if effectstest.LatestOptionPickFor(g, opp.ID) == nil {
		t.Fatalf("the named opponent chooses a pile: %+v", g.PendingChoices)
	}
	// Take the pile of two (the second option): one card is binned.
	effectstest.AnswerOptionPick(t, g, opp.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("hand gained %d cards, want 2", got)
	}
	if got := me.Graveyard.Size() - graves; got != 2 {
		t.Errorf("graveyard gained %d, want 2 (one binned card and the spell)", got)
	}
	if !me.Graveyard.Contains(revealed[0]) {
		t.Error("the card left in the other pile is not in the graveyard")
	}
	if me.Library.Size() == 0 || me.Library.Contains(revealed[2]) {
		t.Error("a revealed card is still in the library")
	}
}

func TestOldB288NaggingThoughtsTakesOneOfTheTopTwoAndMadnessCastsItForTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	top := effectstest.TopOfLibrary(me, "Under", "Sorcery", "{1}")
	next := effectstest.TopOfLibrary(me, "Over", "Sorcery", "{1}")
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "U", "C")
	effectstest.CastFromHandForTest(t, g, me, "Nagging Thoughts", "Sorcery", "{1}{U}", oldB288Nagging, game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("no prompt for the take")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{top}); err != nil {
		t.Fatalf("take: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(top) || !me.Graveyard.Contains(next) {
		t.Errorf("hand has the chosen card: %v; graveyard has the other: %v", me.Hand.Contains(top), me.Graveyard.Contains(next))
	}
}

func TestOldB288NaggingThoughtsHasMadness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	card := effectstest.TorB1LootIntoExile(t, g, "Nagging Thoughts", "Sorcery", oldB288Nagging)
	effectstest.TopOfLibrary(me, "Under", "Sorcery", "{1}")
	effectstest.TopOfLibrary(me, "Over", "Sorcery", "{1}")
	effectstest.DcMana(t, g, me.ID, "{C}{U}")
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true, FromZone: "exile", AlternativeCost: game.AltCostKeyMadness}); err != nil {
		t.Fatalf("the madness cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChooseCardsFor(g, me.ID) == nil {
		t.Error("the madness cast did not resolve")
	}
}

func TestOldB288PiecesOfThePuzzleTakesUpToTwoInstantsOrSorceries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	deep := effectstest.TopOfLibrary(me, "Deep Bear", "Creature — Bear", "{1}{G}")
	bolt := effectstest.TopOfLibrary(me, "Bolt", "Instant", "{R}")
	ritual := effectstest.TopOfLibrary(me, "Ritual", "Sorcery", "{1}")
	opt := effectstest.TopOfLibrary(me, "Opt", "Instant", "{U}")
	land := effectstest.TopOfLibrary(me, "Island", "Basic Land — Island", "")
	bear := effectstest.TopOfLibrary(me, "Bear", "Creature — Bear", "{1}{G}")
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "U", "C", "C")
	effectstest.CastFromHandForTest(t, g, me, "Pieces of the Puzzle", "Sorcery", "{2}{U}", oldB288Pieces, game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("no prompt for the take")
	}
	for _, id := range []uuid.UUID{bear, land} {
		for _, offered := range c.ChooseCards {
			if offered == id {
				t.Error("a card that is neither an instant nor a sorcery was offered")
			}
		}
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{opt, bolt, ritual}); err == nil {
		t.Fatal("three cards were accepted for 'up to two'")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{opt, ritual}); err != nil {
		t.Fatalf("take: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{opt, ritual} {
		if !me.Hand.Contains(id) {
			t.Error("a chosen card is not in hand")
		}
	}
	for _, id := range []uuid.UUID{bolt, land, bear} {
		if !me.Graveyard.Contains(id) {
			t.Error("a revealed card that was not taken is not in the graveyard")
		}
	}
	if !me.Library.Contains(deep) {
		t.Error("the sixth card from the top was moved")
	}
}

func TestOldB288DissensionMakesTwoBlockersFight(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a1 := effectstest.B12Creature(g, opp.ID, "Attacker One", "Creature — Bear", 2, 2)
	a2 := effectstest.B12Creature(g, opp.ID, "Attacker Two", "Creature — Bear", 2, 2)
	b1 := effectstest.B12Creature(g, me.ID, "Blocker One", "Creature — Bear", 3, 3)
	b2 := effectstest.B12Creature(g, me.ID, "Blocker Two", "Creature — Bear", 1, 4)
	idle := effectstest.B12Creature(g, me.ID, "Idle", "Creature — Bear", 1, 1)
	effectstest.CrtBlock(g, b1, a1, me.ID)
	effectstest.CrtBlock(g, b2, a2, me.ID)
	if err := effectstest.CastCatalogSpellErr(t, g, "Dissension in the Ranks", "Instant", oldB288Dissension,
		[]game.TargetRef{{Kind: game.TargetCard, ID: b1}, {Kind: game.TargetCard, ID: idle}}); err == nil {
		t.Fatal("a creature that is not blocking was accepted as a target")
	}
	effectstest.CastCatalogSpell(t, g, "Dissension in the Ranks", "Instant", oldB288Dissension,
		[]game.TargetRef{{Kind: game.TargetCard, ID: b1}, {Kind: game.TargetCard, ID: b2}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.Pr6Marked(g, b1); got != 1 {
		t.Errorf("the first blocker has %d damage, want 1", got)
	}
	if got := effectstest.Pr6Marked(g, b2); got != 3 {
		t.Errorf("the second blocker has %d damage, want 3", got)
	}
}

func TestOldB288WelcomeToTheFoldStealsOnlyASmallCreatureUnlessMadnessPaysX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Bear", 3, 3)
	effectstest.B06AddMana(me, "U", "U", "C", "C")
	effectstest.CastFromHandForTest(t, g, me, "Welcome to the Fold", "Sorcery", "{2}{U}{U}", oldB288Welcome, game.CastSpellParams{Strict: true, Targets: effectstest.B16TargetCard(big)})
	effectstest.PassPriorityAroundTable(t, g)
	if _, c := effectstest.CardWhere(g, opp, big); c.Controller != opp.ID {
		t.Error("a creature with toughness 3 was stolen without madness")
	}
	effectstest.B06AddMana(me, "U", "U", "C", "C")
	effectstest.CastFromHandForTest(t, g, me, "Welcome to the Fold", "Sorcery", "{2}{U}{U}", oldB288Welcome, game.CastSpellParams{Strict: true, Targets: effectstest.B16TargetCard(small)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LayeredCard(t, g, small).Controller; got != me.ID {
		t.Error("a creature with toughness 2 was not stolen")
	}
}

func TestOldB288WelcomeToTheFoldMadnessComparesToughnessWithX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Bear", 3, 3)
	card := effectstest.TorB1LootIntoExile(t, g, "Welcome to the Fold", "Sorcery", oldB288Welcome)
	effectstest.DcMana(t, g, me.ID, "{C}{C}{C}{U}{U}")
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true, FromZone: "exile", AlternativeCost: game.AltCostKeyMadness, XValue: 3, Targets: effectstest.B16TargetCard(big)}); err != nil {
		t.Fatalf("the madness cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LayeredCard(t, g, big).Controller; got != me.ID {
		t.Error("a toughness-3 creature was not stolen with X = 3")
	}
}

func TestOldB288ClipWingsMakesEachOpponentSacrificeAFlier(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	flier := effectstest.MirB1Push(g, opp.ID, "Flier", "Creature — Bird", "", 1, 1, "flying")
	ground := effectstest.B12Creature(g, opp.ID, "Ground", "Creature — Bear", 2, 2)
	mine := effectstest.MirB1Push(g, me.ID, "My Flier", "Creature — Bird", "", 1, 1, "flying")
	effectstest.CastCatalogSpell(t, g, "Clip Wings", "Instant", oldB288ClipWings, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(flier) {
		t.Error("the opponent's flier survived")
	}
	if !g.Battlefield.Contains(ground) || !g.Battlefield.Contains(mine) {
		t.Error("a non-flier or the caster's own flier was sacrificed")
	}
}
