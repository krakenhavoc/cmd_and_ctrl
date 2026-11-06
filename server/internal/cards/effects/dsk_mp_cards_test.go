package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dsk_mp_cards_test.go — Duskmourn: House of Horror, cards M–P (#4466).
// Each test plays the real catalog entry and asserts what its text says.

const (
	dskMPChandelierOracle   = "437553c1-0aa4-4093-b6e5-c175ede5f278"
	dskMPMayhemOracle       = "5236a6b5-0a96-4470-ba03-37b6d7a889bb"
	dskMPSlayerOracle       = "f5bece93-afb3-4fa4-83a8-d7eaac032522"
	dskMPManorOracle        = "c8c632ab-14ec-44e1-ac00-81d48336320d"
	dskMPBoilerbilgesOracle = "c2db05e0-879b-43e4-9009-37a95fd848dc"
	dskMPFloodpitsOracle    = "31196f0b-33ee-4981-ac97-24226e8b405f"
	dskMPPlaythingOracle    = "0b7e950a-89e1-4ec4-94af-5af2f1252b12"
	dskMPBeastieOracle      = "6cf995e9-115a-4659-bace-9fbe75898b48"
	dskMPLighthouseOracle   = "d55f7e20-11c6-44e2-8a21-dca67d3dbc68"
	dskMPPeerOracle         = "8fb7ad08-9bb2-49f1-8ed0-8cc8abca4fe0"
	dskMPPiggyBankOracle    = "a13f4095-300b-434f-aa9a-3608aaae5163"
	dskMPPiranhaFlyOracle   = "58d87cfe-7402-43c3-81ce-13de8cd04005"

	dskMPOverlordTypeLine = "Enchantment Creature — Avatar Horror"
)

func TestDskMPEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		dskMPChandelierOracle:   "Malevolent Chandelier",
		dskMPMayhemOracle:       "Midnight Mayhem",
		dskMPSlayerOracle:       "Most Valuable Slayer",
		dskMPManorOracle:        "Neglected Manor",
		dskMPBoilerbilgesOracle: "Overlord of the Boilerbilges",
		dskMPFloodpitsOracle:    "Overlord of the Floodpits",
		dskMPPlaythingOracle:    "Patched Plaything",
		dskMPBeastieOracle:      "Patchwork Beastie",
		dskMPLighthouseOracle:   "Peculiar Lighthouse",
		dskMPPeerOracle:         "Peer Past the Veil",
		dskMPPiggyBankOracle:    "Piggy Bank",
		dskMPPiranhaFlyOracle:   "Piranha Fly",
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

// dskMPPickIfAsked answers the chooser's open target pick with a card,
// when the engine asked one (a single legal choice may be taken for
// them).
func dskMPPickIfAsked(t *testing.T, g *game.Game, chooser uuid.UUID, ref game.TargetRef) {
	t.Helper()
	p := latestPickTarget(g, chooser)
	if p == nil {
		return
	}
	if err := g.ResolvePickTarget(p.ID, chooser, ref); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
}

// dskMPMill puts a card of the given type line on top of p's library and
// mills it through the engine.
func dskMPMill(t *testing.T, g *game.Game, p *game.Player, name, typeLine string) {
	t.Helper()
	g.WithWriteLock(func() {
		seedLibraryTop(p, name, typeLine, "")
		if err := g.MillNForEffect(p.ID, 1); err != nil {
			t.Fatalf("MillNForEffect: %v", err)
		}
	})
}

// dskMPOverlord casts an Overlord from the active seat's hand, for its
// impending cost when impend is set.
func dskMPOverlord(t *testing.T, g *game.Game, name, oracle, cost string, power, toughness int, impend bool) uuid.UUID {
	t.Helper()
	params := game.CastSpellParams{}
	if impend {
		params.AlternativeCost = AltCostKeyImpending
	}
	return castFromHand(t, g, game.Card{
		Name: name, TypeLine: dskMPOverlordTypeLine, OracleID: oracle,
		ManaCost: cost, Power: power, Toughness: toughness,
	}, params)
}

// --- Neglected Manor, Peculiar Lighthouse ------------------------------------

func TestDskMPTaplandsEnterTappedUnlessAPlayerIsAtThirteen(t *testing.T) {
	for _, tc := range []struct{ name, oracle, produced string }{
		{"Neglected Manor", dskMPManorOracle, "{W|B}"},
		{"Peculiar Lighthouse", dskMPLighthouseOracle, "{U|R}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			land := b12PlayFromHand(t, g, tc.name, "Land", tc.oracle, game.CastSpellParams{})
			if !b13Tapped(t, g, land) {
				t.Error("everyone at 40: the land entered untapped")
			}
			if got := ltProduced(t, g, land); !slices.Equal(got, []string{tc.produced}) {
				t.Errorf("the land adds %v, want %s", got, tc.produced)
			}
			g.Seats[2].Life = 13
			if b13Tapped(t, g, b12PlayFromHand(t, g, tc.name, "Land", tc.oracle, game.CastSpellParams{})) {
				t.Error("an opponent at 13: the land entered tapped")
			}
		})
	}
}

// --- Malevolent Chandelier ------------------------------------------------------

func TestDskMPMalevolentChandelierTucksAGraveyardCardAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	chandelier := b12Push(g, me.ID, "Malevolent Chandelier", "Artifact Creature — Construct", dskMPChandelierOracle, 4, 4)
	card := gyCard(opp, "Their Card", "Sorcery", "{R}", "")
	target := []game.TargetRef{{Kind: game.TargetCard, ID: card}}

	// The table starts in the draw step: no sorcery timing.
	if err := hobB4Activate(t, g, me.ID, chandelier, 0, "{C}{C}", target); err == nil {
		t.Fatal("the ability was activated in the draw step")
	}
	advanceToMain(t, g)
	if err := hobB4Activate(t, g, me.ID, chandelier, 0, "{C}{C}", target); err != nil {
		t.Fatalf("activate in the main phase: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(card) {
		t.Fatal("the card is still in the graveyard")
	}
	bottom, err := opp.Library.Bottom()
	if err != nil || bottom.InstanceID != card {
		t.Errorf("the card is not on the bottom of its owner's library (bottom %v, err %v)", bottom.Name, err)
	}
	if !hasKeywordOf(g, chandelier, "flying") {
		t.Error("Malevolent Chandelier does not have flying")
	}
}

// --- Midnight Mayhem -----------------------------------------------------------

func TestDskMPMidnightMayhemMakesThreeGremlinsAndArmsEveryGremlinYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	myGremlin := b12Creature(g, me.ID, "My Gremlin", "Creature — Gremlin", 2, 2)
	myBear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirGremlin := b12Creature(g, opp.ID, "Their Gremlin", "Creature — Gremlin", 2, 2)

	castCatalogSpell(t, g, "Midnight Mayhem", "Sorcery", dskMPMayhemOracle, nil)
	passPriorityAroundTable(t, g)

	if n := tokensNamedControlledBy(g, "Gremlin", me.ID); n != 3 {
		t.Fatalf("%d Gremlin tokens, want 3", n)
	}
	var tokens []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == "Gremlin" {
			tokens = append(tokens, c.InstanceID)
		}
	}
	for _, id := range append(tokens, myGremlin) {
		for _, kw := range []string{"menace", "lifelink", "haste"} {
			if !hasKeywordOf(g, id, kw) {
				t.Errorf("a Gremlin you control lacks %s", kw)
			}
		}
	}
	if hasKeywordOf(g, myBear, "menace") {
		t.Error("a non-Gremlin gained menace")
	}
	if hasKeywordOf(g, theirGremlin, "menace") {
		t.Error("an opponent's Gremlin gained menace")
	}
}

// --- Most Valuable Slayer --------------------------------------------------------

func TestDskMPMostValuableSlayerPumpsAnAttackerWhenYouAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	slayer := b12Push(g, me.ID, "Most Valuable Slayer", "Creature — Human Warrior", dskMPSlayerOracle, 2, 4)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)

	// The Slayer stays home: "whenever you attack" is any attack of yours.
	declareAttack(t, g, opp.ID, bear)
	dskMPPickIfAsked(t, g, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear})
	passPriorityAroundTable(t, g)

	if p, _ := effectivePTOf(t, g, bear); p != 3 {
		t.Errorf("the attacking Bear has power %d, want 3", p)
	}
	if !hasKeywordOf(g, bear, "first strike") {
		t.Error("the attacking Bear did not gain first strike")
	}
	if p, _ := effectivePTOf(t, g, slayer); p != 2 {
		t.Errorf("the Slayer, not attacking, has power %d, want 2", p)
	}
}

// --- Overlord of the Boilerbilges -------------------------------------------------

func TestDskMPOverlordOfTheBoilerbilgesDealsFourOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dskMPOverlord(t, g, "Overlord of the Boilerbilges", dskMPBoilerbilgesOracle, "{4}{R}{R}", 5, 5, false)
	passPriorityAroundTable(t, g)
	dskMPPickIfAsked(t, g, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	passPriorityAroundTable(t, g)
	if opp.Life != 36 {
		t.Errorf("opponent life %d, want 36", opp.Life)
	}
}

func TestDskMPOverlordOfTheBoilerbilgesImpendedStillBurns(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 4, 4)
	overlord := dskMPOverlord(t, g, "Overlord of the Boilerbilges", dskMPBoilerbilgesOracle, "{4}{R}{R}", 5, 5, true)
	passPriorityAroundTable(t, g)
	dskMPPickIfAsked(t, g, me.ID, game.TargetRef{Kind: game.TargetCard, ID: ogre})
	passPriorityAroundTable(t, g)

	if n := countersOn(g, overlord, game.CounterTime); n != 4 {
		t.Errorf("%d time counters, want 4", n)
	}
	if hasString(effectiveTypes(t, g, overlord), "Creature") {
		t.Error("the impending Overlord is a creature")
	}
	if g.Battlefield.Contains(ogre) {
		t.Error("the 4/4 survived 4 damage")
	}
}

// --- Overlord of the Floodpits ----------------------------------------------------

func TestDskMPOverlordOfTheFloodpitsLootsOnEntryAndFlies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	overlord := dskMPOverlord(t, g, "Overlord of the Floodpits", dskMPFloodpitsOracle, "{3}{U}{U}", 5, 3, false)
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if got := discardOwed(g, me.ID); got != 1 {
		t.Fatalf("discard owed %d, want 1 after the draws", got)
	}
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d before the discard, want %d", got, hand+2)
	}
	discardFromHand(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d after the loot, want %d", got, hand+1)
	}
	if !hasKeywordOf(g, overlord, "flying") {
		t.Error("Overlord of the Floodpits does not have flying")
	}
}

func TestDskMPOverlordOfTheFloodpitsImpendsForFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	overlord := dskMPOverlord(t, g, "Overlord of the Floodpits", dskMPFloodpitsOracle, "{3}{U}{U}", 5, 3, true)
	passPriorityAroundTable(t, g)
	if got := discardOwed(g, me.ID); got != 1 {
		t.Fatalf("the impended Overlord's entry did not loot (discard owed %d)", got)
	}
	discardFromHand(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, overlord, game.CounterTime); n != 4 {
		t.Errorf("%d time counters, want 4", n)
	}
	if hasString(effectiveTypes(t, g, overlord), "Creature") {
		t.Error("the impending Overlord is a creature")
	}
}

// --- Patched Plaything ---------------------------------------------------------------

func TestDskMPPatchedPlaythingTakesTwoCountersOnlyWhenCastFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cast := castFromHand(t, g, game.Card{
		Name: "Patched Plaything", TypeLine: "Artifact Creature — Toy", OracleID: dskMPPlaythingOracle,
		ManaCost: "{2}{W}", Power: 4, Toughness: 3,
	}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if n := countersOn(g, cast, game.CounterMinusOne); n != 2 {
		t.Errorf("cast from hand: %d -1/-1 counters, want 2", n)
	}
	if c := b12Card(t, g, cast); c.CurrentPower() != 2 || c.CurrentToughness() != 1 {
		p, tough := c.CurrentPower(), c.CurrentToughness()
		t.Errorf("cast from hand: %d/%d, want 2/1", p, tough)
	}
	if !hasKeywordOf(g, cast, "double strike") {
		t.Error("Patched Plaything does not have double strike")
	}

	// Put onto the battlefield from a graveyard, it was not cast.
	returned := gyCard(me, "Patched Plaything", "Artifact Creature — Toy", "{2}{W}", dskMPPlaythingOracle)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID},
		game.ZoneRef{Kind: game.ZoneBattlefield}, returned); err != nil {
		t.Fatalf("MoveCardByID to the battlefield: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := countersOn(g, returned, game.CounterMinusOne); n != 0 {
		t.Errorf("not cast: %d -1/-1 counters, want none", n)
	}
}

// --- Patchwork Beastie ---------------------------------------------------------------

func TestDskMPPatchworkBeastieAttacksOnlyWithDelirium(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	beastie := b12Push(g, me.ID, "Patchwork Beastie", "Artifact Creature — Beast", dskMPBeastieOracle, 3, 3)
	dskMPMill(t, g, me, "Old Relic", "Artifact")
	dskMPMill(t, g, me, "Old Bear", "Creature — Bear")
	dskMPMill(t, g, me, "Old Bolt", "Instant")

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(beastie, opp.ID); err == nil {
		t.Fatal("three card types in the graveyard: the Beastie was allowed to attack")
	}
	dskMPMill(t, g, me, "Old Forest", "Land")
	if err := g.DeclareAttacker(beastie, opp.ID); err != nil {
		t.Fatalf("four card types in the graveyard: the Beastie could not attack: %v", err)
	}
}

func TestDskMPPatchworkBeastieMayMillAtUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	b12Push(g, owner.ID, "Patchwork Beastie", "Artifact Creature — Beast", dskMPBeastieOracle, 3, 3)
	grave := owner.Graveyard.Size()
	advanceToUpkeepOf(t, g, 1)
	answerLatestTriggerPrompt(t, g, owner.ID, true)
	passPriorityAroundTable(t, g)
	if got := owner.Graveyard.Size(); got != grave+1 {
		t.Errorf("graveyard %d after the upkeep, want %d", got, grave+1)
	}
}

// --- Peer Past the Veil ---------------------------------------------------------------

func TestDskMPPeerPastTheVeilDrawsOnePerCardTypeAfterDiscarding(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Hand.Cards = nil
	pushCatalogHandCard(me, "Spare Bolt", "Instant", "")
	pushCatalogHandCard(me, "Spare Forest", "Land", "")
	gyCard(me, "Old Bear", "Creature — Bear", "{1}{G}", "")

	// An instant: cast in the draw step. The hand goes, then X counts the
	// creature already there and the instant and land just discarded.
	if err := dskB1CastAtDrawStep(t, g, "Peer Past the Veil", "Instant", dskMPPeerOracle); err != nil {
		t.Fatalf("cast at instant speed: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != 3 {
		t.Errorf("hand %d, want 3 (creature, instant, land)", got)
	}
	for _, c := range me.Hand.Cards {
		if c.Name == "Spare Bolt" || c.Name == "Spare Forest" {
			t.Errorf("%s was not discarded", c.Name)
		}
	}
}

// --- Piggy Bank --------------------------------------------------------------------------

func TestDskMPPiggyBankLeavesATreasure(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pig := b12Push(g, me.ID, "Piggy Bank", "Artifact Creature — Boar Toy", dskMPPiggyBankOracle, 3, 2)
	killCreature(t, g, me.ID, pig)
	passPriorityAroundTable(t, g)
	if n := tokensNamedControlledBy(g, "Treasure", me.ID); n != 1 {
		t.Errorf("%d Treasure tokens, want 1", n)
	}
}

// --- Piranha Fly --------------------------------------------------------------------------

func TestDskMPPiranhaFlyEntersTappedAndFlies(t *testing.T) {
	g := newCatalogGame(t)
	fly := castFromHand(t, g, game.Card{
		Name: "Piranha Fly", TypeLine: "Creature — Fish Insect", OracleID: dskMPPiranhaFlyOracle,
		ManaCost: "{1}{U}", Power: 2, Toughness: 1,
	}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if !b13Tapped(t, g, fly) {
		t.Error("Piranha Fly entered untapped")
	}
	if !hasKeywordOf(g, fly, "flying") {
		t.Error("Piranha Fly does not have flying")
	}
}
