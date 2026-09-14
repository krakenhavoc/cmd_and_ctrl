package deck

// coverage_test.go is the ingestion half of the unimplemented-card
// signal: does the honest predicate reach game.Card through the real
// importer, starting from the Scryfall record as the dump actually
// ships it?
//
// The five records below are the five cards behind the 2026-09-10
// reports — #321 Enduring Curiosity, #324 Anticausal Vestige, #325
// Aang Swift Savior, #332 Lotus Field, #333 Fortune Teller's Talent.
// All five are absent from the effect catalog, all five did nothing
// when cast, and all five were reported as bugs by a player the game
// gave no way to tell that apart from a defect. Oracle text is
// verbatim from the 629 MB default-cards dump.
//
// Same reasoning as printed_keywords_test.go for living here rather
// than in game: asserting on a game.Card whose NeedsEffect the test
// set by hand would pass on a broken importer.

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// enduringCuriosityFull is #319's record with the oracle text
// printed_keywords_test.go had no reason to carry. Flash alone works
// (that was #319's fix); the combat-damage draw and the dies-return
// are what #321 is about.
func enduringCuriosityFull() cards.Card {
	c := enduringCuriosity()
	c.OracleText = "Flash\n" +
		"Whenever a creature you control deals combat damage to a player, draw a card.\n" +
		"When Enduring Curiosity dies, if it was a creature, return it to the " +
		"battlefield under its owner's control. It's an enchantment. (It's not a creature.)"
	return c
}

// anticausalVestige — #324. Warp {4} is the reported half; the
// leaves-the-battlefield clause is the half that makes the card
// unshippable even once warp exists.
func anticausalVestige() cards.Card {
	return cards.Card{
		// Printing IDs are synthetic throughout this file where the
		// dump was not consulted for one; nothing here reads them.
		// The oracle IDs ARE the dump's, because those are the keys
		// the catalog joins on.
		ID:        uuid.New(),
		OracleID:  uuid.MustParse("aceea999-90ab-472c-86a7-48af1542cbcf"),
		Name:      "Anticausal Vestige",
		Layout:    "normal",
		TypeLine:  "Creature — Eldrazi",
		ManaCost:  "{6}",
		Power:     "7",
		Toughness: "5",
		Keywords:  []string{"Warp"},
		OracleText: "When this creature leaves the battlefield, draw a card, then " +
			"you may put a permanent card with mana value less than or equal to " +
			"the number of lands you control from your hand onto the battlefield tapped.\n" +
			"Warp {4} (You may cast this card from your hand for its warp cost. " +
			"Exile this creature at the beginning of the next end step, then you " +
			"may cast it from exile on a later turn.)",
	}
}

// lotusField — #332. Hexproof is printed but the engine does not
// enforce it, "enters tapped" and the ETB sacrifice are both real
// rules, and "add three mana of any one color" has no expression in
// the Produced grammar.
func lotusField() cards.Card {
	return cards.Card{
		ID:       uuid.New(),
		OracleID: uuid.MustParse("134d5b82-7940-4b33-a922-7f9d1f403e50"),
		Name:     "Lotus Field",
		Layout:   "normal",
		TypeLine: "Land",
		Keywords: []string{"Hexproof"},
		OracleText: "Hexproof\nThis land enters tapped.\n" +
			"When this land enters, sacrifice two lands.\n" +
			"{T}: Add three mana of any one color.",
	}
}

// fortuneTellersTalent — #333. The reporter called it a saga; it is
// a Class, Scryfall layout "class", single-faced in all (one) of its
// printings. Levels are the gap either way.
func fortuneTellersTalent() cards.Card {
	return cards.Card{
		ID:       uuid.New(),
		OracleID: uuid.MustParse("1b430f67-3686-4452-8594-b060f1a5a04e"),
		Name:     "Fortune Teller's Talent",
		Layout:   "class",
		TypeLine: "Enchantment — Class",
		ManaCost: "{U}",
		OracleText: "(Gain the next level as a sorcery to add its ability.)\n" +
			"You may look at the top card of your library any time.\n" +
			"{3}{U}: Level 2\n" +
			"As long as you've cast a spell this turn, you may play cards from " +
			"the top of your library.\n" +
			"{2}{U}: Level 3\n" +
			"Spells you cast from anywhere other than your hand cost {2} less to cast.",
	}
}

// clone — #335. Printed 0/0, which is the whole of the reported
// symptom: with no Spec there is no copy effect, so the printed P/T
// stands, and the 704.5f SBA's printed-0 exemption spares it rather
// than sweeping it. The other half of the report, "does not allow
// the selection of any usable target", is the same absence seen from
// the client: a card with no catalog entry carries no target_mode,
// and Board.continueCast fires cast_spell immediately. Clone's copy
// choice was never targeting in the first place (CR 614.1c, an
// as-enters replacement), so no amount of targeting work would have
// produced the prompt the reporter expected.
func clone() cards.Card {
	return cards.Card{
		ID:         uuid.New(),
		OracleID:   uuid.MustParse("42226b87-0746-4ebf-9fd0-108d508462af"),
		Name:       "Clone",
		Layout:     "normal",
		TypeLine:   "Creature — Shapeshifter",
		ManaCost:   "{3}{U}",
		Power:      "0",
		Toughness:  "0",
		OracleText: "You may have this creature enter as a copy of any creature on the battlefield.",
	}
}

// theSeriema — #337. Four printed lines and only the first is cheap:
// SearchLibrary takes an arbitrary predicate, so "a legendary
// creature card" is three lines of Go. Station, the 7+ threshold,
// the type change it implies and the indestructible grant are all
// absent from the engine.
func theSeriema() cards.Card {
	return cards.Card{
		ID:        uuid.New(),
		OracleID:  uuid.MustParse("a6bcec1f-f515-4e63-9e84-8eb04cc582ff"),
		Name:      "The Seriema",
		Layout:    "normal",
		TypeLine:  "Legendary Artifact — Spacecraft",
		ManaCost:  "{1}{W}{W}",
		Power:     "5",
		Toughness: "5",
		Keywords:  []string{"Station"},
		OracleText: "When The Seriema enters, search your library for a legendary " +
			"creature card, reveal it, put it into your hand, then shuffle.\n" +
			"Station (Tap another creature you control: Put charge counters equal " +
			"to its power on this Spacecraft. Station only as a sorcery. It's an " +
			"artifact creature at 7+.)\n" +
			"7+ | Flying\n" +
			"Other tapped legendary creatures you control have indestructible.",
	}
}

// tyLeeChiBlocker — #339 and #340, the same card and the same ETB
// reported twice. Flash works (printed keyword, #319's fix). Prowess
// does not, and neither does the lockdown. #74 gave the untap step a
// hook, but it is the wrong DIRECTION: UntapStepPermission ADDS
// permanents to CR 502.1's set (Seedborn Muse), and "it doesn't
// untap during its controller's next untap step" has to remove one.
// The place for that is untapStepSetLocked, alongside the permission
// leg. The duration is the other half of the problem: #314's
// TurnScopedStatics is swept at cleanup, which is the wrong clock for
// "for as long as you control Ty Lee".
func tyLeeChiBlocker() cards.Card {
	return cards.Card{
		ID:        uuid.New(),
		OracleID:  uuid.MustParse("081ad4e3-cda3-41cc-890f-412611dc9ea0"),
		Name:      "Ty Lee, Chi Blocker",
		Layout:    "normal",
		TypeLine:  "Legendary Creature — Human Performer Ally",
		ManaCost:  "{2}{U}",
		Power:     "2",
		Toughness: "1",
		Keywords:  []string{"Prowess", "Flash"},
		OracleText: "Flash\n" +
			"Prowess (Whenever you cast a noncreature spell, this creature gets " +
			"+1/+1 until end of turn.)\n" +
			"When Ty Lee enters, tap up to one target creature. It doesn't untap " +
			"during its controller's untap step for as long as you control Ty Lee.",
	}
}

// grizzlyBears is the control: a complete, correct card with no
// catalog entry and nothing to say about it.
func grizzlyBears() cards.Card {
	return cards.Card{
		ID:        uuid.New(),
		OracleID:  uuid.New(),
		Name:      "Grizzly Bears",
		Layout:    "normal",
		TypeLine:  "Creature — Bear",
		ManaCost:  "{1}{G}",
		Power:     "2",
		Toughness: "2",
	}
}

// serraAngel is the other control, and the more important one: four
// printed keywords, no oracle text beyond them, no catalog entry —
// and since #317 / #319 / #320 the engine flies it, gives it
// vigilance, and resolves its combat correctly. A signal that fired
// here would be the signal crying wolf.
func serraAngel() cards.Card {
	return cards.Card{
		ID:         uuid.New(),
		OracleID:   uuid.New(),
		Name:       "Serra Angel",
		Layout:     "normal",
		TypeLine:   "Creature — Angel",
		ManaCost:   "{3}{W}{W}",
		Power:      "4",
		Toughness:  "4",
		Keywords:   []string{"Flying", "Vigilance"},
		OracleText: "Flying, vigilance",
	}
}

// island is the land control — a basic taps for mana off its type
// line (CR 305.6) with no Spec anywhere in sight.
func island() cards.Card {
	return cards.Card{
		ID:         uuid.New(),
		OracleID:   uuid.New(),
		Name:       "Island",
		Layout:     "normal",
		TypeLine:   "Basic Land — Island",
		OracleText: "({T}: Add {U}.)",
	}
}

// TestImporterStampsNeedsEffect runs every record through the real
// toGameCard and checks the stamp. NeedsEffect is the Scryfall half
// of the predicate only — the catalog join happens in
// game.Unimplemented — so this test is independent of whether the
// effects package is linked into the test binary.
func TestImporterStampsNeedsEffect(t *testing.T) {
	cases := []struct {
		name string
		card cards.Card
		want bool
	}{
		{"#321 Enduring Curiosity", enduringCuriosityFull(), true},
		{"#324 Anticausal Vestige", anticausalVestige(), true},
		{"#325 Aang, Swift Savior", aangSwiftSavior(), true},
		{"#332 Lotus Field", lotusField(), true},
		{"#333 Fortune Teller's Talent", fortuneTellersTalent(), true},
		{"#335 Clone", clone(), true},
		{"#337 The Seriema", theSeriema(), true},
		{"#339/#340 Ty Lee, Chi Blocker", tyLeeChiBlocker(), true},
		{"vanilla creature", grizzlyBears(), false},
		{"keywords only", serraAngel(), false},
		{"basic land", island(), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := toGameCard(tc.card, false)
			if got.NeedsEffect != tc.want {
				t.Errorf("%s: NeedsEffect = %v, want %v", tc.card.Name, got.NeedsEffect, tc.want)
			}
		})
	}
}

// TestUnimplementedNamesThroughTheImporter is the deck-upload
// summary end to end: a mixed decklist in, the honest subset out, in
// decklist order. With no effects import in this binary
// game.IsCatalogCard is nil, which reads as "no catalog wired" — the
// join is exercised in game's own tests, so what this pins is that
// the names survive ToGameCards and arrive in order.
func TestUnimplementedNamesThroughTheImporter(t *testing.T) {
	list := &List{
		Commanders: []cards.Card{aangSwiftSavior()},
		Mainboard: []cards.Card{
			island(),
			lotusField(),
			serraAngel(),
			island(),
			enduringCuriosityFull(),
			grizzlyBears(),
			lotusField(),
		},
	}
	got := game.UnimplementedNames(list.ToGameCards())
	want := []string{
		// ADR 0034: game.Card.Name is the ACTIVE FACE's name, so a
		// multi-face card reports its front face rather than
		// Scryfall's composite. Same card, named the way it reads
		// on the table.
		"Aang, Swift Savior",
		"Lotus Field",
		"Enduring Curiosity",
	}
	if len(got) != len(want) {
		t.Fatalf("UnimplementedNames = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("UnimplementedNames[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
