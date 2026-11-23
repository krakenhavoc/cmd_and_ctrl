package testsp

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const portentOfCalamityOracle = "0d49187a-1c37-4c3a-afc2-fac2bbc6b29a"

func TestPortentOfCalamityIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(portentOfCalamityOracle)
	if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Portent of Calamity: %+v %v", spec, ok)
	}
}

// portentCast casts X=x with the pool paying {X}{U}.
func portentCast(t *testing.T, g *game.Game, me *game.Player, x int) {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Portent of Calamity", TypeLine: "Sorcery", OracleID: portentOfCalamityOracle,
			ManaCost: "{X}{U}", Colors: []string{"U"}, Owner: me.ID, Controller: me.ID})
	})
	pool := "U"
	for i := 0; i < x; i++ {
		pool += "U"
	}
	effectstest.FloatForTest(g, me, pool)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: x}); err != nil {
		t.Fatalf("cast X=%d: %v", x, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func portentAnswer(t *testing.T, g *game.Game, me *game.Player, wantCandidates int, pick ...uuid.UUID) {
	t.Helper()
	p := effectstest.LatestChooseCardsFor(g, me.ID)
	if p == nil {
		t.Fatal("no type prompt")
	}
	if len(p.ChooseCards) != wantCandidates || p.ChooseMin != 0 || p.ChooseMax != 1 {
		t.Fatalf("prompt offers %d cards in [%d,%d], want %d optional single picks", len(p.ChooseCards), p.ChooseMin, p.ChooseMax, wantCandidates)
	}
	if err := g.ResolveChooseCards(p.ID, me.ID, pick); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
}

func TestPortentOfCalamityFourExiledOffersAFreeCastThenTheRestToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.SunbirdLibrary(me,
		game.Card{Name: "Rock", TypeLine: "Artifact", ManaCost: "{3}"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Bolt", TypeLine: "Instant", ManaCost: "{R}"},
		game.Card{Name: "Extra Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2},
		game.Card{Name: "Untouched", TypeLine: "Sorcery", ManaCost: "{U}"},
	)
	portentCast(t, g, me, 5)
	// Types in order: artifact, creature (two candidates), instant, land.
	portentAnswer(t, g, me, 1, ids[0])
	portentAnswer(t, g, me, 2, ids[1])
	portentAnswer(t, g, me, 1, ids[3])
	portentAnswer(t, g, me, 1, ids[2])
	if !me.Graveyard.Contains(ids[4]) {
		t.Error("the unchosen revealed card did not go to the graveyard")
	}
	if z := effectstest.ZoneOfCard(g, ids[5]); z != game.ZoneLibrary {
		t.Errorf("the sixth card left the library (%v)", z)
	}
	// Four were exiled: one free cast is offered over the spells (not the land).
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, ids[0], ids[1], ids[3])
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, ids[1], game.CastSpellParams{}); err != nil {
		t.Fatalf("free cast of the Bear: %v", err)
	}
	if effectstest.ResolutionCastOfferFor(g, me.ID) != nil {
		t.Fatal("a second free cast is offered")
	}
	for _, id := range []uuid.UUID{ids[0], ids[2], ids[3]} {
		if !effectstest.InHandOf(me, id) {
			t.Errorf("card %v of the rest is not in hand", id)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ids[1]) {
		t.Error("the free Bear did not resolve")
	}
}

func TestPortentOfCalamityFewerThanFourGoesStraightToHandWithNoCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.SunbirdLibrary(me,
		game.Card{Name: "Rock", TypeLine: "Artifact", ManaCost: "{3}"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
	)
	portentCast(t, g, me, 3)
	portentAnswer(t, g, me, 1, ids[0])
	portentAnswer(t, g, me, 1, ids[1])
	portentAnswer(t, g, me, 1) // decline the land
	if effectstest.ResolutionCastOfferFor(g, me.ID) != nil {
		t.Fatal("a free cast was offered with fewer than four exiled")
	}
	if !effectstest.InHandOf(me, ids[0]) || !effectstest.InHandOf(me, ids[1]) {
		t.Error("the two exiled cards are not in hand")
	}
	if !me.Graveyard.Contains(ids[2]) {
		t.Error("the declined land did not go to the graveyard")
	}
}

func TestPortentOfCalamityMultiTypeCardFillsOneSlot(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.SunbirdLibrary(me,
		game.Card{Name: "Golem", TypeLine: "Artifact Creature — Golem", ManaCost: "{4}", Power: 3, Toughness: 3},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2},
	)
	portentCast(t, g, me, 2)
	portentAnswer(t, g, me, 1, ids[0]) // artifact: the Golem
	portentAnswer(t, g, me, 1, ids[1]) // creature: only the Bear is left
	if !effectstest.InHandOf(me, ids[0]) || !effectstest.InHandOf(me, ids[1]) {
		t.Error("both exiled cards should be in hand")
	}
}
