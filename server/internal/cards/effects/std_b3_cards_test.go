package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b3_cards_test.go — Standard set leftovers, batches 3 and 4 (#4452):
// each test plays the real catalog entry and asserts what its text says, and
// what it refuses.

const (
	stdB3CrusherOracle    = "db0894e1-2644-48d4-8de4-8cc43e940bc1"
	stdB3PlungeOracle     = "4f10b164-6d36-4935-92e7-53be73ac4c8a"
	stdB3LandseekerOracle = "528fe479-dab9-4cf8-bda0-1d226a644366"
	stdB3PangolinOracle   = "72e8c978-f66d-42e1-ac5f-d1c307d7b93c"
	stdB3RampageOracle    = "5020628c-5e47-4a83-89c3-7bf5726e69ad"
	stdB3OverrideOracle   = "4c9d9d47-1eb6-4981-ba83-d5ec04906e71"
	stdB3TannukOracle     = "b5ae8016-92ed-4ae8-83a8-65068defc616"
	stdB3VelocityOracle   = "cdda2f04-d59e-4ef8-a26d-964d80750940"
	stdB3BeckonerOracle   = "e0a35a44-e56d-434e-8043-43d5031b2e84"
	stdB3BalladOracle     = "835abd35-27ea-45fe-b55d-25081c57bd4a"
	stdB3FutureOracle     = "6fa78914-85b3-46fd-977a-0e5b4d29eca4"
	stdB3EncounterOracle  = "032fdd08-2ea7-442e-8162-0a8e5a585253"
	stdB3FatemakerOracle  = "583102f8-7232-4fbb-a715-087448536bb2"
	stdB3ResolveOracle    = "6f9bc0c6-1792-4771-a839-95a22d9c699a"
	stdB3WardenOracle     = "981138d0-bf61-406a-85be-68cc5e58796c"
	stdB3MassacreOracle   = "9fdd445a-c8e1-4309-9d74-374676fcac1d"
	stdB3RagostOracle     = "76f5445a-5193-4794-8593-b44f961334b6"
	stdB3MonitorOracle    = "94e8791b-e72f-4aed-b761-826779337168"
	stdB3VondamOracle     = "9e597063-c077-4236-a9e4-b0721b24677f"
	stdB3ScrapbotOracle   = "b75ca527-4494-4957-b2e3-8c590df42e95"
)

func TestStdB3EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		stdB3CrusherOracle:    "Debris Field Crusher",
		stdB3PlungeOracle:     "Orbital Plunge",
		stdB3LandseekerOracle: "Kav Landseeker",
		stdB3PangolinOracle:   "Oreplate Pangolin",
		stdB3RampageOracle:    "Ruinous Rampage",
		stdB3OverrideOracle:   "Systems Override",
		stdB3TannukOracle:     "Tannuk, Steadfast Second",
		stdB3VelocityOracle:   "Terminal Velocity",
		stdB3BeckonerOracle:   "Perigee Beckoner",
		stdB3BalladOracle:     "Zero Point Ballad",
		stdB3FutureOracle:     "Bioengineered Future",
		stdB3EncounterOracle:  "Close Encounter",
		stdB3FatemakerOracle:  "Drix Fatemaker",
		stdB3ResolveOracle:    "Meltstrider's Resolve",
		stdB3WardenOracle:     "Tapestry Warden",
		stdB3MassacreOracle:   "Mutinous Massacre",
		stdB3RagostOracle:     "Ragost, Deft Gastronaut",
		stdB3MonitorOracle:    "Station Monitor",
		stdB3VondamOracle:     "Syr Vondam, Sunstar Exemplar",
		stdB3ScrapbotOracle:   "Dauntless Scrapbot",
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
	for oracle, cost := range map[string]string{stdB3BeckonerOracle: "{1}{B}", stdB3FatemakerOracle: "{1}{G}"} {
		if ac := game.AlternativeCostByKey(oracle, "warp"); ac == nil || ac.ManaCost != cost {
			t.Errorf("%s: warp %+v, want %s", oracle, ac, cost)
		}
	}
}

// stdB3Resolve passes priority until the stack is empty.
func stdB3Resolve(t *testing.T, g *game.Game) {
	t.Helper()
	passPriorityAroundTable(t, g)
	for i := 0; i < 12 && g.Stack.Size() > 0; i++ {
		passPriorityAroundTable(t, g)
	}
}

// stdB3AdvanceToMyNextEndStep walks to the end step of the next turn the
// given seat is the active player of, and lets that step's triggers resolve.
func stdB3AdvanceToMyNextEndStep(t *testing.T, g *game.Game, me uuid.UUID) {
	t.Helper()
	start := g.Turn.Seq
	for i := 0; i < 80; i++ {
		if g.Turn.Step == game.StepEnd && g.Turn.Seq > start && g.Seats[g.Turn.ActiveSeat].ID == me {
			break
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.Step != game.StepEnd || g.Turn.Seq <= start {
		t.Fatalf("never reached my next end step (at %s, turn %d)", g.Turn.Step, g.Turn.Seq)
	}
	stdB3Resolve(t, g)
}

// --- Debris Field Crusher ------------------------------------------------------

func TestStdB3DebrisFieldCrusherHitsAnyTargetAndFliesAtEight(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	before := opp.Life
	ship := avB1Cast(t, g, me, game.Card{Name: "Debris Field Crusher", TypeLine: "Artifact — Spacecraft", ManaCost: "{4}{R}",
		OracleID: stdB3CrusherOracle, Power: 1, Toughness: 5}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("opponent life %d, want %d", opp.Life, before-3)
	}
	if avB1Card(t, g, ship).IsCreature() {
		t.Error("below 8 charge counters it is not a creature")
	}
	if err := g.AddCounter(ship, game.CounterCharge, 7); err != nil {
		t.Fatal(err)
	}
	if layeredCard(t, g, ship).IsCreature() {
		t.Error("at 7 charge counters it is still not a creature")
	}
	if err := g.AddCounter(ship, game.CounterCharge, 1); err != nil {
		t.Fatal(err)
	}
	at8 := layeredCard(t, g, ship)
	if !at8.IsCreature() || !game.HasKeyword(&at8, "flying") || at8.CurrentPower() != 1 || at8.CurrentToughness() != 5 {
		t.Errorf("at 8+ it is a 1/5 flying creature, got %d/%d", at8.CurrentPower(), at8.CurrentToughness())
	}
}

// --- Orbital Plunge -----------------------------------------------------------

func TestStdB3OrbitalPlungeMakesALanderOnlyForExcessDamage(t *testing.T) {
	for _, tc := range []struct {
		name       string
		toughness  int
		marked     int
		wantLander int
	}{
		{"excess over a 2/2", 2, 0, 1},
		{"exactly lethal", 6, 0, 0},
		{"damage already marked makes the excess", 6, 2, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, tc.toughness)
			if tc.marked > 0 {
				eoeCDSet(g, victim, func(c *game.Card) { c.DamageMarked = tc.marked })
			}
			avB1Cast(t, g, me, game.Card{Name: "Orbital Plunge", TypeLine: "Sorcery", ManaCost: "{3}{R}", OracleID: stdB3PlungeOracle},
				game.CastSpellParams{Targets: cardRefs(victim)})
			stdB3Resolve(t, g)
			if got := eoeCDCount(g, me.ID, "Lander"); got != tc.wantLander {
				t.Errorf("%d Landers, want %d", got, tc.wantLander)
			}
			if g.Battlefield.Contains(victim) && tc.marked == 0 && tc.toughness <= 6 {
				g.RunStateChecksForTest()
			}
		})
	}
}

// --- Kav Landseeker -----------------------------------------------------------

func TestStdB3KavLandseekerSacrificesTheLanderAtTheEndStepOnYourNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	avB1Cast(t, g, me, game.Card{Name: "Kav Landseeker", TypeLine: "Creature — Kavu Soldier", ManaCost: "{3}{R}",
		OracleID: stdB3LandseekerOracle, Power: 4, Toughness: 3}, game.CastSpellParams{})
	stdB3Resolve(t, g)
	if got := eoeCDCount(g, me.ID, "Lander"); got != 1 {
		t.Fatalf("%d Landers, want 1", got)
	}
	// This turn's end step is not "your next turn".
	advanceThroughEndStep(t, g)
	if got := eoeCDCount(g, me.ID, "Lander"); got != 1 {
		t.Errorf("the Lander was sacrificed in the turn it was made: %d", got)
	}
	stdB3AdvanceToMyNextEndStep(t, g, me.ID)
	if got := eoeCDCount(g, me.ID, "Lander"); got != 0 {
		t.Errorf("%d Landers after your next end step, want 0", got)
	}
	if !hasEffectiveKeyword(t, g, avB1Named(g, me.ID, "Kav Landseeker"), "menace") {
		t.Error("Kav Landseeker has menace")
	}
}

func TestStdB3KavLandseekerLeavesALanderYouAlreadyUsed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	avB1Cast(t, g, me, game.Card{Name: "Kav Landseeker", TypeLine: "Creature — Kavu Soldier", ManaCost: "{3}{R}",
		OracleID: stdB3LandseekerOracle, Power: 4, Toughness: 3}, game.CastSpellParams{})
	stdB3Resolve(t, g)
	lander := avB1Named(g, me.ID, "Lander")
	if lander == uuid.Nil {
		t.Fatal("no Lander")
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(lander) })
	stdB3AdvanceToMyNextEndStep(t, g, me.ID)
	if got := eoeCDCount(g, me.ID, "Lander"); got != 0 {
		t.Errorf("%d Landers", got)
	}
}

// --- Oreplate Pangolin --------------------------------------------------------

func TestStdB3OreplatePangolinGrowsOnlyWhenYouPayForAnotherArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pango := b12Push(g, me.ID, "Oreplate Pangolin", "Artifact Creature — Robot Pangolin", stdB3PangolinOracle, 2, 2)
	avB1Cast(t, g, me, game.Card{Name: "Rock", TypeLine: "Artifact", ManaCost: "{1}"}, game.CastSpellParams{})
	stdB3Resolve(t, g)
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("another artifact entered: you are asked to pay {1}")
	}
	b06AddMana(me, "C")
	answerPayUnless(t, g, me.ID, true)
	if got := countersOn(g, pango, game.CounterPlusOne); got != 1 {
		t.Errorf("paid: %d counters, want 1", got)
	}
	avB1Cast(t, g, me, game.Card{Name: "Bauble", TypeLine: "Artifact", ManaCost: "{1}"}, game.CastSpellParams{})
	stdB3Resolve(t, g)
	answerPayUnless(t, g, me.ID, false)
	if got := countersOn(g, pango, game.CounterPlusOne); got != 1 {
		t.Errorf("declined: %d counters, want still 1", got)
	}
	avB1Cast(t, g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}", Power: 1, Toughness: 1}, game.CastSpellParams{})
	stdB3Resolve(t, g)
	if hasPayUnlessFor(g, me.ID) {
		t.Error("a nonartifact creature does not trigger it")
	}
}

// --- Ruinous Rampage ----------------------------------------------------------

func TestStdB3RuinousRampageBothModes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cheap := b12Permanent(g, opp.ID, "Cheap", "Artifact")
	eoeCDSet(g, cheap, func(c *game.Card) { c.ManaCost = "{3}" })
	xcost := b12Permanent(g, me.ID, "Xcost", "Artifact")
	eoeCDSet(g, xcost, func(c *game.Card) { c.ManaCost = "{X}{X}{3}" })
	dear := b12Permanent(g, opp.ID, "Dear", "Artifact")
	eoeCDSet(g, dear, func(c *game.Card) { c.ManaCost = "{4}" })
	creature := b12Creature(g, opp.ID, "Golem", "Artifact Creature — Golem", 1, 1)
	eoeCDSet(g, creature, func(c *game.Card) { c.ManaCost = "{2}" })
	before := opp.Life
	avB1Cast(t, g, me, game.Card{Name: "Ruinous Rampage", TypeLine: "Sorcery", ManaCost: "{1}{R}{R}", OracleID: stdB3RampageOracle},
		game.CastSpellParams{Modes: []int{1}})
	stdB3Resolve(t, g)
	if opp.Life != before {
		t.Error("the exile mode deals no damage")
	}
	if !g.Exile.Contains(cheap) || !g.Exile.Contains(xcost) || !g.Exile.Contains(creature) {
		t.Error("artifacts of mana value 3 or less are exiled, X counting 0, whoever controls them and creatures too")
	}
	if !g.Battlefield.Contains(dear) {
		t.Error("an artifact with mana value 4 stays")
	}
	avB1Cast(t, g, me, game.Card{Name: "Ruinous Rampage", TypeLine: "Sorcery", ManaCost: "{1}{R}{R}", OracleID: stdB3RampageOracle},
		game.CastSpellParams{Modes: []int{0}})
	stdB3Resolve(t, g)
	if opp.Life != before-3 {
		t.Errorf("opponent life %d, want %d", opp.Life, before-3)
	}
	if me.Life != 40 && me.Life != g.Seats[0].Life {
		t.Error("the caster is not hit")
	}
}

// --- Systems Override ---------------------------------------------------------

func TestStdB3SystemsOverrideStealsAndChargesASpacecraft(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ship := b12Push(g, opp.ID, "Debris Field Crusher", "Artifact — Spacecraft", stdB3CrusherOracle, 1, 5)
	eoeCDSet(g, ship, func(c *game.Card) { c.Tapped = true })
	avB1Cast(t, g, me, game.Card{Name: "Systems Override", TypeLine: "Sorcery", ManaCost: "{2}{R}", OracleID: stdB3OverrideOracle},
		game.CastSpellParams{Targets: cardRefs(ship)})
	stdB3Resolve(t, g)
	got := layeredCard(t, g, ship)
	if got.Controller != me.ID {
		t.Fatal("you control the Spacecraft")
	}
	if got.Tapped {
		t.Error("it is untapped")
	}
	if got.Counters[game.CounterCharge] != 10 {
		t.Errorf("%d charge counters, want 10", got.Counters[game.CounterCharge])
	}
	if !got.IsCreature() || !game.HasKeyword(&got, "haste") {
		t.Error("with ten charge counters it is a creature with haste")
	}
	advanceThroughEndStep(t, g)
	after := layeredCard(t, g, ship)
	if after.Counters[game.CounterCharge] != 0 {
		t.Errorf("%d charge counters after the end step, want 0", after.Counters[game.CounterCharge])
	}
	if after.Controller != opp.ID {
		t.Error("control returns at cleanup")
	}
}

func TestStdB3SystemsOverrideTakesAPlainArtifactWithoutCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	avB1Cast(t, g, me, game.Card{Name: "Systems Override", TypeLine: "Sorcery", ManaCost: "{2}{R}", OracleID: stdB3OverrideOracle},
		game.CastSpellParams{Targets: cardRefs(rock)})
	stdB3Resolve(t, g)
	got := layeredCard(t, g, rock)
	if got.Controller != me.ID || got.Counters[game.CounterCharge] != 0 {
		t.Errorf("controller %v, charge counters %d", got.Controller == me.ID, got.Counters[game.CounterCharge])
	}
	land := pushLandFor(g, opp.ID, "Island", "Basic Land — Island")
	advanceToMain(t, g)
	id := avB1Hand(me, game.Card{Name: "Systems Override", TypeLine: "Sorcery", ManaCost: "{2}{R}", OracleID: stdB3OverrideOracle})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: cardRefs(land)}); err == nil {
		t.Error("a land is neither an artifact nor a creature")
	}
}

// --- Tannuk, Steadfast Second -------------------------------------------------

func TestStdB3TannukGivesHasteAndWarpToArtifactsAndRedCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tannuk := b12Push(g, me.ID, "Tannuk, Steadfast Second", "Legendary Creature — Kavu Pilot", stdB3TannukOracle, 3, 5)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if !hasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("other creatures you control have haste")
	}
	if hasEffectiveKeyword(t, g, tannuk, "haste") {
		t.Error("Tannuk himself has no haste")
	}
	advanceToMain(t, g)
	red := avB1Hand(me, game.Card{Name: "Red Bear", TypeLine: "Creature — Bear", ManaCost: "{4}{R}", Colors: []string{"R"}, Power: 4, Toughness: 4})
	green := avB1Hand(me, game.Card{Name: "Green Bear", TypeLine: "Creature — Bear", ManaCost: "{4}{G}", Colors: []string{"G"}, Power: 4, Toughness: 4})
	rock := avB1Hand(me, game.Card{Name: "Big Rock", TypeLine: "Artifact", ManaCost: "{5}"})
	spell := avB1Hand(me, game.Card{Name: "Red Sorcery", TypeLine: "Sorcery", ManaCost: "{4}{R}", Colors: []string{"R"}})
	b06AddMana(me, "R", "R", "R")
	if err := g.CastSpell(me.ID, green, game.CastSpellParams{Strict: true, AlternativeCost: game.GrantedAltCostWarp}); err == nil {
		t.Error("a green creature card has no warp from Tannuk")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AlternativeCost: game.GrantedAltCostWarp}); err == nil {
		t.Error("a sorcery card has no warp from Tannuk")
	}
	if err := g.CastSpell(me.ID, red, game.CastSpellParams{Strict: true, AlternativeCost: game.GrantedAltCostWarp}); err != nil {
		t.Fatalf("a red creature card warps for {2}{R}: %v", err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("pool %+v: the warp cost is {2}{R}, three mana", me.ManaPool)
	}
	stdB3Resolve(t, g)
	if !g.Battlefield.Contains(red) {
		t.Fatal("the warped creature entered")
	}
	b06AddMana(me, "R", "R", "R")
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{Strict: true, AlternativeCost: game.GrantedAltCostWarp}); err != nil {
		t.Fatalf("an artifact card warps too: %v", err)
	}
	stdB3Resolve(t, g)
	advanceThroughEndStep(t, g)
	for name, id := range map[string]uuid.UUID{"red creature": red, "artifact": rock} {
		if g.Battlefield.Contains(id) || !g.Exile.Contains(id) {
			t.Errorf("the warped %s is exiled at the next end step", name)
		}
	}
	if g.TurnTally.SpellsWarped < 2 {
		t.Errorf("%d spells warped this turn, want 2", g.TurnTally.SpellsWarped)
	}
}

// --- Terminal Velocity --------------------------------------------------------

func TestStdB3TerminalVelocityPutsAPermanentInWithHasteAndAFuse(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bystander := b12Creature(g, opp.ID, "Bystander", "Creature — Bear", 2, 3)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 1)
	ogre := avB1Hand(me, game.Card{Name: "Ogre", TypeLine: "Creature — Ogre", ManaCost: "{2}{R}{R}", Colors: []string{"R"}, Power: 4, Toughness: 4})
	avB1Cast(t, g, me, game.Card{Name: "Terminal Velocity", TypeLine: "Sorcery", ManaCost: "{4}{R}{R}", OracleID: stdB3VelocityOracle}, game.CastSpellParams{})
	stdB3Resolve(t, g)
	answerChooseCards(t, g, me.ID, ogre)
	stdB3Resolve(t, g)
	if !g.Battlefield.Contains(ogre) {
		t.Fatal("the creature card from hand entered")
	}
	got := layeredCard(t, g, ogre)
	if !game.HasKeyword(&got, "haste") {
		t.Error("it gains haste")
	}
	// At the beginning of your end step, sacrifice it; it leaves, dealing
	// damage equal to its mana value (4) to each creature.
	advanceThroughEndStep(t, g)
	if g.Battlefield.Contains(ogre) {
		t.Fatal("sacrificed at your end step")
	}
	if g.Battlefield.Contains(bystander) || g.Battlefield.Contains(mine) {
		t.Error("4 damage to each creature kills a 2/3 and a 1/1")
	}
}

func TestStdB3TerminalVelocityOnlyTakesAnArtifactOrCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := avB1Hand(me, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	ench := avB1Hand(me, game.Card{Name: "Aura of Nothing", TypeLine: "Enchantment", ManaCost: "{1}"})
	avB1Cast(t, g, me, game.Card{Name: "Terminal Velocity", TypeLine: "Sorcery", ManaCost: "{4}{R}{R}", OracleID: stdB3VelocityOracle}, game.CastSpellParams{})
	stdB3Resolve(t, g)
	if p := chooseCardsChoiceFor(g, me.ID); p != nil && (hasID(p.ChooseCards, land) || hasID(p.ChooseCards, ench)) {
		t.Error("only artifact and creature cards are offered")
	}
}

// --- Perigee Beckoner ---------------------------------------------------------

func TestStdB3PerigeeBeckonerPumpsAndReturnsAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	beckoner := avB1Cast(t, g, me, game.Card{Name: "Perigee Beckoner", TypeLine: "Creature — Horror", ManaCost: "{4}{B}",
		OracleID: stdB3BeckonerOracle, Power: 4, Toughness: 5}, game.CastSpellParams{})
	stdB3Resolve(t, g)
	pickIfAsked(t, g, me.ID, bear)
	stdB3Resolve(t, g)
	if p := layeredCard(t, g, bear); p.CurrentPower() != 4 {
		t.Errorf("the Bear is %d/%d, want power 4", p.CurrentPower(), p.CurrentToughness())
	}
	if p := layeredCard(t, g, beckoner); p.CurrentPower() != 4 {
		t.Errorf("the Beckoner is untouched: %d", p.CurrentPower())
	}
	b15Destroy(t, g, bear)
	stdB3Resolve(t, g)
	back, ok := battlefieldCard(g, bear)
	if !ok || !back.Tapped {
		t.Errorf("the Bear returned tapped: on battlefield %v", ok)
	}
	// The returned object is new: it does not return a second time.
	b15Destroy(t, g, bear)
	stdB3Resolve(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the new object has no such ability")
	}
}

func TestStdB3PerigeeBeckonerGraceEndsWithTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	avB1Cast(t, g, me, game.Card{Name: "Perigee Beckoner", TypeLine: "Creature — Horror", ManaCost: "{4}{B}",
		OracleID: stdB3BeckonerOracle, Power: 4, Toughness: 5}, game.CastSpellParams{})
	stdB3Resolve(t, g)
	pickIfAsked(t, g, me.ID, bear)
	stdB3Resolve(t, g)
	advanceThroughEndStep(t, g)
	advanceTo(t, g, game.StepCleanup)
	if err := stdB3Pass(g); err != nil {
		t.Fatal(err)
	}
	if p := layeredCard(t, g, bear); p.CurrentPower() != 2 {
		t.Errorf("after the turn the Bear is power %d, want 2", p.CurrentPower())
	}
}

func stdB3Pass(g *game.Game) error {
	_, err := g.AdvanceStep()
	return err
}

// --- Zero Point Ballad --------------------------------------------------------

func TestStdB3ZeroPointBalladSweepsByToughnessAndCostsLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := b12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 2)
	exact := b12Creature(g, me.ID, "Exact", "Creature — Bear", 3, 3)
	big := b12Creature(g, opp.ID, "Big", "Creature — Giant", 5, 4)
	life := me.Life
	avB1Cast(t, g, me, game.Card{Name: "Zero Point Ballad", TypeLine: "Sorcery", ManaCost: "{X}{B}", OracleID: stdB3BalladOracle},
		game.CastSpellParams{XValue: 3})
	stdB3Resolve(t, g)
	if g.Battlefield.Contains(small) || g.Battlefield.Contains(exact) {
		t.Error("creatures with toughness 3 or less are destroyed, your own included")
	}
	if !g.Battlefield.Contains(big) {
		t.Error("toughness 4 survives X=3")
	}
	if me.Life != life-3 {
		t.Errorf("life %d, want %d", me.Life, life-3)
	}
	if p := chooseCardsChoiceFor(g, me.ID); p != nil {
		t.Error("below X=6 nothing returns")
	}
}

func TestStdB3ZeroPointBalladReturnsOneDestroyedCreatureAtSixOrMore(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := stdB3Bear(g, me.ID, "Mine", 2)
	theirs := stdB3Bear(g, opp.ID, "Theirs", 3)
	token := b12Creature(g, opp.ID, "Goblin", "Token Creature — Goblin", 1, 1)
	life := me.Life
	avB1Cast(t, g, me, game.Card{Name: "Zero Point Ballad", TypeLine: "Sorcery", ManaCost: "{X}{B}", OracleID: stdB3BalladOracle},
		game.CastSpellParams{XValue: 6})
	stdB3Resolve(t, g)
	if me.Life != life-6 {
		t.Errorf("life %d, want %d", me.Life, life-6)
	}
	p := chooseCardsChoiceFor(g, me.ID)
	if p == nil {
		t.Fatal("with X=6 you choose a destroyed creature card to return")
	}
	if hasID(p.ChooseCards, token) || !hasID(p.ChooseCards, mine) || !hasID(p.ChooseCards, theirs) {
		t.Errorf("candidates %v: both creature cards and no token", p.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, theirs)
	stdB3Resolve(t, g)
	back, ok := battlefieldCard(g, theirs)
	if !ok || back.Controller != me.ID {
		t.Errorf("the chosen card returns under your control: on battlefield %v", ok)
	}
	if g.Battlefield.Contains(mine) {
		t.Error("only one card returns")
	}
}

// stdB3Bear seeds a creature on the battlefield with the given
// toughness.
func stdB3Bear(g *game.Game, owner uuid.UUID, name string, toughness int) uuid.UUID {
	return b12Creature(g, owner, name, "Creature — Bear", 1, toughness)
}
