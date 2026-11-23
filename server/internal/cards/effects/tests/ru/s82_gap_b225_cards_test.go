package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b225_cards_test.go — Commander Legends: Battle for Baldur's Gate
// white creatures, GAP batch 225 (#5948). Each test plays the real catalog
// entry through the engine.

const (
	gap225NagaOracle      = "67dec976-bcf5-4995-8da1-cd570862d3cf"
	gap225HammersOracle   = "97e260c9-3e78-402b-8d97-af4b515b278b"
	gap225RecruiterOracle = "9ecbc7de-9399-42f4-9c39-b9255080f48c"
	gap225StalwartOracle  = "9ad12c75-e97d-404a-bfa9-26ff1d0bb506"
	gap225LuluOracle      = "d99197b9-143a-462f-ace4-f35228ba082e"
	gap225MagusOracle     = "ed450f48-e69d-48d7-82cc-79cf89b2a1fb"
	gap225OrderOracle     = "16c46eaf-7a31-4957-ab7c-ff07fbbd9ddc"
	gap225PegasusOracle   = "e71863a7-0de1-4ab5-95e8-c39e6810d899"
	gap225RasaadOracle    = "3848ad74-8ee7-4546-9394-30db88e1d1e6"
	gap225ChwingaOracle   = "be1f0d72-14b5-41be-a37a-7abeb52ec0e5"
	gap225HarperOracle    = "6474f6b3-4e8e-4056-a01c-2f5e5e59c8de"
	gap225HawkOracle      = "0f96c48f-aceb-4554-8b06-cca162151a64"
)

func TestGapB225EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		gap225NagaOracle:           "Guardian Naga",
		gap225NagaOracle + "#1":    "Banishing Coils",
		gap225HammersOracle:        "Hammers of Moradin",
		gap225RecruiterOracle:      "Harper Recruiter",
		gap225StalwartOracle:       "Icewind Stalwart",
		gap225LuluOracle:           "Lulu, Loyal Hollyphant",
		gap225MagusOracle:          "Magus of the Balance",
		gap225OrderOracle:          "Order of Whiteclay",
		gap225PegasusOracle:        "Pegasus Guardian",
		gap225PegasusOracle + "#1": "Rescue the Foal",
		gap225RasaadOracle:         "Rasaad yn Bashir",
		gap225ChwingaOracle:        "Rescuer Chwinga",
		gap225HarperOracle:         "Roving Harper",
		gap225HawkOracle:           "Scouting Hawk",
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Roving Harper ---

func TestGapB225RovingHarperDrawsACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedCheapLibrary(g, me, 3)
	before := me.Hand.Size()
	effectstest.B12PlayFromHand(t, g, "Roving Harper", "Creature — Elf Scout", gap225HarperOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d, want %d (the Harper was put in hand, cast, and drew a card)", got, before+1)
	}
}

// --- Scouting Hawk ---

// gap225Choose answers the latest choose_cards prompt addressed to chooser.
func gap225Choose(t *testing.T, g *game.Game, chooser uuid.UUID, ids ...uuid.UUID) {
	t.Helper()
	c := effectstest.LatestChooseCards(g, chooser)
	if c == nil {
		t.Fatalf("no choose_cards prompt for %s", chooser)
	}
	if err := g.ResolveChooseCards(c.ID, chooser, ids); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
}

func gap225Plains(p *game.Player) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{InstanceID: id, Name: "Plains", TypeLine: "Basic Land — Plains", Owner: p.ID, Controller: p.ID})
	return id
}

func TestGapB225ScoutingHawkFetchesAPlainsOnlyWhenAnOpponentHasMoreLands(t *testing.T) {
	for _, behind := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		plains := gap225Plains(me)
		if behind {
			effectstest.AcLands(g, opp.ID, 2, "Their Land")
		}
		effectstest.B12PlayFromHand(t, g, "Scouting Hawk", "Creature — Bird", gap225HawkOracle, game.CastSpellParams{})
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.TakeSearchPrompts(t, g)
		onBF := g.Battlefield.Contains(plains)
		if onBF != behind {
			t.Errorf("behind=%v: Plains on the battlefield = %v", behind, onBF)
		}
		if behind {
			if c, ok := effectstest.BattlefieldCard(g, plains); !ok || !c.Tapped {
				t.Error("the Plains did not enter tapped")
			}
		}
	}
}

// --- Order of Whiteclay ---

func TestGapB225OrderOfWhiteclayNeedsToBeTappedAndReturnsASmallCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	order := effectstest.PushCatalogPermanent(g, me.ID, "Order of Whiteclay", "Creature — Kithkin Cleric", gap225OrderOracle, false)
	small := effectstest.PushGraveyardCardTyped(me, "Small Bear", "Creature — Bear")
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Big Bear", TypeLine: "Creature — Bear", ManaCost: "{4}{G}", Owner: me.ID, Controller: me.ID})
	effectstest.B06AddMana(me, "W", "W", "W")
	params := game.ActivateAbilityParams{Targets: effectstest.CardRefs(small)}
	if err := g.ActivateCatalogAbility(me.ID, order, 0, params); err == nil {
		t.Fatal("activated {Q} while untapped")
	}
	effectstest.AvB4SetTapped(g, order)
	if err := g.ActivateCatalogAbility(me.ID, order, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(small)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(small) {
		t.Error("the small creature did not return")
	}
	if effectstest.B16Tapped(t, g, order) {
		t.Error("the Order stayed tapped: {Q} untaps it")
	}
}

func TestGapB225OrderOfWhiteclayCannotPayWhileSummoningSick(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	order := effectstest.PushCatalogPermanent(g, me.ID, "Order of Whiteclay", "Creature — Kithkin Cleric", gap225OrderOracle, true)
	small := effectstest.PushGraveyardCardTyped(me, "Small Bear", "Creature — Bear")
	effectstest.AvB4SetTapped(g, order)
	effectstest.B06AddMana(me, "W", "W", "W")
	if err := g.ActivateCatalogAbility(me.ID, order, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(small)}); err == nil {
		t.Error("a summoning-sick creature paid {Q}")
	}
}

// --- Icewind Stalwart ---

func TestGapB225IcewindStalwartBlinksANonWarrior(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	warrior := effectstest.B12Creature(g, me.ID, "A Warrior", "Creature — Human Warrior", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "A Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterPlusOne, 1) })
	effectstest.B12PlayFromHand(t, g, "Icewind Stalwart", "Creature — Tiefling Warrior", gap225StalwartOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestPickTarget(g, me.ID)
	if c == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTargets(c.ID, me.ID, effectstest.CardRefs(warrior)); err == nil {
		t.Fatal("a Warrior was a legal target")
	}
	if err := g.ResolvePickTargets(c.ID, me.ID, effectstest.CardRefs(bear)); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear is the same object: it was not exiled and returned")
	}
	if n := effectstest.NamedOnBattlefield(g, "A Bear", me.ID); n != 1 {
		t.Error("the Bear did not come back")
	}
}

// --- Rescuer Chwinga ---

func TestGapB225RescuerChwingaMayReturnAnotherPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	effectstest.B12PlayFromHand(t, g, "Rescuer Chwinga", "Creature — Elemental Spirit", gap225ChwingaOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if pick == nil {
		t.Fatal("no permanent choice")
	}
	if err := g.ResolveOwnPermanents(pick.ID, me.ID, []uuid.UUID{rock}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the Rock was not returned")
	}
	if effectstest.B12ZoneOf(g, rock) != game.ZoneHand {
		t.Errorf("the Rock is in %v, want hand", effectstest.B12ZoneOf(g, rock))
	}
}

// --- Hammers of Moradin ---

func TestGapB225HammersOfMoradinTapsACreatureOfEachOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hammers := effectstest.B12Push(g, me.ID, "Hammers of Moradin", "Creature — Dwarf Cleric", gap225HammersOracle, 3, 3)
	var theirs []uuid.UUID
	for _, opp := range g.Seats[1:] {
		theirs = append(theirs, effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2))
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(hammers, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.HobAnswerOrder(t, g, me.ID)
	effectstest.HobAnswerTargets(t, g, me.ID, theirs...)
	effectstest.HobAnswerOrder(t, g, me.ID)
	for _, c := range g.PendingChoices {
		t.Logf("pending: %v %v", c.Kind, c.Reason)
	}
	for i := 0; i < 8; i++ {
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.LatestConfirmFor(g, me.ID) == nil {
			break
		}
		effectstest.AnswerConfirmFor(t, g, me, false)
	}
	for _, id := range theirs {
		if !effectstest.B16Tapped(t, g, id) {
			t.Errorf("creature %s was not tapped", id)
		}
	}
}

// --- Harper Recruiter ---

func TestGapB225HarperRecruiterTakesOneCardPerClass(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedCheapLibrary(g, me, 6)
	push := func(name, tl string) uuid.UUID {
		id := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: tl, Owner: me.ID, Controller: me.ID})
		return id
	}
	cleric := push("A Cleric", "Creature — Human Cleric")
	wizard := push("A Wizard", "Creature — Human Wizard")
	wizard2 := push("Another Wizard", "Creature — Human Wizard")
	bear := push("A Bear", "Creature — Bear")
	harper := effectstest.B12Push(g, me.ID, "Harper Recruiter", "Creature — Human Warrior", gap225RecruiterOracle, 3, 1)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(harper, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChooseCards(g, me.ID)
	if c == nil {
		t.Fatal("no choose_cards prompt")
	}
	// Two Wizards fill one slot only.
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{wizard, wizard2}); err == nil {
		t.Error("two Wizards were accepted for one slot")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{cleric, wizard}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{cleric, wizard} {
		if effectstest.B12ZoneOf(g, id) != game.ZoneHand {
			t.Errorf("%s did not reach the hand", id)
		}
	}
	for _, id := range []uuid.UUID{wizard2, bear} {
		if effectstest.B12ZoneOf(g, id) != game.ZoneLibrary {
			t.Errorf("%s left the library", id)
		}
	}
}

// --- Lulu, Loyal Hollyphant ---

func TestGapB225LuluCountersAndUntapsTappedCreaturesWithRevolt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Lulu, Loyal Hollyphant", "Legendary Creature — Elephant Angel", gap225LuluOracle, false)
	tapped := effectstest.B12Creature(g, me.ID, "Tapped Bear", "Creature — Bear", 2, 2)
	fresh := effectstest.B12Creature(g, me.ID, "Fresh Bear", "Creature — Bear", 2, 2)
	effectstest.AvB4SetTapped(g, tapped)
	effectstest.RevoltFor(t, g, me.ID)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.BattlefieldPlusOnes(g, tapped); n != 1 {
		t.Errorf("tapped creature has %d counters, want 1", n)
	}
	if n := effectstest.BattlefieldPlusOnes(g, fresh); n != 0 {
		t.Errorf("untapped creature has %d counters, want 0", n)
	}
	if effectstest.B16Tapped(t, g, tapped) {
		t.Error("the tapped creature stayed tapped")
	}
}

func TestGapB225LuluDoesNothingWithoutRevolt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Lulu, Loyal Hollyphant", "Legendary Creature — Elephant Angel", gap225LuluOracle, false)
	tapped := effectstest.B12Creature(g, me.ID, "Tapped Bear", "Creature — Bear", 2, 2)
	effectstest.AvB4SetTapped(g, tapped)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	if effectstest.TriggerOnStackFrom(g, "Lulu, Loyal Hollyphant") {
		t.Fatal("triggered with no permanent having left")
	}
}

// --- Pegasus Guardian // Rescue the Foal ---

func TestGapB225PegasusGuardianMakesAPegasusWithRevolt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Pegasus Guardian", "Creature — Pegasus", gap225PegasusOracle, 3, 3)
	effectstest.RevoltFor(t, g, me.ID)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountNamed(g, "Pegasus"); n != 1 {
		t.Errorf("%d Pegasus tokens, want 1", n)
	}
}

func TestGapB225RescueTheFoalBlinksYourCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "A Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterPlusOne, 1) })
	effectstest.CastCatalogSpell(t, g, "Rescue the Foal", "Instant — Adventure", gap225PegasusOracle+"#1", effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear kept its identity: it was not exiled and returned")
	}
	if effectstest.NamedOnBattlefield(g, "A Bear", me.ID) != 1 {
		t.Error("the Bear did not return")
	}
}

// --- Guardian Naga // Banishing Coils ---

func TestGapB225GuardianNagaPreventsDamageOnlyDuringYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	naga := effectstest.B12Push(g, me.ID, "Guardian Naga", "Creature — Snake", gap225NagaOracle, 5, 6)
	effectstest.Pr6Damage(t, g, naga, naga, 3)
	if d := effectstest.Pr6Marked(g, naga); d != 0 {
		t.Errorf("Naga took %d damage on its controller's turn", d)
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.Pr6Damage(t, g, naga, naga, 3)
	if d := effectstest.Pr6Marked(g, naga); d != 3 {
		t.Errorf("Naga took %d damage on another player's turn, want 3", d)
	}
}

func TestGapB225BanishingCoilsExilesAnArtifactOrEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	coils := effectstest.HandSpell(me, "Banishing Coils", "Instant — Adventure", "{0}")
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == coils {
				me.Hand.Cards[i].OracleID = gap225NagaOracle + "#1"
			}
		}
	})
	if err := g.CastSpell(me.ID, coils, game.CastSpellParams{Targets: effectstest.CardTarget(bear), Strict: true}); err == nil {
		t.Error("Banishing Coils targeted a creature")
	}
	effectstest.CastCatalogSpell(t, g, "Banishing Coils", "Instant — Adventure", gap225NagaOracle+"#1", effectstest.CardTarget(rock))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, rock) != game.ZoneExile {
		t.Errorf("the Rock is in %v, want exile", effectstest.B12ZoneOf(g, rock))
	}
}

// --- Rasaad yn Bashir ---

func TestGapB225RasaadAssignsDamageByToughnessAndDoublesItWithTheInitiative(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rasaad := effectstest.B12Push(g, me.ID, "Rasaad yn Bashir", "Legendary Creature — Human Monk", gap225RasaadOracle, 0, 3)
	wall := effectstest.B12Creature(g, me.ID, "Wall", "Creature — Wall", 0, 4)
	if !effectstest.GaLayered(t, g, wall).Effective().DamageByToughnessAlways {
		t.Fatal("another creature you control does not assign combat damage by toughness")
	}
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 5, 5)
	if effectstest.GaLayered(t, g, theirs).Effective().DamageByToughnessAlways {
		t.Error("an opposing creature assigns damage by toughness")
	}
	if err := g.SetInitiative(me.ID); err != nil {
		t.Fatalf("SetInitiative: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(rasaad, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LayeredCard(t, g, rasaad).CurrentToughness(); got != 6 {
		t.Errorf("Rasaad's toughness is %d, want 6", got)
	}
	if got := effectstest.LayeredCard(t, g, wall).CurrentToughness(); got != 8 {
		t.Errorf("the Wall's toughness is %d, want 8", got)
	}
	if got := effectstest.LayeredCard(t, g, theirs).CurrentToughness(); got != 5 {
		t.Errorf("an opposing creature's toughness changed to %d", got)
	}
}

// --- Magus of the Balance ---

func TestGapB225MagusBalancesLandsCardsAndCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	magus := effectstest.PushCatalogPermanent(g, me.ID, "Magus of the Balance", "Creature — Human Wizard", gap225MagusOracle, false)
	myLands := effectstest.AcLands(g, me.ID, 4, "My Land")
	oppLands := effectstest.AcLands(g, opp.ID, 2, "Their Land")
	for _, s := range g.Seats[2:] {
		effectstest.AcLands(g, s.ID, 3, "Land")
	}
	myBears := []uuid.UUID{
		effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2),
		effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2),
	}
	effectstest.B06AddMana(me, "W", "W", "W", "W", "W")
	if err := g.ActivateCatalogAbility(me.ID, magus, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	// Answer every prompt: keep the first lands/creatures in the list, give up the rest.
	for i := 0; i < 40; i++ {
		effectstest.PassPriorityAroundTable(t, g)
		answered := false
		for _, s := range g.Seats {
			if c := effectstest.LatestChooseCards(g, s.ID); c != nil && len(c.ChooseCards) > 0 {
				n := c.ChooseMin
				if n == 0 {
					n = 1
				}
				if n > len(c.ChooseCards) {
					n = len(c.ChooseCards)
				}
				gap225Choose(t, g, s.ID, c.ChooseCards[:n]...)
				answered = true
				break
			}
		}
		if !answered {
			break
		}
	}
	count := func(p *game.Player, pred func(game.Card) bool) int {
		n := 0
		for _, c := range g.Battlefield.Cards {
			if c.Controller == p.ID && pred(c) {
				n++
			}
		}
		return n
	}
	for _, p := range g.Seats {
		if got := count(p, game.Card.IsLand); got != 2 {
			t.Errorf("%s controls %d lands, want 2", p.Name, got)
		}
	}
	if got := count(me, game.Card.IsCreature); got != 0 {
		t.Errorf("I control %d creatures, want 0 (the Magus is sacrificed; the fewest is 0)", got)
	}
	_, _, _ = myLands, oppLands, myBears
}
