package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// granted_mana_spent_readers_test.go — ADR 0109 §11 (#1552), the card
// half: the printed sunburst cards on the keyword, the readers that one
// permanent grants over another's cast (Lux Artillery, Coin of Mastery,
// Kalain, Solar Array), "no mana was spent" (Satoru, Freestrider
// Commando), the spend riders with a duration or a count (Generator
// Servant, Carnelian Orb, Opal Palace) and Tyvar Kell's emblem. The
// engine contract is pinned in game/granted_mana_spent_readers_test.go.

const (
	gmrLuxArtilleryOracle      = "cbd76b22-d04e-48e7-bcab-c7f5466d67d8"
	gmrCoinOfMasteryOracle     = "d78518ee-df79-48d1-b9d5-4f968b441899"
	gmrSatoruOracle            = "7555c429-5f2d-4171-b6b0-8e3c8da7f314"
	gmrOpalPalaceOracle        = "aa6723a2-75da-49f5-a1ba-cbfa82c55301"
	gmrKalainOracle            = "4fca09ac-8134-43d9-a84b-686db5e2bf69"
	gmrGeneratorServantOracle  = "68ac061a-e4a3-46df-9169-62f6a8481584"
	gmrCarnelianOrbOracle      = "651c967c-8f71-4eb5-b22f-545e55ea050e"
	gmrTyvarKellOracle         = "14c8a590-88ec-4982-ad7b-36a219ff6de7"
	gmrFreestriderOracle       = "09117016-5fd1-4590-9867-aad65d3097e3"
	gmrSolarArrayOracle        = "e8e2f273-5e74-4f16-8d49-2e86e9c9f2dc"
	gmrEngineeredExplosives    = "95fd897e-9086-42c4-8e0d-61bb02333c5f"
	gmrPentadPrismOracle       = "72e8d67a-41e6-4074-bcbd-b54ead995237"
	gmrHeliophialOracle        = "6c3be36d-2461-4e38-9ab4-4394cd4996e7"
	gmrInfusedArrowsOracle     = "942216f7-6803-4ab3-9239-463002590f6f"
	gmrSpinalParasiteOracle    = "159427f7-27fe-490d-ac78-fed092952f51"
	gmrArcboundWandererOracle  = "03436524-197a-4941-a2a1-c7c4b71c4709"
	gmrClearwaterGobletOracle  = "ffecd3e8-aa28-41a1-8154-a54ae3ce8016"
	gmrOpalineBracersOracle    = "a4d6cb66-698d-4b76-b922-f799ad9d8805"
	gmrLunarAvengerOracle      = "ab42fb8b-c463-45db-b71d-ed847c9abf1f"
	gmrSawtoothThresherOracle  = "3a8f382e-0bc6-4e79-87e2-9f99d907d234"
	gmrSuncrusherOracle        = "4e81ead0-df8b-4d9d-b1cb-d1d9dc503d4c"
	gmrSolarionOracle          = "4022e540-8fa0-4eee-bcf1-df5046289f2b"
	gmrBatonOfCourageOracle    = "a20bb8b3-b724-4fec-b179-ca9e6adc4718"
	gmrSkyreachMantaOracle     = "58eab745-9afa-4db7-9d8a-f5eb0b9dca6c"
	gmrSuntouchedMyrOracle     = "75d4f61a-8276-43ad-99e6-0d756a597e4b"
	gmrHalfTreasureSourceKinds = game.ManaSourceTreasure | game.ManaSourceArtifact
)

// gmrFloatFrom floats one mana of each colour in `spec`, made by a
// source of `kinds` — a Treasure's mana is artifact mana too.
func gmrFloatFrom(g *game.Game, p *game.Player, spec string, kinds game.ManaSourceKinds) {
	g.WithWriteLock(func() {
		for _, r := range spec {
			p.ManaPool.AddMana(game.ManaToken{Color: string(r), Source: uuid.New(), SourceKinds: kinds})
		}
	})
}

// gmrCreature casts a vanilla creature of `typeLine` and `cost` from
// hand under strict mana and resolves it.
func gmrCreature(t *testing.T, g *game.Game, p *game.Player, typeLine, cost string) uuid.UUID {
	t.Helper()
	id := castFromHandForTest(t, g, p, "Test Creature", typeLine, cost, "", strict)
	passPriorityAroundTable(t, g)
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
		{"Suntouched Myr", "Artifact Creature — Myr", "{3}", gmrSuntouchedMyrOracle, "WUB", game.CounterPlusOne, 0},
		{"Skyreach Manta", "Artifact Creature — Fish", "{5}", gmrSkyreachMantaOracle, "WUBRG", game.CounterPlusOne, 0},
		{"Pentad Prism", "Artifact", "{2}", gmrPentadPrismOracle, "WU", game.CounterCharge, 0},
		{"Clearwater Goblet", "Artifact", "{5}", gmrClearwaterGobletOracle, "WUBRG", game.CounterCharge, 0},
		{"Opaline Bracers", "Artifact — Equipment", "{4}", gmrOpalineBracersOracle, "WUBR", game.CounterCharge, 0},
		{"Sawtooth Thresher", "Artifact Creature — Construct", "{6}", gmrSawtoothThresherOracle, "WUBRGC", game.CounterPlusOne, 0},
		{"Lunar Avenger", "Artifact Creature — Golem", "{7}", gmrLunarAvengerOracle, "WUBRGCC", game.CounterPlusOne, 0},
		{"Heliophial", "Artifact", "{5}", gmrHeliophialOracle, "WUBRC", game.CounterCharge, 0},
		{"Suncrusher", "Artifact Creature — Construct", "{9}", gmrSuncrusherOracle, "WUBRGCCCC", game.CounterPlusOne, 0},
		{"Solarion", "Artifact Creature — Construct", "{7}", gmrSolarionOracle, "WUBCCCC", game.CounterPlusOne, 0},
		{"Baton of Courage", "Artifact", "{3}", gmrBatonOfCourageOracle, "WUB", game.CounterCharge, 0},
		{"Infused Arrows", "Artifact", "{4}", gmrInfusedArrowsOracle, "WUBR", game.CounterCharge, 0},
		{"Spinal Parasite", "Artifact Creature — Insect", "{5}", gmrSpinalParasiteOracle, "WUBRG", game.CounterPlusOne, 0},
		{"Arcbound Wanderer", "Artifact Creature — Golem", "{6}", gmrArcboundWandererOracle, "WUBRGC", game.CounterPlusOne, 0},
		{"Engineered Explosives", "Artifact", "{X}", gmrEngineeredExplosives, "UR", game.CounterCharge, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			advanceToMain(t, g)
			floatForTest(g, me, tc.pay)
			want := 0
			for _, r := range tc.pay {
				if r != 'C' {
					want++
				}
			}
			params := strict
			params.XValue = tc.x
			id := castFromHandForTest(t, g, me, tc.name, tc.typeLine, tc.cost, tc.oracle, params)
			passPriorityAroundTable(t, g)
			if got := counterCount(g, id, tc.kind); got != want {
				t.Errorf("%s counters = %d, want %d — one per colour spent", tc.kind, got, want)
			}
		})
	}
}

// Engineered Explosives with two charge counters destroys each nonland
// permanent with mana value 2, and nothing else.
func TestEngineeredExplosivesDestroysTheManaValueOfItsCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	two := b12Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	three := b12Push(g, opp.ID, "Ogre", "Creature — Ogre", "", 3, 3)
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
	floatForTest(g, me, "UR")
	params := strict
	params.XValue = 2
	ee := castFromHandForTest(t, g, me, "Engineered Explosives", "Artifact", "{X}", gmrEngineeredExplosives, params)
	passPriorityAroundTable(t, g)
	floatForTest(g, me, "CC")
	b16Activate(t, g, me.ID, ee, 0, game.ActivateAbilityParams{Strict: true})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(two) {
		t.Error("the mana value 2 Bear survived two charge counters")
	}
	if !g.Battlefield.Contains(three) {
		t.Error("the mana value 3 Ogre was destroyed")
	}
}

// Pentad Prism cashes its charge counters for mana of any colour.
func TestPentadPrismRemovesAChargeCounterForMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	floatForTest(g, me, "WU")
	prism := castFromHandForTest(t, g, me, "Pentad Prism", "Artifact", "{2}", gmrPentadPrismOracle, strict)
	passPriorityAroundTable(t, g)
	activateManaFor(t, g, me.ID, prism, 0, game.ManaAbilityParams{Colors: []string{"G"}})
	if got := counterCount(g, prism, game.CounterCharge); got != 1 {
		t.Errorf("charge counters = %d after one activation, want 1", got)
	}
	if got := gmrPoolCount(me, "G"); got != 1 {
		t.Errorf("pool has %d {G}, want 1", got)
	}
}

// Heliophial's damage is its last-known charge counters, after the
// sacrifice paid for it.
func TestHeliophialDealsItsCountersAfterTheSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	floatForTest(g, me, "WUBCC")
	phial := castFromHandForTest(t, g, me, "Heliophial", "Artifact", "{5}", gmrHeliophialOracle, strict)
	passPriorityAroundTable(t, g)
	life := opp.Life
	floatForTest(g, me, "CC")
	b16Activate(t, g, me.ID, phial, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	})
	passPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 3 {
		t.Errorf("Heliophial dealt %d, want 3 — its three charge counters", got)
	}
}

// Infused Arrows' X is the number of counters removed.
func TestInfusedArrowsGivesMinusXForTheCountersRemoved(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	ogre := b12Push(g, opp.ID, "Ogre", "Creature — Ogre", "", 3, 3)
	floatForTest(g, me, "WUBR")
	arrows := castFromHandForTest(t, g, me, "Infused Arrows", "Artifact", "{4}", gmrInfusedArrowsOracle, strict)
	passPriorityAroundTable(t, g)
	b16Activate(t, g, me.ID, arrows, 0, game.ActivateAbilityParams{
		Strict:           true,
		Targets:          []game.TargetRef{{Kind: game.TargetCard, ID: ogre}},
		CounterSourceIDs: []uuid.UUID{arrows},
		CounterCounts:    []int{3},
	})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) {
		t.Error("a 3/3 survived -3/-3")
	}
	if got := counterCount(g, arrows, game.CounterCharge); got != 1 {
		t.Errorf("charge counters left = %d, want 1", got)
	}
}

// Spinal Parasite removes one counter from a permanent that has one
// kind, without asking.
func TestSpinalParasiteRemovesACounterFromTheTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	rock := b12Push(g, opp.ID, "Charged Rock", "Artifact", "", 0, 0)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(rock, game.CounterCharge, 3) })
	floatForTest(g, me, "WUBRG")
	parasite := castFromHandForTest(t, g, me, "Spinal Parasite", "Artifact Creature — Insect", "{5}", gmrSpinalParasiteOracle, strict)
	passPriorityAroundTable(t, g)
	b16Activate(t, g, me.ID, parasite, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}},
	})
	passPriorityAroundTable(t, g)
	if got := counterCount(g, rock, game.CounterCharge); got != 2 {
		t.Errorf("charge counters on the target = %d, want 2", got)
	}
	if got := counterCount(g, parasite, game.CounterPlusOne); got != 3 {
		t.Errorf("+1/+1 counters left on the Parasite = %d, want 3", got)
	}
}

// Modular: when Arcbound Wanderer dies, its +1/+1 counters may go onto
// a target artifact creature (CR 702.43a), counted from its last-known
// information.
func TestArcboundWandererMovesItsCountersWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	golem := b12Push(g, me.ID, "Golem", "Artifact Creature — Golem", "", 3, 3)
	floatForTest(g, me, "WUBRCC")
	wanderer := castFromHandForTest(t, g, me, "Arcbound Wanderer", "Artifact Creature — Golem", "{6}", gmrArcboundWandererOracle, strict)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(wanderer); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	if pick := latestPickTarget(g, me.ID); pick != nil {
		answerTriggerTargets(t, g, me.ID, golem)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, golem, game.CounterPlusOne); got != 4 {
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
		{"Etched Oracle, which prints it too", "Artifact Creature — Wizard", etchedOracleOracle, 6},
		{"a creature that is not an artifact", "Creature — Bear", "", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			advanceToMain(t, g)
			seedPermanentWithOracle(g, me.ID, "Lux Artillery", "Artifact", gmrLuxArtilleryOracle)
			floatForTest(g, me, "WUB")
			id := castFromHandForTest(t, g, me, "Spell", tc.typeLine, "{3}", tc.oracle, strict)
			passPriorityAroundTable(t, g)
			if got := counterCount(g, id, game.CounterPlusOne); got != tc.want {
				t.Errorf("+1/+1 counters = %d, want %d", got, tc.want)
			}
		})
	}
}

// Solar Array's mana makes the next artifact spell this turn gain
// sunburst — including the spell that mana paid for.
func TestSolarArrayGivesTheNextArtifactSpellSunburst(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	array := seedPermanentWithOracle(g, me.ID, "Solar Array", "Artifact", gmrSolarArrayOracle)
	activateManaFor(t, g, me.ID, array, 0, game.ManaAbilityParams{Colors: []string{"W"}})
	floatForTest(g, me, "UB")
	golem := castFromHandForTest(t, g, me, "Golem", "Artifact Creature — Golem", "{3}", "", strict)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, golem, game.CounterPlusOne); got != 3 {
		t.Errorf("+1/+1 counters = %d, want 3 — W from the Array, U and B", got)
	}
	// The trigger was spent: the next artifact this turn gets nothing.
	floatForTest(g, me, "WUB")
	second := castFromHandForTest(t, g, me, "Second Golem", "Artifact Creature — Golem", "{3}", "", strict)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, second, game.CounterPlusOne); got != 0 {
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
		{"Coin of Mastery, Treasure mana", gmrCoinOfMasteryOracle, "Artifact", gmrHalfTreasureSourceKinds, 2},
		{"Coin of Mastery, a mana rock's mana", gmrCoinOfMasteryOracle, "Artifact", game.ManaSourceArtifact, 2},
		{"Coin of Mastery, land mana", gmrCoinOfMasteryOracle, "Artifact", game.ManaSourceLand, 0},
		{"Kalain, Treasure mana", gmrKalainOracle, "Legendary Creature — Human Elf Bard", gmrHalfTreasureSourceKinds, 2},
		{"Kalain, a mana rock's mana", gmrKalainOracle, "Legendary Creature — Human Elf Bard", game.ManaSourceArtifact, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			advanceToMain(t, g)
			seedPermanentWithOracle(g, me.ID, "Reader", tc.typeLine, tc.oracle)
			gmrFloatFrom(g, me, "RR", tc.kinds)
			floatForTest(g, me, "G")
			bear := gmrCreature(t, g, me, "Creature — Bear", "{2}{G}")
			if got := counterCount(g, bear, game.CounterPlusOne); got != tc.want {
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
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			advanceToMain(t, g)
			seedPermanentWithOracle(g, me.ID, "Satoru, the Infiltrator", "Legendary Creature — Human Ninja Rogue", gmrSatoruOracle)
			hand := me.Hand.Size()
			switch tc.how {
			case "reanimate":
				id := uuid.New()
				g.WithWriteLock(func() {
					me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Bear", TypeLine: "Creature — Bear",
						ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
				})
				b30Reanimate(t, g, id, me.ID)
				hand = me.Hand.Size()
				passPriorityAroundTable(t, g)
			case "free":
				gmrCreature(t, g, me, "Artifact Creature — Construct", "{0}")

			case "paid":
				floatForTest(g, me, "GG")
				gmrCreature(t, g, me, "Creature — Bear", "{1}{G}")
			case "mixed":
				floatForTest(g, me, "GG")
				gmrCreature(t, g, me, "Creature — Bear", "{1}{G}")
				id := uuid.New()
				g.WithWriteLock(func() {
					me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Bear", TypeLine: "Creature — Bear",
						ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
				})
				// No boundary between the cast creature's resolution and
				// this: the same batch.
				b30Reanimate(t, g, id, me.ID)
				g.RunStateChecksForTest()
				passPriorityAroundTable(t, g)

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
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			advanceToMain(t, g)
			floatForTest(g, me, tc.pay)
			id := castFromHandForTest(t, g, me, "Freestrider Commando", "Creature — Centaur Mercenary", tc.cost, gmrFreestriderOracle, strict)
			passPriorityAroundTable(t, g)
			if got := counterCount(g, id, game.CounterPlusOne); got != tc.want {
				t.Errorf("+1/+1 counters = %d, want %d", got, tc.want)
			}
		})
	}
	t.Run("put onto the battlefield", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		advanceToMain(t, g)
		id := uuid.New()
		g.WithWriteLock(func() {
			me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Freestrider Commando", TypeLine: "Creature — Centaur Mercenary",
				ManaCost: "{2}{G}", OracleID: gmrFreestriderOracle, Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
		})
		b30Reanimate(t, g, id, me.ID)
		if got := counterCount(g, id, game.CounterPlusOne); got != 2 {
			t.Errorf("+1/+1 counters = %d, want 2", got)
		}
	})
}

// --- spend riders with a duration or a count ------------------------

// Generator Servant's mana gives a creature spell haste until end of
// turn; the haste is gone once the turn ends.
func TestGeneratorServantHasteLastsUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	servant := seedPermanentWithOracle(g, me.ID, "Generator Servant", "Creature — Elemental", gmrGeneratorServantOracle)
	activateManaFor(t, g, me.ID, servant, 0, game.ManaAbilityParams{})
	bear := gmrCreature(t, g, me, "Creature — Bear", "{2}")
	if !effectiveAbilitiesContain(t, g, bear, "haste") {
		t.Fatal("the creature cast with the Servant's mana has no haste")
	}
	endTurn(t, g)
	if effectiveAbilitiesContain(t, g, bear, "haste") {
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
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			advanceToMain(t, g)
			orb := seedPermanentWithOracle(g, me.ID, "Carnelian Orb of Dragonkind", "Artifact", gmrCarnelianOrbOracle)
			activateManaFor(t, g, me.ID, orb, 0, game.ManaAbilityParams{})
			id := gmrCreature(t, g, me, tc.typeLine, "{R}")
			if got := effectiveAbilitiesContain(t, g, id, "haste"); got != tc.haste {
				t.Errorf("haste = %v, want %v", got, tc.haste)
			}
		})
	}
}

// Opal Palace's coloured mana, spent on the commander, puts a +1/+1
// counter on it per time it has been cast from the command zone — the
// cast it paid for included.
func TestOpalPalaceCountsTheCommandersCommandZoneCasts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	palace := seedPermanentWithOracle(g, me.ID, "Opal Palace", "Land", gmrOpalPalaceOracle)
	cmdr := uuid.New()
	g.WithWriteLock(func() {
		me.Command.PushTop(game.Card{InstanceID: cmdr, Name: "Test Commander", TypeLine: "Legendary Creature — Elf",
			ManaCost: "{G}", Colors: []string{"G"}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID, IsCommander: true})
		me.CommanderCasts[cmdr] = 1 // cast once before
	})
	floatForTest(g, me, "C")
	activateManaFor(t, g, me.ID, palace, 1, game.ManaAbilityParams{Colors: []string{"G"}})
	floatForTest(g, me, "CC") // the commander tax for the second cast
	if err := g.CastSpell(me.ID, cmdr, game.CastSpellParams{FromZone: "command", Strict: true}); err != nil {
		t.Fatalf("cast the commander: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, cmdr, game.CounterPlusOne); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2 — two casts from the command zone", got)
	}
}

// --- Tyvar Kell -------------------------------------------------------

// Tyvar Kell's Elves tap for {B}, and its emblem gives an Elf spell
// haste until end of turn and draws two.
func TestTyvarKellGrantAndEmblem(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	tyvar := seedPermanentWithOracle(g, me.ID, "Tyvar Kell", "Legendary Planeswalker — Tyvar", gmrTyvarKellOracle)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(tyvar, game.CounterLoyalty, 6) })
	elf := b12Push(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", "", 1, 1)
	if err := tapGrantedMana(t, g, me.ID, elf, ""); err != nil {
		t.Fatalf("the Elf has no granted {T}: Add {B}: %v", err)
	}
	if got := gmrPoolCount(me, "B"); got != 1 {
		t.Fatalf("pool has %d {B}, want 1", got)
	}

	b16Activate(t, g, me.ID, tyvar, 2, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)

	hand := me.Hand.Size()
	floatForTest(g, me, "G")
	warrior := castFromHandForTest(t, g, me, "Elf Warrior", "Creature — Elf Warrior", "{1}{G}", "", strict)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
	if !effectiveAbilitiesContain(t, g, warrior, "haste") {
		t.Fatal("the Elf did not gain haste")
	}
	endTurn(t, g)
	if effectiveAbilitiesContain(t, g, warrior, "haste") {
		t.Error("the emblem's haste lasted past the end of the turn")
	}
}

// gmrPoolCount is how many mana of `color` is floating in p's pool.
func gmrPoolCount(p *game.Player, color string) int {
	n := 0
	for _, t := range p.ManaPool {
		if t.Color == color {
			n++
		}
	}
	return n
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
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			advanceToMain(t, g)
			domri := seedPermanentWithOracle(g, me.ID, "Domri, Chaos Bringer", "Legendary Planeswalker — Domri", gmrDomriChaosBringerOracle)
			g.WithWriteLock(func() { _ = g.AddCounterForEffect(domri, game.CounterLoyalty, 5) })
			b16Activate(t, g, me.ID, domri, 0, game.ActivateAbilityParams{})
			passPriorityAroundTable(t, g)
			pick := riderLatestManaPick(g, me.ID)
			if pick == nil {
				t.Fatal("the +1 queued no {R} or {G} pick")
			}
			if err := g.ResolveManaChoice(pick.ID, me.ID, "R"); err != nil {
				t.Fatalf("ResolveManaChoice: %v", err)
			}
			id := castFromHandForTest(t, g, me, "Spell", tc.typeLine, "{R}", "", strict)
			passPriorityAroundTable(t, g)
			open := pendingOfKind(g, game.PendingChoiceEntryRiot) != nil
			if open != tc.riot {
				t.Fatalf("riot question open = %v, want %v", open, tc.riot)
			}
			if !tc.riot {
				return
			}
			answerRiotFor(t, g, false)
			passPriorityAroundTable(t, g)
			if !effectiveAbilitiesContain(t, g, id, "haste") {
				t.Error("the creature took haste from riot and does not have it")
			}
		})
	}
}
