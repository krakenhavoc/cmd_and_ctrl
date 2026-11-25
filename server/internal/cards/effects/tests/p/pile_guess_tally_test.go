package testsp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pile_guess_tally_test.go — #6175, the turn's tally of spells and
// abilities that caused one or more players to guess or to group cards
// or permanents into a pile (CR 700.3), which The Celestial Toymaker
// reads at each end step.

// pileGuessCount is the tally as a card would read it.
func pileGuessCount(g *game.Game) int {
	n := 0
	g.WithWriteLock(func() { n = g.PilesOrGuessesThisTurnForEffect() })
	return n
}

// A split counts the spell that asked for it, once, through the
// choose-an-opponent step that precedes it.
func TestPileOrGuessTallyCountsASplit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.SeedFactOrFictionLibrary(me, "FF One", "FF Two", "FF Three", "FF Four", "FF Five")
	if n := pileGuessCount(g); n != 0 {
		t.Fatalf("setup: tally %d, want 0", n)
	}
	effectstest.CastCatalogSpell(t, g, "Fact or Fiction", "Instant", effectstest.FactOrFictionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	if n := pileGuessCount(g); n != 1 {
		t.Errorf("tally %d after Fact or Fiction's split, want 1", n)
	}
}

// One spell, two piles: Hostile Negotiations makes a pile and then
// another, and is one spell for the count.
func TestPileOrGuessTallyCountsOnePerSpellNotPerPile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "HN One", "HN Two", "HN Three", "HN Four", "HN Five", "HN Six")
	effectstest.CastCatalogSpell(t, g, "Hostile Negotiations", "Sorcery", pgHostileNegotiations, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := pileGuessCount(g); n != 1 {
		t.Errorf("tally %d after Hostile Negotiations' two piles, want 1", n)
	}
}

// The face-down pile family counts too: Riddles in the Dark separates
// the cards itself and an opponent chooses a pile.
func TestPileOrGuessTallyCountsAFaceDownSplit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.SeedLibrary(me, "R One", "R Two", "R Three", "R Four")
	effectstest.CastCatalogSpell(t, g, "Riddles in the Dark", "Instant", pgRiddlesInTheDark, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	if n := pileGuessCount(g); n != 1 {
		t.Errorf("tally %d after Riddles in the Dark, want 1", n)
	}
}

// A guess counts: The Toymaker's Trap asks an opponent to guess the
// number its controller chose.
func TestPileOrGuessTallyCountsAGuess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "The Toymaker's Trap", "Enchantment", pgToymakersTrap, false)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	if n := pileGuessCount(g); n != 1 {
		t.Errorf("tally %d after The Toymaker's Trap asked for a guess, want 1", n)
	}
}

// A turn with neither counts nothing, and the tally is a TURN tally: it
// is empty again on the next turn.
func TestPileOrGuessTallyIsEmptyWithoutOneAndResetsEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.SeedFactOrFictionLibrary(me, "FF One", "FF Two")
	effectstest.CastCatalogSpell(t, g, "Divination", "Sorcery", effectstest.DivinationOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := pileGuessCount(g); n != 0 {
		t.Fatalf("tally %d after a spell that makes no pile, want 0", n)
	}

	effectstest.CastCatalogSpell(t, g, "Fact or Fiction", "Instant", effectstest.FactOrFictionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	if n := pileGuessCount(g); n != 1 {
		t.Fatalf("tally %d after the split, want 1", n)
	}
	// Answer what is open so the turn can end, then cross the boundary.
	if split := effectstest.LatestRevealPickFor(g, opp.ID); split != nil {
		if err := g.ResolveRevealPick(split.ID, opp.ID, nil); err != nil {
			t.Fatalf("ResolveRevealPick: %v", err)
		}
	}
	if pick := effectstest.LatestOptionPickFor(g, me.ID); pick != nil {
		effectstest.AnswerOptionPick(t, g, me.ID, 0)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if n := pileGuessCount(g); n != 0 {
		t.Errorf("tally %d on the next turn, want 0", n)
	}
}

const (
	pgHostileNegotiations = "c5cbd852-c42f-4592-a0bb-a8bffa2dd2f4"
	pgRiddlesInTheDark    = "78bed291-0ec7-466f-bc7a-9dae7d0b56ae"
	pgToymakersTrap       = "8e93ea21-5e3f-448e-a175-96c8438c328e"
)

// pgPileWords and pgGuessWords are the audit's reading of "group cards
// or permanents into a pile" and "guess".
var (
	pgPileWords  = regexp.MustCompile(`(?i)\bpiles?\b`)
	pgGuessWords = regexp.MustCompile(`(?i)\bguess(?:es|ed|ing)?\b`)
)

// pgReviewed is every catalog card whose printed text says "pile" or
// "guess", and how it is accounted for. A card that MAKES a pile or
// ASKS for a guess notes itself in the tally (effects.NotePileOrGuess),
// through a shared helper or in its own file; a card that only READS
// the tally, or that says the word without making either, is listed
// with that reason.
//
// The audit below fails on a card that is not here, so a new pile or
// guess card cannot be added without someone deciding which it is.
var pgReviewed = map[string]string{
	// The shared split (effects.PileSplit) notes the tally for all of these.
	"437b2dab-15e0-4b9a-a204-58622d37a3b3": "Fact or Fiction — PileSplit",
	"f25af42d-f1bf-4bd3-aded-ecadafdbf6e6": "Do or Die — PileSplit",
	"2274c7ae-5a40-4fd4-a4ac-6f56b23034e4": "Jace, Architect of Thought — PileSplit",
	"eedb1527-548a-435a-a3e1-8ad7d9d6d8a9": "Intrude on the Mind — PileSplit",
	"591edddc-3736-44cf-9719-117c24ff3980": "Make an Example — PileSplit",
	"7ee394a2-84de-42e8-834e-6c1030be0aef": "Sphinx of Clear Skies — PileSplit",
	"48ad5bd6-9568-4813-be29-99aef8b9327a": "Split the Spoils — PileSplit",
	"0ba134d8-ee7d-48ec-8dc6-57942b8e9261": "Liliana of the Veil — PileSplit",
	"4b4f268d-34ae-416d-a412-66f4fe37659c": "Sphinx of Uthuun — PileSplit",
	"a3d048c8-e8df-4b50-a422-fc94a69f44bb": "Boneyard Parley — PileSplit",
	"33befbd6-56ac-45a7-a4d0-448167162707": "Epiphany at the Drownyard — PileSplit",
	// The face-down pile family, which is PileSplit underneath.
	"8b2b00ab-f1c5-4057-9957-d7daac95a847": "Atris, Oracle of Half-Truths — face-down piles",
	"9baff93e-4ef9-404d-884b-f651a11f3742": "Fortune's Favor — face-down piles",
	"fe04a81b-413c-4bc8-bb83-29978da9a43d": "Sauron's Ransom — face-down piles",
	"9938d178-0ce6-45c0-b317-fd5c54231579": "Curator of Destinies — face-down piles",
	"78bed291-0ec7-466f-bc7a-9dae7d0b56ae": "Riddles in the Dark — face-down piles",
	// Piles built in the card's own file; each notes the tally there.
	"3e0e3676-e634-4e64-bdfd-4aaa5885461f": "Abstract Performance — notes its own piles",
	"bac0fcee-9c1a-46b7-86c8-ffbfcc1e96de": "Become Anonymous — notes its own pile",
	"4f7e4aff-ef68-44fe-a0dc-b9ab0ecd7146": "Expose the Culprit — notes its pile mode",
	"c5cbd852-c42f-4592-a0bb-a8bffa2dd2f4": "Hostile Negotiations — notes its own piles",
	"6f8a4998-35c7-4e30-a9e7-7d4ec933238a": "Triumph of Saint Katherine — notes its own pile",
	"3bf20e68-0e19-4242-bfba-a8a1bb244685": "Ghastly Conscription — notes its face-down pile",
	"39b359c1-48b8-43a7-b972-54764eadd2c7": "Jeskai Infiltrator — notes its face-down pile",
	// Guesses, noted where the guess is asked.
	"3fb5c9a6-3fab-4109-875b-83c92518c5eb": "Gollum, Scheming Guide — notes the opponent's guess",
	"43e1b053-2cb9-4b1d-9e38-100cfa7267b6": "The Seventh Doctor — notes the defending player's guess",
	"8e93ea21-5e3f-448e-a175-96c8438c328e": "The Toymaker's Trap — notes the opponent's guess",
	// The reader, which also makes piles of its own.
	"388fe92a-0bc0-4c0a-a7c6-910c5e71e342": "The Celestial Toymaker — notes its attack piles and reads the tally",
}

// Every catalog card whose printed text says "pile" or "guess" is
// accounted for above, so a new one is a reviewer's decision rather
// than a silent gap in The Celestial Toymaker's count (#6175).
func TestEveryPileOrGuessCardIsReviewed(t *testing.T) {
	const dir = "../../../coverage/testdata/oracle"
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading the oracle fixtures: %v", err)
	}
	var missing []string
	seen := map[string]bool{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatalf("reading %s: %v", e.Name(), err)
		}
		var fx struct {
			Name  string `json:"name"`
			Text  string `json:"text"`
			Faces []struct {
				Name string `json:"name"`
				Text string `json:"text"`
			} `json:"faces"`
		}
		if err := json.Unmarshal(b, &fx); err != nil {
			t.Fatalf("parsing %s: %v", e.Name(), err)
		}
		text := fx.Text
		for _, f := range fx.Faces {
			text += "\n" + f.Text
		}
		if !pgPileWords.MatchString(text) && !pgGuessWords.MatchString(text) {
			continue
		}
		oracleID := strings.TrimSuffix(e.Name(), ".json")
		seen[oracleID] = true
		if _, ok := pgReviewed[oracleID]; !ok {
			missing = append(missing, oracleID+"\t"+fx.Name)
		}
	}
	if len(missing) > 0 {
		t.Errorf("%d catalog card(s) say \"pile\" or \"guess\" and are not reviewed for the #6175 tally.\n"+
			"Each one either notes itself with effects.NotePileOrGuess (it makes a pile or asks for a guess)\n"+
			"or is listed in pgReviewed with the reason it does not:\n  %s", len(missing), strings.Join(missing, "\n  "))
	}
	for oracleID, why := range pgReviewed {
		if !seen[oracleID] {
			t.Errorf("pgReviewed lists %s (%s), whose text no longer says \"pile\" or \"guess\"; drop the row", oracleID, why)
		}
	}
}
