package game

// coverage_test.go pins the honest predicate behind the
// unimplemented-card signal. The interesting assertions are the
// NEGATIVE ones: a signal that fires on everything is a signal
// nobody reads, and most of a real Commander deck has no catalog
// entry and does not need one.

import (
	"reflect"
	"testing"
)

// TestNeedsCatalogEffectSilentOnCardsThatWork covers everything the
// engine already carries out without a Spec. Every one of these
// would be flagged by the naive "not in the catalog" test, and every
// flag would be a lie.
func TestNeedsCatalogEffectSilentOnCardsThatWork(t *testing.T) {
	cases := []struct {
		name     string
		typeLine string
		texts    []string
	}{{
		name:     "vanilla creature",
		typeLine: "Creature — Bear",
		texts:    []string{""},
	}, {
		// Since #317 / #319 / #320 these are enforced for every card
		// in the dump, not just for catalog cards.
		name:     "one keyword",
		typeLine: "Creature — Bird",
		texts:    []string{"Flying"},
	}, {
		name:     "comma-separated keyword line",
		typeLine: "Creature — Angel",
		texts:    []string{"Flying, vigilance"},
	}, {
		name:     "multi-word keywords",
		typeLine: "Creature — Knight",
		texts:    []string{"First strike, deathtouch, lifelink"},
	}, {
		name:     "keyword lines stacked",
		typeLine: "Creature — Elemental",
		texts:    []string{"Flash\nTrample\nHaste"},
	}, {
		// Reminder text names two more keywords and must not be read
		// as a rule of its own.
		name:     "keyword with reminder text",
		typeLine: "Creature — Drake",
		texts: []string{"Flying (This creature can't be blocked except by " +
			"creatures with flying or reach.)"},
	}, {
		// A basic land's mana ability is granted by the type line
		// (CR 305.6); the printed text is reminder only, and
		// ManaAbilitiesForCard synthesises the ability.
		name:     "basic Forest",
		typeLine: "Basic Land — Forest",
		texts:    []string{"({T}: Add {G}.)"},
	}, {
		name:     "snow-covered basic",
		typeLine: "Basic Snow Land — Island",
		texts:    []string{""},
	}, {
		name:     "basic with empty oracle text",
		typeLine: "Basic Land — Mountain",
		texts:    []string{""},
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if NeedsCatalogEffect(tc.typeLine, tc.texts...) {
				t.Errorf("flagged a card the engine handles: %q %q", tc.typeLine, tc.texts)
			}
		})
	}
}

// TestNeedsCatalogEffectCatchesRealRules covers the other side —
// text the engine has no generic path for. The five named cases are
// the five cards behind the 2026-09-10 reports.
func TestNeedsCatalogEffectCatchesRealRules(t *testing.T) {
	cases := []struct {
		name     string
		typeLine string
		texts    []string
	}{{
		name:     "#321 Enduring Curiosity",
		typeLine: "Enchantment Creature — Cat Glimmer",
		texts: []string{"Flash\nWhenever a creature you control deals combat " +
			"damage to a player, draw a card.\nWhen Enduring Curiosity dies, " +
			"if it was a creature, return it to the battlefield under its " +
			"owner's control. It's an enchantment. (It's not a creature.)"},
	}, {
		name:     "#324 Anticausal Vestige",
		typeLine: "Creature — Eldrazi",
		texts: []string{"When this creature leaves the battlefield, draw a " +
			"card, then you may put a permanent card with mana value less " +
			"than or equal to the number of lands you control from your hand " +
			"onto the battlefield tapped.\nWarp {4} (You may cast this card " +
			"from your hand for its warp cost. Exile this creature at the " +
			"beginning of the next end step, then you may cast it from exile " +
			"on a later turn.)"},
	}, {
		// A transform DFC: Scryfall leaves the top-level text null and
		// puts the rules on the faces, so the union is what tells the
		// truth here.
		name:     "#325 Aang, Swift Savior",
		typeLine: "Legendary Creature — Human Avatar Ally // Legendary Creature — Avatar Spirit Ally",
		texts: []string{
			"",
			"Flash\nFlying\nWhen Aang enters, airbend up to one other target " +
				"creature or spell. (Exile it. While it's exiled, its owner may " +
				"cast it for {2} rather than its mana cost.)\nWaterbend {8}: " +
				"Transform Aang.",
			"Reach, trample\nWhenever Aang and La attack, put a +1/+1 counter " +
				"on each tapped creature you control.",
		},
	}, {
		name:     "#332 Lotus Field",
		typeLine: "Land",
		texts: []string{"Hexproof\nThis land enters tapped.\nWhen this land " +
			"enters, sacrifice two lands.\n{T}: Add three mana of any one color."},
	}, {
		name:     "#333 Fortune Teller's Talent",
		typeLine: "Enchantment — Class",
		texts: []string{"(Gain the next level as a sorcery to add its " +
			"ability.)\nYou may look at the top card of your library any " +
			"time.\n{3}{U}: Level 2\nAs long as you've cast a spell this turn, " +
			"you may play cards from the top of your library.\n{2}{U}: Level " +
			"3\nSpells you cast from anywhere other than your hand cost {2} " +
			"less to cast."},
	}, {
		name:     "plain spell",
		typeLine: "Instant",
		texts:    []string{"Lightning Bolt deals 3 damage to any target."},
	}, {
		// A non-basic land gets no synthetic mana ability, so an
		// uncatalogued one taps for nothing at all.
		name:     "uncatalogued dual land",
		typeLine: "Land",
		texts:    []string{"This land enters tapped.\n{T}: Add {W} or {U}."},
	}, {
		// canonicalKeywords is closed on purpose: a keyword outside
		// it is a keyword nothing in the engine acts on.
		name:     "keyword the engine does not enforce",
		typeLine: "Creature — Human Wizard",
		texts:    []string{"Ward {2}"},
	}, {
		name:     "keyword with a cost",
		typeLine: "Artifact — Equipment",
		texts:    []string{"Equipped creature gets +1/+1.\nEquip {2}"},
	}, {
		// Wastes is all reminder text, so the line scan finds
		// nothing — but basicLandColor knows only the five colours,
		// so without a Spec it produces no mana whatsoever.
		name:     "Wastes",
		typeLine: "Basic Land — Wastes",
		texts:    []string{"({T}: Add {C}.)"},
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !NeedsCatalogEffect(tc.typeLine, tc.texts...) {
				t.Errorf("missed a card the engine cannot run: %q %q", tc.typeLine, tc.texts)
			}
		})
	}
}

// TestUnimplementedJoinsTheCatalog proves the two halves of the
// predicate are independent: needing an effect is not the same thing
// as lacking one, and having a Spec is not the same thing as being
// complete without one.
func TestUnimplementedJoinsTheCatalog(t *testing.T) {
	const registered = "9d2460c3-8eeb-4f35-b6f6-748c478664c7"
	prev := IsCatalogCard
	IsCatalogCard = func(oracleID string) bool { return oracleID == registered }
	t.Cleanup(func() { IsCatalogCard = prev })

	cases := []struct {
		name string
		card Card
		want bool
	}{{
		name: "prints rules and has no Spec",
		card: Card{Name: "Lotus Field", OracleID: "134d5b82", NeedsEffect: true},
		want: true,
	}, {
		name: "prints rules and has a Spec",
		card: Card{Name: "Enduring Curiosity", OracleID: registered, NeedsEffect: true},
		want: false,
	}, {
		name: "needs nothing, has no Spec",
		card: Card{Name: "Forest", OracleID: "b34bb2dc", NeedsEffect: false},
		want: false,
	}, {
		// Tokens, fixtures and the demo seed never see the importer.
		// Staying quiet about them is the point.
		name: "never went through deck import",
		card: Card{Name: "Treasure"},
		want: false,
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Unimplemented(tc.card); got != tc.want {
				t.Errorf("Unimplemented(%s) = %v, want %v", tc.card.Name, got, tc.want)
			}
		})
	}
}

// TestUnimplementedNamesDedupes — a deck runs several copies and the
// player wants to hear about the card once, in the order it appears.
func TestUnimplementedNamesDedupes(t *testing.T) {
	prev := IsCatalogCard
	IsCatalogCard = func(string) bool { return false }
	t.Cleanup(func() { IsCatalogCard = prev })

	got := UnimplementedNames([]Card{
		{Name: "Lotus Field", NeedsEffect: true},
		{Name: "Forest"},
		{Name: "Lotus Field", NeedsEffect: true},
		{Name: "Enduring Curiosity", NeedsEffect: true},
		{Name: "Island"},
	})
	want := []string{"Lotus Field", "Enduring Curiosity"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("UnimplementedNames = %v, want %v", got, want)
	}
}

// TestStripReminderTextIsTotal — oracle text is data, and the
// predicate must not panic or loop on shapes no real card prints.
// The two unbalanced cases are asymmetric on purpose: an unclosed
// "(" swallows the rest and so errs toward silence, while an
// unmatched ")" is kept as ordinary text and so errs toward
// flagging. Neither occurs in the dump; both are defined.
func TestStripReminderTextIsTotal(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", ""},
		{"Flying", "Flying"},
		{"Flying (reminder)", "Flying "},
		{"a (b (c) d) e", "a  e"},
		{"unclosed ( swallows\nthe rest", "unclosed "},
		{"stray ) stays", "stray ) stays"},
	}
	for _, tc := range cases {
		if got := stripReminderText(tc.in); got != tc.want {
			t.Errorf("stripReminderText(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
