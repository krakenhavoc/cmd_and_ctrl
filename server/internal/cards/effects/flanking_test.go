package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// flanking_test.go — #5757. Flanking (CR 702.25a) shrinks only a blocker
// WITHOUT flanking, so the keyword has to be on the real permanents: the
// cards here come through the deck importer from their Scryfall records,
// never with a Keywords slice set by hand.

const (
	flankingCrusaderOracle = "9e60c102-412e-4956-a7bb-a2cd737d6692" // Zhalfirin Crusader
	flankingNimbusOracle   = "4e92c705-19c2-42df-ad12-808d272c505e" // Knight of the Holy Nimbus
	flankingOutriderOracle = "644b7aba-a7b8-4861-9e83-73329cf85be2" // Outrider en-Kor
	flankingSidarOracle    = "331b6834-6bb4-40fa-b80b-22d67b34734f" // Sidar Kondo of Jamuraa
)

// flankingImport is a card as the deck importer hands it to a game: its
// Scryfall record, with the printed text from the coverage oracle fixture
// when the card has one, run through deck.ToGameCard. An empty oracle is a
// card the catalog does not know.
func flankingImport(t *testing.T, oracle, name, typeLine, cost string, power, toughness string, keywords ...string) game.Card {
	t.Helper()
	id := uuid.New()
	text := ""
	if oracle != "" {
		id = uuid.MustParse(oracle)
		if text = spellText(t, oracle); text == "" {
			t.Fatalf("%s has no oracle fixture", name)
		}
	} else if _, ok := Lookup(id.String()); ok {
		t.Fatalf("%s must not be in the catalog", name)
	}
	return deck.ToGameCard(cards.Card{
		ID: uuid.New(), OracleID: id, Name: name, Layout: "normal",
		TypeLine: typeLine, ManaCost: cost, Power: power, Toughness: toughness,
		Keywords: keywords, OracleText: text,
	}, false)
}

func flankingCrusader(t *testing.T) game.Card {
	return flankingImport(t, flankingCrusaderOracle, "Zhalfirin Crusader", "Creature — Human Knight", "{1}{W}{W}", "2", "2", "Flanking")
}

func flankingNimbus(t *testing.T) game.Card {
	return flankingImport(t, flankingNimbusOracle, "Knight of the Holy Nimbus", "Creature — Human Rebel Knight", "{W}{W}", "2", "2", "Flanking")
}

func flankingBears(t *testing.T) game.Card {
	return flankingImport(t, "", "Grizzly Bears", "Creature — Bear", "{1}{G}", "2", "2")
}

// flankingBlocks has the active seat's `attacker` attack `defender`, has
// each of `blockers` block it, locks the blocks in and resolves what
// triggered, stopping in the declare blockers step.
func flankingBlocks(t *testing.T, g *game.Game, defender, attacker uuid.UUID, blockers ...uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, defender); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	for _, b := range blockers {
		if err := g.DeclareBlocker(b, attacker); err != nil {
			t.Fatal(err)
		}
	}
	if !finishBlockDeclarations(t, g) {
		t.Fatal("no block declaration was left to finish")
	}
	passPriorityAroundTable(t, g)
	if g.Turn.Step != game.StepDeclareBlockers {
		t.Fatalf("combat left the declare blockers step (now %v)", g.Turn.Step)
	}
}

// The importer stamps Scryfall's "Flanking" as the canonical token, and
// every catalog row with the trigger declares it as well, so a card that
// never went through an import has it too.
func TestFlankingIsAKeywordARealCardHas(t *testing.T) {
	if got, ok := game.CanonicalKeyword("Flanking"); !ok || got != game.KeywordFlanking {
		t.Fatalf(`CanonicalKeyword("Flanking") = %q, %v; want %q, true`, got, ok, game.KeywordFlanking)
	}
	imported := flankingNimbus(t)
	if !game.HasKeyword(&imported, game.KeywordFlanking) {
		t.Errorf("the imported Knight of the Holy Nimbus has keywords %q, want flanking", imported.Keywords)
	}
	for _, oracle := range []string{flankingCrusaderOracle, flankingNimbusOracle, flankingOutriderOracle, flankingSidarOracle} {
		c := game.Card{OracleID: oracle, Name: oracle, TypeLine: "Creature — Knight"}
		if !game.HasKeyword(&c, game.KeywordFlanking) {
			t.Errorf("%s: the catalog card does not have flanking", oracle)
		}
	}
}

// Every catalog row that carries the flanking trigger declares the token,
// and every row that declares the token carries the trigger, one per
// printed instance: a card with the keyword and no trigger would be a
// badge with nothing behind it, and one with the trigger and no keyword
// would shrink a flanking creature that blocks it.
func TestEveryFlankingRowDeclaresTheKeyword(t *testing.T) {
	found := 0
	for _, spec := range All() {
		triggers := 0
		for _, tr := range spec.Triggered {
			if strings.HasSuffix(tr.Key, " — flanking") {
				triggers++
			}
		}
		declared := 0
		for _, kw := range spec.PrintedKeywords {
			if kw == game.KeywordFlanking {
				declared++
			}
		}
		if triggers == 0 && declared == 0 {
			continue
		}
		found++
		if declared == 0 {
			t.Errorf("%s: has %d flanking trigger(s) and does not declare the keyword", spec.Name, triggers)
		}
		if triggers == 0 {
			t.Errorf("%s: declares flanking and has no flanking trigger", spec.Name)
		}
	}
	if found < 4 {
		t.Fatalf("found %d flanking cards, want at least the four of #5757", found)
	}
}

// The issue's case: a real Zhalfirin Crusader blocked by a real Knight of
// the Holy Nimbus. The Knight has flanking, so it keeps its toughness
// (CR 702.25a); a Grizzly Bears blocking beside it does not, and gets
// -1/-1 until end of turn.
func TestFlankingBlockerWithFlankingKeepsItsToughness(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crusader := apaPush(g, me.ID, me.ID, flankingCrusader(t))
	knight := apaPush(g, opp.ID, opp.ID, flankingNimbus(t))
	bears := apaPush(g, opp.ID, opp.ID, flankingBears(t))
	flankingBlocks(t, g, opp.ID, crusader, knight, bears)
	if got := apaLive(g, knight).CurrentToughness(); got != 2 {
		t.Errorf("Knight of the Holy Nimbus, which has flanking, has toughness %d, want 2", got)
	}
	if got := apaLive(g, bears).CurrentToughness(); got != 1 {
		t.Errorf("Grizzly Bears, without flanking, has toughness %d, want 1", got)
	}
}

// A blocker without flanking still gets -1/-1 when it blocks alone.
func TestFlankingBlockerWithoutFlankingGetsMinusOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crusader := apaPush(g, me.ID, me.ID, flankingCrusader(t))
	bears := apaPush(g, opp.ID, opp.ID, flankingBears(t))
	flankingBlocks(t, g, opp.ID, crusader, bears)
	if got := apaLive(g, bears).CurrentPower(); got != 1 {
		t.Errorf("Grizzly Bears has power %d, want 1", got)
	}
	if got := apaLive(g, bears).CurrentToughness(); got != 1 {
		t.Errorf("Grizzly Bears has toughness %d, want 1", got)
	}
}

// A blocker given flanking by an effect has it (CR 613.1f), so it is not
// shrunk either.
func TestFlankingGrantedToABlockerCounts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crusader := apaPush(g, me.ID, me.ID, flankingCrusader(t))
	bears := apaPush(g, opp.ID, opp.ID, flankingBears(t))
	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(bears),
			[]game.Mod{game.AddKeywordsMod(game.KeywordFlanking)}, g.UntilEndOfTurnDuration(),
			"test — the Bears gain flanking")
	})
	if !game.HasKeyword(apaLive(g, bears), game.KeywordFlanking) {
		t.Fatal("the Bears were given flanking and do not have it")
	}
	flankingBlocks(t, g, opp.ID, crusader, bears)
	if got := apaLive(g, bears).CurrentToughness(); got != 2 {
		t.Errorf("Grizzly Bears, given flanking, has toughness %d, want 2", got)
	}
}

// CR 702.25b: a creature with two instances of flanking has two triggers,
// so a blocker without flanking gets -1/-1 twice, and a real Knight of the
// Holy Nimbus blocking beside it still gets nothing. No card prints
// flanking twice, so the attacker is a test row with two.
func TestFlankingTwoInstancesTriggerSeparately(t *testing.T) {
	const oracle = "test-5757-double-flanking"
	registerForTest(t, Spec{
		OracleID:        oracle,
		Name:            "Double Flanker",
		PrintedKeywords: []string{game.KeywordFlanking, game.KeywordFlanking},
		Triggered:       []game.TriggeredAbility{Flanking("Double Flanker"), Flanking("Double Flanker")},
	})
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := apaPush(g, me.ID, me.ID, apaCreature("Double Flanker", oracle, 2, 2))
	big := apaPush(g, opp.ID, opp.ID, apaCreature("Big Blocker", "", 2, 5))
	knight := apaPush(g, opp.ID, opp.ID, flankingNimbus(t))
	flankingBlocks(t, g, opp.ID, attacker, big, knight)
	if got := apaLive(g, big).CurrentToughness(); got != 3 {
		t.Errorf("a 2/5 blocker without flanking has toughness %d, want 3 (-1/-1 twice)", got)
	}
	switch k := apaLive(g, knight); {
	case k == nil:
		t.Error("Knight of the Holy Nimbus left the battlefield; it has flanking and should not have shrunk")
	case k.CurrentToughness() != 2:
		t.Errorf("Knight of the Holy Nimbus has toughness %d, want 2", k.CurrentToughness())
	}
}
