package testsg

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gmrCreature casts a vanilla creature of `typeLine` and `cost` from
// hand under strict mana and resolves it.
func gmrCreature(t *testing.T, g *game.Game, p *game.Player, typeLine, cost string) uuid.UUID {
	t.Helper()
	id := effectstest.CastFromHandForTest(t, g, p, "Test Creature", typeLine, cost, "", effectstest.Strict)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatalf("the %s did not reach the battlefield", typeLine)
	}
	return id
}

// --- sunburst --------------------------------------------------------

// Every printed sunburst card enters with one counter per colour spent:
// +1/+1 for a creature, charge for a noncreature (CR 702.44a).
func TestPrintedSunburstCardsEnterWithACounterPerColour(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, cost, oracle, pay, kind string
		x                                       int
	}{
		{"Suntouched Myr", "Artifact Creature — Myr", "{3}", effectstest.GmrSuntouchedMyrOracle, "WUB", game.CounterPlusOne, 0},
		{"Skyreach Manta", "Artifact Creature — Fish", "{5}", effectstest.GmrSkyreachMantaOracle, "WUBRG", game.CounterPlusOne, 0},
		{"Pentad Prism", "Artifact", "{2}", effectstest.GmrPentadPrismOracle, "WU", game.CounterCharge, 0},
		{"Clearwater Goblet", "Artifact", "{5}", effectstest.GmrClearwaterGobletOracle, "WUBRG", game.CounterCharge, 0},
		{"Opaline Bracers", "Artifact — Equipment", "{4}", effectstest.GmrOpalineBracersOracle, "WUBR", game.CounterCharge, 0},
		{"Sawtooth Thresher", "Artifact Creature — Construct", "{6}", effectstest.GmrSawtoothThresherOracle, "WUBRGC", game.CounterPlusOne, 0},
		{"Lunar Avenger", "Artifact Creature — Golem", "{7}", effectstest.GmrLunarAvengerOracle, "WUBRGCC", game.CounterPlusOne, 0},
		{"Heliophial", "Artifact", "{5}", effectstest.GmrHeliophialOracle, "WUBRC", game.CounterCharge, 0},
		{"Suncrusher", "Artifact Creature — Construct", "{9}", effectstest.GmrSuncrusherOracle, "WUBRGCCCC", game.CounterPlusOne, 0},
		{"Solarion", "Artifact Creature — Construct", "{7}", effectstest.GmrSolarionOracle, "WUBCCCC", game.CounterPlusOne, 0},
		{"Baton of Courage", "Artifact", "{3}", effectstest.GmrBatonOfCourageOracle, "WUB", game.CounterCharge, 0},
		{"Infused Arrows", "Artifact", "{4}", effectstest.GmrInfusedArrowsOracle, "WUBR", game.CounterCharge, 0},
		{"Spinal Parasite", "Artifact Creature — Insect", "{5}", effectstest.GmrSpinalParasiteOracle, "WUBRG", game.CounterPlusOne, 0},
		{"Arcbound Wanderer", "Artifact Creature — Golem", "{6}", effectstest.GmrArcboundWandererOracle, "WUBRGC", game.CounterPlusOne, 0},
		{"Engineered Explosives", "Artifact", "{X}", effectstest.GmrEngineeredExplosives, "UR", game.CounterCharge, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.AdvanceToMain(t, g)
			effectstest.FloatForTest(g, me, tc.pay)
			want := 0
			for _, r := range tc.pay {
				if r != 'C' {
					want++
				}
			}
			params := effectstest.Strict
			params.XValue = tc.x
			id := effectstest.CastFromHandForTest(t, g, me, tc.name, tc.typeLine, tc.cost, tc.oracle, params)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.CounterCount(g, id, tc.kind); got != want {
				t.Errorf("%s counters = %d, want %d — one per colour spent", tc.kind, got, want)
			}
		})
	}
}

// Engineered Explosives with two charge counters destroys each nonland
// permanent with mana value 2, and nothing else.
func TestEngineeredExplosivesDestroysTheManaValueOfItsCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	two := effectstest.B12Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	three := effectstest.B12Push(g, opp.ID, "Ogre", "Creature — Ogre", "", 3, 3)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			switch g.Battlefield.Cards[i].InstanceID {
			case two:
				g.Battlefield.Cards[i].ManaCost = "{1}{G}"
			case three:
				g.Battlefield.Cards[i].ManaCost = "{2}{R}"
			}
		}
	})
	effectstest.FloatForTest(g, me, "UR")
	params := effectstest.Strict
	params.XValue = 2
	ee := effectstest.CastFromHandForTest(t, g, me, "Engineered Explosives", "Artifact", "{X}", effectstest.GmrEngineeredExplosives, params)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FloatForTest(g, me, "CC")
	effectstest.B16Activate(t, g, me.ID, ee, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(two) {
		t.Error("the mana value 2 Bear survived two charge counters")
	}
	if !g.Battlefield.Contains(three) {
		t.Error("the mana value 3 Ogre was destroyed")
	}
}

// Pentad Prism cashes its charge counters for mana of any colour.
func TestPentadPrismRemovesAChargeCounterForMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "WU")
	prism := effectstest.CastFromHandForTest(t, g, me, "Pentad Prism", "Artifact", "{2}", effectstest.GmrPentadPrismOracle, effectstest.Strict)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ActivateManaFor(t, g, me.ID, prism, 0, game.ManaAbilityParams{Colors: []string{"G"}})
	if got := effectstest.CounterCount(g, prism, game.CounterCharge); got != 1 {
		t.Errorf("charge counters = %d after one activation, want 1", got)
	}
	if got := effectstest.GmrPoolCount(me, "G"); got != 1 {
		t.Errorf("pool has %d {G}, want 1", got)
	}
}

// Heliophial's damage is its last-known charge counters, after the
// sacrifice paid for it.
func TestHeliophialDealsItsCountersAfterTheSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "WUBCC")
	phial := effectstest.CastFromHandForTest(t, g, me, "Heliophial", "Artifact", "{5}", effectstest.GmrHeliophialOracle, effectstest.Strict)
	effectstest.PassPriorityAroundTable(t, g)
	life := opp.Life
	effectstest.FloatForTest(g, me, "CC")
	effectstest.B16Activate(t, g, me.ID, phial, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 3 {
		t.Errorf("Heliophial dealt %d, want 3 — its three charge counters", got)
	}
}

// Infused Arrows' X is the number of counters removed.
func TestInfusedArrowsGivesMinusXForTheCountersRemoved(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	ogre := effectstest.B12Push(g, opp.ID, "Ogre", "Creature — Ogre", "", 3, 3)
	effectstest.FloatForTest(g, me, "WUBR")
	arrows := effectstest.CastFromHandForTest(t, g, me, "Infused Arrows", "Artifact", "{4}", effectstest.GmrInfusedArrowsOracle, effectstest.Strict)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B16Activate(t, g, me.ID, arrows, 0, game.ActivateAbilityParams{
		Strict:           true,
		Targets:          []game.TargetRef{{Kind: game.TargetCard, ID: ogre}},
		CounterSourceIDs: []uuid.UUID{arrows},
		CounterCounts:    []int{3},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) {
		t.Error("a 3/3 survived -3/-3")
	}
	if got := effectstest.CounterCount(g, arrows, game.CounterCharge); got != 1 {
		t.Errorf("charge counters left = %d, want 1", got)
	}
}

// Spinal Parasite removes one counter from a permanent that has one
// kind, without asking.
func TestSpinalParasiteRemovesACounterFromTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	rock := effectstest.B12Push(g, opp.ID, "Charged Rock", "Artifact", "", 0, 0)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(rock, game.CounterCharge, 3) })
	effectstest.FloatForTest(g, me, "WUBRG")
	parasite := effectstest.CastFromHandForTest(t, g, me, "Spinal Parasite", "Artifact Creature — Insect", "{5}", effectstest.GmrSpinalParasiteOracle, effectstest.Strict)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B16Activate(t, g, me.ID, parasite, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, rock, game.CounterCharge); got != 2 {
		t.Errorf("charge counters on the target = %d, want 2", got)
	}
	if got := effectstest.CounterCount(g, parasite, game.CounterPlusOne); got != 3 {
		t.Errorf("+1/+1 counters left on the Parasite = %d, want 3", got)
	}
}

// Modular: when Arcbound Wanderer dies, its +1/+1 counters may go onto
// a target artifact creature (CR 702.43a), counted from its last-known
// information.
func TestArcboundWandererMovesItsCountersWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	golem := effectstest.B12Push(g, me.ID, "Golem", "Artifact Creature — Golem", "", 3, 3)
	effectstest.FloatForTest(g, me, "WUBRCC")
	wanderer := effectstest.CastFromHandForTest(t, g, me, "Arcbound Wanderer", "Artifact Creature — Golem", "{6}", effectstest.GmrArcboundWandererOracle, effectstest.Strict)
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(wanderer); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
		effectstest.AnswerTriggerTargets(t, g, me.ID, golem)
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, golem, game.CounterPlusOne); got != 4 {
		t.Errorf("+1/+1 counters on the Golem = %d, want the Wanderer's 4", got)
	}
}

// --- granted sunburst ----------------------------------------------

// Lux Artillery gives an artifact creature spell sunburst; an Etched
// Oracle cast under it has two instances and counts the colours twice
// (CR 702.44d). A non-artifact creature is not given it.
func TestLuxArtilleryGivesArtifactCreatureSpellsSunburst(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		want                   int
	}{
		{"an artifact creature", "Artifact Creature — Golem", "", 3},
		{"Etched Oracle, which prints it too", "Artifact Creature — Wizard", effectstest.EtchedOracleOracle, 6},
		{"a creature that is not an artifact", "Creature — Bear", "", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.AdvanceToMain(t, g)
			effectstest.SeedPermanentWithOracle(g, me.ID, "Lux Artillery", "Artifact", effectstest.GmrLuxArtilleryOracle)
			effectstest.FloatForTest(g, me, "WUB")
			id := effectstest.CastFromHandForTest(t, g, me, "Spell", tc.typeLine, "{3}", tc.oracle, effectstest.Strict)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.CounterCount(g, id, game.CounterPlusOne); got != tc.want {
				t.Errorf("+1/+1 counters = %d, want %d", got, tc.want)
			}
		})
	}
}

// Solar Array's mana makes the next artifact spell this turn gain
// sunburst — including the spell that mana paid for.
func TestSolarArrayGivesTheNextArtifactSpellSunburst(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	array := effectstest.SeedPermanentWithOracle(g, me.ID, "Solar Array", "Artifact", effectstest.GmrSolarArrayOracle)
	effectstest.ActivateManaFor(t, g, me.ID, array, 0, game.ManaAbilityParams{Colors: []string{"W"}})
	effectstest.FloatForTest(g, me, "UB")
	golem := effectstest.CastFromHandForTest(t, g, me, "Golem", "Artifact Creature — Golem", "{3}", "", effectstest.Strict)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, golem, game.CounterPlusOne); got != 3 {
		t.Errorf("+1/+1 counters = %d, want 3 — W from the Array, U and B", got)
	}
	// The trigger was spent: the next artifact this turn gets nothing.
	effectstest.FloatForTest(g, me, "WUB")
	second := effectstest.CastFromHandForTest(t, g, me, "Second Golem", "Artifact Creature — Golem", "{3}", "", effectstest.Strict)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, second, game.CounterPlusOne); got != 0 {
		t.Errorf("the second artifact got %d counters, want 0", got)
	}
}

// --- another permanent reads the entering spell's payment -----------

// Coin of Mastery counts the mana from artifact sources — a Treasure's
// included — and Kalain counts only the Treasure mana, on OTHER
// creatures.
func TestCoinOfMasteryAndKalainCountTheEnteringSpellsMana(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine string
		kinds                  game.ManaSourceKinds
		want                   int
	}{
		{"Coin of Mastery, Treasure mana", effectstest.GmrCoinOfMasteryOracle, "Artifact", effectstest.GmrHalfTreasureSourceKinds, 2},
		{"Coin of Mastery, a mana rock's mana", effectstest.GmrCoinOfMasteryOracle, "Artifact", game.ManaSourceArtifact, 2},
		{"Coin of Mastery, land mana", effectstest.GmrCoinOfMasteryOracle, "Artifact", game.ManaSourceLand, 0},
		{"Kalain, Treasure mana", effectstest.GmrKalainOracle, "Legendary Creature — Human Elf Bard", effectstest.GmrHalfTreasureSourceKinds, 2},
		{"Kalain, a mana rock's mana", effectstest.GmrKalainOracle, "Legendary Creature — Human Elf Bard", game.ManaSourceArtifact, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.AdvanceToMain(t, g)
			effectstest.SeedPermanentWithOracle(g, me.ID, "Reader", tc.typeLine, tc.oracle)
			effectstest.GmrFloatFrom(g, me, "RR", tc.kinds)
			effectstest.FloatForTest(g, me, "G")
			bear := gmrCreature(t, g, me, "Creature — Bear", "{2}{G}")
			if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != tc.want {
				t.Errorf("+1/+1 counters = %d, want %d", got, tc.want)
			}
		})
	}
}

// --- "no mana was spent" --------------------------------------------

// Satoru draws for a creature that was not cast, and for one cast
// without spending mana; it does not draw for one cast with mana.
func TestSatoruReadsWhetherManaWasSpent(t *testing.T) {
	for _, tc := range []struct {
		name string
		how  string
		draw int
	}{
		{"reanimated", "reanimate", 1},
		{"cast for {0}", "free", 1},
		{"cast with mana", "paid", 0},
		// CR 603.4 over the whole set: a creature reanimated in the same
		// event batch as one cast with mana does not draw.
		{"reanimated alongside one cast with mana", "mixed", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.AdvanceToMain(t, g)
			effectstest.SeedPermanentWithOracle(g, me.ID, "Satoru, the Infiltrator", "Legendary Creature — Human Ninja Rogue", effectstest.GmrSatoruOracle)
			hand := me.Hand.Size()
			switch tc.how {
			case "reanimate":
				id := uuid.New()
				g.WithWriteLock(func() {
					me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Bear", TypeLine: "Creature — Bear",
						ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
				})
				effectstest.B30Reanimate(t, g, id, me.ID)
				hand = me.Hand.Size()
				effectstest.PassPriorityAroundTable(t, g)
			case "free":
				gmrCreature(t, g, me, "Artifact Creature — Construct", "{0}")

			case "paid":
				effectstest.FloatForTest(g, me, "GG")
				gmrCreature(t, g, me, "Creature — Bear", "{1}{G}")
			case "mixed":
				effectstest.FloatForTest(g, me, "GG")
				gmrCreature(t, g, me, "Creature — Bear", "{1}{G}")
				id := uuid.New()
				g.WithWriteLock(func() {
					me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Bear", TypeLine: "Creature — Bear",
						ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
				})
				// No boundary between the cast creature's resolution and
				// this: the same batch.
				effectstest.B30Reanimate(t, g, id, me.ID)
				g.RunStateChecksForTest()
				effectstest.PassPriorityAroundTable(t, g)

			}
			if got := me.Hand.Size() - hand; got != tc.draw {
				t.Errorf("drew %d, want %d", got, tc.draw)
			}
		})
	}
}

// Freestrider Commando gets its two counters when it was not cast or
// cost nothing, and not when mana paid for it.
func TestFreestriderCommandoCountersWhenNoManaWasSpent(t *testing.T) {
	for _, tc := range []struct {
		name string
		cost string
		pay  string
		want int
	}{
		{"cast with mana", "{2}{G}", "GGG", 0},
		{"cast for nothing", "{0}", "", 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.AdvanceToMain(t, g)
			effectstest.FloatForTest(g, me, tc.pay)
			id := effectstest.CastFromHandForTest(t, g, me, "Freestrider Commando", "Creature — Centaur Mercenary", tc.cost, effectstest.GmrFreestriderOracle, effectstest.Strict)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.CounterCount(g, id, game.CounterPlusOne); got != tc.want {
				t.Errorf("+1/+1 counters = %d, want %d", got, tc.want)
			}
		})
	}
	t.Run("put onto the battlefield", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.AdvanceToMain(t, g)
		id := uuid.New()
		g.WithWriteLock(func() {
			me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Freestrider Commando", TypeLine: "Creature — Centaur Mercenary",
				ManaCost: "{2}{G}", OracleID: effectstest.GmrFreestriderOracle, Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
		})
		effectstest.B30Reanimate(t, g, id, me.ID)
		if got := effectstest.CounterCount(g, id, game.CounterPlusOne); got != 2 {
			t.Errorf("+1/+1 counters = %d, want 2", got)
		}
	})
}

// --- spend riders with a duration or a count ------------------------

// Generator Servant's mana gives a creature spell haste until end of
// turn; the haste is gone once the turn ends.
func TestGeneratorServantHasteLastsUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	servant := effectstest.SeedPermanentWithOracle(g, me.ID, "Generator Servant", "Creature — Elemental", effectstest.GmrGeneratorServantOracle)
	effectstest.ActivateManaFor(t, g, me.ID, servant, 0, game.ManaAbilityParams{})
	bear := gmrCreature(t, g, me, "Creature — Bear", "{2}")
	if !effectstest.EffectiveAbilitiesContain(t, g, bear, "haste") {
		t.Fatal("the creature cast with the Servant's mana has no haste")
	}
	effectstest.EndTurn(t, g)
	if effectstest.EffectiveAbilitiesContain(t, g, bear, "haste") {
		t.Error("the haste lasted past the end of the turn")
	}
}

// Carnelian Orb's mana hastes a Dragon and nothing else.
func TestCarnelianOrbHastesOnlyADragon(t *testing.T) {
	for _, tc := range []struct {
		typeLine string
		haste    bool
	}{
		{"Creature — Dragon", true},
		{"Creature — Goblin", false},
	} {
		t.Run(tc.typeLine, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.AdvanceToMain(t, g)
			orb := effectstest.SeedPermanentWithOracle(g, me.ID, "Carnelian Orb of Dragonkind", "Artifact", effectstest.GmrCarnelianOrbOracle)
			effectstest.ActivateManaFor(t, g, me.ID, orb, 0, game.ManaAbilityParams{})
			id := gmrCreature(t, g, me, tc.typeLine, "{R}")
			if got := effectstest.EffectiveAbilitiesContain(t, g, id, "haste"); got != tc.haste {
				t.Errorf("haste = %v, want %v", got, tc.haste)
			}
		})
	}
}

// Opal Palace's coloured mana, spent on the commander, puts a +1/+1
// counter on it per time it has been cast from the command zone — the
// cast it paid for included.
func TestOpalPalaceCountsTheCommandersCommandZoneCasts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	palace := effectstest.SeedPermanentWithOracle(g, me.ID, "Opal Palace", "Land", effectstest.GmrOpalPalaceOracle)
	cmdr := uuid.New()
	g.WithWriteLock(func() {
		me.Command.PushTop(game.Card{InstanceID: cmdr, Name: "Test Commander", TypeLine: "Legendary Creature — Elf",
			ManaCost: "{G}", Colors: []string{"G"}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID, IsCommander: true})
		me.CommanderCasts[cmdr] = 1 // cast once before
	})
	effectstest.FloatForTest(g, me, "C")
	effectstest.ActivateManaFor(t, g, me.ID, palace, 1, game.ManaAbilityParams{Colors: []string{"G"}})
	effectstest.FloatForTest(g, me, "CC") // the commander tax for the second cast
	if err := g.CastSpell(me.ID, cmdr, game.CastSpellParams{FromZone: "command", Strict: true}); err != nil {
		t.Fatalf("cast the commander: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, cmdr, game.CounterPlusOne); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2 — two casts from the command zone", got)
	}
}

// --- Tyvar Kell -------------------------------------------------------

// Tyvar Kell's Elves tap for {B}, and its emblem gives an Elf spell
// haste until end of turn and draws two.
func TestTyvarKellGrantAndEmblem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	tyvar := effectstest.SeedPermanentWithOracle(g, me.ID, "Tyvar Kell", "Legendary Planeswalker — Tyvar", effectstest.GmrTyvarKellOracle)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(tyvar, game.CounterLoyalty, 6) })
	elf := effectstest.B12Push(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", "", 1, 1)
	if err := effectstest.TapGrantedMana(t, g, me.ID, elf, ""); err != nil {
		t.Fatalf("the Elf has no granted {T}: Add {B}: %v", err)
	}
	if got := effectstest.GmrPoolCount(me, "B"); got != 1 {
		t.Fatalf("pool has %d {B}, want 1", got)
	}

	effectstest.B16Activate(t, g, me.ID, tyvar, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)

	hand := me.Hand.Size()
	effectstest.FloatForTest(g, me, "G")
	warrior := effectstest.CastFromHandForTest(t, g, me, "Elf Warrior", "Creature — Elf Warrior", "{1}{G}", "", effectstest.Strict)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, warrior, "haste") {
		t.Fatal("the Elf did not gain haste")
	}
	effectstest.EndTurn(t, g)
	if effectstest.EffectiveAbilitiesContain(t, g, warrior, "haste") {
		t.Error("the emblem's haste lasted past the end of the turn")
	}
}

// --- Domri, Chaos Bringer ---------------------------------------------

const gmrDomriChaosBringerOracle = "2a5408ed-8b47-4896-97e4-aa102a4b85c9"

// Domri's +1 mana, spent on a creature spell, gives it riot: the creature
// asks the riot question as it enters (ADR 0109 §10 and §11). Spent on
// an artifact, it gives nothing.
func TestDomriChaosBringerManaGivesACreatureSpellRiot(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		riot           bool
	}{
		{"a creature spell", "Creature — Bear", true},
		{"an artifact spell", "Artifact", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.AdvanceToMain(t, g)
			domri := effectstest.SeedPermanentWithOracle(g, me.ID, "Domri, Chaos Bringer", "Legendary Planeswalker — Domri", gmrDomriChaosBringerOracle)
			g.WithWriteLock(func() { _ = g.AddCounterForEffect(domri, game.CounterLoyalty, 5) })
			effectstest.B16Activate(t, g, me.ID, domri, 0, game.ActivateAbilityParams{})
			effectstest.PassPriorityAroundTable(t, g)
			pick := effectstest.RiderLatestManaPick(g, me.ID)
			if pick == nil {
				t.Fatal("the +1 queued no {R} or {G} pick")
			}
			if err := g.ResolveManaChoice(pick.ID, me.ID, "R"); err != nil {
				t.Fatalf("ResolveManaChoice: %v", err)
			}
			id := effectstest.CastFromHandForTest(t, g, me, "Spell", tc.typeLine, "{R}", "", effectstest.Strict)
			effectstest.PassPriorityAroundTable(t, g)
			open := effectstest.PendingOfKind(g, game.PendingChoiceEntryRiot) != nil
			if open != tc.riot {
				t.Fatalf("riot question open = %v, want %v", open, tc.riot)
			}
			if !tc.riot {
				return
			}
			effectstest.AnswerRiotFor(t, g, false)
			effectstest.PassPriorityAroundTable(t, g)
			if !effectstest.EffectiveAbilitiesContain(t, g, id, "haste") {
				t.Error("the creature took haste from riot and does not have it")
			}
		})
	}
}
