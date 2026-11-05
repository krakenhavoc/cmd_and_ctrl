package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b5_cards_test.go — Star Trek, batch 5 (#4973). Each test plays the
// real catalog entry through the engine and asserts what its text says.

const (
	trekB5TribblesOracle   = "fc76eddd-8f2a-45e3-95d2-3074d971effb"
	trekB5UnaOracle        = "0529c6fa-a35f-4d33-b274-93446ce68419"
	trekB5EnterpriseOracle = "5e2958fd-ede6-4563-babc-4924d3e95754"
	trekB5ExcelsiorOracle  = "9f4ae54b-f77d-43df-87d5-792ac0905058"
	trekB5DroneOracle      = "8e90e417-65ca-48bc-a32d-b93976519fe9"
	trekB5MissionOracle    = "535c0280-fdb5-4ff8-8608-c2f98b8a7ffc"
	trekB5CetaceanOracle   = "64e39ec9-358a-4aa0-a7e9-164e8aa0b236"
	trekB5CloakOracle      = "668a9b6b-731c-4127-9c88-1c6112ffd590"
	trekB5TelepathOracle   = "8656291d-8de7-4341-b1ee-0fcd65d629a4"
)

func TestTrekB5EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB5TribblesOracle:   "Troubling Tribbles",
		trekB5UnaOracle:        "Una Chin-Riley, Number One",
		trekB5EnterpriseOracle: "U.S.S. Enterprise-A, Free Spirit",
		trekB5ExcelsiorOracle:  "U.S.S. Excelsior, Cutting Edge",
		trekB5DroneOracle:      "Vexing Drone",
		trekB5MissionOracle:    "Anthropological Mission",
		trekB5CetaceanOracle:   "Cetacean Ops",
		trekB5CloakOracle:      "Cloak and Phaser",
		trekB5TelepathOracle:   "Cloistered Telepath",
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

// trekB5ToMain walks seat 0's first turn into its precombat main phase.
func trekB5ToMain(t *testing.T, g *game.Game) {
	t.Helper()
	advanceTo(t, g, game.StepPrecombatMain)
}

// trekB5CountTokensNamed counts the tokens named `name` that `controller`
// controls.
func trekB5CountTokensNamed(t *testing.T, g *game.Game, controller uuid.UUID, name string) []game.Card {
	t.Helper()
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name {
			v := layeredCard(t, g, c.InstanceID)
			if v.IsToken() {
				out = append(out, v)
			}
		}
	}
	return out
}

// trekB5OpponentTargets has seat 1 cast a Doom Blade at `victim` while
// seat 0 holds priority with an empty stack, and reports whether a ward
// payment prompt reached seat 1.
func trekB5OpponentTargets(t *testing.T, g *game.Game, victim uuid.UUID) bool {
	t.Helper()
	opp := g.Seats[1]
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	castAtWardedCreature(t, g, opp, victim)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	return hasPayUnlessFor(g, opp.ID)
}

// trekB5CastNoncreature casts a plain sorcery from seat 0's hand.
func trekB5CastNoncreature(t *testing.T, g *game.Game) {
	t.Helper()
	castFromHandForTest(t, g, g.Seats[0], "Plain Sorcery", "Sorcery", "", "", game.CastSpellParams{})
}

// trekB5Settle resolves the stack, answering every discard prompt for
// seat 0 off the top of the hand, and returns how many it answered.
func trekB5Settle(t *testing.T, g *game.Game) int {
	t.Helper()
	me := g.Seats[0]
	discards := 0
	for i := 0; i < 8; i++ {
		passPriorityAroundTable(t, g)
		if discardChoiceFor(g, me.ID) == nil {
			return discards
		}
		discardFromHand(t, g, me.ID)
		discards++
	}
	t.Fatal("the stack never settled")
	return discards
}

// --- Troubling Tribbles ------------------------------------------------------

func trekB5Tribbles(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b27Push(g, owner, "Troubling Tribbles", "Creature — Tribble", trekB5TribblesOracle, "{2}{W}", 0, 1, "W")
}

func TestTrekB5TribblesCopyThemselvesEachUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	trekB5Tribbles(g, me.ID)

	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if n := tmtB1Count(g, me.ID, "Troubling Tribbles"); n != 2 {
		t.Fatalf("%d Tribbles after one upkeep, want 2", n)
	}
	copies := trekB5CountTokensNamed(t, g, me.ID, "Troubling Tribbles")
	if len(copies) != 1 || copies[0].CurrentPower() != 0 || copies[0].CurrentToughness() != 1 || !copies[0].IsCreature() {
		t.Fatalf("copies %+v, want one 0/1 creature token named Troubling Tribbles", copies)
	}

	// The copy carries the trigger too, so the herd doubles.
	advanceTo(t, g, game.StepDraw)
	advanceToUpkeepOf(t, g, 0)
	// Two Tribbles trigger at once; their controller orders them.
	answerAnyTriggerOrderPrompt(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if n := tmtB1Count(g, me.ID, "Troubling Tribbles"); n != 4 {
		t.Fatalf("%d Tribbles after two upkeeps, want 4", n)
	}
}

func TestTrekB5TribblesDoNotTriggerOnAnOpponentsUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	trekB5Tribbles(g, me.ID)
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if n := tmtB1Count(g, me.ID, "Troubling Tribbles"); n != 1 {
		t.Fatalf("%d Tribbles on an opponent's upkeep, want 1", n)
	}
}

func TestTrekB5TribblesCantBlock(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tribbles := trekB5Tribbles(g, opp.ID)
	other := pushVanillaCreature(g, opp.ID, "Blocker", 2, 2)
	attacker := pushVanillaCreature(g, me.ID, "Attacker", 1, 1)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(tribbles, attacker); err == nil {
		t.Error("Troubling Tribbles blocked")
	}
	if err := g.DeclareBlocker(other, attacker); err != nil {
		t.Errorf("another creature could not block: %v", err)
	}
}

// --- Una Chin-Riley, Number One ---------------------------------------------

func trekB5UnaAttack(t *testing.T, g *game.Game, otherPower int) (una, attacker uuid.UUID) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	una = b27Push(g, me.ID, "Una Chin-Riley, Number One", "Legendary Creature — Illyrian Officer", trekB5UnaOracle, "{4}{W}", 4, 5, "W")
	attacker = b12Creature(g, me.ID, "Bear", "Creature — Bear", otherPower, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	return una, attacker
}

func TestTrekB5UnaPumpsAnAttackerAtTotalPowerEight(t *testing.T) {
	g := newCatalogGame(t)
	una, attacker := trekB5UnaAttack(t, g, 4)
	if n := triggersOnStackFrom(g, una); n != 1 && latestPickTarget(g, g.Seats[0].ID) == nil {
		t.Fatalf("Una did not trigger at total power 8 (triggers %d)", n)
	}
	answerPickTargetIfOpen(t, g, attacker)
	passPriorityAroundTable(t, g)
	c := layeredCard(t, g, attacker)
	if c.CurrentPower() != 6 || c.CurrentToughness() != 2 || !game.HasKeyword(&c, "first strike") {
		t.Errorf("the attacker is %d/%d first strike=%v, want 6/2 with first strike",
			c.CurrentPower(), c.CurrentToughness(), game.HasKeyword(&c, "first strike"))
	}
	if u := layeredCard(t, g, una); u.CurrentPower() != 4 {
		t.Errorf("Una (not attacking) is %d power, want 4", u.CurrentPower())
	}
}

func TestTrekB5UnaDoesNotTriggerBelowTotalPowerEight(t *testing.T) {
	g := newCatalogGame(t)
	una, attacker := trekB5UnaAttack(t, g, 3)
	if n := triggersOnStackFrom(g, una); n != 0 || latestPickTarget(g, g.Seats[0].ID) != nil {
		t.Fatalf("Una triggered at total power 7 (triggers %d)", n)
	}
	passPriorityAroundTable(t, g)
	if c := layeredCard(t, g, attacker); c.CurrentPower() != 3 {
		t.Errorf("the attacker is %d power, want 3", c.CurrentPower())
	}
}

// --- U.S.S. Enterprise-A, Free Spirit -----------------------------------------

func TestTrekB5EnterpriseMakesThreeRedOfficers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "U.S.S. Enterprise-A, Free Spirit", "Legendary Artifact — Spacecraft", trekB5EnterpriseOracle, nil)
	passPriorityAroundTable(t, g)
	officers := trekB5CountTokensNamed(t, g, me.ID, "Officer")
	if len(officers) != 3 {
		t.Fatalf("%d Officer tokens, want 3", len(officers))
	}
	for _, o := range officers {
		cols := o.EffectiveColors()
		if o.CurrentPower() != 1 || o.CurrentToughness() != 1 || len(cols) != 1 || cols[0] != "R" || !o.IsCreature() {
			t.Errorf("Officer %d/%d colours %v, want a 1/1 red creature", o.CurrentPower(), o.CurrentToughness(), cols)
		}
	}
}

func TestTrekB5EnterpriseStationsToAFlyingLifelinkerWithWard(t *testing.T) {
	g, me, ship, crew := stationBoard(t, "U.S.S. Enterprise-A, Free Spirit", "Legendary Artifact — Spacecraft", trekB5EnterpriseOracle, 5)
	station(t, g, me, ship, crew)
	passPriorityAroundTable(t, g)
	if c := layeredCard(t, g, ship); c.IsCreature() || game.HasKeyword(&c, "flying") {
		t.Fatal("five charge counters is below 6+")
	}
	second := b12Creature(g, me.ID, "Ensign", "Creature — Human", 1, 1)
	station(t, g, me, ship, second)
	passPriorityAroundTable(t, g)
	c := layeredCard(t, g, ship)
	if !c.IsCreature() || c.CurrentPower() != 5 || c.CurrentToughness() != 5 ||
		!game.HasKeyword(&c, "flying") || !game.HasKeyword(&c, "lifelink") {
		t.Fatalf("at 6+ want a 5/5 flying lifelink artifact creature; creature=%v %d/%d flying=%v lifelink=%v",
			c.IsCreature(), c.CurrentPower(), c.CurrentToughness(), game.HasKeyword(&c, "flying"), game.HasKeyword(&c, "lifelink"))
	}
	if !trekB5OpponentTargets(t, g, ship) {
		t.Fatal("an opponent's spell targeting the Enterprise-A asked for no ward payment")
	}
	answerPayUnless(t, g, g.Seats[1].ID, false)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ship) {
		t.Error("the unpaid ward did not counter the spell")
	}
}

// --- U.S.S. Excelsior, Cutting Edge ------------------------------------------

func TestTrekB5ExcelsiorDestroysATappedOpposingCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tapped := b12Creature(g, opp.ID, "Tapped Bear", "Creature — Bear", 2, 2)
	untapped := b12Creature(g, opp.ID, "Untapped Bear", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "My Tapped Bear", "Creature — Bear", 2, 2)
	tmtB4Tap(g, tapped)
	tmtB4Tap(g, mine)
	castCatalogSpell(t, g, "U.S.S. Excelsior, Cutting Edge", "Legendary Artifact — Spacecraft", trekB5ExcelsiorOracle, nil)
	passPriorityAroundTable(t, g)
	if pick := latestPickTarget(g, me.ID); pick != nil {
		for _, id := range pick.PickTargetCards {
			if id == untapped || id == mine {
				t.Errorf("%s was offered", cardNameFor(g, id))
			}
		}
	}
	answerPickTargetIfOpen(t, g, tapped)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tapped) {
		t.Error("the tapped opposing creature survived")
	}
	if !g.Battlefield.Contains(untapped) || !g.Battlefield.Contains(mine) {
		t.Error("a creature that was not the target left")
	}
}

func TestTrekB5ExcelsiorIsAFlierAtEight(t *testing.T) {
	g, me, ship, crew := stationBoard(t, "U.S.S. Excelsior, Cutting Edge", "Legendary Artifact — Spacecraft", trekB5ExcelsiorOracle, 7)
	station(t, g, me, ship, crew)
	passPriorityAroundTable(t, g)
	if c := layeredCard(t, g, ship); c.IsCreature() || game.HasKeyword(&c, "flying") {
		t.Fatal("seven charge counters is below 8+")
	}
	second := b12Creature(g, me.ID, "Ensign", "Creature — Human", 1, 1)
	station(t, g, me, ship, second)
	passPriorityAroundTable(t, g)
	c := layeredCard(t, g, ship)
	if !c.IsCreature() || c.CurrentPower() != 4 || c.CurrentToughness() != 4 || !game.HasKeyword(&c, "flying") {
		t.Errorf("at 8+ want a 4/4 flying artifact creature; creature=%v %d/%d flying=%v",
			c.IsCreature(), c.CurrentPower(), c.CurrentToughness(), game.HasKeyword(&c, "flying"))
	}
}

// --- Vexing Drone ----------------------------------------------------------------

func TestTrekB5VexingDroneTapsAnOpposingCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Vexing Drone", "Artifact Creature — Borg", trekB5DroneOracle, nil)
	passPriorityAroundTable(t, g)
	if pick := latestPickTarget(g, me.ID); pick != nil {
		for _, id := range pick.PickTargetCards {
			if id == mine {
				t.Error("a creature I control was offered")
			}
		}
	}
	answerPickTargetIfOpen(t, g, theirs)
	passPriorityAroundTable(t, g)
	if !layeredCard(t, g, theirs).Tapped {
		t.Error("the opposing creature is untapped")
	}
	if layeredCard(t, g, mine).Tapped {
		t.Error("my creature was tapped")
	}
}

// --- Anthropological Mission -------------------------------------------------

func TestTrekB5MissionLootsUntilThreeQuestCountersThenDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mission := pushCatalogPermanent(g, me.ID, "Anthropological Mission", "Enchantment", trekB5MissionOracle, false)
	trekB5ToMain(t, g)

	for i, wantDiscards := range []int{1, 1, 0, 0} {
		before := me.Hand.Size()
		trekB5CastNoncreature(t, g)
		discards := trekB5Settle(t, g)
		if discards != wantDiscards {
			t.Errorf("cast %d: %d discards, want %d", i+1, discards, wantDiscards)
		}
		// The sorcery is pushed and cast (net 0), the trigger draws one,
		// and the discard, while it applies, takes one.
		if got, want := me.Hand.Size(), before+1-wantDiscards; got != want {
			t.Errorf("cast %d: hand %d, want %d", i+1, got, want)
		}
		if got := counterOn(g, mission, "quest"); got != i+1 {
			t.Errorf("cast %d: %d quest counters, want %d", i+1, got, i+1)
		}
	}
}

func TestTrekB5MissionIgnoresCreatureSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mission := pushCatalogPermanent(g, me.ID, "Anthropological Mission", "Enchantment", trekB5MissionOracle, false)
	trekB5ToMain(t, g)
	castFromHandForTest(t, g, me, "Plain Bear", "Creature — Bear", "", "", game.CastSpellParams{})
	if n := trekB5Settle(t, g); n != 0 {
		t.Errorf("%d discards off a creature spell", n)
	}
	if got := counterOn(g, mission, "quest"); got != 0 {
		t.Errorf("%d quest counters off a creature spell, want 0", got)
	}
}

// --- Cetacean Ops --------------------------------------------------------------------

func TestTrekB5CetaceanOpsGrowsOnNoncreatureSpellsAndHasWard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ops := pushCatalogPermanent(g, me.ID, "Cetacean Ops", "Creature — Whale Officer", trekB5CetaceanOracle, false)
	trekB5ToMain(t, g)
	trekB5CastNoncreature(t, g)
	passPriorityAroundTable(t, g)
	castFromHandForTest(t, g, me, "Plain Bear", "Creature — Bear", "", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if got := counterOn(g, ops, game.CounterPlusOne); got != 1 {
		t.Fatalf("%d +1/+1 counters after one noncreature and one creature spell, want 1", got)
	}
	if !trekB5OpponentTargets(t, g, ops) {
		t.Fatal("an opponent's spell targeting Cetacean Ops asked for no ward payment")
	}
}

// --- Cloak and Phaser -----------------------------------------------------------

func TestTrekB5CloakAndPhaserCostsLessPerDifferentlyNamedLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Permanent(g, me.ID, "Island", "Basic Land — Island")
	b12Permanent(g, me.ID, "Island", "Basic Land — Island")
	b12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	b12Permanent(g, me.ID, "Mosswort Bridge", "Land")
	trekB5ToMain(t, g)

	// Three names: {8}{U} costs {5}{U}.
	floatMana(t, g, me, "{C}{C}{C}{C}{U}")
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Cloak and Phaser", TypeLine: "Sorcery", ManaCost: "{8}{U}",
			OracleID: trekB5CloakOracle, Owner: me.ID, Controller: me.ID,
		})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast for {4}{U} with three differently named lands")
	}
	floatMana(t, g, me, "{C}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast for {5}{U} with three differently named lands: %v", err)
	}
	before := me.Hand.Size()
	if n := trekB5Settle(t, g); n != 1 {
		t.Errorf("%d discards, want 1", n)
	}
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d, want %d (draw two, discard one)", got, before+1)
	}
	romulans := trekB5CountTokensNamed(t, g, me.ID, "Romulan")
	if len(romulans) != 2 {
		t.Fatalf("%d Romulan tokens, want 2", len(romulans))
	}
	for _, r := range romulans {
		cols := r.EffectiveColors()
		if r.CurrentPower() != 2 || r.CurrentToughness() != 2 || len(cols) != 1 || cols[0] != "B" || !r.IsCreature() {
			t.Errorf("Romulan %d/%d colours %v, want a 2/2 black creature", r.CurrentPower(), r.CurrentToughness(), cols)
		}
	}
}

// --- Cloistered Telepath -----------------------------------------------------

func TestTrekB5TelepathHasFlashAndLootsOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	// The harness parks on seat 0's draw step: not a main phase, so only
	// flash lets a creature spell be cast here.
	if g.Turn.Step != game.StepDraw {
		t.Fatalf("at %s, want the draw step", g.Turn.Step)
	}
	floatMana(t, g, me, "{C}{U}")
	castFromHandForTest(t, g, me, "Cloistered Telepath", "Creature — Talosian Advisor", "{1}{U}", trekB5TelepathOracle, game.CastSpellParams{Strict: true})
	before := me.Hand.Size()
	if n := trekB5Settle(t, g); n != 1 {
		t.Errorf("%d discards, want 1", n)
	}
	if got := me.Hand.Size(); got != before {
		t.Errorf("hand %d, want %d (draw one, discard one)", got, before)
	}
	if tmtB1Count(g, me.ID, "Cloistered Telepath") != 1 {
		t.Error("the Telepath is not on the battlefield")
	}
}

func TestTrekB5TelepathWithoutFlashWouldBeRefusedAtThisTiming(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	floatMana(t, g, me, "{C}{U}")
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Plain Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{U}",
			Owner: me.ID, Controller: me.ID,
		})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("a creature without flash was cast in the draw step; the flash test proves nothing")
	}
}
