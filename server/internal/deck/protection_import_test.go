package deck

// protection_import_test.go — #662, the INGESTION half.
//
// Protection is the only keyword whose token carries a parameter, and
// Scryfall's `keywords` array does not carry the parameter: every
// protection card is tagged with the bare word "Protection" and the
// quality lives in the oracle text. So the importer cannot take the
// array at its word, for exactly the reason ADR 0038 §6 gives for
// "Hexproof from" — a bare token would be a badge with nothing behind
// it.
//
// These tests start from the Scryfall record as the dump ships it and
// run the real importer, because a test that set Card.Keywords by
// hand would pass on a build where the importer stamped nothing.

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// baneslayerAngel is the Scryfall record, trimmed to the fields the
// importer reads. Note the array: "Protection" is bare, and the two
// qualities are one comma-separated item of one oracle line.
func baneslayerAngel() cards.Card {
	return cards.Card{
		ID:        uuid.MustParse("55f36e7b-21a0-4d9b-9b0f-4a6f4e0e5d42"),
		OracleID:  uuid.MustParse("0e11792b-7fe5-4208-aa0b-e5d09b2b65fe"),
		Name:      "Baneslayer Angel",
		Layout:    "normal",
		TypeLine:  "Creature — Angel",
		ManaCost:  "{3}{W}{W}",
		Power:     "5",
		Toughness: "5",
		Keywords:  []string{"Flying", "First strike", "Lifelink", "Protection"},
		OracleText: "Flying, first strike, lifelink, protection from Demons and " +
			"from Dragons",
	}
}

// sphinxOfTheGuildpact prints a quality the closed grammar cannot
// parse. It must import with NO protection at all, and it must stay
// flagged by the ADR 0037 coverage signal.
func sphinxOfTheGuildpact() cards.Card {
	return cards.Card{
		ID:        uuid.MustParse("1f1f4d35-6e4c-4b55-9a8f-96c7f0b1f1aa"),
		OracleID:  uuid.MustParse("6f4c1b1e-4c3d-4a6f-8a2b-0c9f1d2e3f40"),
		Name:      "Sphinx of the Guildpact",
		Layout:    "normal",
		TypeLine:  "Artifact Creature — Sphinx",
		ManaCost:  "{5}",
		Power:     "5",
		Toughness: "5",
		Keywords:  []string{"Flying", "Hexproof", "Hexproof from", "Protection"},
		OracleText: "Flying\nSphinx of the Guildpact is all colors.\n" +
			"Protection from monocolored",
	}
}

// korFirewalker is the ordinary single-quality case.
func korFirewalker() cards.Card {
	return cards.Card{
		ID:        uuid.MustParse("2b2b4d35-6e4c-4b55-9a8f-96c7f0b1f1bb"),
		OracleID:  uuid.MustParse("7f4c1b1e-4c3d-4a6f-8a2b-0c9f1d2e3f41"),
		Name:      "Kor Firewalker",
		Layout:    "normal",
		TypeLine:  "Creature — Kor Soldier",
		ManaCost:  "{W}{W}",
		Power:     "2",
		Toughness: "2",
		Keywords:  []string{"Protection"},
		OracleText: "Protection from red\nWhenever a player casts a red spell, " +
			"you may gain 1 life.",
	}
}

func TestProtectionImportsAsParameterisedTokens(t *testing.T) {
	list := &List{Mainboard: []cards.Card{baneslayerAngel(), korFirewalker(), sphinxOfTheGuildpact()}}
	byName := map[string][]string{}
	for _, c := range list.ToGameCards() {
		byName[c.Name] = c.Keywords
	}

	want := map[string][]string{
		// CR 702.16m: one printed clause, two abilities. The bare
		// "Protection" from the array is dropped.
		"Baneslayer Angel": {"flying", "first strike", "lifelink",
			"protection from Demons", "protection from Dragons"},
		"Kor Firewalker": {"protection from red"},
		// The grammar cannot test "monocolored" against a source, so
		// the card carries NEITHER the broad token nor a narrow one —
		// and its hexproof is dropped by the ADR 0038 §6 rule for the
		// same reason.
		"Sphinx of the Guildpact": {"flying"},
	}
	for name, exp := range want {
		got, ok := byName[name]
		if !ok {
			t.Fatalf("%s missing from the imported deck", name)
		}
		if !sameSet(got, exp) {
			t.Errorf("%s: keywords = %v, want %v", name, got, exp)
		}
	}
}

// TestImportedProtectionIsEnforced closes the loop: the token the
// importer stamps is the one the engine's reader parses. An
// uncatalogued Kor Firewalker really does turn a red Bolt away.
func TestImportedProtectionIsEnforced(t *testing.T) {
	list := &List{Mainboard: []cards.Card{korFirewalker()}}
	imported := list.ToGameCards()
	if len(imported) != 1 {
		t.Fatalf("imported %d cards", len(imported))
	}
	c := imported[0]
	qs := game.ProtectionQualities(&c)
	if len(qs) != 1 || qs[0].Value != "R" || qs[0].Printed != "red" {
		t.Fatalf("the engine reads %+v off the imported card", qs)
	}
	if !game.ProtectedFrom(&c, &game.Characteristic{Colors: []string{"R"}}) {
		t.Error("an imported Kor Firewalker is not protected from red")
	}
	if game.ProtectedFrom(&c, &game.Characteristic{Colors: []string{"W"}}) {
		t.Error("...nor should it be protected from white")
	}
}

// TestACardPrintingOnlyProtectionIsNotFlagged is the ADR 0037 side of
// the same coin, in both directions: a quality the engine enforces
// makes the line a keyword line, and one it does not leaves the card
// flagged as unimplemented.
func TestACardPrintingOnlyProtectionIsNotFlagged(t *testing.T) {
	if game.NeedsCatalogEffect("Creature — Kor Soldier", "Protection from red") {
		t.Error("a card whose whole text is an enforced protection needs no catalog Spec")
	}
	if !game.NeedsCatalogEffect("Artifact Creature — Sphinx", "Protection from monocolored") {
		t.Error("a protection the grammar cannot parse must keep the card flagged")
	}
	if !game.NeedsCatalogEffect("Creature — Kor Soldier", "Protection") {
		t.Error("a bare Protection line names no quality and is not a keyword line")
	}
}
