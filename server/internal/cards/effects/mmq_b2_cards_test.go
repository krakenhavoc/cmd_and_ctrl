package effects

import (
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mmq_b2_cards_test.go — Mercadian Masques, batch 2 (#5619). Each test
// plays the real catalog entry and asserts what its text says. The
// seeding and casting helpers are batch 1's (mmq_b1_cards_test.go).

const (
	mmqB2RushwoodLegateOracle  = "f3623759-bf3a-4846-9ef7-1b732ad4355f"
	mmqB2CaptainOracle         = "30255f36-2e26-4990-8b54-27f38130d45b"
	mmqB2CommanderOracle       = "05baf38c-677a-42e3-bb8c-a52a0a478984"
	mmqB2SkyMarshalOracle      = "52fcc5b6-6ada-457f-be76-fa674c74b4fd"
	mmqB2PersuaderOracle       = "415e9484-0788-4c00-bb1f-19308511114b"
	mmqB2BruteOracle           = "f60d03ae-f45f-45b2-a0f7-108acb0a45db"
	mmqB2KidnappersOracle      = "938c4f2e-6dda-4ac4-b50c-40dc165b4eaa"
	mmqB2EnforcerOracle        = "a2f3139e-80f6-43fb-bdb7-d45d6eb2a652"
	mmqB2SlaverOracle          = "e7989ef3-d2a5-41c6-944c-f969d850f557"
	mmqB2OverlordOracle        = "c2223cdf-419d-4153-9551-705557461664"
	mmqB2UndertakerOracle      = "a8ffbf6d-44e9-44e4-846c-4f894078130e"
	mmqB2BouncerOracle         = "b53a8cde-494c-43d0-8d08-7bdfe2e84064"
	mmqB2DevoutWitnessOracle   = "83647f5b-2b0f-4f9d-83c5-75c52460f35b"
	mmqB2TonicPeddlerOracle    = "9039f59a-915b-4b16-a5c4-1a5fbfb0b639"
	mmqB2DeepwoodDrummerOracle = "b2bbb4e5-05b1-49e5-ad02-7384e115e06b"
	mmqB2SeismicMageOracle     = "a1f7e43a-f126-47d8-bd59-339ae052e81b"
	mmqB2ReverentMantraOracle  = "b4f7b1eb-6cd4-40f9-b259-3c39291edf27"
)

func TestMmqB2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		mmqB2RushwoodLegateOracle:  "Rushwood Legate",
		mmqB2CaptainOracle:         "Ramosian Captain",
		mmqB2CommanderOracle:       "Ramosian Commander",
		mmqB2SkyMarshalOracle:      "Ramosian Sky Marshal",
		mmqB2PersuaderOracle:       "Cateran Persuader",
		mmqB2BruteOracle:           "Cateran Brute",
		mmqB2KidnappersOracle:      "Cateran Kidnappers",
		mmqB2EnforcerOracle:        "Cateran Enforcer",
		mmqB2SlaverOracle:          "Cateran Slaver",
		mmqB2OverlordOracle:        "Cateran Overlord",
		mmqB2UndertakerOracle:      "Undertaker",
		mmqB2BouncerOracle:         "Waterfront Bouncer",
		mmqB2DevoutWitnessOracle:   "Devout Witness",
		mmqB2TonicPeddlerOracle:    "Tonic Peddler",
		mmqB2DeepwoodDrummerOracle: "Deepwood Drummer",
		mmqB2SeismicMageOracle:     "Seismic Mage",
		mmqB2ReverentMantraOracle:  "Reverent Mantra",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// mmqB2Seats is the active seat and the next one, in a main phase.
func mmqB2Seats(t *testing.T, g *game.Game) (me, opp *game.Player) {
	t.Helper()
	advanceToMain(t, g)
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

// --- Rushwood Legate: free if an opponent has an Island and you a Forest ----

func TestMmqB2RushwoodLegateIsFreeOnlyWithTheirIslandAndYourForest(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	legate := handCardFull(me, "Rushwood Legate", "Creature — Dryad", "{2}{G}", mmqB2RushwoodLegateOracle, []string{"G"})
	seedLand(g, them.ID, "Island", "Basic Land — Island", "")
	if err := mmqB1Cast(g, me.ID, legate, "free", nil, nil); err == nil {
		t.Fatal("cast free with an opponent's Island but no Forest of your own")
	}
	seedLand(g, me.ID, "Forest", "Basic Land — Forest", "")
	if err := mmqB1Cast(g, me.ID, legate, "free", nil, nil); err != nil {
		t.Fatalf("free cast with an opponent's Island and your Forest: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(legate) {
		t.Fatal("the Legate did not enter")
	}
}

func TestMmqB2RushwoodLegateIgnoresYourOwnIsland(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	legate := handCardFull(me, "Rushwood Legate", "Creature — Dryad", "{2}{G}", mmqB2RushwoodLegateOracle, []string{"G"})
	seedLand(g, me.ID, "Forest", "Basic Land — Forest", "")
	seedLand(g, me.ID, "Island", "Basic Land — Island", "")
	if err := mmqB1Cast(g, me.ID, legate, "free", nil, nil); err == nil {
		t.Error("cast free off your own Island")
	}
}

// --- the Ramosian and Cateran searchers -------------------------------------

var mmqB2Searchers = []struct {
	name, oracle, typeLine, subtype, keyword string
	pay, max                                 int
	color                                    string
}{
	{"Ramosian Captain", mmqB2CaptainOracle, "Creature — Human Rebel", "Rebel", "first strike", 5, 4, "W"},
	{"Ramosian Commander", mmqB2CommanderOracle, "Creature — Human Rebel", "Rebel", "", 6, 5, "W"},
	{"Ramosian Sky Marshal", mmqB2SkyMarshalOracle, "Creature — Human Rebel", "Rebel", "flying", 7, 6, "W"},
	{"Cateran Persuader", mmqB2PersuaderOracle, "Creature — Human Mercenary", "Mercenary", "", 1, 1, "B"},
	{"Cateran Brute", mmqB2BruteOracle, "Creature — Horror Mercenary", "Mercenary", "", 2, 2, "B"},
	{"Cateran Kidnappers", mmqB2KidnappersOracle, "Creature — Human Mercenary", "Mercenary", "", 3, 3, "B"},
	{"Cateran Enforcer", mmqB2EnforcerOracle, "Creature — Horror Mercenary", "Mercenary", "fear", 4, 4, "B"},
	{"Cateran Slaver", mmqB2SlaverOracle, "Creature — Horror Mercenary", "Mercenary", "swampwalk", 5, 5, "B"},
	{"Cateran Overlord", mmqB2OverlordOracle, "Creature — Horror Mercenary", "Mercenary", "", 6, 6, "B"},
}

// mmqB2CostOfValue is a mana cost of colour `color` with mana value mv.
func mmqB2CostOfValue(mv int, color string) string {
	if mv <= 1 {
		return "{" + color + "}"
	}
	return "{" + strconv.Itoa(mv-1) + "}{" + color + "}"
}

func TestMmqB2SearchersFindTheirSubtypeWithinTheirManaValue(t *testing.T) {
	for _, tc := range mmqB2Searchers {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, _ := mmqB2Seats(t, g)
			seedSearchLibrary(me,
				game.Card{Name: "Edge Fit", TypeLine: "Creature — Human " + tc.subtype, ManaCost: mmqB2CostOfValue(tc.max, tc.color)},
				game.Card{Name: "Too Big", TypeLine: "Creature — Human " + tc.subtype, ManaCost: mmqB2CostOfValue(tc.max+1, tc.color)},
				game.Card{Name: "Kindred Spell", TypeLine: "Kindred Sorcery — " + tc.subtype, ManaCost: "{" + tc.color + "}"},
				game.Card{Name: "Wrong Type", TypeLine: "Creature — Human Soldier", ManaCost: "{" + tc.color + "}"},
			)
			row := 0
			if tc.name == "Cateran Overlord" {
				row = 1
			}
			id := mmqB1Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2, tc.color)
			dcMana(t, g, me.ID, strings.Repeat("{C}", tc.pay-1))
			if err := g.ActivateCatalogAbility(me.ID, id, row, game.ActivateAbilityParams{Strict: true}); err == nil {
				t.Fatalf("activated for one less than {%d}", tc.pay)
			}
			dcMana(t, g, me.ID, "{C}")
			if err := g.ActivateCatalogAbility(me.ID, id, row, game.ActivateAbilityParams{Strict: true}); err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
			if !b16Tapped(t, g, id) {
				t.Error("the searcher did not tap")
			}
			passPriorityAroundTable(t, g)
			c := searchChoiceFor(g, me.ID)
			if c == nil {
				t.Fatal("the search names a quality, so it asks")
			}
			for _, name := range []string{"Too Big", "Kindred Spell", "Wrong Type"} {
				if searchOptionNamed(g, c, name) != uuid.Nil {
					t.Errorf("%s was offered", name)
				}
			}
			answerSearchNamed(t, g, me.ID, "Edge Fit")
			if findBattlefieldByName(g, "Edge Fit") == uuid.Nil {
				t.Errorf("the %s with mana value %d did not enter the battlefield", tc.subtype, tc.max)
			}
		})
	}
}

func TestMmqB2SearchersHaveTheirPrintedKeywords(t *testing.T) {
	for _, tc := range mmqB2Searchers {
		if tc.keyword == "" {
			continue
		}
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := mmqB1Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 2, 2, tc.color)
			if !hasEffectiveKeyword(t, g, id, tc.keyword) {
				t.Errorf("%s does not have %s", tc.name, tc.keyword)
			}
		})
	}
}

// --- Cateran Overlord: sacrifice a creature to regenerate --------------------

func TestMmqB2CateranOverlordSacrificesACreatureToRegenerate(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := mmqB2Seats(t, g)
	overlord := mmqB1Push(g, me.ID, "Cateran Overlord", "Creature — Horror Mercenary", mmqB2OverlordOracle, 7, 5, "B")
	fodder := mmqB1Push(g, me.ID, "Fodder", "Creature — Bear", "", 2, 2, "G")
	if err := g.ActivateCatalogAbility(me.ID, overlord, 0, game.ActivateAbilityParams{
		Strict: true, SacrificeIDs: []uuid.UUID{fodder},
	}); err != nil {
		t.Fatalf("regenerate: %v", err)
	}
	if g.Battlefield.Contains(fodder) {
		t.Fatal("the creature was not sacrificed as the cost")
	}
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(overlord) })
	if !g.Battlefield.Contains(overlord) {
		t.Fatal("the regenerated Overlord was destroyed")
	}
	if !b16Tapped(t, g, overlord) {
		t.Error("regeneration taps the Overlord")
	}
}

// --- the spellshapers: mana, {T}, discard a card ----------------------------

// mmqB2Shape activates a spellshaper's only ability with `mana` floated
// (nothing more when empty: a refused activation leaves the pool as it
// was), discarding `discard` (none when uuid.Nil).
func mmqB2Shape(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID, mana string, discard uuid.UUID, targets []game.TargetRef) error {
	t.Helper()
	if mana != "" {
		dcMana(t, g, me.ID, mana)
	}
	params := game.ActivateAbilityParams{Strict: true, Targets: targets}
	if discard != uuid.Nil {
		params.DiscardIDs = []uuid.UUID{discard}
	}
	return g.ActivateCatalogAbility(me.ID, id, 0, params)
}

// mmqB2Discards puts a card in hand, activates with it as the discard,
// and checks it went to the graveyard and the shaper tapped.
func mmqB2Discards(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID, mana string, targets []game.TargetRef) {
	t.Helper()
	fodder := handCardFull(me, "Fodder", "Sorcery", "{1}", "", nil)
	if err := mmqB2Shape(t, g, me, id, mana, fodder, targets); err != nil {
		t.Fatalf("activation: %v", err)
	}
	if me.Hand.Contains(fodder) || !me.Graveyard.Contains(fodder) {
		t.Fatal("the card was not discarded as the cost")
	}
	if !b16Tapped(t, g, id) {
		t.Error("the spellshaper did not tap")
	}
	passPriorityAroundTable(t, g)
}

func TestMmqB2SpellshapersNeedADiscard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := mmqB2Seats(t, g)
	peddler := mmqB1Push(g, me.ID, "Tonic Peddler", "Creature — Human Spellshaper", mmqB2TonicPeddlerOracle, 1, 1, "W")
	if err := mmqB2Shape(t, g, me, peddler, "{W}", uuid.Nil, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err == nil {
		t.Error("Tonic Peddler was activated without discarding a card")
	}
}

func TestMmqB2UndertakerReturnsACreatureCardFromYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := mmqB2Seats(t, g)
	shaper := mmqB1Push(g, me.ID, "Undertaker", "Creature — Human Spellshaper", mmqB2UndertakerOracle, 1, 1, "B")
	dead := pushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	spell := pushGraveyardCardTyped(me, "Old Sorcery", "Sorcery")
	if err := mmqB2Shape(t, g, me, shaper, "{B}", handCardFull(me, "Spare", "Sorcery", "{1}", "", nil), cardRefs(spell)); err == nil {
		t.Fatal("Undertaker targeted a noncreature card")
	}
	mmqB2Discards(t, g, me, shaper, "", cardRefs(dead))
	if !me.Hand.Contains(dead) || me.Graveyard.Contains(dead) {
		t.Error("the creature card did not return to hand")
	}
}

func TestMmqB2WaterfrontBouncerReturnsACreatureToItsOwnersHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := mmqB2Seats(t, g)
	shaper := mmqB1Push(g, me.ID, "Waterfront Bouncer", "Creature — Merfolk Spellshaper", mmqB2BouncerOracle, 1, 1, "U")
	theirs := mmqB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2, "G")
	mmqB2Discards(t, g, me, shaper, "{U}", cardRefs(theirs))
	if g.Battlefield.Contains(theirs) || !opp.Hand.Contains(theirs) {
		t.Error("the creature did not return to its owner's hand")
	}
}

func TestMmqB2DevoutWitnessDestroysAnArtifactOrEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := mmqB2Seats(t, g)
	shaper := mmqB1Push(g, me.ID, "Devout Witness", "Creature — Human Spellshaper", mmqB2DevoutWitnessOracle, 2, 2, "W")
	bear := mmqB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2, "G")
	aura := mmqB1Push(g, opp.ID, "Their Enchantment", "Enchantment", "", 0, 0, "W")
	if err := mmqB2Shape(t, g, me, shaper, "{C}{W}", handCardFull(me, "Spare", "Sorcery", "{1}", "", nil), cardRefs(bear)); err == nil {
		t.Fatal("Devout Witness targeted a creature")
	}
	mmqB2Discards(t, g, me, shaper, "", cardRefs(aura))
	if g.Battlefield.Contains(aura) {
		t.Error("the enchantment survived")
	}
}

func TestMmqB2SeismicMageDestroysALand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := mmqB2Seats(t, g)
	shaper := mmqB1Push(g, me.ID, "Seismic Mage", "Creature — Human Spellshaper", mmqB2SeismicMageOracle, 1, 1, "R")
	land := seedLand(g, opp.ID, "Forest", "Basic Land — Forest", "")
	mmqB2Discards(t, g, me, shaper, "{C}{C}{R}", cardRefs(land))
	if g.Battlefield.Contains(land) {
		t.Error("the land survived")
	}
}

func TestMmqB2DeepwoodDrummerGivesPlusTwoPlusTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := mmqB2Seats(t, g)
	shaper := mmqB1Push(g, me.ID, "Deepwood Drummer", "Creature — Human Spellshaper", mmqB2DeepwoodDrummerOracle, 1, 1, "G")
	bear := mmqB1Push(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2, "G")
	mmqB2Discards(t, g, me, shaper, "{G}", cardRefs(bear))
	if p, tough := ptOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the bear is %d/%d, want 4/4", p, tough)
	}
}

func TestMmqB2TonicPeddlerGivesAnyPlayerThreeLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := mmqB2Seats(t, g)
	shaper := mmqB1Push(g, me.ID, "Tonic Peddler", "Creature — Human Spellshaper", mmqB2TonicPeddlerOracle, 1, 1, "W")
	mine, theirs := me.Life, opp.Life
	mmqB2Discards(t, g, me, shaper, "{W}", []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	if opp.Life != theirs+3 || me.Life != mine {
		t.Errorf("life gained: theirs %d, mine %d; want 3 and 0", opp.Life-theirs, me.Life-mine)
	}
}

// --- Reverent Mantra: exile a white card instead ----------------------------

func TestMmqB2ReverentMantraProtectsEveryCreatureFromTheChosenColor(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	mantra := handCardFull(me, "Reverent Mantra", "Instant", "{3}{W}", mmqB2ReverentMantraOracle, []string{"W"})
	red := handCardFull(me, "Red Card", "Instant", "{R}", "", []string{"R"})
	white := handCardFull(me, "White Card", "Instant", "{W}", "", []string{"W"})
	mine := mmqB1Push(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2, "G")
	theirs := mmqB1Push(g, them.ID, "Their Bear", "Creature — Bear", "", 2, 2, "G")

	if err := mmqB1Cast(g, me.ID, mantra, "pitch", []uuid.UUID{red}, nil); err == nil {
		t.Fatal("a red card paid \"exile a white card\"")
	}
	if err := mmqB1Cast(g, me.ID, mantra, "pitch", []uuid.UUID{white}, nil); err != nil {
		t.Fatalf("Reverent Mantra for a white card: %v", err)
	}
	if me.Hand.Contains(white) || !me.Hand.Contains(red) {
		t.Fatal("the cost took the wrong card")
	}
	passPriorityAroundTable(t, g)
	answerColor(t, g, me.ID, "B")
	for _, id := range []uuid.UUID{mine, theirs} {
		if !p1g2HasKeyword(g, id, "protection from black") {
			t.Errorf("%s has no protection from black", id)
		}
	}
	if p1g2HasKeyword(g, mine, "protection from red") {
		t.Error("protection from an unchosen colour")
	}
	dftMPPassTurn(t, g)
	if p1g2HasKeyword(g, theirs, "protection from black") {
		t.Error("the protection outlasted the turn")
	}
}
