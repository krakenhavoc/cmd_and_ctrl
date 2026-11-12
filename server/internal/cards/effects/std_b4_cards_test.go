package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b4_cards_test.go — Standard set leftovers, batch 4 (#4452).

const stdB3AdagiaOracle = "70d35dbd-1d91-4a2a-a643-6870d168f4f5"

// stdB3Exile exiles a battlefield permanent.
func stdB3Exile(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneExile}, id); err != nil {
		t.Fatal(err)
	}
	stdB3Resolve(t, g)
}

// stdB3AbilityIndex finds an activated ability by a fragment of its label.
func stdB3AbilityIndex(t *testing.T, g *game.Game, id uuid.UUID, fragment string) int {
	t.Helper()
	c := layeredCard(t, g, id)
	for i, a := range game.ActivatedAbilitiesForCard(c) {
		if strings.Contains(a.Label, fragment) {
			return i
		}
	}
	return -1
}

// --- Bioengineered Future -----------------------------------------------------

func TestStdB3BioengineeredFutureCountsLandsThatEnteredThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	base := g.TurnTallyFor(me.ID).LandsEntered
	b31PlayLand(t, g, "Forest", "Basic Land — Forest", "")
	want := g.TurnTallyFor(me.ID).LandsEntered
	if want != base+1 {
		t.Fatalf("playing a land moved the tally %d -> %d", base, want)
	}
	avB1Cast(t, g, me, game.Card{Name: "Bioengineered Future", TypeLine: "Enchantment", ManaCost: "{1}{G}{G}", OracleID: stdB3FutureOracle}, game.CastSpellParams{})
	stdB3Resolve(t, g)
	if got := eoeCDCount(g, me.ID, "Lander"); got != 1 {
		t.Errorf("%d Landers, want 1", got)
	}
	// #5213: the Lander is an artifact, not a land, so its entering
	// is not counted.
	if got := g.TurnTallyFor(me.ID).LandsEntered; got != want {
		t.Errorf("the Lander token moved the lands-entered tally %d -> %d", want, got)
	}
	bear := avB1Cast(t, g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}", Power: 2, Toughness: 2}, game.CastSpellParams{})
	stdB3Resolve(t, g)
	if got := countersOn(g, bear, game.CounterPlusOne); got != want {
		t.Errorf("a land entered before the enchantment did: %d counters, want %d (the Forest only, not the Lander)", got, want)
	}
	opp := g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
	if got := countersOn(g, theirs, game.CounterPlusOne); got != 0 {
		t.Errorf("only your creatures: %d", got)
	}
}

func TestStdB3BioengineeredFutureGivesNothingWithoutALandThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	base := g.TurnTallyFor(me.ID).LandsEntered
	avB1Cast(t, g, me, game.Card{Name: "Bioengineered Future", TypeLine: "Enchantment", ManaCost: "{1}{G}{G}", OracleID: stdB3FutureOracle}, game.CastSpellParams{})
	stdB3Resolve(t, g)
	if got := eoeCDCount(g, me.ID, "Lander"); got != 1 {
		t.Fatalf("%d Landers, want 1", got)
	}
	bear := avB1Cast(t, g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}", Power: 2, Toughness: 2}, game.CastSpellParams{})
	stdB3Resolve(t, g)
	// #5213: its own Lander is not a land, so it adds no counter.
	if got := g.TurnTallyFor(me.ID).LandsEntered; got != base {
		t.Errorf("the Lander token moved the lands-entered tally %d -> %d", base, got)
	}
	if got := countersOn(g, bear, game.CounterPlusOne); got != base {
		t.Errorf("one counter per land entered this turn: %d counters, want %d (the Lander adds none)", got, base)
	}
}

// --- Close Encounter ----------------------------------------------------------

func TestStdB3CloseEncounterDealsTheChosenCreaturesPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := b12Creature(g, me.ID, "Ogre", "Creature — Ogre", 4, 4)
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Giant", 1, 6)
	avB1Cast(t, g, me, game.Card{Name: "Close Encounter", TypeLine: "Instant", ManaCost: "{1}{G}", OracleID: stdB3EncounterOracle},
		game.CastSpellParams{Targets: cardRefs(victim), RevealIDs: []uuid.UUID{ogre}})
	// The creature is pumped in response: its power as the spell resolves counts.
	_ = g.AddCounter(ogre, game.CounterPlusOne, 1)
	stdB3Resolve(t, g)
	if c, ok := battlefieldCard(g, victim); !ok || c.DamageMarked != 5 {
		t.Errorf("the victim has %d damage, want 5", c.DamageMarked)
	}
	if !g.Battlefield.Contains(ogre) {
		t.Error("nothing happens to the chosen creature")
	}
}

func TestStdB3CloseEncounterUsesTheLastKnownPowerOfAChosenCreatureThatLeft(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := b12Creature(g, me.ID, "Ogre", "Creature — Ogre", 4, 4)
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Giant", 1, 9)
	avB1Cast(t, g, me, game.Card{Name: "Close Encounter", TypeLine: "Instant", ManaCost: "{1}{G}", OracleID: stdB3EncounterOracle},
		game.CastSpellParams{Targets: cardRefs(victim), RevealIDs: []uuid.UUID{ogre}})
	b15Destroy(t, g, ogre)
	stdB3Resolve(t, g)
	if c, ok := battlefieldCard(g, victim); !ok || c.DamageMarked != 4 {
		t.Errorf("the victim has %d damage, want 4", c.DamageMarked)
	}
}

func TestStdB3CloseEncounterNeedsAChosenCreatureOfYours(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 7, 7)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	advanceToMain(t, g)
	for name, params := range map[string]game.CastSpellParams{
		"nothing chosen":          {Targets: cardRefs(theirs)},
		"an opponent's creature":  {Targets: cardRefs(mine), RevealIDs: []uuid.UUID{theirs}},
		"a noncreature permanent": {Targets: cardRefs(theirs), RevealIDs: []uuid.UUID{rock}},
	} {
		params.Strict = true
		id := avB1Hand(me, game.Card{Name: "Close Encounter", TypeLine: "Instant", ManaCost: "{1}{G}", OracleID: stdB3EncounterOracle})
		b06AddMana(me, "G", "G")
		if err := g.CastSpell(me.ID, id, params); err == nil {
			t.Errorf("%s: the cast was accepted", name)
		}
	}
}

func TestStdB3CloseEncounterCanChooseAWarpedCreatureCardInExile(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Giant", 1, 9)
	beckoner := avB1Cast(t, g, me, game.Card{Name: "Perigee Beckoner", TypeLine: "Creature — Horror", ManaCost: "{4}{B}",
		OracleID: stdB3BeckonerOracle, Power: 4, Toughness: 5}, game.CastSpellParams{AlternativeCost: "warp"})
	stdB3Resolve(t, g)
	// Not yet exiled: it is a creature on the battlefield, not a warped card.
	advanceThroughEndStep(t, g)
	if !g.Exile.Contains(beckoner) {
		t.Fatal("the warped Beckoner is in exile")
	}
	var warped []game.Card
	g.WithWriteLock(func() { warped = g.WarpedCardsOwnedInExileForEffect(me.ID) })
	if len(warped) != 1 || warped[0].InstanceID != beckoner {
		t.Fatalf("warped cards in exile: %+v", warped)
	}
	id := avB1Hand(me, game.Card{Name: "Close Encounter", TypeLine: "Instant", ManaCost: "{1}{G}", OracleID: stdB3EncounterOracle})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: cardRefs(victim), RevealIDs: []uuid.UUID{beckoner}}); err != nil {
		t.Fatalf("choosing a warped card: %v", err)
	}
	stdB3Resolve(t, g)
	if c, ok := battlefieldCard(g, victim); !ok || c.DamageMarked != 4 {
		t.Errorf("the victim has %d damage, want 4", c.DamageMarked)
	}
	if !g.Exile.Contains(beckoner) {
		t.Error("the warped card stays in exile")
	}
}

func TestStdB3CloseEncounterRefusesAnExiledCardThatWasNotWarped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Giant", 1, 9)
	exiled := b12Creature(g, me.ID, "Exiled Bear", "Creature — Bear", 3, 3)
	stdB3Exile(t, g, exiled)
	advanceToMain(t, g)
	id := avB1Hand(me, game.Card{Name: "Close Encounter", TypeLine: "Instant", ManaCost: "{1}{G}", OracleID: stdB3EncounterOracle})
	b06AddMana(me, "G", "G")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: cardRefs(victim), RevealIDs: []uuid.UUID{exiled}}); err == nil {
		t.Error("a card in exile that was not exiled by warp can't be chosen")
	}
}

// --- Drix Fatemaker -----------------------------------------------------------

func TestStdB3DrixFatemakerGivesTrampleToCreaturesWithCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	fate := avB1Cast(t, g, me, game.Card{Name: "Drix Fatemaker", TypeLine: "Creature — Drix Wizard", ManaCost: "{3}{G}",
		OracleID: stdB3FatemakerOracle, Power: 3, Toughness: 2}, game.CastSpellParams{})
	stdB3Resolve(t, g)
	pickIfAsked(t, g, me.ID, bear)
	stdB3Resolve(t, g)
	if got := countersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the Bear has %d counters, want 1", got)
	}
	if !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("a creature you control with a +1/+1 counter has trample")
	}
	if hasEffectiveKeyword(t, g, other, "trample") || hasEffectiveKeyword(t, g, fate, "trample") {
		t.Error("creatures without a counter do not")
	}
	_ = g.AddCounter(fate, game.CounterPlusOne, 1)
	if !hasEffectiveKeyword(t, g, fate, "trample") {
		t.Error("the Fatemaker itself, once it has a counter")
	}
	theirs := b12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 2, 2)
	_ = g.AddCounter(theirs, game.CounterPlusOne, 1)
	if hasEffectiveKeyword(t, g, theirs, "trample") {
		t.Error("not an opponent's creature")
	}
}

// --- Meltstrider's Resolve ----------------------------------------------------

func TestStdB3MeltstridersResolveFightsAndGrowsToughness(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 3, 3)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 5)
	avB1Cast(t, g, me, game.Card{Name: "Meltstrider's Resolve", TypeLine: "Enchantment — Aura", ManaCost: "{G}", OracleID: stdB3ResolveOracle},
		game.CastSpellParams{Targets: cardRefs(mine)})
	stdB3Resolve(t, g)
	pickIfAsked(t, g, me.ID, theirs)
	stdB3Resolve(t, g)
	if c, ok := battlefieldCard(g, theirs); !ok || c.DamageMarked != 3 {
		t.Errorf("the enchanted creature dealt %d to the target, want 3", c.DamageMarked)
	}
	if c, ok := battlefieldCard(g, mine); !ok || c.DamageMarked != 2 {
		t.Errorf("the target dealt %d back, want 2", c.DamageMarked)
	}
	if got := layeredCard(t, g, mine); got.CurrentToughness() != 5 || got.CurrentPower() != 3 {
		t.Errorf("enchanted creature is %d/%d, want 3/5", got.CurrentPower(), got.CurrentToughness())
	}
}

func TestStdB3MeltstridersResolveWorksWithNoTargetAndEnchantsOnlyYourCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 3, 3)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 5)
	advanceToMain(t, g)
	id := avB1Hand(me, game.Card{Name: "Meltstrider's Resolve", TypeLine: "Enchantment — Aura", ManaCost: "{G}", OracleID: stdB3ResolveOracle})
	b06AddMana(me, "G")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: cardRefs(theirs)}); err == nil {
		t.Error("enchant creature YOU control")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: cardRefs(mine)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	stdB3Resolve(t, g)
	pickIfAsked(t, g, me.ID)
	stdB3Resolve(t, g)
	if c, _ := battlefieldCard(g, mine); c.DamageMarked != 0 {
		t.Errorf("up to one target: with none nothing fights (%d damage)", c.DamageMarked)
	}
}

// --- Tapestry Warden ----------------------------------------------------------

func TestStdB3TapestryWardenStationsAndAssignsByToughness(t *testing.T) {
	g, me, ship, crew := stationBoard(t, "Debris Field Crusher", "Artifact — Spacecraft", stdB3CrusherOracle, 1)
	warden := b12Push(g, me.ID, "Tapestry Warden", "Artifact Creature — Robot Soldier", stdB3WardenOracle, 3, 4)
	if !hasEffectiveKeyword(t, g, warden, "vigilance") {
		t.Error("Tapestry Warden has vigilance")
	}
	hand := layeredCard(t, g, crew)
	if hand.CurrentPower() != 1 {
		t.Errorf("the crew's power is untouched: %d", hand.CurrentPower())
	}
	if !hand.Effective().DamageByToughness || !layeredCard(t, g, warden).Effective().DamageByToughness {
		t.Error("creatures you control assign combat damage by toughness when it is greater")
	}
	station(t, g, me, ship, crew)
	stdB3Resolve(t, g)
	stdB3Resolve(t, g)
	if got := layeredCard(t, g, ship).Counters[game.CounterCharge]; got != 3 {
		t.Errorf("a 1/3 stations for its toughness: %d charge counters, want 3", got)
	}
}

func TestStdB3TapestryWardenLeavesBiggerPowerAlone(t *testing.T) {
	g, me, ship, crew := stationBoard(t, "Debris Field Crusher", "Artifact — Spacecraft", stdB3CrusherOracle, 5)
	b12Push(g, me.ID, "Tapestry Warden", "Artifact Creature — Robot Soldier", stdB3WardenOracle, 3, 4)
	station(t, g, me, ship, crew)
	stdB3Resolve(t, g)
	stdB3Resolve(t, g)
	if got := layeredCard(t, g, ship).Counters[game.CounterCharge]; got != 5 {
		t.Errorf("a 5/3 stations for its power: %d charge counters, want 5", got)
	}
}

func TestStdB3TapestryWardenAlsoStationsAPlanet(t *testing.T) {
	g, me, land, crew := stationBoard(t, "Adagia, Windswept Bastion", "Land — Planet", stdB3AdagiaOracle, 1)
	b12Push(g, me.ID, "Tapestry Warden", "Artifact Creature — Robot Soldier", stdB3WardenOracle, 3, 4)
	station(t, g, me, land, crew)
	stdB3Resolve(t, g)
	stdB3Resolve(t, g)
	if got := layeredCard(t, g, land).Counters[game.CounterCharge]; got != 3 {
		t.Errorf("a Planet stationed by a 1/3 gets %d charge counters, want 3", got)
	}
}

func TestStdB3TapestryWardenIsOnlyForYourCreatures(t *testing.T) {
	g, me, ship, _ := stationBoard(t, "Debris Field Crusher", "Artifact — Spacecraft", stdB3CrusherOracle, 1)
	opp := g.Seats[1]
	b12Push(g, opp.ID, "Tapestry Warden", "Artifact Creature — Robot Soldier", stdB3WardenOracle, 3, 4)
	station(t, g, me, ship, avB1Named(g, me.ID, "Station Hand"))
	stdB3Resolve(t, g)
	stdB3Resolve(t, g)
	if got := layeredCard(t, g, ship).Counters[game.CounterCharge]; got != 1 {
		t.Errorf("an opponent's Warden does nothing for you: %d charge counters, want 1", got)
	}
}

// --- Mutinous Massacre --------------------------------------------------------

func TestStdB3MutinousMassacreDestroysTheChosenParityThenStealsEverything(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mk := func(owner uuid.UUID, name, cost string) uuid.UUID {
		id := b12Creature(g, owner, name, "Creature — Bear", 2, 2)
		eoeCDSet(g, id, func(c *game.Card) { c.ManaCost = cost })
		return id
	}
	odd := mk(opp.ID, "Odd", "{1}")
	even := mk(opp.ID, "Even", "{2}")
	zero := mk(opp.ID, "Zero", "")
	xcost := mk(opp.ID, "X", "{X}{1}")
	myOdd := mk(me.ID, "MyOdd", "{3}")
	myEven := mk(me.ID, "MyEven", "{4}")
	eoeCDSet(g, myEven, func(c *game.Card) { c.Tapped = true })
	avB1Cast(t, g, me, game.Card{Name: "Mutinous Massacre", TypeLine: "Sorcery", ManaCost: "{3}{B}{B}{R}{R}", OracleID: stdB3MassacreOracle}, game.CastSpellParams{})
	stdB3Resolve(t, g)
	answerOptionPick(t, g, me.ID, 0) // odd
	stdB3Resolve(t, g)
	for name, id := range map[string]uuid.UUID{"odd": odd, "x": xcost, "my odd": myOdd} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s (odd mana value) survived", name)
		}
	}
	for name, id := range map[string]uuid.UUID{"even": even, "zero": zero, "my even": myEven} {
		c := layeredCard(t, g, id)
		if c.Controller != me.ID {
			t.Errorf("%s is not under your control", name)
		}
		if c.Tapped {
			t.Errorf("%s was not untapped", name)
		}
		if !game.HasKeyword(&c, "haste") {
			t.Errorf("%s has no haste", name)
		}
	}
	advanceThroughEndStep(t, g)
	advanceTo(t, g, game.StepCleanup)
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatal(err)
	}
	if got := layeredCard(t, g, even); got.Controller != opp.ID {
		t.Error("control reverts at the end of the turn")
	}
}

// --- Ragost, Deft Gastronaut --------------------------------------------------

func TestStdB3RagostMakesArtifactsFoodsThatGainLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ragost := b12Push(g, me.ID, "Ragost, Deft Gastronaut", "Legendary Creature — Lobster Citizen", stdB3RagostOracle, 2, 2)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	theirs := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	enchant := b12Permanent(g, me.ID, "Ward", "Enchantment")
	if !layeredCard(t, g, rock).HasSubtype("Food") {
		t.Error("artifacts you control are Foods")
	}
	if layeredCard(t, g, theirs).HasSubtype("Food") || layeredCard(t, g, enchant).HasSubtype("Food") {
		t.Error("only your artifacts")
	}
	if !layeredCard(t, g, rock).IsArtifact() {
		t.Error("they keep their other types")
	}
	idx := stdB3AbilityIndex(t, g, rock, "gain 3 life")
	if idx < 0 {
		t.Fatal("the artifact has the Food ability")
	}
	advanceToMain(t, g)
	b06AddMana(me, "C", "C")
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, rock, idx, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	stdB3Resolve(t, g)
	if me.Life != life+3 || g.Battlefield.Contains(rock) {
		t.Errorf("life %d (want %d), rock still there: %v", me.Life, life+3, g.Battlefield.Contains(rock))
	}
	// Ragost's own ability sacrifices a Food for 3 damage to each opponent.
	food := b12Permanent(g, me.ID, "Food Rock", "Artifact")
	b06AddMana(me, "C")
	oppLife := opp.Life
	ridx := stdB3AbilityIndex(t, g, ragost, "3 damage to each opponent")
	if ridx < 0 {
		t.Fatal("Ragost has his ability")
	}
	if err := g.ActivateCatalogAbility(me.ID, ragost, ridx, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{food}}); err != nil {
		t.Fatalf("Ragost activate: %v", err)
	}
	stdB3Resolve(t, g)
	if opp.Life != oppLife-3 {
		t.Errorf("opponent life %d, want %d", opp.Life, oppLife-3)
	}
	if g.Battlefield.Contains(food) {
		t.Error("the Food was sacrificed")
	}
}

func TestStdB3RagostUntapsAtAnEndStepIfYouGainedLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ragost := b12Push(g, me.ID, "Ragost, Deft Gastronaut", "Legendary Creature — Lobster Citizen", stdB3RagostOracle, 2, 2)
	eoeCDSet(g, ragost, func(c *game.Card) { c.Tapped = true })
	advanceThroughEndStep(t, g)
	if c, _ := battlefieldCard(g, ragost); !c.Tapped {
		t.Error("no life gained this turn: Ragost stays tapped")
	}
	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	ragost2 := b12Push(g2, me2.ID, "Ragost, Deft Gastronaut", "Legendary Creature — Lobster Citizen", stdB3RagostOracle, 2, 2)
	eoeCDSet(g2, ragost2, func(c *game.Card) { c.Tapped = true })
	g2.WithWriteLock(func() { _ = g2.ChangePlayerLifeForEffect(uuid.Nil, me2.ID, 2) })
	advanceThroughEndStep(t, g2)
	if c, _ := battlefieldCard(g2, ragost2); c.Tapped {
		t.Error("you gained life this turn: Ragost untaps at the end step")
	}
}

// --- Station Monitor ----------------------------------------------------------

func TestStdB3StationMonitorMakesADroneOnTheSecondSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Station Monitor", "Creature — Lizard Artificer", stdB3MonitorOracle, 2, 2)
	avB1Cast(t, g, me, game.Card{Name: "First", TypeLine: "Sorcery", ManaCost: "{1}"}, game.CastSpellParams{})
	stdB3Resolve(t, g)
	if got := eoeCDCount(g, me.ID, "Drone"); got != 0 {
		t.Errorf("the first spell makes %d Drones, want 0", got)
	}
	avB1Cast(t, g, me, game.Card{Name: "Second", TypeLine: "Sorcery", ManaCost: "{1}"}, game.CastSpellParams{})
	stdB3Resolve(t, g)
	drone := avB1Named(g, me.ID, "Drone")
	if drone == uuid.Nil {
		t.Fatal("the second spell makes a Drone")
	}
	d := layeredCard(t, g, drone)
	if d.CurrentPower() != 1 || d.CurrentToughness() != 1 || !d.IsArtifact() || !game.HasKeyword(&d, "flying") {
		t.Errorf("the Drone is a %d/%d flying artifact creature", d.CurrentPower(), d.CurrentToughness())
	}
	avB1Cast(t, g, me, game.Card{Name: "Third", TypeLine: "Sorcery", ManaCost: "{1}"}, game.CastSpellParams{})
	stdB3Resolve(t, g)
	if got := eoeCDCount(g, me.ID, "Drone"); got != 1 {
		t.Errorf("the third spell makes no more: %d Drones", got)
	}
}

// --- Syr Vondam, Sunstar Exemplar ---------------------------------------------

func TestStdB3SyrVondamGrowsAndGainsLifeWhenYourOtherCreaturesDieOrAreExiled(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	syr := b12Push(g, me.ID, "Syr Vondam, Sunstar Exemplar", "Legendary Creature — Human Knight", stdB3VondamOracle, 2, 2)
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 1, 1)
	b := b12Creature(g, me.ID, "B", "Creature — Bear", 1, 1)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
	life := me.Life
	b15Destroy(t, g, a)
	stdB3Resolve(t, g)
	if got := countersOn(g, syr, game.CounterPlusOne); got != 1 || me.Life != life+1 {
		t.Errorf("after a death: %d counters, life %d (want 1, %d)", got, me.Life, life+1)
	}
	stdB3Exile(t, g, b)
	stdB3Resolve(t, g)
	if got := countersOn(g, syr, game.CounterPlusOne); got != 2 || me.Life != life+2 {
		t.Errorf("after an exile: %d counters, life %d (want 2, %d)", got, me.Life, life+2)
	}
	b15Destroy(t, g, theirs)
	stdB3Resolve(t, g)
	if got := countersOn(g, syr, game.CounterPlusOne); got != 2 {
		t.Errorf("an opponent's creature does not count: %d", got)
	}
	hand := avB1Hand(me, game.Card{Name: "Bounced", TypeLine: "Creature — Bear", Power: 1, Toughness: 1})
	_ = hand
	bounce := b12Creature(g, me.ID, "Bounce", "Creature — Bear", 1, 1)
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(bounce) })
	stdB3Resolve(t, g)
	if got := countersOn(g, syr, game.CounterPlusOne); got != 2 {
		t.Errorf("a bounce is neither death nor exile: %d", got)
	}
}

func TestStdB3SyrVondamDestroysOnlyWhenHeLeavesWithPowerFourOrMore(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	syr := b12Push(g, me.ID, "Syr Vondam, Sunstar Exemplar", "Legendary Creature — Human Knight", stdB3VondamOracle, 2, 2)
	victim := b12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	land := pushLandFor(g, opp.ID, "Swamp", "Basic Land — Swamp")
	b15Destroy(t, g, syr)
	stdB3Resolve(t, g)
	pickIfAsked(t, g, me.ID, victim)
	stdB3Resolve(t, g)
	if !g.Battlefield.Contains(victim) {
		t.Error("at power 2 nothing triggers")
	}

	g2 := newCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	syr2 := b12Push(g2, me2.ID, "Syr Vondam, Sunstar Exemplar", "Legendary Creature — Human Knight", stdB3VondamOracle, 2, 2)
	eoeCDSet(g2, syr2, func(c *game.Card) { c.Counters = map[string]int{game.CounterPlusOne: 2} })
	victim2 := b12Permanent(g2, opp2.ID, "Their Enchantment", "Enchantment")
	land2 := pushLandFor(g2, opp2.ID, "Swamp", "Basic Land — Swamp")
	b15Destroy(t, g2, syr2)
	stdB3Resolve(t, g2)
	if p := latestPickTarget(g2, me2.ID); p != nil && hasID(p.PickTargetCards, land2) {
		t.Error("a land is not a legal target")
	}
	pickIfAsked(t, g2, me2.ID, victim2)
	stdB3Resolve(t, g2)
	if g2.Battlefield.Contains(victim2) {
		t.Error("at power 4 the nonland permanent is destroyed")
	}
	_ = land
}

// --- Dauntless Scrapbot -------------------------------------------------------

func TestStdB3DauntlessScrapbotExilesOpponentGraveyardsAndMakesALander(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, opp2 := g.Seats[0], g.Seats[1], g.Seats[2]
	mine := avB1Hand(me, game.Card{Name: "My Dead", TypeLine: "Creature — Bear"})
	theirs := avB1Hand(opp, game.Card{Name: "Their Dead", TypeLine: "Creature — Bear"})
	theirs2 := avB1Hand(opp2, game.Card{Name: "Their Dead Too", TypeLine: "Creature — Bear"})
	for id, owner := range map[uuid.UUID]*game.Player{mine: me, theirs: opp, theirs2: opp2} {
		if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: owner.ID}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: owner.ID}, id); err != nil {
			t.Fatal(err)
		}
	}
	avB1Cast(t, g, me, game.Card{Name: "Dauntless Scrapbot", TypeLine: "Artifact Creature — Robot", ManaCost: "{3}",
		OracleID: stdB3ScrapbotOracle, Power: 3, Toughness: 1}, game.CastSpellParams{})
	stdB3Resolve(t, g)
	if !me.Graveyard.Contains(mine) {
		t.Error("your own graveyard stays")
	}
	if opp.Graveyard.Contains(theirs) || opp2.Graveyard.Contains(theirs2) || !g.Exile.Contains(theirs) || !g.Exile.Contains(theirs2) {
		t.Error("each opponent's graveyard is exiled")
	}
	if got := eoeCDCount(g, me.ID, "Lander"); got != 1 {
		t.Errorf("%d Landers, want 1", got)
	}
}
