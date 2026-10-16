package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// extraction_family_test.go — #4033: Surgical Extraction, Extirpate,
// Eradicate, Scour, Splinter, Sowing Salt and Quash on the graveyard, hand
// and library search. "Any number" leaves every match to the caster; "all"
// finds every match in the graveyard and leaves the hidden zones' matches
// to the caster (CR 701.23b).

const (
	surgicalExtractionOracle = "3bfa088c-9a47-4123-aee5-4de44e355bef"
	extirpateOracle          = "fa857e9d-dfe4-4348-9c1f-16aa240d69cd"
	eradicateOracle          = "782ff740-ff43-4c07-af0b-433cb9770661"
	scourOracle              = "f18622da-81f3-4854-a1bb-8d658751d8db"
	splinterOracle           = "90f01e9f-68f1-4e41-97dc-8de990a91504"
	sowingSaltOracle         = "131d2b70-959d-4814-b0c8-38031f141718"
	quashOracle              = "255131a2-616c-4c8d-a36a-14e5b5ab345e"
)

type extractionTable struct {
	g             *game.Game
	me, opp, opp2 *game.Player
	cards         map[string]uuid.UUID
}

// extractionSetUp gives the opponent `name` twice in the graveyard, once in
// the hand and once in the library, beside a card of another name in each
// zone and a basic Forest in the graveyard.
func extractionSetUp(t *testing.T, name, typeLine string) extractionTable {
	t.Helper()
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	x := extractionTable{
		g: g, me: g.Seats[seat], opp: g.Seats[(seat+1)%len(g.Seats)], opp2: g.Seats[(seat+2)%len(g.Seats)],
		cards: map[string]uuid.UUID{},
	}
	g.WithWriteLock(func() { x.opp.Hand.Cards = nil })
	x.put("gy", x.opp.Graveyard, name, typeLine)
	x.put("gy 2", x.opp.Graveyard, name, typeLine)
	x.put("hand", x.opp.Hand, name, typeLine)
	x.put("lib", x.opp.Library, name, typeLine)
	x.put("gy other", x.opp.Graveyard, "Shock", "Instant")
	x.put("hand other", x.opp.Hand, "Counterspell", "Instant")
	x.put("lib other", x.opp.Library, "Opt", "Instant")
	x.put("gy forest", x.opp.Graveyard, "Forest", "Basic Land — Forest")
	return x
}

func (x extractionTable) put(key string, zone *game.Zone, name, typeLine string) {
	id := uuid.New()
	zone.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: x.opp.ID, Controller: x.opp.ID})
	x.cards[key] = id
}

// permanent puts a permanent of `name` onto the battlefield under the
// opponent's control.
func (x extractionTable) permanent(key, name, typeLine string, colors ...string) uuid.UUID {
	id := pushBattlefieldCardWithTimestamp(x.g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Colors: colors,
		Power: 2, Toughness: 2, Owner: x.opp.ID, Controller: x.opp.ID,
	})
	x.cards[key] = id
	return id
}

func (x extractionTable) exiled(t *testing.T, keys ...string) {
	t.Helper()
	for _, k := range keys {
		if !x.g.Exile.Contains(x.cards[k]) {
			t.Errorf("%s was not exiled", k)
		}
	}
}

func (x extractionTable) notExiled(t *testing.T, keys ...string) {
	t.Helper()
	for _, k := range keys {
		if x.g.Exile.Contains(x.cards[k]) {
			t.Errorf("%s was exiled", k)
		}
	}
}

// searches counts the searches of the opponent's library that shuffled it.
func (x extractionTable) searches() int {
	n := 0
	for _, ev := range x.g.Events {
		if ev.Kind == game.EventSearchLibrary && ev.Target == x.opp.ID && ev.Label == game.SearchShuffledLabel {
			n++
		}
	}
	return n
}

func (x extractionTable) target(key string) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: x.cards[key]}}
}

// candidates is the open search prompt's candidate set, by key.
func (x extractionTable) candidates(t *testing.T) []string {
	t.Helper()
	c := searchChoiceFor(x.g, x.me.ID)
	if c == nil {
		t.Fatal("no search prompt for the caster")
	}
	var keys []string
	for k, id := range x.cards {
		if slices.Contains(c.SearchCards, id) {
			keys = append(keys, k)
		}
	}
	slices.Sort(keys)
	return keys
}

func wantKeys(t *testing.T, got []string, want ...string) {
	t.Helper()
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Errorf("candidates %v, want %v", got, want)
	}
}

// Surgical Extraction chooses the target without exiling it: the target
// is a candidate like every other card with its name, and "any number"
// lets the caster leave any of them, the target included.
func TestSurgicalExtractionTakesAnyNumberIncludingNone(t *testing.T) {
	x := extractionSetUp(t, "Lightning Bolt", "Instant")
	castCatalogSpell(t, x.g, "Surgical Extraction", "Instant", surgicalExtractionOracle, x.target("gy"))
	passPriorityAroundTable(t, x.g)

	if x.g.Exile.Contains(x.cards["gy"]) {
		t.Fatal("the target was exiled before the search")
	}
	wantKeys(t, x.candidates(t), "gy", "gy 2", "hand", "lib")
	if c := searchChoiceFor(x.g, x.me.ID); x.g.SearchFloorLocked(c) != 0 {
		t.Error("any number: finding nothing must be legal")
	}
	answerSearchByID(t, x.g, x.me.ID, x.cards["gy 2"], x.cards["hand"])
	x.exiled(t, "gy 2", "hand")
	x.notExiled(t, "gy", "lib", "gy other", "hand other", "lib other")
	if x.searches() != 1 {
		t.Errorf("%d searches of their library, want 1", x.searches())
	}
}

// "Other than a basic land card", in any graveyard; and the Phyrexian
// symbol may be paid with 2 life and no mana (CR 107.4f).
func TestSurgicalExtractionTargetsAndPhyrexianMana(t *testing.T) {
	x := extractionSetUp(t, "Lightning Bolt", "Instant")
	if err := castCatalogSpellErr(t, x.g, "Surgical Extraction", "Instant", surgicalExtractionOracle, x.target("gy forest")); err == nil {
		t.Fatal("a basic land card was a legal target")
	}
	life := x.me.Life
	tdCastFromHand(t, x.g, game.Card{Name: "Surgical Extraction", TypeLine: "Instant", OracleID: surgicalExtractionOracle,
		ManaCost: "{B/P}"}, game.CastSpellParams{Strict: true, PhyrexianLife: 1, Targets: x.target("gy")})
	if got := life - x.me.Life; got != game.PhyrexianLifePerSymbol {
		t.Errorf("paid %d life for {B/P}, want %d", got, game.PhyrexianLifePerSymbol)
	}
	passPriorityAroundTable(t, x.g)
	answerSearchFailToFind(t, x.g, x.me.ID)
	x.notExiled(t, "gy", "gy 2", "hand", "lib")
	if x.searches() != 1 {
		t.Errorf("%d searches, want 1: the library is shuffled whatever was taken", x.searches())
	}
}

// Extirpate has split second, and "all": both graveyard copies, the target
// among them, are exiled without being offered; the hand and library copies
// are the caster's to take or leave.
func TestExtirpateFindsEveryGraveyardCopyUnderSplitSecond(t *testing.T) {
	x := extractionSetUp(t, "Lightning Bolt", "Instant")
	spell := castCatalogSpell(t, x.g, "Extirpate", "Instant", extirpateOracle, x.target("gy"))
	requireSplitSecondOn(t, x.g, spell)
	passPriorityAroundTable(t, x.g)

	wantKeys(t, x.candidates(t), "hand", "lib")
	answerSearchByID(t, x.g, x.me.ID, x.cards["lib"])
	x.exiled(t, "gy", "gy 2", "lib")
	x.notExiled(t, "hand", "gy other", "hand other", "lib other", "gy forest")
	if x.searches() != 1 {
		t.Errorf("%d searches, want 1", x.searches())
	}
}

// With no copy in the hand or the library there is nothing to choose:
// the graveyard copies go and the library is still shuffled.
func TestExtirpateWithOnlyGraveyardCopiesDoesNotAsk(t *testing.T) {
	x := extractionSetUp(t, "Lightning Bolt", "Instant")
	x.g.WithWriteLock(func() {
		_, _ = x.opp.Hand.Remove(x.cards["hand"])
		_, _ = x.opp.Library.Remove(x.cards["lib"])
	})
	castCatalogSpell(t, x.g, "Extirpate", "Instant", extirpateOracle, x.target("gy"))
	passPriorityAroundTable(t, x.g)
	if c := searchChoiceFor(x.g, x.me.ID); c != nil {
		t.Fatalf("asked with nothing to choose: %+v", c)
	}
	x.exiled(t, "gy", "gy 2")
	if x.searches() != 1 {
		t.Errorf("%d searches, want 1", x.searches())
	}
}

// Eradicate exiles the creature, then every graveyard card with its name,
// and offers the hand and library copies.
func TestEradicateExilesTheCreatureAndEveryGraveyardCopy(t *testing.T) {
	x := extractionSetUp(t, "Grizzly Bears", "Creature — Bear")
	x.permanent("bears", "Grizzly Bears", "Creature — Bear", "G")
	castCatalogSpell(t, x.g, "Eradicate", "Sorcery", eradicateOracle, x.target("bears"))
	passPriorityAroundTable(t, x.g)

	x.exiled(t, "bears")
	wantKeys(t, x.candidates(t), "hand", "lib")
	answerSearchByID(t, x.g, x.me.ID, x.cards["hand"], x.cards["lib"])
	x.exiled(t, "gy", "gy 2", "hand", "lib")
	x.notExiled(t, "gy other", "hand other", "lib other")
	if x.searches() != 1 {
		t.Errorf("%d searches, want 1", x.searches())
	}
}

// "Nonblack", and "its controller": a creature the opponent controls but
// another player owns sends the search through the controller's zones.
func TestEradicateSearchesTheControllersZonesAndSparesBlack(t *testing.T) {
	x := extractionSetUp(t, "Grizzly Bears", "Creature — Bear")
	x.permanent("black", "Vampire", "Creature — Vampire", "B")
	if err := castCatalogSpellErr(t, x.g, "Eradicate", "Sorcery", eradicateOracle, x.target("black")); err == nil {
		t.Fatal("a black creature was a legal target")
	}
	stolen := pushBattlefieldCardWithTimestamp(x.g, game.Card{
		InstanceID: uuid.New(), Name: "Grizzly Bears", TypeLine: "Creature — Bear", Colors: []string{"G"},
		Power: 2, Toughness: 2, Owner: x.opp2.ID, Controller: x.opp.ID,
	})
	ownerCopy := uuid.New()
	x.opp2.Graveyard.PushTop(game.Card{InstanceID: ownerCopy, Name: "Grizzly Bears", TypeLine: "Creature — Bear",
		Owner: x.opp2.ID, Controller: x.opp2.ID})
	castCatalogSpell(t, x.g, "Eradicate", "Sorcery", eradicateOracle, []game.TargetRef{{Kind: game.TargetCard, ID: stolen}})
	passPriorityAroundTable(t, x.g)
	answerSearchFailToFind(t, x.g, x.me.ID)
	if !x.g.Exile.Contains(stolen) {
		t.Error("the creature was not exiled")
	}
	x.exiled(t, "gy", "gy 2")
	if x.g.Exile.Contains(ownerCopy) {
		t.Error("the owner's graveyard was searched instead of the controller's")
	}
}

// Scour on an enchantment, with nothing in the hidden zones: no prompt.
func TestScourExilesTheEnchantmentAndItsGraveyardCopies(t *testing.T) {
	x := extractionSetUp(t, "Pacifism", "Enchantment — Aura")
	x.g.WithWriteLock(func() {
		_, _ = x.opp.Hand.Remove(x.cards["hand"])
		_, _ = x.opp.Library.Remove(x.cards["lib"])
	})
	x.permanent("ench", "Pacifism", "Enchantment")
	x.permanent("art", "Sol Ring", "Artifact")
	if err := castCatalogSpellErr(t, x.g, "Scour", "Instant", scourOracle, x.target("art")); err == nil {
		t.Fatal("an artifact was a legal target")
	}
	castCatalogSpell(t, x.g, "Scour", "Instant", scourOracle, x.target("ench"))
	passPriorityAroundTable(t, x.g)
	if c := searchChoiceFor(x.g, x.me.ID); c != nil {
		t.Fatalf("asked with nothing to choose: %+v", c)
	}
	x.exiled(t, "ench", "gy", "gy 2")
	x.notExiled(t, "art")
	if x.searches() != 1 {
		t.Errorf("%d searches, want 1", x.searches())
	}
}

func TestSplinterExilesTheArtifactAndEveryCopy(t *testing.T) {
	x := extractionSetUp(t, "Sol Ring", "Artifact")
	x.permanent("art", "Sol Ring", "Artifact")
	x.permanent("ench", "Pacifism", "Enchantment")
	if err := castCatalogSpellErr(t, x.g, "Splinter", "Sorcery", splinterOracle, x.target("ench")); err == nil {
		t.Fatal("an enchantment was a legal target")
	}
	castCatalogSpell(t, x.g, "Splinter", "Sorcery", splinterOracle, x.target("art"))
	passPriorityAroundTable(t, x.g)
	x.exiled(t, "art")
	wantKeys(t, x.candidates(t), "hand", "lib")
	answerSearchByID(t, x.g, x.me.ID, x.cards["hand"], x.cards["lib"])
	x.exiled(t, "gy", "gy 2", "hand", "lib")
}

// Sowing Salt: a basic land is not a legal target; a nonbasic land goes,
// with every graveyard copy.
func TestSowingSaltExilesANonbasicLandAndItsCopies(t *testing.T) {
	x := extractionSetUp(t, "Command Tower", "Land")
	x.permanent("tower", "Command Tower", "Land")
	x.permanent("forest", "Forest", "Basic Land — Forest")
	if err := castCatalogSpellErr(t, x.g, "Sowing Salt", "Sorcery", sowingSaltOracle, x.target("forest")); err == nil {
		t.Fatal("a basic land was a legal target")
	}
	castCatalogSpell(t, x.g, "Sowing Salt", "Sorcery", sowingSaltOracle, x.target("tower"))
	passPriorityAroundTable(t, x.g)
	x.exiled(t, "tower")
	wantKeys(t, x.candidates(t), "hand", "lib")
	answerSearchFailToFind(t, x.g, x.me.ID)
	x.exiled(t, "gy", "gy 2")
	x.notExiled(t, "forest", "hand", "lib", "gy forest")
}

// Quash counters the spell, which lands in the graveyard and is exiled
// with the other graveyard copies; the hidden copies are offered.
func TestQuashCountersAndExilesEveryGraveyardCopy(t *testing.T) {
	x := extractionSetUp(t, "Lightning Bolt", "Instant")
	bolt := uuid.New()
	x.opp.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant",
		Owner: x.opp.ID, Controller: x.opp.ID})
	x.cards["cast"] = bolt
	advanceToMain(t, x.g)
	if err := x.g.CastSpell(x.opp.ID, bolt, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the Bolt: %v", err)
	}
	castCatalogSpell(t, x.g, "Quash", "Instant", quashOracle, x.target("cast"))
	passPriorityAroundTable(t, x.g)

	if !x.opp.Graveyard.Contains(x.cards["cast"]) {
		t.Fatal("the Bolt was not countered into its owner's graveyard")
	}
	wantKeys(t, x.candidates(t), "hand", "lib")
	answerSearchByID(t, x.g, x.me.ID, x.cards["hand"])
	x.exiled(t, "cast", "gy", "gy 2", "hand")
	x.notExiled(t, "lib", "gy other", "hand other")
	if x.searches() != 1 {
		t.Errorf("%d searches, want 1", x.searches())
	}
}

// "Target instant or sorcery spell": a creature spell is not one.
func TestQuashRefusesACreatureSpell(t *testing.T) {
	x := extractionSetUp(t, "Grizzly Bears", "Creature — Bear")
	x.cards["cast"] = tdCastFromHand(t, x.g, game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2}, game.CastSpellParams{})
	if err := castCatalogSpellErr(t, x.g, "Quash", "Instant", quashOracle, x.target("cast")); err == nil {
		t.Error("a creature spell was a legal target")
	}
}

// Every card of the family is catalogued as Full.
func TestExtractionFamilyIsComplete(t *testing.T) {
	for _, oid := range []string{surgicalExtractionOracle, extirpateOracle, eradicateOracle, scourOracle,
		splinterOracle, sowingSaltOracle, quashOracle} {
		spec, ok := Lookup(oid)
		if !ok || spec.Completeness != CompletenessFull {
			t.Errorf("%s: registered %v, completeness %v", oid, ok, spec.Completeness)
		}
	}
}
