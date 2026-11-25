package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b262_cards_test.go — Kaladesh, batch 262 (#5539). Each test plays
// the real catalog entry through the engine and asserts what the printed text
// says.

const (
	oldB262Crane    = "1983478a-cedb-4053-adab-ac5f1a163296"
	oldB262Gremlins = "4766f232-8275-4256-b756-8dd5c5e825a3"
	oldB262Kingpin  = "680da572-6cd5-482a-b01d-32944566d8c4"
	oldB262Herald   = "6bb7d985-b033-46a2-9bbb-6f7ee62aaaff"
	oldB262Mechanic = "b47d37ee-2bb2-4ca8-b245-94d2cc4b6684"
	oldB262Bruiser  = "f432998f-7737-4f41-86b5-a4372bc5db2b"
)

func TestOldB262AllRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{
		oldB262Crane:    "Glint-Nest Crane",
		oldB262Gremlins: "Salivating Gremlins",
		oldB262Kingpin:  "Contraband Kingpin",
		oldB262Herald:   "Herald of the Fair",
		oldB262Mechanic: "Aviary Mechanic",
		oldB262Bruiser:  "Embraal Bruiser",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s registration: %+v", name, spec)
		}
	}
}

func TestOldB262PrintedKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for _, tc := range []struct {
		name, typeLine, oracle, kw string
	}{
		{"Glint-Nest Crane", "Creature — Bird", oldB262Crane, "flying"},
		{"Contraband Kingpin", "Creature — Aetherborn Rogue", oldB262Kingpin, "lifelink"},
	} {
		id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 1, 3)
		if !effectstest.HasEffectiveKeyword(t, g, id, tc.kw) {
			t.Errorf("%s lacks %s", tc.name, tc.kw)
		}
	}
}

func TestOldB262GlintNestCraneTakesAnArtifactFromTheTopFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lib := effectstest.DftMPLibraryTop(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2},
		game.Card{Name: "Rock", TypeLine: "Artifact"},
		game.Card{Name: "Wand", TypeLine: "Artifact — Equipment"},
		game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Fifth", TypeLine: "Artifact"},
	)
	before := me.Library.Size()
	handBefore := me.Hand.Size()
	effectstest.CastRingCreature(t, g, "Glint-Nest Crane", "Creature — Bird", oldB262Crane, 1, 3)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, lib[2])
	oldB262OrderTheBottom(t, g, me, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(lib[2]) {
		t.Fatal("the chosen artifact is not in hand")
	}
	// The hand gained the artifact and lost the Crane it cast.
	if me.Hand.Size() != handBefore+1 {
		t.Errorf("hand = %d, want %d", me.Hand.Size(), handBefore+1)
	}
	if me.Library.Size() != before-1 {
		t.Errorf("library = %d, want %d", me.Library.Size(), before-1)
	}
	bottom := effectstest.LibraryBottomIDs(me, 3)
	for _, i := range []int{0, 1, 3} {
		if me.Hand.Contains(lib[i]) || !effectstest.HasID(bottom, lib[i]) {
			t.Errorf("card %d should be on the library's bottom", i)
		}
	}
	// The fifth card was never looked at: it is now the top of the library.
	if top := me.Library.Cards[me.Library.Size()-1].InstanceID; top != lib[4] {
		t.Errorf("the top card is not the unseen fifth card")
	}
}

func TestOldB262GlintNestCraneOnlyOffersArtifactsAndMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lib := effectstest.DftMPLibraryTop(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2},
		game.Card{Name: "Rock", TypeLine: "Artifact"},
		game.Card{Name: "Spell", TypeLine: "Instant"},
		game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
	)
	effectstest.CastRingCreature(t, g, "Glint-Nest Crane", "Creature — Bird", oldB262Crane, 1, 3)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatal("no choose_cards prompt")
	}
	if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != lib[1] {
		t.Errorf("offered %v, want only the artifact", pick.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID)
	oldB262OrderTheBottom(t, g, me, 4)
	effectstest.PassPriorityAroundTable(t, g)
	bottom := effectstest.LibraryBottomIDs(me, 4)
	for i, id := range lib {
		if me.Hand.Contains(id) {
			t.Errorf("card %d went to hand after declining", i)
		}
		if !effectstest.HasID(bottom, id) {
			t.Errorf("card %d is not on the bottom", i)
		}
	}
}

func TestOldB262GlintNestCraneWithNoArtifactPutsAllFourOnTheBottom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lib := effectstest.DftMPLibraryTop(me,
		game.Card{Name: "A", TypeLine: "Instant"},
		game.Card{Name: "B", TypeLine: "Instant"},
		game.Card{Name: "C", TypeLine: "Instant"},
		game.Card{Name: "D", TypeLine: "Instant"},
	)
	handBefore := me.Hand.Size()
	effectstest.CastRingCreature(t, g, "Glint-Nest Crane", "Creature — Bird", oldB262Crane, 1, 3)
	effectstest.PassPriorityAroundTable(t, g)
	oldB262OrderTheBottom(t, g, me, 4)
	effectstest.PassPriorityAroundTable(t, g)
	// The hand lost the Crane it cast and took nothing.
	if me.Hand.Size() != handBefore {
		t.Errorf("hand = %d, want %d: nothing should be taken", me.Hand.Size(), handBefore)
	}
	bottom := effectstest.LibraryBottomIDs(me, 4)
	for _, id := range lib {
		if !effectstest.HasID(bottom, id) {
			t.Errorf("a looked-at card is not among the bottom four")
		}
	}
}

// oldB262OrderTheBottom answers the "in any order" prompt with the cards as
// offered and checks it holds exactly n of them.
func oldB262OrderTheBottom(t *testing.T, g *game.Game, me *game.Player, n int) {
	t.Helper()
	prompt := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no put-in-library prompt for the rest")
	}
	if len(prompt.ScryCards) != n {
		t.Fatalf("the prompt orders %d cards, want %d", len(prompt.ScryCards), n)
	}
	if err := g.ResolvePutInLibrary(prompt.ID, me.ID, prompt.ScryCards, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
}

func TestOldB262SalivatingGremlinsPumpsForEachArtifactYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gremlins := effectstest.B12Push(g, me.ID, "Salivating Gremlins", "Creature — Gremlin", oldB262Gremlins, 2, 3)
	effectstest.AdvanceToMain(t, g)
	if p := effectstest.EffectivePower(t, g, gremlins); p != 2 {
		t.Fatalf("power %d before any artifact, want 2", p)
	}
	// An artifact an opponent controls entering does nothing.
	effectstest.EnterCard(t, g, opp.ID, game.Card{Name: "Their Rock", TypeLine: "Artifact"})
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, gremlins); p != 2 {
		t.Errorf("power %d after an opponent's artifact, want 2", p)
	}
	// A nonartifact of mine does nothing.
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, gremlins); p != 2 {
		t.Errorf("power %d after a nonartifact, want 2", p)
	}
	// My artifact: +2/+0 and trample.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TreasureToken(), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, gremlins); p != 4 {
		t.Errorf("power %d after my artifact, want 4", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, gremlins, "trample") {
		t.Error("no trample after my artifact")
	}
	// A second artifact stacks.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TreasureToken(), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, gremlins); p != 6 {
		t.Errorf("power %d after two artifacts, want 6", p)
	}
	// It ends with the turn.
	effectstest.NextTurnOfMine(t, g)
	if p := effectstest.EffectivePower(t, g, gremlins); p != 2 {
		t.Errorf("power %d next turn, want 2", p)
	}
	if effectstest.HasEffectiveKeyword(t, g, gremlins, "trample") {
		t.Error("trample outlasted the turn")
	}
}

func TestOldB262ContrabandKingpinScriesWhenAnArtifactYouControlEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.Msc18Library(me, 6)
	effectstest.B12Push(g, me.ID, "Contraband Kingpin", "Creature — Aetherborn Rogue", oldB262Kingpin, 1, 4)
	effectstest.AdvanceToMain(t, g)
	effectstest.EnterCard(t, g, opp.ID, game.Card{Name: "Their Rock", TypeLine: "Artifact"})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceScry, me.ID); got != nil {
		t.Fatal("scried for an opponent's artifact")
	}
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceScry, me.ID); got != nil {
		t.Fatal("scried for a nonartifact")
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TreasureToken(), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	scry := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceScry, me.ID)
	if scry == nil || len(scry.ScryCards) != 1 {
		t.Fatalf("scry prompt = %+v, want scry 1", scry)
	}
}

func TestOldB262HeraldOfTheFairPumpsATargetCreatureYouControlUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ally := effectstest.B12Creature(g, me.ID, "Ally", "Creature — Bear", 2, 2)
	herald := effectstest.CastHoldCreature(t, g, "Herald of the Fair", "Creature — Human", oldB262Herald, 3, 2)
	effectstest.AnswerPickTarget(t, g, ally)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, ally); p != 3 || tough != 3 {
		t.Errorf("the ally is %d/%d, want 3/3", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, herald); p != 3 || tough != 2 {
		t.Errorf("the Herald is %d/%d, want 3/2", p, tough)
	}
	effectstest.NextTurnOfMine(t, g)
	if p, tough := effectstest.PtOf(t, g, ally); p != 2 || tough != 2 {
		t.Errorf("the ally is %d/%d next turn, want 2/2", p, tough)
	}
}

func TestOldB262HeraldOfTheFairCanTargetItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	herald := effectstest.CastHoldCreature(t, g, "Herald of the Fair", "Creature — Human", oldB262Herald, 3, 2)
	effectstest.AnswerPickTarget(t, g, herald)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, herald); p != 4 || tough != 3 {
		t.Errorf("the Herald is %d/%d, want 4/3", p, tough)
	}
}

func TestOldB262HeraldOfTheFairCannotTargetAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Herald of the Fair", "Creature — Human", oldB262Herald, 3, 2)
	ch := effectstest.PickTargetPrompt(t, g)
	err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: theirs})
	if err == nil {
		t.Error("the Herald targeted an opponent's creature")
	}
}

func TestOldB262AviaryMechanicMayReturnAnotherPermanent(t *testing.T) {
	for _, decline := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		rock := effectstest.B12Push(g, me.ID, "Rock", "Artifact", "", 0, 0)
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		mech := effectstest.CastRingCreature(t, g, "Aviary Mechanic", "Creature — Dwarf Artificer", oldB262Mechanic, 2, 2)
		effectstest.PassPriorityAroundTable(t, g)
		pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
		if pick == nil {
			t.Fatal("no permanent choice")
		}
		if effectstest.HasID(pick.ChooseCards, mech) {
			t.Error("the Mechanic itself was offered")
		}
		if !effectstest.HasID(pick.ChooseCards, rock) || !effectstest.HasID(pick.ChooseCards, bear) {
			t.Errorf("offered %v, want the artifact and the creature", pick.ChooseCards)
		}
		picks := []uuid.UUID{rock}
		if decline {
			picks = nil
		}
		if err := g.ResolveOwnPermanents(pick.ID, me.ID, picks); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := me.Hand.Contains(rock); got == decline {
			t.Errorf("decline=%v: the artifact is in hand = %v", decline, got)
		}
		if me.Hand.Contains(bear) || me.Hand.Contains(mech) {
			t.Errorf("decline=%v: a permanent that was not picked went to hand", decline)
		}
	}
}

func TestOldB262AviaryMechanicAloneOffersNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mech := effectstest.CastRingCreature(t, g, "Aviary Mechanic", "Creature — Dwarf Artificer", oldB262Mechanic, 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, mech); !ok {
		t.Error("the Mechanic left the battlefield")
	}
	if me.Hand.Contains(mech) {
		t.Error("the Mechanic returned itself")
	}
}

func TestOldB262EmbraalBruiserEntersTappedAndHasMenaceWithAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bruiser := effectstest.CastRingCreature(t, g, "Embraal Bruiser", "Creature — Human Warrior", oldB262Bruiser, 3, 1)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, bruiser)
	if !ok || !c.Tapped {
		t.Fatalf("the Bruiser entered untapped or is missing: %+v", c)
	}
	if effectstest.HasEffectiveKeyword(t, g, bruiser, "menace") {
		t.Error("menace with no artifact")
	}
	rock := effectstest.B12Push(g, me.ID, "Rock", "Artifact", "", 0, 0)
	if !effectstest.HasEffectiveKeyword(t, g, bruiser, "menace") {
		t.Error("no menace with an artifact")
	}
	// An opponent's artifact does not count.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(rock) })
	effectstest.B12Push(g, g.Seats[1].ID, "Their Rock", "Artifact", "", 0, 0)
	if effectstest.HasEffectiveKeyword(t, g, bruiser, "menace") {
		t.Error("menace from an opponent's artifact")
	}
}
