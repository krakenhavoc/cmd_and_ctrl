package effects

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b8_cards_test.go — Star Trek, batch 8 (#4973). Each test plays the
// real catalog entry through the engine: beam me up printed (Open
// Communications) and granted (Scotty, Miracle Worker), Marooned's "can't
// be beamed up", and the rest of the batch.

const (
	trekB8MakeItSoOracle  = "a9368972-4697-4a10-97e3-324149a26fba"
	trekB8MaroonedOracle  = "74e02bdf-a074-4356-9d06-0df01dd2df66"
	trekB8SpockOracle     = "869bde03-8813-43a9-9df5-a21c62511b4b"
	trekB8UhuraOracle     = "d89f34e4-8628-4faf-89e6-c9c16023ae8b"
	trekB8OpenCommsOracle = "b08c5b86-4f84-46a3-877f-bbf71ec17adb"
	trekB8BirdOracle      = "e614f97b-da8e-4569-9bdd-f76371c66836"
	trekB8RulesOracle     = "5068eba6-ccf0-45f2-b0db-40f787ee3308"
	trekB8ScottyOracle    = "8e5057ae-165c-4fe4-b712-a638b7eea038"
	trekB8SevenOracle     = "af9c9b8a-0713-4904-8e86-2b3331ec5495"
)

func TestTrekB8EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB8MakeItSoOracle:  "Make It So",
		trekB8MaroonedOracle:  "Marooned",
		trekB8SpockOracle:     "Mr. Spock, Science Officer",
		trekB8UhuraOracle:     "Nyota Uhura, Communications",
		trekB8OpenCommsOracle: "Open Communications",
		trekB8BirdOracle:      "Romulan Bird-of-Prey",
		trekB8RulesOracle:     "Rules of Acquisition",
		trekB8ScottyOracle:    "Scotty, Miracle Worker",
		trekB8SevenOracle:     "Seven of Nine, Unassimilated",
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

// trekB8Main walks the fresh table to seat 0's precombat main phase.
func trekB8Main(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := newCatalogGame(t)
	advanceTo(t, g, game.StepPrecombatMain)
	return g, g.Seats[0], g.Seats[1]
}

// trekB8Float floats a cost's mana into p's pool, each generic {N} as N
// colourless.
func trekB8Float(t *testing.T, g *game.Game, p *game.Player, cost string) {
	t.Helper()
	mana := regexp.MustCompile(`\{(\d+)\}`).ReplaceAllStringFunc(cost, func(m string) string {
		n, _ := strconv.Atoi(strings.Trim(m, "{}"))
		return strings.Repeat("{C}", n)
	})
	floatMana(t, g, p, mana)
}

func trekB8Card(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }

// --- beam me up: Open Communications ------------------------------------

func TestTrekB8OpenCommunicationsDrawsFromHand(t *testing.T) {
	g, me, _ := trekB8Main(t)
	seedLibrary(me, "One", "Two")
	hand := me.Hand.Size()
	trekB8Float(t, g, me, "{U}")
	castFromHandForTest(t, g, me, "Open Communications", "Sorcery", "{U}", trekB8OpenCommsOracle, game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
}

func trekB8OpenCommsInGraveyard(g *game.Game, me *game.Player) uuid.UUID {
	return pushToGraveyard(g, me, game.Card{Name: "Open Communications", TypeLine: "Sorcery", ManaCost: "{U}", OracleID: trekB8OpenCommsOracle})
}

func TestTrekB8OpenCommunicationsBeamsUpFromTheGraveyard(t *testing.T) {
	g, me, _ := trekB8Main(t)
	seedLibrary(me, "One", "Two")
	bear := b12Creature(g, me.ID, "Away Team", "Creature — Human", 2, 2)
	card := trekB8OpenCommsInGraveyard(g, me)
	trekB8Float(t, g, me, "{2}{U}")
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: game.AltCostKeyBeamMeUp, AltCostIDs: []uuid.UUID{bear}, Strict: true,
	}); err != nil {
		t.Fatalf("beam me up cast: %v", err)
	}
	if !me.Hand.Contains(bear) {
		t.Fatal("the creature was not returned to hand as the cost")
	}
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d: the spell draws a card", me.Hand.Size(), hand+1)
	}
	if me.Graveyard.Contains(card) || !g.Exile.Contains(card) {
		t.Error("a beamed-up spell must be exiled as it leaves the stack")
	}
}

func TestTrekB8BeamMeUpNeedsACreatureYouControl(t *testing.T) {
	g, me, opp := trekB8Main(t)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	card := trekB8OpenCommsInGraveyard(g, me)
	trekB8Float(t, g, me, "{2}{U}")
	for name, ids := range map[string][]uuid.UUID{"no creature": nil, "an opponent's creature": {theirs}} {
		if err := g.CastSpell(me.ID, card, game.CastSpellParams{
			FromZone: "graveyard", AlternativeCost: game.AltCostKeyBeamMeUp, AltCostIDs: ids, Strict: true,
		}); err == nil {
			t.Fatalf("beam me up was cast returning %s", name)
		}
	}
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err == nil {
		t.Fatal("the card was cast from the graveyard for its mana cost")
	}
}

// --- Marooned -------------------------------------------------------------

func trekB8CastMarooned(t *testing.T, g *game.Game, me *game.Player, host uuid.UUID, pay bool) {
	t.Helper()
	trekB8Float(t, g, me, "{U}{U}{U}{U}")
	castFromHandForTest(t, g, me, "Marooned", "Enchantment — Aura", "{U}", trekB8MaroonedOracle,
		game.CastSpellParams{Targets: []game.TargetRef{trekB8Card(host)}, Strict: true})
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, pay)
	passPriorityAroundTable(t, g)
}

func TestTrekB8MaroonedPaidTapsTheEnchantedPermanent(t *testing.T) {
	g, me, opp := trekB8Main(t)
	ship := b12Permanent(g, opp.ID, "Their Shuttle", "Artifact")
	trekB8CastMarooned(t, g, me, ship, true)
	if !tmtB1Card(t, g, ship).Tapped {
		t.Error("paid {3}: the enchanted artifact was not tapped")
	}
}

func TestTrekB8MaroonedDeclinedLeavesItUntapped(t *testing.T) {
	g, me, opp := trekB8Main(t)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	trekB8CastMarooned(t, g, me, bear, false)
	if tmtB1Card(t, g, bear).Tapped {
		t.Error("declined: the creature was tapped anyway")
	}
}

func TestTrekB8MaroonedRefusesALand(t *testing.T) {
	g, me, _ := trekB8Main(t)
	land := b12Permanent(g, me.ID, "Island", "Basic Land — Island")
	trekB8Float(t, g, me, "{U}")
	if _, err := castFromHandErr(g, me, "Marooned", "Enchantment — Aura", "{U}", trekB8MaroonedOracle,
		game.CastSpellParams{Targets: []game.TargetRef{trekB8Card(land)}, Strict: true}); err == nil {
		t.Fatal("Marooned enchanted a land")
	}
}

// The enchanted creature can't pay a beam me up cost; another can.
func TestTrekB8MaroonedCreatureCantBeBeamedUp(t *testing.T) {
	g, me, _ := trekB8Main(t)
	stuck := b12Creature(g, me.ID, "Stranded Ensign", "Creature — Human", 1, 1)
	free := b12Creature(g, me.ID, "Away Team", "Creature — Human", 2, 2)
	trekB8CastMarooned(t, g, me, stuck, false)
	var marked bool
	g.ReadSnapshot(func() { marked = g.CantBeBeamedUpForEffect(stuck) })
	if !marked {
		t.Fatal("the enchanted creature is not marked as unable to be beamed up")
	}
	card := trekB8OpenCommsInGraveyard(g, me)
	trekB8Float(t, g, me, "{2}{U}")
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: game.AltCostKeyBeamMeUp, AltCostIDs: []uuid.UUID{stuck}, Strict: true,
	}); err == nil {
		t.Fatal("the Marooned creature was beamed up")
	}
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: game.AltCostKeyBeamMeUp, AltCostIDs: []uuid.UUID{free}, Strict: true,
	}); err != nil {
		t.Fatalf("another creature could not be beamed up: %v", err)
	}
}

// --- Mr. Spock, Science Officer ---------------------------------------------

func TestTrekB8SpockDrawsTwoInstead(t *testing.T) {
	g, me, _ := trekB8Main(t)
	b27Push(g, me.ID, "Mr. Spock, Science Officer", "Legendary Creature — Vulcan Human Scientist", trekB8SpockOracle, "{2}{U}{U}", 2, 4, "U")
	seedLibrary(me, "One", "Two", "Three", "Four")
	trekB8Float(t, g, me, "{U}")
	castFromHandForTest(t, g, me, "Open Communications", "Sorcery", "{U}", trekB8OpenCommsOracle, game.CastSpellParams{Strict: true})
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d, want %d: one draw became two", got, hand+2)
	}
}

// --- Nyota Uhura, Communications ------------------------------------------

func TestTrekB8UhuraChoosesANumberAsSheEnters(t *testing.T) {
	g, me, _ := trekB8Main(t)
	trekB8Float(t, g, me, "{2}{U}")
	uhura := castFromHandForTest(t, g, me, "Nyota Uhura, Communications", "Legendary Creature — Human Officer", "{2}{U}", trekB8UhuraOracle, game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	prompt := answerNumber(t, g, me.ID, 4)
	if prompt.Min != 0 || prompt.Max != 10 {
		t.Errorf("number prompt %d..%d, want 0..10", prompt.Min, prompt.Max)
	}
	if n := g.ChosenNumberOf(uhura); n != 4 {
		t.Errorf("chosen number %d, want 4", n)
	}
}

// Uhura is seat 1's, with 1 chosen: my one-drop draws her controller a
// card, my two-drop does not, and her controller's own spell does not.
func TestTrekB8UhuraDrawsOnAnOpponentsSpellOfTheChosenManaValue(t *testing.T) {
	g, me, opp := trekB8Main(t)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Nyota Uhura, Communications", TypeLine: "Legendary Creature — Human Officer",
		OracleID: trekB8UhuraOracle, ManaCost: "{2}{U}", Power: 2, Toughness: 4, Owner: opp.ID, Controller: opp.ID, ChosenNumber: 1,
	})
	seedLibrary(opp, "One", "Two", "Three")
	hand := opp.Hand.Size()
	trekB8Float(t, g, me, "{1}{R}")
	castFromHandForTest(t, g, me, "Two-Drop", "Sorcery", "{1}{R}", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if opp.Hand.Size() != hand {
		t.Fatalf("a mana value 2 spell drew a card (hand %d, want %d)", opp.Hand.Size(), hand)
	}
	trekB8Float(t, g, me, "{R}")
	castFromHandForTest(t, g, me, "One-Drop", "Sorcery", "{R}", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if opp.Hand.Size() != hand+1 {
		t.Errorf("a mana value 1 spell: hand %d, want %d", opp.Hand.Size(), hand+1)
	}
}

// --- Romulan Bird-of-Prey ---------------------------------------------------

func TestTrekB8BirdOfPreyCloakedChargesOnNoncreatureSpells(t *testing.T) {
	g, me, _ := trekB8Main(t)
	bird := b27Push(g, me.ID, "Romulan Bird-of-Prey", "Artifact — Spacecraft", trekB8BirdOracle, "{1}{U}", 4, 4, "U")
	if ab := effectiveAbilities(t, g, bird); !containsString(ab, "hexproof") {
		t.Errorf("abilities %v, want hexproof while it isn't a creature", ab)
	}
	trekB8Float(t, g, me, "{1}{G}")
	castFromHandForTest(t, g, me, "Grizzly", "Creature — Bear", "{1}{G}", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if n := tmtB1Card(t, g, bird).Counters[game.CounterCharge]; n != 0 {
		t.Fatalf("a creature spell put %d charge counters on it", n)
	}
	trekB8Float(t, g, me, "{R}")
	castFromHandForTest(t, g, me, "Spark", "Sorcery", "{R}", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if n := tmtB1Card(t, g, bird).Counters[game.CounterCharge]; n != 2 {
		t.Errorf("%d charge counters after a noncreature spell, want 2", n)
	}
}

// At eight charge counters it is a 4/4 flying creature: no hexproof, and
// the intervening if stops the cast trigger.
func TestTrekB8BirdOfPreyDecloaksAtEight(t *testing.T) {
	g, me, _ := trekB8Main(t)
	bird := b27Push(g, me.ID, "Romulan Bird-of-Prey", "Artifact — Spacecraft", trekB8BirdOracle, "{1}{U}", 4, 4, "U")
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bird, game.CounterCharge, 8) })
	c := tmtB1Card(t, g, bird)
	if !c.IsCreature() || c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Fatalf("at 8 charge counters: creature %v %d/%d, want a 4/4 creature", c.IsCreature(), c.CurrentPower(), c.CurrentToughness())
	}
	ab := effectiveAbilities(t, g, bird)
	if !containsString(ab, "flying") || containsString(ab, "hexproof") {
		t.Errorf("abilities %v, want flying and no hexproof", ab)
	}
	trekB8Float(t, g, me, "{R}")
	castFromHandForTest(t, g, me, "Spark", "Sorcery", "{R}", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if n := tmtB1Card(t, g, bird).Counters[game.CounterCharge]; n != 8 {
		t.Errorf("%d charge counters, want 8: it is a creature, so the trigger does nothing", n)
	}
}

// --- Make It So ---------------------------------------------------------------

func TestTrekB8MakeItSoUntapsAndTargetPlayerDraws(t *testing.T) {
	g, me, opp := trekB8Main(t)
	mine := b12Permanent(g, me.ID, "My Relic", "Artifact")
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(mine) })
	seedLibrary(opp, "One", "Two", "Three")
	hand := opp.Hand.Size()
	trekB8Float(t, g, me, "{3}{U}")
	castFromHandForTest(t, g, me, "Make It So", "Sorcery", "{3}{U}", trekB8MakeItSoOracle, game.CastSpellParams{
		Modes: []int{0}, Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine, Slot: 0}, {Kind: game.TargetPlayer, ID: opp.ID, Slot: 1}},
	})
	passPriorityAroundTable(t, g)
	if tmtB1Card(t, g, mine).Tapped {
		t.Error("the permanent was not untapped")
	}
	if opp.Hand.Size() != hand+2 {
		t.Errorf("target player's hand %d, want %d", opp.Hand.Size(), hand+2)
	}
}

func TestTrekB8MakeItSoRefusesAPermanentYouDontControl(t *testing.T) {
	g, me, opp := trekB8Main(t)
	theirs := b12Permanent(g, opp.ID, "Their Relic", "Artifact")
	trekB8Float(t, g, me, "{3}{U}")
	if _, err := castFromHandErr(g, me, "Make It So", "Sorcery", "{3}{U}", trekB8MakeItSoOracle, game.CastSpellParams{
		Modes: []int{0}, Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs, Slot: 0}, {Kind: game.TargetPlayer, ID: me.ID, Slot: 1}},
	}); err == nil {
		t.Fatal("Make It So targeted a permanent an opponent controls")
	}
}

func TestTrekB8MakeItSoMakesAnOfficerAndDraws(t *testing.T) {
	g, me, _ := trekB8Main(t)
	seedLibrary(me, "One", "Two", "Three")
	trekB8Float(t, g, me, "{3}{U}")
	castFromHandForTest(t, g, me, "Make It So", "Sorcery", "{3}{U}", trekB8MakeItSoOracle, game.CastSpellParams{Modes: []int{1}, Strict: true})
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+2)
	}
	officers := tmtB1Named(g, "Officer")
	if len(officers) != 1 {
		t.Fatalf("%d Officers, want 1", len(officers))
	}
	o := tmtB1Card(t, g, officers[0])
	if o.Controller != me.ID || o.CurrentPower() != 1 || o.CurrentToughness() != 1 || len(o.Colors) != 1 || o.Colors[0] != "R" {
		t.Errorf("Officer %+v, want my 1/1 red", o)
	}
}

// --- Rules of Acquisition -------------------------------------------------

func TestTrekB8RulesOfAcquisitionThirdRule(t *testing.T) {
	g, me, opp := trekB8Main(t)
	theirs := b12Permanent(g, opp.ID, "Their Relic", "Artifact")
	seedLibrary(opp, "One", "Two", "Three", "Four")
	hand := opp.Hand.Size()
	trekB8Float(t, g, me, "{3}{U}{U}")
	castFromHandForTest(t, g, me, "Rules of Acquisition", "Sorcery", "{3}{U}{U}", trekB8RulesOracle, game.CastSpellParams{
		Modes: []int{0}, Strict: true, Targets: []game.TargetRef{trekB8Card(theirs)},
	})
	passPriorityAroundTable(t, g)
	if c := tmtB1Card(t, g, theirs); c.Controller != me.ID {
		t.Error("I did not gain control of the permanent")
	}
	if opp.Hand.Size() != hand+3 {
		t.Errorf("that opponent's hand %d, want %d", opp.Hand.Size(), hand+3)
	}
	// No duration: it is still mine next turn.
	advanceToNextSeatsTurn(t, g)
	if c := tmtB1Card(t, g, theirs); c.Controller != me.ID {
		t.Error("control went back at end of turn")
	}
}

func TestTrekB8RulesOfAcquisitionNinetyEighthRule(t *testing.T) {
	g, me, opp := trekB8Main(t)
	mine := b12Permanent(g, me.ID, "My Relic", "Artifact")
	seedLibrary(me, "One", "Two", "Three", "Four")
	trekB8Float(t, g, me, "{3}{U}{U}")
	castFromHandForTest(t, g, me, "Rules of Acquisition", "Sorcery", "{3}{U}{U}", trekB8RulesOracle, game.CastSpellParams{
		Modes: []int{1}, Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID, Slot: 0}, {Kind: game.TargetCard, ID: mine, Slot: 1}},
	})
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if c := tmtB1Card(t, g, mine); c.Controller != opp.ID {
		t.Error("the opponent did not gain control")
	}
	if me.Hand.Size() != hand+3 {
		t.Errorf("my hand %d, want %d", me.Hand.Size(), hand+3)
	}
}

func TestTrekB8RulesOfAcquisitionRefusesALandAndYourself(t *testing.T) {
	g, me, opp := trekB8Main(t)
	land := b12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	mine := b12Permanent(g, me.ID, "My Relic", "Artifact")
	trekB8Float(t, g, me, "{3}{U}{U}")
	if _, err := castFromHandErr(g, me, "Rules of Acquisition", "Sorcery", "{3}{U}{U}", trekB8RulesOracle, game.CastSpellParams{
		Modes: []int{0}, Strict: true, Targets: []game.TargetRef{trekB8Card(land)},
	}); err == nil {
		t.Fatal("the 3rd Rule took a land")
	}
	if _, err := castFromHandErr(g, me, "Rules of Acquisition", "Sorcery", "{3}{U}{U}", trekB8RulesOracle, game.CastSpellParams{
		Modes: []int{1}, Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID, Slot: 0}, {Kind: game.TargetCard, ID: mine, Slot: 1}},
	}); err == nil {
		t.Fatal("the 98th Rule targeted its caster as the opponent")
	}
}

// --- Scotty, Miracle Worker -------------------------------------------------

func trekB8Scotty(g *game.Game, p *game.Player) uuid.UUID {
	return b27Push(g, p.ID, "Scotty, Miracle Worker", "Legendary Creature — Human Artificer", trekB8ScottyOracle, "{1}{U}{U}", 2, 3, "U")
}

func trekB8GraveSorcery(g *game.Game, p *game.Player) uuid.UUID {
	return pushToGraveyard(g, p, game.Card{Name: "Divination", TypeLine: "Sorcery", ManaCost: "{2}{U}"})
}

func TestTrekB8ScottyBeamsUpASorceryForItsManaCost(t *testing.T) {
	g, me, _ := trekB8Main(t)
	scotty := trekB8Scotty(g, me)
	crew := b12Creature(g, me.ID, "Redshirt", "Creature — Human", 1, 1)
	card := trekB8GraveSorcery(g, me)
	keys := graveyardOfferKeys(g, me, card)
	if len(keys) != 1 || keys[0] != game.AltCostKeyGrantedBeamMeUp {
		t.Fatalf("graveyard offers %v, want only the granted beam me up", keys)
	}
	trekB8Float(t, g, me, "{2}{U}")
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: game.AltCostKeyGrantedBeamMeUp, AltCostIDs: []uuid.UUID{scotty}, Strict: true,
	}); err == nil {
		t.Fatal("Scotty was beamed up to cast a spell his own way")
	}
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: game.AltCostKeyGrantedBeamMeUp, AltCostIDs: []uuid.UUID{crew}, Strict: true,
	}); err != nil {
		t.Fatalf("granted beam me up cast: %v", err)
	}
	if !me.Hand.Contains(crew) {
		t.Fatal("the other creature was not returned to hand as the cost")
	}
	if n := me.ManaPool.Total(); n != 0 {
		t.Errorf("%d mana left, want 0: the price is its mana cost", n)
	}
	passPriorityAroundTable(t, g)
	if me.Graveyard.Contains(card) || !g.Exile.Contains(card) {
		t.Error("the beamed-up spell must be exiled as it leaves the stack")
	}
}

// Outside its controller's main phase, and for a creature card, nothing is
// offered.
func TestTrekB8ScottyOnlyDuringYourMainPhaseAndOnlyInstantsAndSorceries(t *testing.T) {
	g := newCatalogGame(t) // seat 0's draw step
	me := g.Seats[0]
	trekB8Scotty(g, me)
	b12Creature(g, me.ID, "Redshirt", "Creature — Human", 1, 1)
	instant := pushToGraveyard(g, me, game.Card{Name: "Opt", TypeLine: "Instant", ManaCost: "{U}"})
	bear := pushToGraveyard(g, me, graveBear())
	if keys := graveyardOfferKeys(g, me, instant); len(keys) != 0 {
		t.Errorf("offers %v in the draw step, want none", keys)
	}
	advanceTo(t, g, game.StepPrecombatMain)
	if keys := graveyardOfferKeys(g, me, instant); len(keys) != 1 || keys[0] != game.AltCostKeyGrantedBeamMeUp {
		t.Errorf("offers %v for an instant in my main phase, want the granted beam me up", keys)
	}
	if keys := graveyardOfferKeys(g, me, bear); len(keys) != 0 {
		t.Errorf("offers %v for a creature card, want none", keys)
	}
}

// --- Seven of Nine, Unassimilated -------------------------------------------

func trekB8Lands(g *game.Game, p *game.Player, names ...string) {
	for _, n := range names {
		b12Permanent(g, p.ID, n, "Land")
	}
}

func TestTrekB8SevenOfNineLootsWithFewLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b27Push(g, me.ID, "Seven of Nine, Unassimilated", "Legendary Artifact Creature — Human Officer", trekB8SevenOracle, "{1}{U}", 0, 4, "U")
	trekB8Lands(g, me, "Island", "Island", "Plains", "Swamp", "Mountain")
	seedLibrary(me, "One", "Two", "Three")
	advanceTo(t, g, game.StepPrecombatMain)
	passPriorityAroundTable(t, g)
	if discardOwed(g, me.ID) != 1 {
		t.Errorf("discard owed %d, want 1: four differently named lands", discardOwed(g, me.ID))
	}
}

func TestTrekB8SevenOfNineJustDrawsWithFiveNames(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b27Push(g, me.ID, "Seven of Nine, Unassimilated", "Legendary Artifact Creature — Human Officer", trekB8SevenOracle, "{1}{U}", 0, 4, "U")
	trekB8Lands(g, me, "Island", "Plains", "Swamp", "Mountain", "Forest")
	seedLibrary(me, "One", "Two", "Three")
	hand := me.Hand.Size()
	advanceTo(t, g, game.StepPrecombatMain)
	passPriorityAroundTable(t, g)
	if discardOwed(g, me.ID) != 0 || me.Hand.Size() != hand+1 {
		t.Errorf("discard owed %d, hand %d: want a plain draw to %d", discardOwed(g, me.ID), me.Hand.Size(), hand+1)
	}
}

func TestTrekB8SevenOfNineBecomesSevenNineWithSevenNames(t *testing.T) {
	g, me, _ := trekB8Main(t)
	seven := b27Push(g, me.ID, "Seven of Nine, Unassimilated", "Legendary Artifact Creature — Human Officer", trekB8SevenOracle, "{1}{U}", 0, 4, "U")
	trekB8Lands(g, me, "Island", "Plains", "Swamp", "Mountain", "Forest", "Wastes")
	trekB8Float(t, g, me, "{6}{U}")
	if err := g.ActivateCatalogAbility(me.ID, seven, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with six differently named lands")
	}
	trekB8Lands(g, me, "Island") // a second Island is not a new name
	if err := g.ActivateCatalogAbility(me.ID, seven, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with a duplicate name counted as the seventh")
	}
	trekB8Lands(g, me, "Desert")
	if err := g.ActivateCatalogAbility(me.ID, seven, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate with seven names: %v", err)
	}
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(seven, game.CounterPlusOne, 1) })
	c := tmtB1Card(t, g, seven)
	if c.CurrentPower() != 8 || c.CurrentToughness() != 10 {
		t.Errorf("Seven of Nine is %d/%d, want base 7/9 plus a counter", c.CurrentPower(), c.CurrentToughness())
	}
	advanceToNextSeatsTurn(t, g)
	if c := tmtB1Card(t, g, seven); c.CurrentPower() != 8 {
		t.Errorf("the base change ended (power %d): it has no duration", c.CurrentPower())
	}
}
