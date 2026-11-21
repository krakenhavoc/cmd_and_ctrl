package testsex

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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

func flankingCrusader(t *testing.T) game.Card {
	return effectstest.FlankingImport(t, flankingCrusaderOracle, "Zhalfirin Crusader", "Creature — Human Knight", "{1}{W}{W}", "2", "2", "Flanking")
}

func flankingNimbus(t *testing.T) game.Card {
	return effectstest.FlankingImport(t, flankingNimbusOracle, "Knight of the Holy Nimbus", "Creature — Human Rebel Knight", "{W}{W}", "2", "2", "Flanking")
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

// #5784: the retired flanking row stays on the cards that shipped with
// it, one per printed instance, so a restore point holding its trigger
// still resolves through the catalog (ADR 0041 P9) and the #522 parity
// count is unchanged. No other card carries it: flanking is the keyword
// token now, and a row beside it on a new card would be dead weight,
// never a second trigger. The list only shrinks.
func TestOnlyTheLegacyFlankingCardsKeepTheRetiredRow(t *testing.T) {
	legacy := map[string]bool{}
	for _, name := range []string{
		"Burning Shield Askari", "Cadaverous Knight", "Fallen Askari", "Femeref Knight",
		"Jolrael's Centaur", "Knight of the Holy Nimbus", "Knight of Valor", "Mtenda Herder",
		"Outrider en-Kor", "Searing Spear Askari", "Shadow Rider", "Sidar Jabari",
		"Sidar Kondo of Jamuraa", "Suq'Ata Lancer", "Teferi's Honor Guard", "Telim'Tor",
		"Zhalfirin Commander", "Zhalfirin Crusader", "Zhalfirin Knight",
	} {
		legacy[name] = true
	}
	seen := 0
	for _, spec := range effects.All() {
		rows := 0
		for _, tr := range spec.Triggered {
			if strings.HasSuffix(tr.Key, " — flanking") {
				rows++
				if tr.Key != spec.Name+" — flanking" || tr.Keyword != game.KeywordFlanking {
					t.Errorf("%s: retired row keyed %q, keyword %q", spec.Name, tr.Key, tr.Keyword)
				}
			}
		}
		declared := 0
		for _, kw := range spec.PrintedKeywords {
			if kw == game.KeywordFlanking {
				declared++
			}
		}
		switch {
		case legacy[spec.Name]:
			seen++
			if rows != 1 || declared != 1 {
				t.Errorf("%s: %d retired row(s) and %d declared flanking, want 1 and 1", spec.Name, rows, declared)
			}
		case rows > 0:
			t.Errorf("%s: carries the retired flanking row; declare the keyword alone (#5784)", spec.Name)
		}
	}
	if seen != len(legacy) {
		t.Fatalf("found %d of the %d legacy flanking cards", seen, len(legacy))
	}
}

// The retired row never triggers, whatever blocks: the keyword path is
// the only one that does (#5784).
func TestRetiredFlankingRowNeverTriggers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crusader := effectstest.ApaPush(g, me.ID, me.ID, flankingCrusader(t))
	bears := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.FlankingBears(t))
	row := effects.RetiredFlankingRow("Zhalfirin Crusader")
	g.WithWriteLock(func() {
		ev := game.Event{Kind: game.EventBlock, Actor: opp.ID, CardID: bears, Target: crusader}
		if row.AppliesTo(ev, effectstest.FindBattlefieldCardForTest(g, crusader), game.Characteristic{}, g) {
			t.Error("the retired flanking row triggered")
		}
	})
}

// The issue's case: a real Zhalfirin Crusader blocked by a real Knight of
// the Holy Nimbus. The Knight has flanking, so it keeps its toughness
// (CR 702.25a); a Grizzly Bears blocking beside it does not, and gets
// -1/-1 until end of turn.
func TestFlankingBlockerWithFlankingKeepsItsToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crusader := effectstest.ApaPush(g, me.ID, me.ID, flankingCrusader(t))
	knight := effectstest.ApaPush(g, opp.ID, opp.ID, flankingNimbus(t))
	bears := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.FlankingBears(t))
	effectstest.FlankingBlocks(t, g, opp.ID, crusader, knight, bears)
	if got := effectstest.ApaLive(g, knight).CurrentToughness(); got != 2 {
		t.Errorf("Knight of the Holy Nimbus, which has flanking, has toughness %d, want 2", got)
	}
	if got := effectstest.ApaLive(g, bears).CurrentToughness(); got != 1 {
		t.Errorf("Grizzly Bears, without flanking, has toughness %d, want 1", got)
	}
}

// A blocker without flanking still gets -1/-1 when it blocks alone.
func TestFlankingBlockerWithoutFlankingGetsMinusOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crusader := effectstest.ApaPush(g, me.ID, me.ID, flankingCrusader(t))
	bears := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.FlankingBears(t))
	effectstest.FlankingBlocks(t, g, opp.ID, crusader, bears)
	if got := effectstest.ApaLive(g, bears).CurrentPower(); got != 1 {
		t.Errorf("Grizzly Bears has power %d, want 1", got)
	}
	if got := effectstest.ApaLive(g, bears).CurrentToughness(); got != 1 {
		t.Errorf("Grizzly Bears has toughness %d, want 1", got)
	}
}

// A blocker given flanking by an effect has it (CR 613.1f), so it is not
// shrunk either.
func TestFlankingGrantedToABlockerCounts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crusader := effectstest.ApaPush(g, me.ID, me.ID, flankingCrusader(t))
	bears := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.FlankingBears(t))
	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(bears),
			[]game.Mod{game.AddKeywordsMod(game.KeywordFlanking)}, g.UntilEndOfTurnDuration(),
			"test — the Bears gain flanking")
	})
	if !game.HasKeyword(effectstest.ApaLive(g, bears), game.KeywordFlanking) {
		t.Fatal("the Bears were given flanking and do not have it")
	}
	effectstest.FlankingBlocks(t, g, opp.ID, crusader, bears)
	if got := effectstest.ApaLive(g, bears).CurrentToughness(); got != 2 {
		t.Errorf("Grizzly Bears, given flanking, has toughness %d, want 2", got)
	}
}

// CR 702.25b: a creature with two instances of flanking has two triggers,
// so a blocker without flanking gets -1/-1 twice, and a real Knight of the
// Holy Nimbus blocking beside it still gets nothing. No card prints
// flanking twice, so the attacker is a test card that declares it twice;
// since #5784 the two tokens are the two triggers, with no row of its own.
func TestFlankingTwoInstancesTriggerSeparately(t *testing.T) {
	const oracle = "test-5757-double-flanking"
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID:        oracle,
		Name:            "Double Flanker",
		PrintedKeywords: []string{game.KeywordFlanking, game.KeywordFlanking},
	})
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Double Flanker", oracle, 2, 2))
	if got := game.FlankingCount(effectstest.ApaLive(g, attacker)); got != 2 {
		t.Fatalf("the Double Flanker has %d instances of flanking, want 2", got)
	}
	big := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Big Blocker", "", 2, 5))
	knight := effectstest.ApaPush(g, opp.ID, opp.ID, flankingNimbus(t))
	effectstest.FlankingBlocks(t, g, opp.ID, attacker, big, knight)
	if got := effectstest.ApaLive(g, big).CurrentToughness(); got != 3 {
		t.Errorf("a 2/5 blocker without flanking has toughness %d, want 3 (-1/-1 twice)", got)
	}
	switch k := effectstest.ApaLive(g, knight); {
	case k == nil:
		t.Error("Knight of the Holy Nimbus left the battlefield; it has flanking and should not have shrunk")
	case k.CurrentToughness() != 2:
		t.Errorf("Knight of the Holy Nimbus has toughness %d, want 2", k.CurrentToughness())
	}
}
