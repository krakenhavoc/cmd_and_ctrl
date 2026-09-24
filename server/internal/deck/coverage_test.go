package deck

// coverage_test.go is the ingestion half of the unimplemented-card
// signal: does the honest predicate reach game.Card through the real
// importer, starting from the Scryfall record as the dump actually
// ships it?
//
// The records below are the cards behind real reports. First the
// 2026-09-10 playtest — #321 Enduring Curiosity, #324 Anticausal
// Vestige, #325 Aang Swift Savior, #332 Lotus Field, #333 Fortune
// Teller's Talent — then the in-app "my trigger didn't fire" set,
// #366, #369, #373 and #508. Every one of them is absent from the
// effect catalog, every one did nothing when cast, and every one was
// reported as a bug by a player the game gave no way to tell that
// apart from a defect. Oracle text is verbatim from the dump.
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
// printings.
//
// Registered since S46 (#757, ADR 0071): levels are a designation
// with one gate on printed abilities. It stays in this table because
// NeedsEffect is the SCRYFALL half of the predicate — "this text
// needs a Spec" — and that is as true after the Spec exists as
// before. The catalog join is game.Unimplemented's job, and it now
// answers false for this card.
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

// theSeriema — #337. Four printed lines, all working: the ETB tutor
// (SearchLibrary takes an arbitrary predicate), the 7+ threshold and
// the artifact-creature type change it implies (a charge-counter gate
// over layers 4, 6 and 7b, S46 / ADR 0071), the indestructible grant,
// and since #759 the station ability itself, on #758's tap-another
// cost.
//
// It stays in this table for fortuneTellersTalent's reason: this
// fixture pins the Scryfall half of the predicate, which does not
// change when a Spec is written.
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
// reported twice. NeedsEffect is the Scryfall half of the predicate
// only, so the stamp stays true now that the card is catalogued: #1313
// shipped its "doesn't untap for as long as you control Ty Lee" hold
// (ADR 0058's 2026-09-23 amendment). Prowess is still #706.
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

// --- the in-app "my trigger didn't fire" set ----------------------
//
// #366, #369, #373 and #508 were filed from the client over four
// playtest days, name four unrelated cards and four unrelated
// triggers, and are one fact: none of the four is in the effect
// catalog, so none of them has a TriggeredAbility, so there was
// never a trigger to fire. The value of pinning them here is that
// the signal is the ONLY thing standing between a player and that
// same report — see the note on TestImporterStampsNeedsEffect.

// brassTunnelGrinder — #366, "enters an artifact but no prompt or
// resolution works for its ETB". A transform DFC, so Scryfall leaves
// the top-level oracle text empty and puts both halves on the faces:
// this record is the one in the file that fails if oracleTexts ever
// stops unioning the faces, because the top level alone reads as a
// complete card with nothing to say.
//
// Note the name: the catalog and two of the docs spell it "Tunnel
// Grinder", the dump spells it "Tunnel-Grinder", and a hyphen is why
// a repo-wide grep for the reported name comes back empty twice over.
func brassTunnelGrinder() cards.Card {
	return cards.Card{
		ID:            uuid.MustParse("d61d8895-7f2e-4c77-951f-4f1a49e96f57"),
		OracleID:      uuid.MustParse("af1553eb-4f9f-4335-9078-56649bd8d8fc"),
		Name:          "Brass's Tunnel-Grinder // Tecutlan, the Searing Rift",
		Layout:        "transform",
		TypeLine:      "Legendary Artifact // Legendary Land — Cave",
		ColorIdentity: []string{"R"},
		Keywords:      []string{"Transform", "Discover"},
		CardFaces: []cards.CardFace{
			{
				Name:     "Brass's Tunnel-Grinder",
				TypeLine: "Legendary Artifact",
				ManaCost: "{2}{R}",
				OracleText: "When Brass's Tunnel-Grinder enters, discard any number of cards, " +
					"then draw that many cards plus one.\n" +
					"At the beginning of your end step, if you descended this turn, put a " +
					"bore counter on Brass's Tunnel-Grinder. Then if there are three or more " +
					"bore counters on it, remove those counters and transform it. (You " +
					"descended if a permanent card was put into your graveyard from anywhere.)",
			},
			{
				Name:     "Tecutlan, the Searing Rift",
				TypeLine: "Legendary Land — Cave",
				OracleText: "(Transforms from Brass's Tunnel-Grinder.)\n{T}: Add {R}.\n" +
					"Whenever you cast a permanent spell using mana produced by Tecutlan, " +
					"discover X, where X is that spell's mana value.",
			},
		},
	}
}

// monumentToEndurance — #369, "discarding a card does not trigger it
// with selection". The discard half of what the card wants exists and
// works: EventDiscardCard is emitted by the one discard path and Mary
// Read and Anne Bonny consumes it in the same game. What is missing is
// the card file, and behind it the per-source "choose one that hasn't
// been chosen this turn" tally (#764, ADR 0065 §5).
//
// The bullets are the interesting shape for the scanner: a modal line
// starts with "•", which is not a keyword, so each one reads as a
// rule — correctly.
func monumentToEndurance() cards.Card {
	return cards.Card{
		ID:       uuid.MustParse("d21433ba-0a14-42bc-ad0b-a4ef823a3295"),
		OracleID: uuid.MustParse("e69e8de4-b521-4888-8074-17f1efe2f345"),
		Name:     "Monument to Endurance",
		Layout:   "normal",
		TypeLine: "Artifact",
		ManaCost: "{3}",
		Keywords: []string{"Treasure"},
		OracleText: "Whenever you discard a card, choose one that hasn't been chosen this turn —\n" +
			"• Draw a card.\n" +
			"• Create a Treasure token.\n" +
			"• Each opponent loses 3 life.",
	}
}

// theMightyThorJaneFoster — #373, "her triggered ability did not
// trigger for choosing a creature and exile and then return". Flying
// is printed and genuinely works; the two trigger lines under it are
// the report. ADR 0027:150 deferred this card on one word — "flicker"
// — and that deferral has since expired (the flicker helpers and
// ReturnFromExile{Tapped: true} both exist, and attack triggers are
// routine), so the card is no longer blocked, only unwritten.
func theMightyThorJaneFoster() cards.Card {
	return cards.Card{
		ID:            uuid.MustParse("082cc8cc-bbea-4ca7-a0e8-da1f865d6626"),
		OracleID:      uuid.MustParse("57d02dc8-e22e-4874-9f02-490a2528a28f"),
		Name:          "The Mighty Thor, Jane Foster",
		Layout:        "normal",
		TypeLine:      "Legendary Creature — Human God Hero",
		ManaCost:      "{1}{W}{U}",
		Power:         "3",
		Toughness:     "3",
		Colors:        []string{"U", "W"},
		ColorIdentity: []string{"U", "W"},
		Keywords:      []string{"Flying"},
		OracleText: "Flying\n" +
			"Whenever The Mighty Thor attacks, exile up to one target nontoken artifact " +
			"or creature, then return that card to the battlefield tapped under its " +
			"owner's control.\n" +
			"Whenever an Equipment you control enters, draw a card.",
	}
}

// ambrosiaWhiteheart — #508, "entering the battlefield did not
// trigger its ability to return target permanent". Flash works. The
// ETB bounce is expressible today and the landfall pump is not — it
// wants an until-end-of-turn continuous effect (#279), which is why
// the deck triage filed the whole card as blocked rather than
// shipping half of it (ADR 0037 §5).
//
// "Landfall —" is an ability word: printed in italics, rules-free by
// CR 207.2c. It shares a line with the trigger it labels, so the line
// is a rule either way and the scanner never has to know that.
func ambrosiaWhiteheart() cards.Card {
	return cards.Card{
		ID:            uuid.MustParse("f2596767-7d19-4110-86ed-3cfc93ac7483"),
		OracleID:      uuid.MustParse("2bcc9f11-5b12-433e-9680-0f4b18aa521c"),
		Name:          "Ambrosia Whiteheart",
		Layout:        "normal",
		TypeLine:      "Legendary Creature — Bird",
		ManaCost:      "{1}{W}",
		Power:         "2",
		Toughness:     "2",
		Colors:        []string{"W"},
		ColorIdentity: []string{"W"},
		Keywords:      []string{"Flash", "Landfall"},
		OracleText: "Flash\n" +
			"When Ambrosia Whiteheart enters, you may return another permanent you " +
			"control to its owner's hand.\n" +
			"Landfall — Whenever a land you control enters, Ambrosia Whiteheart gets " +
			"+1/+0 until end of turn.",
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
		{"#366 Brass's Tunnel-Grinder", brassTunnelGrinder(), true},
		{"#369 Monument to Endurance", monumentToEndurance(), true},
		{"#373 The Mighty Thor, Jane Foster", theMightyThorJaneFoster(), true},
		{"#508 Ambrosia Whiteheart", ambrosiaWhiteheart(), true},
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
