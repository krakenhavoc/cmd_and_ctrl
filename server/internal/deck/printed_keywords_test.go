package deck

// printed_keywords_test.go is the end-to-end regression suite for
// issues #317 / #319 / #320 — all three reported from the same
// playtest (game 422b603e-131f-447b-86bf-6ffba7e22e81), all three
// the same root cause: printed keywords never travelled from
// Scryfall to game.Card, so they existed only for the handful of
// cards with a hand-written effect-catalog entry.
//
// The tests live in the deck package rather than in game because
// the bug is an INGESTION bug: asserting on a game.Card whose
// Keywords slice the test set by hand would pass on the broken
// build. Each one starts from the Scryfall record as the dump
// actually ships it, runs the real importer, and then drives the
// real engine — deck imports game, so this direction is the only
// one available without an import cycle.

import (
	"errors"
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// skycoachConductor is the Scryfall record for the creature in
// #317, trimmed to the fields the importer reads and otherwise
// copied verbatim from the dump (id f9092a09, set sos). A
// "prepare"-layout card: two faces, top-level mana_cost carrying
// both halves, and — the part that matters — the `keywords` array
// at the TOP level with the faces carrying none of their own.
func skycoachConductor() cards.Card {
	return cards.Card{
		ID:        uuid.MustParse("f9092a09-78a8-421c-868b-115175f4c252"),
		OracleID:  uuid.MustParse("788d3cfa-7706-4728-9c48-cf7bc963d002"),
		Name:      "Skycoach Conductor // All Aboard",
		Layout:    "prepare",
		TypeLine:  "Creature — Bird Pilot // Instant",
		ManaCost:  "{2}{U} // {U}",
		Power:     "2",
		Toughness: "3",
		Keywords:  []string{"Flying", "Vigilance", "Prepared", "Flash"},
		CardFaces: []cards.CardFace{
			{
				Name:     "Skycoach Conductor",
				TypeLine: "Creature — Bird Pilot",
				OracleText: "Flash\nFlying, vigilance\nThis creature enters prepared. " +
					"(While it's prepared, you may cast a copy of its spell. Doing so unprepares it.)",
			},
			{
				Name:       "All Aboard",
				TypeLine:   "Instant",
				OracleText: "Exile target non-Pilot creature you control, then return that card to the battlefield under its owner's control.",
			},
		},
	}
}

// enduringCuriosity is the Scryfall record for the creature in
// #319 (id 8616629e, set dsk). Single-faced, one keyword, no
// catalog entry.
func enduringCuriosity() cards.Card {
	return cards.Card{
		ID:        uuid.MustParse("8616629e-08f9-41ad-bfec-f86c8096f1cb"),
		OracleID:  uuid.MustParse("9d2460c3-8eeb-4f35-b6f6-748c478664c7"),
		Name:      "Enduring Curiosity",
		Layout:    "normal",
		TypeLine:  "Enchantment Creature — Cat Glimmer",
		ManaCost:  "{2}{U}{U}",
		Power:     "4",
		Toughness: "3",
		Keywords:  []string{"Flash"},
	}
}

// aangSwiftSavior is the Scryfall record for the commander in #320
// (id de89fec4, set tla). A transform DFC, so `keywords` is the
// UNION over both faces: flash and flying come off the front, reach
// and trample off the back. Only the front face is ever cast from
// the command zone, so the back face's keywords must not ride along.
func aangSwiftSavior() cards.Card {
	return cards.Card{
		ID:            uuid.MustParse("de89fec4-f5c8-4513-8504-ac9bafb44054"),
		OracleID:      uuid.MustParse("cbf09050-39d0-463b-96db-9e22011ae0d8"),
		Name:          "Aang, Swift Savior // Aang and La, Ocean's Fury",
		Layout:        "transform",
		TypeLine:      "Legendary Creature — Human Avatar Ally // Legendary Creature — Avatar Spirit Ally",
		ColorIdentity: []string{"U", "W"},
		Keywords:      []string{"Flying", "Reach", "Airbend", "Transform", "Trample", "Flash", "Waterbend"},
		CardFaces: []cards.CardFace{
			{
				Name:     "Aang, Swift Savior",
				TypeLine: "Legendary Creature — Human Avatar Ally",
				OracleText: "Flash\nFlying\nWhen Aang enters, airbend up to one other target creature or spell. " +
					"(Exile it. While it's exiled, its owner may cast it for {2} rather than its mana cost.)\n" +
					"Waterbend {8}: Transform Aang.",
			},
			{
				Name:       "Aang and La, Ocean's Fury",
				TypeLine:   "Legendary Creature — Avatar Spirit Ally",
				OracleText: "Reach, trample\nWhenever Aang and La attack, put a +1/+1 counter on each tapped creature you control.",
			},
		},
	}
}

// filler is a second player's minimum legal deck — AddPlayer
// rejects an empty one.
func filler() []game.Card {
	return (&List{Mainboard: []cards.Card{{
		ID:       uuid.New(),
		Name:     "Island",
		TypeLine: "Basic Land — Island",
	}}}).ToGameCards()
}

// startedGame seats the supplied deck at seat 0 and a filler deck at
// seat 1, then starts the game. Returns the game and the seat-0
// player.
func startedGame(t *testing.T, deck []game.Card) (*game.Game, *game.Player) {
	t.Helper()
	g := game.NewGame()
	p, err := g.AddPlayer("reporter", deck)
	if err != nil {
		t.Fatalf("AddPlayer: %v", err)
	}
	if _, err := g.AddPlayer("opponent", filler()); err != nil {
		t.Fatalf("AddPlayer opponent: %v", err)
	}
	if err := g.Start(rand.New(rand.NewPCG(1, 2))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	return g, p
}

// moveTo moves one of the seat's cards into the given zone, from
// wherever Start happened to leave it — a small deck is dealt into
// the opening hand, a larger one stays in the library, and the test
// should not care which.
func moveTo(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID, dst game.ZoneRef) {
	t.Helper()
	for _, kind := range []game.ZoneKind{game.ZoneLibrary, game.ZoneHand} {
		if kind == dst.Kind {
			zone := p.Library
			if kind == game.ZoneHand {
				zone = p.Hand
			}
			for _, c := range zone.Cards {
				if c.InstanceID == id {
					return // already where the caller wants it
				}
			}
		}
	}
	for _, src := range []game.ZoneRef{
		{Kind: game.ZoneLibrary, Owner: p.ID},
		{Kind: game.ZoneHand, Owner: p.ID},
	} {
		if err := g.MoveCardByID(src, dst, id); err == nil {
			return
		}
	}
	t.Fatalf("card %s is in neither library nor hand", id)
}

// battlefieldCard finds a card on the battlefield by instance ID.
func battlefieldCard(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c
		}
	}
	t.Fatalf("card %s not on the battlefield", id)
	return game.Card{}
}

// TestVigilantAttackerDoesNotTap reproduces #317: "The bird pilot
// creature has vigilance but is tapping during declare attackers".
// Skycoach Conductor has printed vigilance and no catalog entry, so
// before the printed-keyword pipeline landed HasKeyword said false
// and DeclareAttacker tapped it as an attack cost (CR 508.1f, and
// CR 702.20 for the exemption it was not getting).
func TestVigilantAttackerDoesNotTap(t *testing.T) {
	list := &List{Mainboard: []cards.Card{skycoachConductor()}}
	deck := list.ToGameCards()
	bird := deck[0].InstanceID

	g, p := startedGame(t, deck)
	opponent := g.Seats[1].ID

	moveTo(t, g, p, bird, game.ZoneRef{Kind: game.ZoneBattlefield})
	// Put the cursor where the reporter was — their own declare-
	// attackers step — and clear summoning sickness the way the
	// untap step they passed through would have.
	g.Turn = game.Turn{
		Number:         3,
		ActiveSeat:     0,
		PriorityHolder: 0,
		Phase:          game.PhaseCombat,
		Step:           game.StepDeclareAttackers,
	}
	if err := g.UntapAll(p.ID); err != nil {
		t.Fatalf("UntapAll: %v", err)
	}

	if err := g.DeclareAttacker(bird, opponent); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if got := battlefieldCard(t, g, bird); got.Tapped {
		t.Errorf("Skycoach Conductor tapped on attack despite printed vigilance (#317)")
	}
}

// TestFlashFromHandCastsAtInstantSpeed reproduces #319: "Enduring
// Curiosity should be castable while I am holding priority here but
// it is not allowing me to do so". The reporter held priority during
// an opponent's combat; the server's sorcery-speed gate refused
// because HasKeyword(card, "flash") was false for every card
// outside the effect catalog.
func TestFlashFromHandCastsAtInstantSpeed(t *testing.T) {
	list := &List{Mainboard: []cards.Card{enduringCuriosity()}}
	deck := list.ToGameCards()
	cat := deck[0].InstanceID

	g, p := startedGame(t, deck)
	moveTo(t, g, p, cat, game.ZoneRef{Kind: game.ZoneHand, Owner: p.ID})
	// Opponent's combat, reporter holding priority — the exact
	// window in the #319 log (seq 95, declare_blockers on seat 0's
	// turn).
	g.Turn = game.Turn{
		Number:         4,
		ActiveSeat:     1,
		PriorityHolder: 0,
		Phase:          game.PhaseCombat,
		Step:           game.StepDeclareBlockers,
	}

	err := g.CastSpell(p.ID, cat, game.CastSpellParams{FromZone: "hand"})
	if errors.Is(err, game.ErrSorcerySpeedRequired) {
		t.Fatalf("Enduring Curiosity refused at instant speed despite printed flash (#319)")
	}
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
}

// TestFlashCommanderCastsFromCommandZone reproduces #320: "When
// trying to cast my commander Aang Swift Savior who has flash I get
// sorcery speed only." The command-zone cast path shares CastSpell's
// one sorcery-speed gate with the hand path, so the same missing
// keyword refused it — the server answered `sorcery speed required`
// twice at 15:05:40 and 15:05:46 in the reporter's log.
func TestFlashCommanderCastsFromCommandZone(t *testing.T) {
	list := &List{Commanders: []cards.Card{aangSwiftSavior()}}
	deck := list.ToGameCards()
	aang := deck[0].InstanceID

	g, p := startedGame(t, deck)
	// AddPlayer routes commanders to the command zone; confirm the
	// test is exercising that path and not a hand cast.
	if len(p.Command.Cards) != 1 || p.Command.Cards[0].InstanceID != aang {
		t.Fatalf("commander not in the command zone: %+v", p.Command.Cards)
	}
	g.Turn = game.Turn{
		Number:         4,
		ActiveSeat:     1,
		PriorityHolder: 0,
		Phase:          game.PhaseCombat,
		Step:           game.StepDeclareAttackers,
	}

	err := g.CastSpell(p.ID, aang, game.CastSpellParams{FromZone: "command"})
	if errors.Is(err, game.ErrSorcerySpeedRequired) {
		t.Fatalf("commander refused at instant speed despite printed flash (#320)")
	}
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
}

// TestPrintedKeywordsStampedOnGameCard pins the ingestion contract
// itself: which of Scryfall's keyword strings reach game.Card, in
// what form, and which are dropped.
func TestPrintedKeywordsStampedOnGameCard(t *testing.T) {
	list := &List{
		Commanders: []cards.Card{aangSwiftSavior()},
		Mainboard:  []cards.Card{skycoachConductor(), enduringCuriosity()},
	}
	byName := map[string][]string{}
	for _, c := range list.ToGameCards() {
		byName[c.Name] = c.Keywords
	}

	want := map[string][]string{
		// Single-faced: Scryfall's list, lowercased, canonical only.
		"Enduring Curiosity": {"flash"},
		// Multi-faced: "prepared" is not a keyword the engine knows,
		// and all three canonical ones are printed on the front face.
		// Keyed on the FRONT FACE's name since ADR 0034 — game.Card
		// carries the active face's name, not Scryfall's composite.
		"Skycoach Conductor": {"flying", "vigilance", "flash"},
		// Multi-faced: Scryfall's top-level array is the union over
		// both faces. Reach and trample belong to the back face and
		// must NOT ride along on the front; airbend, transform and
		// waterbend are not canonical keywords at all.
		"Aang, Swift Savior": {"flying", "flash"},
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

// sameSet compares two keyword slices as sets — the importer's
// output order follows Scryfall's array, which is not a contract.
func sameSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	seen := map[string]int{}
	for _, s := range a {
		seen[s]++
	}
	for _, s := range b {
		seen[s]--
		if seen[s] < 0 {
			return false
		}
	}
	return true
}
