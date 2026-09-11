package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mana_pipeline_test.go — the S32 mana-pipeline batch (#352): the ten
// Signets, Cabal Coffers, Gaea's Cradle, Temple of the False God, Mox
// Opal, Exotic Orchard, Reflecting Pool, Mox Amber, Delighted
// Halfling, Shrine of the Forsaken Gods, Eldrazi Temple, and the
// Fellwar Stone retrofit.
//
// Four seams, and each card here is chosen to prove one of them
// against the printed text rather than against the implementation:
//
//   - a mana component in a mana ability's cost (Signets, Coffers);
//   - an activation gate (Temple, Mox Opal, Shrine);
//   - a produced string computed at activation, derived (Orchard,
//     Pool, Fellwar, Amber) or scaled (Coffers, Cradle);
//   - spend restrictions that actually bind at SPEND time (Halfling,
//     Shrine, Eldrazi Temple).
//
// The last group carries the weight. A restriction stamped but not
// enforced makes every one of those cards stronger than printed,
// which is the #259 rule, so each has a test that pays for something
// it should NOT be able to pay for.

const (
	azoriusSignetOracle    = "e018773f-95b3-49a3-9674-6f04ddef2092"
	dimirSignetOracle      = "7d881c57-0bd9-4c57-aa4a-b10808b86143"
	selesnyaSignetOracle   = "1436dd81-496e-42a5-b210-fb5b9cdf073f"
	cabalCoffersOracle     = "7358e164-5704-4e78-9b21-6a9bf2a968ce"
	gaeasCradleOracle      = "7c427c3d-ecd8-45ef-bebd-8f10f4a311db"
	templeFalseGodOracle   = "cfdd5dc6-593e-495a-8cfe-3a56b3c4c7df"
	moxOpalOracle          = "de2440de-e948-4811-903c-0bbe376ff64d"
	moxAmberOracle         = "7a43bd27-fdd8-41f0-9bc4-92568f3408f1"
	exoticOrchardOracle    = "27b047e3-0d41-45e2-98e9-9391d7923a1e"
	reflectingPoolOracle   = "67f43ac6-2a58-4b53-b5d7-0330e2a252e2"
	delightedHalflingOracl = "f9d3b046-0b95-4103-a630-4b3fb88bb60b"
	shrineForsakenOracle   = "8ea46945-d5ab-4209-b473-4769e7b8b962"
	eldraziTempleOracle    = "7fab8d65-af51-47d3-8f10-2676bf6e8ba3"
)

// seedLand puts a land on the battlefield with an explicit
// ProducedMana list, which is how the deck importer stamps Scryfall's
// `produced_mana` and therefore what the derivation reads for lands
// outside the catalog.
func seedLand(g *game.Game, owner uuid.UUID, name, typeLine string, produces ...string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID:   id,
		Name:         name,
		TypeLine:     typeLine,
		ProducedMana: produces,
		Owner:        owner,
		Controller:   owner,
	})
	return id
}

// manaPickFor returns the latest queued colour pick for a chooser.
func manaPickFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	var out *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceMana && c.Chooser == chooser {
			out = c
		}
	}
	return out
}

// --- the Signet cycle --------------------------------------------

func TestAzoriusSignetFiltersOneManaIntoTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	signet := seedPermanentWithOracle(g, me.ID, "Azorius Signet", "Artifact", azoriusSignetOracle)
	me.ManaPool.AddMana(game.ManaToken{Color: "G"})

	if err := g.ActivateManaAbility(me.ID, signet, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if len(me.ManaPool) != 2 {
		t.Fatalf("pool = %v, want {W}{U}", me.ManaPool)
	}
	if me.ManaPool[0].Color != "W" || me.ManaPool[1].Color != "U" {
		t.Errorf("pool = %v, want {W} then {U}", me.ManaPool)
	}
}

// The {1} is a cost, so an empty pool is a refusal and not a free
// activation. Without this the Signet would be a Sol Ring.
func TestDimirSignetNeedsTheOneManaFirst(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	signet := seedPermanentWithOracle(g, me.ID, "Dimir Signet", "Artifact", dimirSignetOracle)

	if err := g.ActivateManaAbility(me.ID, signet, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("a Signet activated on an empty pool")
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == signet && c.Tapped {
			t.Error("the rejected activation tapped the Signet")
		}
	}
}

// A Signet's commander-identity narrowing must not kick in: the
// produced string is two named colours, not a pipe. A mono-white
// commander must not shrink a Selesnya Signet's {G}.
func TestSelesnyaSignetIgnoresCommanderIdentity(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	riderGiveCommander(g, me, "{W}")
	signet := seedPermanentWithOracle(g, me.ID, "Selesnya Signet", "Artifact", selesnyaSignetOracle)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})

	if err := g.ActivateManaAbility(me.ID, signet, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if len(me.ManaPool) != 2 {
		t.Fatalf("pool = %v, want {G}{W}", me.ManaPool)
	}
	if me.ManaPool[0].Color != "G" {
		t.Errorf("pool = %v, want green first — the Signet's colours are printed, not derived", me.ManaPool)
	}
}

// --- Cabal Coffers: a mana cost AND a scaled output ---------------

func TestCabalCoffersScalesWithSwampsAndCostsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	coffers := seedPermanentWithOracle(g, me.ID, "Cabal Coffers", "Land", cabalCoffersOracle)
	for i := 0; i < 3; i++ {
		seedLand(g, me.ID, "Swamp", "Basic Land — Swamp", "B")
	}
	// An opponent's Swamps are not yours.
	seedLand(g, g.Seats[1].ID, "Swamp", "Basic Land — Swamp", "B")
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})

	if err := g.ActivateManaAbility(me.ID, coffers, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if len(me.ManaPool) != 3 {
		t.Fatalf("pool = %v, want three {B} (the {2} was spent)", me.ManaPool)
	}
	for _, tok := range me.ManaPool {
		if tok.Color != "B" {
			t.Errorf("token %v, want black", tok)
		}
	}
}

func TestCabalCoffersWithoutTheTwoManaIsRefused(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	coffers := seedPermanentWithOracle(g, me.ID, "Cabal Coffers", "Land", cabalCoffersOracle)
	seedLand(g, me.ID, "Swamp", "Basic Land — Swamp", "B")
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})

	if err := g.ActivateManaAbility(me.ID, coffers, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("Cabal Coffers activated for {1}")
	}
	if len(me.ManaPool) != 1 {
		t.Errorf("pool = %v, want the one floating mana untouched", me.ManaPool)
	}
}

// --- Gaea's Cradle: the pure scaled shape ------------------------

func TestGaeasCradleCountsCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cradle := seedPermanentWithOracle(g, me.ID, "Gaea's Cradle", "Legendary Land", gaeasCradleOracle)
	for i := 0; i < 2; i++ {
		seedPermanentWithOracle(g, me.ID, "Bear", "Creature — Bear", "")
	}

	if err := g.ActivateManaAbility(me.ID, cradle, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if len(me.ManaPool) != 2 {
		t.Errorf("pool = %v, want two {G}", me.ManaPool)
	}
}

// --- Temple of the False God and Mox Opal: the gate ---------------

func TestTempleOfTheFalseGodNeedsFiveLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	temple := seedPermanentWithOracle(g, me.ID, "Temple of the False God", "Land", templeFalseGodOracle)
	for i := 0; i < 3; i++ {
		seedLand(g, me.ID, "Wastes", "Basic Land — Wastes", "C")
	}

	// Four lands including the Temple — one short.
	if err := g.ActivateManaAbility(me.ID, temple, 0, game.ManaAbilityParams{}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("err = %v, want ErrConditionNotMet", err)
	}
	if len(me.ManaPool) != 0 {
		t.Fatalf("pool = %v, want empty — a failed gate pays nothing", me.ManaPool)
	}

	// The fifth land switches it on. The Temple counts itself.
	seedLand(g, me.ID, "Wastes", "Basic Land — Wastes", "C")
	if err := g.ActivateManaAbility(me.ID, temple, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility at five lands: %v", err)
	}
	if len(me.ManaPool) != 2 {
		t.Errorf("pool = %v, want {C}{C}", me.ManaPool)
	}
}

func TestMoxOpalNeedsMetalcraft(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opal := seedPermanentWithOracle(g, me.ID, "Mox Opal", "Legendary Artifact", moxOpalOracle)

	if err := g.ActivateManaAbility(me.ID, opal, 0, game.ManaAbilityParams{}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("err = %v, want ErrConditionNotMet — a turn-one any-colour Mox", err)
	}
	// Two more artifacts: the Opal counts itself, so that is three.
	seedPermanentWithOracle(g, me.ID, "Sol Ring", "Artifact", "")
	seedPermanentWithOracle(g, me.ID, "Ornithopter", "Artifact Creature — Thopter", "")
	if err := g.ActivateManaAbility(me.ID, opal, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility with metalcraft: %v", err)
	}
	if pick := manaPickFor(g, me.ID); pick == nil || len(pick.ColorOptions) != 5 {
		t.Errorf("pick = %v, want a full five-colour choice", pick)
	}
}

// --- Exotic Orchard, Reflecting Pool, Fellwar Stone --------------

func TestExoticOrchardDerivesOpponentColours(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	orchard := seedPermanentWithOracle(g, me.ID, "Exotic Orchard", "Land", exoticOrchardOracle)
	seedLand(g, them.ID, "Island", "Basic Land — Island", "U")
	seedLand(g, them.ID, "Mountain", "Basic Land — Mountain", "R")
	// Colorless is not a colour: an opposing Wastes adds nothing.
	seedLand(g, them.ID, "Wastes", "Basic Land — Wastes", "C")
	// Your own Forest is irrelevant — the card reads opponents only.
	seedLand(g, me.ID, "Forest", "Basic Land — Forest", "G")

	if err := g.ActivateManaAbility(me.ID, orchard, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := manaPickFor(g, me.ID)
	if pick == nil {
		t.Fatal("no colour pick queued")
	}
	want := map[string]bool{"U": true, "R": true}
	if len(pick.ColorOptions) != 2 {
		t.Fatalf("options = %v, want exactly {U,R}", pick.ColorOptions)
	}
	for _, o := range pick.ColorOptions {
		if !want[o] {
			t.Errorf("options = %v, contains %q which no opponent's land could produce", pick.ColorOptions, o)
		}
	}
}

// The printed card with no opposing lands produces nothing.
func TestExoticOrchardWithNoOpposingLandsProducesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	orchard := seedPermanentWithOracle(g, me.ID, "Exotic Orchard", "Land", exoticOrchardOracle)

	if err := g.ActivateManaAbility(me.ID, orchard, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if len(me.ManaPool) != 0 || manaPickFor(g, me.ID) != nil {
		t.Errorf("pool = %v, pick = %v; want nothing produced", me.ManaPool, manaPickFor(g, me.ID))
	}
}

// "Any TYPE", so Reflecting Pool can make {C} where Exotic Orchard
// could not. That one word is the whole difference between the cards.
func TestReflectingPoolIncludesColorlessBecauseItSaysType(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pool := seedPermanentWithOracle(g, me.ID, "Reflecting Pool", "Land", reflectingPoolOracle)
	seedLand(g, me.ID, "Wastes", "Basic Land — Wastes", "C")
	seedLand(g, me.ID, "Plains", "Basic Land — Plains", "W")

	if err := g.ActivateManaAbility(me.ID, pool, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := manaPickFor(g, me.ID)
	if pick == nil {
		t.Fatal("no colour pick queued")
	}
	sawC := false
	for _, o := range pick.ColorOptions {
		if o == "C" {
			sawC = true
		}
	}
	if !sawC {
		t.Errorf("options = %v, want {C} among them", pick.ColorOptions)
	}
}

// A lone Reflecting Pool produces nothing — the printed ruling, and
// what the recursion guard yields.
func TestLoneReflectingPoolProducesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pool := seedPermanentWithOracle(g, me.ID, "Reflecting Pool", "Land", reflectingPoolOracle)

	if err := g.ActivateManaAbility(me.ID, pool, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if len(me.ManaPool) != 0 || manaPickFor(g, me.ID) != nil {
		t.Error("a lone Reflecting Pool produced mana")
	}
}

// The retrofit. Before #352 Fellwar Stone offered all five colours
// regardless of the board, which is stronger than printed.
func TestFellwarStoneNoLongerOffersColoursNoOpponentCanMake(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	stone := seedPermanentWithOracle(g, me.ID, "Fellwar Stone", "Artifact", fellwarStoneOracle)
	seedLand(g, them.ID, "Forest", "Basic Land — Forest", "G")

	if err := g.ActivateManaAbility(me.ID, stone, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if len(me.ManaPool) != 1 || me.ManaPool[0].Color != "G" {
		t.Fatalf("pool = %v, pick = %v; a single-colour derivation should drop straight in as {G}",
			me.ManaPool, manaPickFor(g, me.ID))
	}
}

// --- Mox Amber ----------------------------------------------------

func TestMoxAmberReadsLegendaryPermanentColours(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	amber := seedPermanentWithOracle(g, me.ID, "Mox Amber", "Legendary Artifact", moxAmberOracle)

	// Nothing legendary but the Mox itself (an artifact, not a
	// creature or planeswalker) → no mana.
	if err := g.ActivateManaAbility(me.ID, amber, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if len(me.ManaPool) != 0 || manaPickFor(g, me.ID) != nil {
		t.Fatal("Mox Amber produced mana with no legendary creatures out")
	}

	// A legendary green creature switches it on.
	legend := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: legend,
		Name:       "Yeva, Nature's Herald",
		TypeLine:   "Legendary Creature — Elf Shaman",
		ManaCost:   "{2}{G}{G}",
		Owner:      me.ID,
		Controller: me.ID,
	})
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == amber {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	if err := g.ActivateManaAbility(me.ID, amber, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility with a legend out: %v", err)
	}
	if len(me.ManaPool) != 1 || me.ManaPool[0].Color != "G" {
		t.Errorf("pool = %v, want one {G}", me.ManaPool)
	}
}

// --- restricted mana: the cards, end to end ----------------------

// Delighted Halfling's coloured half is legendary-spells-only. The
// second half of this test is the one that matters: the mana must NOT
// pay for a non-legendary spell.
func TestDelightedHalflingManaIsLockedToLegendarySpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	halfling := seedPermanentWithOracle(g, me.ID, "Delighted Halfling",
		"Creature — Halfling Citizen", delightedHalflingOracl)
	// Summoning sickness: give it haste so ability 1 is activatable.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == halfling {
				g.Battlefield.Cards[i].Keywords = []string{"haste"}
			}
		}
	})

	if err := g.ActivateManaAbility(me.ID, halfling, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := manaPickFor(g, me.ID)
	if pick == nil {
		t.Fatal("no colour pick queued")
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "G"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if len(me.ManaPool) != 1 {
		t.Fatalf("pool = %v, want one token", me.ManaPool)
	}

	cost, err := game.ParseCost("{G}")
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}
	ordinary := game.Card{Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid", ManaCost: "{G}"}
	legendary := game.Card{Name: "Yeva, Nature's Herald", TypeLine: "Legendary Creature — Elf Shaman", ManaCost: "{G}"}

	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(ordinary)) {
		t.Error("Halfling mana paid for a non-legendary creature — stronger than printed")
	}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(legendary)) {
		t.Error("Halfling mana refused a legendary creature")
	}
}

// The Halfling's colourless half has no restriction and no gate, so
// it pays for anything.
func TestDelightedHalflingColorlessHalfIsUnrestricted(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	halfling := seedPermanentWithOracle(g, me.ID, "Delighted Halfling",
		"Creature — Halfling Citizen", delightedHalflingOracl)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == halfling {
				g.Battlefield.Cards[i].Keywords = []string{"haste"}
			}
		}
	})

	if err := g.ActivateManaAbility(me.ID, halfling, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if len(me.ManaPool) != 1 || len(me.ManaPool[0].Restrictions) != 0 {
		t.Errorf("pool = %v, want one unrestricted {C}", me.ManaPool)
	}
}

// Eldrazi Temple: colorless AND Eldrazi, and the mana works for a
// cast or an activation alike.
func TestEldraziTempleManaIsLockedToColorlessEldrazi(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	temple := seedPermanentWithOracle(g, me.ID, "Eldrazi Temple", "Land", eldraziTempleOracle)

	if err := g.ActivateManaAbility(me.ID, temple, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if len(me.ManaPool) != 2 {
		t.Fatalf("pool = %v, want {C}{C}", me.ManaPool)
	}

	cost, _ := game.ParseCost("{C}{C}")
	colorlessEldrazi := game.Card{Name: "Ulamog", TypeLine: "Legendary Creature — Eldrazi", ManaCost: "{10}"}
	coloredEldrazi := game.Card{Name: "Void Winnower", TypeLine: "Creature — Eldrazi", ManaCost: "{7}{B}"}
	colorlessGolem := game.Card{Name: "Karn's Sylex", TypeLine: "Legendary Artifact", ManaCost: "{1}"}

	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(colorlessEldrazi)) {
		t.Error("Eldrazi Temple mana refused a colorless Eldrazi")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(coloredEldrazi)) {
		t.Error("a black Eldrazi is not a colorless Eldrazi")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(colorlessGolem)) {
		t.Error("a colorless artifact is not an Eldrazi — stronger than printed")
	}
	// "or activate abilities of colorless Eldrazi": no purpose tag,
	// so an activation is fine too.
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(colorlessEldrazi)) {
		t.Error("Eldrazi Temple mana refused a colorless Eldrazi's ability")
	}
}

// Shrine of the Forsaken Gods is the two-drawback card: a gate and a
// restriction on the same ability.
func TestShrineOfTheForsakenGodsGatesAndRestricts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	shrine := seedPermanentWithOracle(g, me.ID, "Shrine of the Forsaken Gods", "Land", shrineForsakenOracle)

	// The gate first: six lands is not seven.
	for i := 0; i < 5; i++ {
		seedLand(g, me.ID, "Wastes", "Basic Land — Wastes", "C")
	}
	if err := g.ActivateManaAbility(me.ID, shrine, 1, game.ManaAbilityParams{}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("err = %v, want ErrConditionNotMet at six lands", err)
	}

	seedLand(g, me.ID, "Wastes", "Basic Land — Wastes", "C")
	if err := g.ActivateManaAbility(me.ID, shrine, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility at seven lands: %v", err)
	}
	if len(me.ManaPool) != 2 {
		t.Fatalf("pool = %v, want {C}{C}", me.ManaPool)
	}

	cost, _ := game.ParseCost("{C}{C}")
	colorless := game.Card{Name: "Wurmcoil Engine", TypeLine: "Artifact Creature — Phyrexian Wurm", ManaCost: "{6}"}
	colored := game.Card{Name: "Shivan Dragon", TypeLine: "Creature — Dragon", ManaCost: "{4}{R}{R}"}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(colorless)) {
		t.Error("Shrine mana refused a colorless spell")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(colored)) {
		t.Error("Shrine mana paid for a coloured spell — stronger than printed")
	}
	// "colorless SPELLS" — casting only, not activating.
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(colorless)) {
		t.Error("Shrine mana paid an activation cost; the card says 'cast'")
	}
}

// --- cross-cutting: the whole cycle registers ---------------------

func TestAllTenSignetsAreRegistered(t *testing.T) {
	for _, oracle := range []string{
		"7d881c57-0bd9-4c57-aa4a-b10808b86143", // Dimir
		"3adb7681-977f-4a32-9ec8-51481b958268", // Rakdos
		"2fda4fe7-8b0c-489c-a000-6d358e614e34", // Izzet
		"de3dcb5d-775a-479f-99f5-d1883ed9b1b5", // Orzhov
		"e018773f-95b3-49a3-9674-6f04ddef2092", // Azorius
		"41c84665-1f99-40ab-aaca-1188649eb263", // Boros
		"44503105-3e13-408d-a44f-37d503c61d72", // Simic
		"1cf51f50-24e4-48d0-95b3-1dad3ffa4bf5", // Golgari
		"d36e0c9f-c025-4dfe-9644-9cad2461ce38", // Gruul
		"1436dd81-496e-42a5-b210-fb5b9cdf073f", // Selesnya
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("oracle %s is not registered", oracle)
			continue
		}
		if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Cost.Mana != "{1}" {
			t.Errorf("%s: mana abilities = %+v, want one costing {1}", spec.Name, spec.ManaAbilities)
		}
	}
}
