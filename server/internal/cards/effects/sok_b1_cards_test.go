package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sok_b1_cards_test.go — Saviors of Kamigawa, batch 1 (#5601). Each test
// plays the real catalog entry and asserts what its text says.

const (
	sokB1HandOfHonorOracle     = "feb65f2e-6f71-4f29-bf31-b210c85f3ab3"
	sokB1HandOfCrueltyOracle   = "ddbd4d8c-9763-43d7-882f-6edf8efffb0a"
	sokB1MothriderOracle       = "d6b54f58-7775-4180-b2e5-1b57da65a409"
	sokB1InnerChamberOracle    = "ea5e8920-1343-4335-b516-6db43c3bfeb0"
	sokB1CavekeeperOracle      = "56551cba-3871-4746-a11f-7ab8aaf46cf6"
	sokB1KannushiOracle        = "b92bd85f-20c4-410e-af19-77d987d7dc02"
	sokB1EmptyGravesOracle     = "bff92baa-fc0e-4da4-8bd5-e57e4963d642"
	sokB1NightsoilOracle       = "2efd1c86-41ba-4b86-bf63-e60e3263cb06"
	sokB1KurosTakenOracle      = "8b787ba5-0b87-449f-98cd-9454445e2f48"
	sokB1DawnbladeOracle       = "cd954a2a-5799-4c65-af8a-47adb1ce5e91"
	sokB1NikkoOracle           = "f5cf3f74-e497-46d0-acfb-3b57ee27a314"
	sokB1YukiOracle            = "8f867a2c-f17b-4608-bca7-5b49d1387bc7"
	sokB1KiriOracle            = "e09661d4-858b-43a8-b6a4-ba76f7a69777"
	sokB1HaruOracle            = "fead5048-3ce5-46d9-9c04-0a52f1554030"
	sokB1KemuriOracle          = "54db986e-cf8a-4d2c-8ce9-13da11b11464"
	sokB1SerowOracle           = "7e605a8c-d6d4-410a-aca4-52576d08d81c"
	sokB1FreeRidersOracle      = "0e381b7e-2d32-4a9a-a327-e9dd1236dc1b"
	sokB1OniOracle             = "f551093d-40b7-43ac-8023-4cfbb6c646da"
	sokB1SkullCollectorOracle  = "42ec62b1-c899-4ebd-b786-28edf76ce8e0"
	sokB1GlitterfangOracle     = "226dca3d-9da3-4c3f-89d0-67264fd19c51"
	sokB1ScoutOracle           = "0c319f45-cb0f-4641-acd8-145768acb3fc"
	sokB1OboroOracle           = "645fb11b-d684-4bec-8532-8fa97e8f7b28"
	sokB1BreezecallerOracle    = "257ee9a0-5ee0-49c4-8619-e75e1134d7ac"
	sokB1DrillmasterOracle     = "aa228a01-7de3-4539-9ebb-b3c546f72b41"
	sokB1CaptiveFlameOracle    = "fada1102-bbb7-4c90-a72e-6c595c08a55d"
	sokB1BriarknitOracle       = "913dcf97-0804-4b21-8daa-f2b481dc0f2f"
	sokB1DreamcatcherOracle    = "61529c3a-9b7e-4e73-9f6e-58c992ef1ffc"
	sokB1RedeemerOracle        = "b6c22cbe-61e0-4d66-a78d-7016e84c303d"
	sokB1RaiderOracle          = "3d772c44-a756-406d-a83f-f671a81513b0"
	sokB1FurysFireOracle       = "5b3b512e-6849-4d2e-9352-a36bc1646a92"
	sokB1InnerFireOracle       = "27348fe5-21d4-40f7-9feb-bba2c2321008"
	sokB1SeekOracle            = "7537b482-7626-4b65-8b67-53b99baf1872"
	sokB1InnerCalmOracle       = "b7bdbae5-549f-403b-83ca-4f9a1ac93e93"
	sokB1OneWithNothingOracle  = "ba995901-bb32-45ed-8af8-f98784b5a9b0"
	sokB1GazeOracle            = "91a69925-1e9b-433a-ad58-88bfd3c4bc46"
	sokB1SpiralingEmbersOracle = "25a2f4c8-cdab-49a5-80fd-a75230829f3f"
	sokB1PresenceOracle        = "6728064e-1d4e-4d4f-a0ab-127c23478f3b"
	sokB1DosanOracle           = "59aeeb18-d263-426f-aebc-0f687b09801b"
	sokB1SunderOracle          = "09f75f28-09b9-4a6e-a1b0-671208d274d3"
	sokB1PathOfAngersOracle    = "9e89db98-fc9f-4574-b5bc-ee421fa29428"
)

func TestSokB1EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		sokB1HandOfHonorOracle:     "Hand of Honor",
		sokB1HandOfCrueltyOracle:   "Hand of Cruelty",
		sokB1MothriderOracle:       "Araba Mothrider",
		sokB1InnerChamberOracle:    "Inner-Chamber Guard",
		sokB1CavekeeperOracle:      "Ronin Cavekeeper",
		sokB1KannushiOracle:        "Promised Kannushi",
		sokB1EmptyGravesOracle:     "Kami of Empty Graves",
		sokB1NightsoilOracle:       "Nightsoil Kami",
		sokB1KurosTakenOracle:      "Kuro's Taken",
		sokB1DawnbladeOracle:       "Kitsune Dawnblade",
		sokB1NikkoOracle:           "Nikko-Onna",
		sokB1YukiOracle:            "Yuki-Onna",
		sokB1KiriOracle:            "Kiri-Onna",
		sokB1HaruOracle:            "Haru-Onna",
		sokB1KemuriOracle:          "Kemuri-Onna",
		sokB1SerowOracle:           "Stampeding Serow",
		sokB1FreeRidersOracle:      "Eiganjo Free-Riders",
		sokB1OniOracle:             "Oni of Wild Places",
		sokB1SkullCollectorOracle:  "Skull Collector",
		sokB1GlitterfangOracle:     "Glitterfang",
		sokB1ScoutOracle:           "Sakura-Tribe Scout",
		sokB1OboroOracle:           "Oboro, Palace in the Clouds",
		sokB1BreezecallerOracle:    "Oboro Breezecaller",
		sokB1DrillmasterOracle:     "Akki Drillmaster",
		sokB1CaptiveFlameOracle:    "Captive Flame",
		sokB1BriarknitOracle:       "Briarknit Kami",
		sokB1DreamcatcherOracle:    "Dreamcatcher",
		sokB1RedeemerOracle:        "Ghost-Lit Redeemer",
		sokB1RaiderOracle:          "Ghost-Lit Raider",
		sokB1FurysFireOracle:       "Shinen of Fury's Fire",
		sokB1InnerFireOracle:       "Inner Fire",
		sokB1SeekOracle:            "Seek the Horizon",
		sokB1InnerCalmOracle:       "Inner Calm, Outer Strength",
		sokB1OneWithNothingOracle:  "One with Nothing",
		sokB1GazeOracle:            "Gaze of Adamaro",
		sokB1SpiralingEmbersOracle: "Spiraling Embers",
		sokB1PresenceOracle:        "Presence of the Wise",
		sokB1DosanOracle:           "Dosan's Oldest Chant",
		sokB1SunderOracle:          "Sunder from Within",
		sokB1PathOfAngersOracle:    "Path of Anger's Flame",
	}
	if len(want) != 40 {
		t.Fatalf("the batch registers 40 cards, the table lists %d", len(want))
	}
	for oracle, name := range want {
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

// --- shared test helpers ------------------------------------------------------------

// sokB1Cast casts a catalog creature with a printed power and toughness
// from the active player's hand in a main phase. Nothing is passed: the
// caller settles the spell and its enters trigger.
func sokB1Cast(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	advanceToMain(t, g)
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	g.WithWriteLock(func() {
		active.Hand.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
			Power: power, Toughness: toughness, Owner: active.ID, Controller: active.ID,
		})
	})
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

// sokB1CastAndTarget casts the creature and answers its enters trigger's
// target prompt with `pick`.
func sokB1CastAndTarget(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int, pick game.TargetRef) uuid.UUID {
	t.Helper()
	id := sokB1Cast(t, g, name, typeLine, oracle, power, toughness)
	me := g.Seats[g.Turn.ActiveSeat]
	p := passUntilPickTarget(t, g, me.ID)
	if p == nil {
		t.Fatalf("%s: no target prompt for the enters trigger", name)
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, pick); err != nil {
		t.Fatalf("%s: ResolvePickTarget: %v", name, err)
	}
	return id
}

// sokB1ExpectOnnaReturn casts a non-Spirit spell (no trigger) and then a
// Spirit spell, says yes, and checks the Onna went back to its owner's
// hand.
func sokB1ExpectOnnaReturn(t *testing.T, g *game.Game, name string, onna uuid.UUID) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	if !g.Battlefield.Contains(onna) {
		t.Fatalf("%s is not on the battlefield", name)
	}
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	if latestTriggerPrompt(g, me.ID) != nil {
		t.Errorf("%s: a Bear spell asked to return it", name)
	}
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "A Spirit", "Creature — Spirit", "", nil)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(onna) || !me.Hand.Contains(onna) {
		t.Errorf("%s did not return to its owner's hand on a Spirit spell", name)
	}
}

// --- printed keywords -----------------------------------------------------------------

func TestSokB1CardsHaveTheirPrintedKeywords(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Hand of Honor", "Creature — Human Samurai", sokB1HandOfHonorOracle, []string{"protection from black"}},
		{"Hand of Cruelty", "Creature — Human Samurai", sokB1HandOfCrueltyOracle, []string{"protection from white"}},
		{"Araba Mothrider", "Creature — Human Samurai", sokB1MothriderOracle, []string{"flying"}},
		{"Stampeding Serow", "Creature — Antelope Beast", sokB1SerowOracle, []string{"trample"}},
		{"Eiganjo Free-Riders", "Creature — Human Soldier", sokB1FreeRidersOracle, []string{"flying"}},
		{"Oni of Wild Places", "Creature — Demon Spirit", sokB1OniOracle, []string{"haste"}},
		{"Glitterfang", "Creature — Spirit", sokB1GlitterfangOracle, []string{"haste"}},
		{"Oboro Breezecaller", "Creature — Moonfolk Wizard", sokB1BreezecallerOracle, []string{"flying"}},
		{"Shinen of Fury's Fire", "Creature — Spirit", sokB1FurysFireOracle, []string{"haste"}},
	} {
		g := newCatalogGame(t)
		id := b12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
		for _, kw := range tc.keywords {
			if !hasKeywordOnBattlefield(t, g, id, kw) {
				t.Errorf("%s lacks %s", tc.name, kw)
			}
		}
	}
}

// --- bushido ----------------------------------------------------------------------------

func TestSokB1BushidoCreaturesGrowWhenBlocked(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, tough, n        int
	}{
		{"Hand of Honor", "Creature — Human Samurai", sokB1HandOfHonorOracle, 2, 2, 1},
		{"Hand of Cruelty", "Creature — Human Samurai", sokB1HandOfCrueltyOracle, 2, 2, 1},
		{"Araba Mothrider", "Creature — Human Samurai", sokB1MothriderOracle, 1, 1, 1},
		{"Inner-Chamber Guard", "Creature — Human Samurai", sokB1InnerChamberOracle, 0, 2, 2},
		{"Ronin Cavekeeper", "Creature — Human Samurai", sokB1CavekeeperOracle, 4, 3, 2},
		{"Kuro's Taken", "Creature — Rat Samurai", sokB1KurosTakenOracle, 1, 1, 1},
		{"Kitsune Dawnblade", "Creature — Fox Samurai", sokB1DawnbladeOracle, 2, 3, 1},
	} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.tough)
		g.WithWriteLock(func() {
			g.EmitEvent(game.Event{Kind: game.EventBecomesBlocked, Actor: opp.ID, Source: id, CardID: id, Target: id})
		})
		passPriorityAroundTable(t, g)
		if p, tough := effectivePTOf(t, g, id); p != tc.power+tc.n || tough != tc.tough+tc.n {
			t.Errorf("%s is %d/%d after becoming blocked, want %d/%d", tc.name, p, tough, tc.power+tc.n, tc.tough+tc.n)
		}
	}
}

// --- soulshift --------------------------------------------------------------------------

func TestSokB1SoulshiftersReturnASmallerSpirit(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, cost string
		n                            int
	}{
		{"Promised Kannushi", "Creature — Human Druid", sokB1KannushiOracle, "{G}", 7},
		{"Kami of Empty Graves", "Creature — Spirit", sokB1EmptyGravesOracle, "{3}{B}", 3},
		{"Nightsoil Kami", "Creature — Spirit", sokB1NightsoilOracle, "{4}{G}{G}", 5},
	} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		kami := pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: tc.name, TypeLine: tc.typeLine, OracleID: tc.oracle,
			ManaCost: tc.cost, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
		})
		fits, tooBig := uuid.New(), uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: fits, Name: "Fits", TypeLine: "Creature — Spirit",
			ManaCost: chkB1GenericCost(tc.n), Owner: me.ID, Controller: me.ID})
		me.Graveyard.PushTop(game.Card{InstanceID: tooBig, Name: "Too Big", TypeLine: "Creature — Spirit",
			ManaCost: chkB1GenericCost(tc.n + 1), Owner: me.ID, Controller: me.ID})
		killCreature(t, g, me.ID, kami)
		p := passUntilPickTarget(t, g, me.ID)
		if p == nil {
			t.Fatalf("%s: no soulshift target prompt", tc.name)
		}
		if !hasID(p.PickTargetCards, fits) || hasID(p.PickTargetCards, tooBig) {
			t.Errorf("%s: soulshift offers %v, want the mana value %d Spirit %v", tc.name, p.PickTargetCards, tc.n, fits)
		}
		pickCard(t, g, me.ID, fits)
		answerLatestTriggerPrompt(t, g, me.ID, true)
		passPriorityAroundTable(t, g)
		if !me.Hand.Contains(fits) {
			t.Errorf("%s: soulshift did not return the Spirit card", tc.name)
		}
	}
}

// --- regeneration -----------------------------------------------------------------------

func TestSokB1RegeneratorsSurviveDestruction(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Kuro's Taken", "Creature — Rat Samurai", sokB1KurosTakenOracle},
		{"Skull Collector", "Creature — Ogre Warrior", sokB1SkullCollectorOracle},
	} {
		g := newCatalogGame(t)
		advanceToMain(t, g)
		me := g.Seats[g.Turn.ActiveSeat]
		id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 3, 3)
		if err := ravB2Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{}, "B"); err == nil {
			t.Fatalf("%s: regenerated for {B} alone", tc.name)
		}
		ravB2MustActivate(t, g, me.ID, id, 0, game.ActivateAbilityParams{}, "C")
		destroyIt(t, g, id)
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s was destroyed through its regeneration shield", tc.name)
		}
	}
}

// --- Kitsune Dawnblade ----------------------------------------------------------------------

func TestSokB1KitsuneDawnbladeMayTapACreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	sokB1CastAndTarget(t, g, "Kitsune Dawnblade", "Creature — Fox Samurai", sokB1DawnbladeOracle, 2, 3, game.TargetRef{Kind: game.TargetCard, ID: bear})
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !isTapped(g, bear) {
		t.Error("the target was not tapped")
	}
}

// --- the Onna cycle ---------------------------------------------------------------------------

func TestSokB1HaruOnnaDrawsAndReturns(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	hand := me.Hand.Size()
	onna := sokB1Cast(t, g, "Haru-Onna", "Creature — Spirit", sokB1HaruOracle, 2, 1)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one draw as it entered", hand, me.Hand.Size())
	}
	sokB1ExpectOnnaReturn(t, g, "Haru-Onna", onna)
}

func TestSokB1NikkoAndYukiOnnaDestroyTheirTarget(t *testing.T) {
	for _, tc := range []struct{ name, oracle, legalType, illegalType string }{
		{"Nikko-Onna", sokB1NikkoOracle, "Enchantment", "Artifact"},
		{"Yuki-Onna", sokB1YukiOracle, "Artifact", "Enchantment"},
	} {
		g := newCatalogGame(t)
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		legal := b12Permanent(g, opp.ID, "Legal", tc.legalType)
		illegal := b12Permanent(g, opp.ID, "Illegal", tc.illegalType)
		onna := sokB1Cast(t, g, tc.name, "Creature — Spirit", tc.oracle, 2, 2)
		me := g.Seats[g.Turn.ActiveSeat]
		p := passUntilPickTarget(t, g, me.ID)
		if p == nil {
			t.Fatalf("%s: no target prompt", tc.name)
		}
		if !hasID(p.PickTargetCards, legal) || hasID(p.PickTargetCards, illegal) {
			t.Errorf("%s offers %v, want only the %s", tc.name, p.PickTargetCards, tc.legalType)
		}
		pickCard(t, g, me.ID, legal)
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(legal) || !opp.Graveyard.Contains(legal) {
			t.Errorf("%s did not destroy the %s", tc.name, tc.legalType)
		}
		sokB1ExpectOnnaReturn(t, g, tc.name, onna)
	}
}

func TestSokB1KiriOnnaBouncesACreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	onna := sokB1CastAndTarget(t, g, "Kiri-Onna", "Creature — Spirit", sokB1KiriOracle, 2, 2, game.TargetRef{Kind: game.TargetCard, ID: bear})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Hand.Contains(bear) {
		t.Error("the target did not return to its owner's hand")
	}
	sokB1ExpectOnnaReturn(t, g, "Kiri-Onna", onna)
}

func TestSokB1KemuriOnnaMakesATargetPlayerDiscard(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	d2dSetHand(opp, 2)
	onna := sokB1CastAndTarget(t, g, "Kemuri-Onna", "Creature — Spirit", sokB1KemuriOracle, 3, 3, playerTarget(opp.ID))
	passPriorityAroundTable(t, g)
	c := discardChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("no discard prompt for the target player")
	}
	if c.ChooseMin != 1 || c.ChooseMax != 1 {
		t.Errorf("discard bounds [%d,%d], want [1,1]", c.ChooseMin, c.ChooseMax)
	}
	answerDiscard(t, g, opp.ID, opp.Hand.Cards[0].InstanceID)
	if opp.Hand.Size() != 1 {
		t.Errorf("the target holds %d cards, want 1", opp.Hand.Size())
	}
	passPriorityAroundTable(t, g)
	sokB1ExpectOnnaReturn(t, g, "Kemuri-Onna", onna)
}

// --- the upkeep returners ---------------------------------------------------------------

func TestSokB1UpkeepReturnsACreatureOfItsColour(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle, cost, sameCost string }{
		{"Stampeding Serow", "Creature — Antelope Beast", sokB1SerowOracle, "{2}{G}{G}", "{1}{G}"},
		{"Eiganjo Free-Riders", "Creature — Human Soldier", sokB1FreeRidersOracle, "{3}{W}", "{1}{W}"},
		{"Oni of Wild Places", "Creature — Demon Spirit", sokB1OniOracle, "{5}{R}", "{1}{R}"},
		{"Skull Collector", "Creature — Ogre Warrior", sokB1SkullCollectorOracle, "{1}{B}{B}", "{1}{B}"},
	} {
		g := newCatalogGame(t)
		me := g.Seats[1]
		id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 3, 3)
		findBattlefieldCardForTest(g, id).ManaCost = tc.cost
		same := b12Creature(g, me.ID, "Same Colour", "Creature — Bear", 2, 2)
		findBattlefieldCardForTest(g, same).ManaCost = tc.sameCost
		blue := b12Creature(g, me.ID, "Blue One", "Creature — Bird", 1, 1)
		findBattlefieldCardForTest(g, blue).ManaCost = "{1}{U}"
		advanceToUpkeepOf(t, g, 1)
		passPriorityAroundTable(t, g)
		p := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
		if p == nil {
			t.Fatalf("%s queued no return pick", tc.name)
		}
		if !hasID(p.ChooseCards, id) || !hasID(p.ChooseCards, same) || hasID(p.ChooseCards, blue) {
			t.Errorf("%s offered %v, want itself and the same-colour creature, not the blue one", tc.name, p.ChooseCards)
		}
		if err := g.ResolveOwnPermanents(p.ID, me.ID, []uuid.UUID{same}); err != nil {
			t.Fatalf("ResolveOwnPermanents: %v", err)
		}
		if !me.Hand.Contains(same) || !g.Battlefield.Contains(id) {
			t.Errorf("%s: the chosen creature did not return, or the wrong one did", tc.name)
		}
	}
}

// --- Glitterfang ----------------------------------------------------------------------------

func TestSokB1GlitterfangReturnsAtTheEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	id := pushCatalogPermanent(g, me.ID, "Glitterfang", "Creature — Spirit", sokB1GlitterfangOracle, true)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) || !me.Hand.Contains(id) {
		t.Error("Glitterfang did not return to its owner's hand at the end step")
	}
}

// --- lands and activated abilities --------------------------------------------------------

func TestSokB1SakuraTribeScoutPutsALandFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	scout := pushCatalogPermanent(g, me.ID, "Sakura-Tribe Scout", "Creature — Snake Shaman Scout", sokB1ScoutOracle, false)
	land := handCardForTest(me, "Forest", "Basic Land — Forest", "")
	rock := handCardForTest(me, "Some Rock", "Artifact", "")
	if err := g.ActivateCatalogAbility(me.ID, scout, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	if !tappedOnBattlefield(t, g, scout) {
		t.Error("the {T} cost was not paid")
	}
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no land prompt")
	}
	if pick.ChooseMin != 0 || hasID(pick.ChooseCards, rock) {
		t.Errorf("bounds from %d with %v offered, want an optional pick of land cards only", pick.ChooseMin, pick.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, land)
	if !g.Battlefield.Contains(land) {
		t.Fatal("the land did not reach the battlefield")
	}
	if n := g.LandsPlayedThisTurnFor(me.ID); n != 0 {
		t.Errorf("land drops used = %d, want 0", n)
	}
}

func TestSokB1OboroTapsForBlueAndReturnsToHand(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	oboro := pushCatalogPermanent(g, me.ID, "Oboro, Palace in the Clouds", "Legendary Land", sokB1OboroOracle, false)
	if err := g.ActivateManaAbility(me.ID, oboro, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if n := me.ManaPool.CountColor("U"); n != 1 || me.ManaPool.Total() != 1 {
		t.Errorf("pool holds %d {U} of %d, want 1 of 1", n, me.ManaPool.Total())
	}

	g = newCatalogGame(t)
	advanceToMain(t, g)
	me = g.Seats[g.Turn.ActiveSeat]
	oboro = pushCatalogPermanent(g, me.ID, "Oboro, Palace in the Clouds", "Legendary Land", sokB1OboroOracle, false)
	ravB2MustActivate(t, g, me.ID, oboro, 0, game.ActivateAbilityParams{}, "C")
	if g.Battlefield.Contains(oboro) || !me.Hand.Contains(oboro) {
		t.Error("Oboro did not return to its owner's hand")
	}
}

func TestSokB1OboroBreezecallerReturnsALandToUntapALand(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	caller := b12Push(g, me.ID, "Oboro Breezecaller", "Creature — Moonfolk Wizard", sokB1BreezecallerOracle, 1, 1)
	tapped := ccLand(g, me.ID, "Island", "Basic Land — Island")
	findBattlefieldCardForTest(g, tapped).Tapped = true
	back := ccLand(g, me.ID, "Plains", "Basic Land — Plains")
	if err := ravB2Activate(t, g, me.ID, caller, 0, game.ActivateAbilityParams{Targets: cardTarget(tapped)}, "C", "C"); err == nil {
		t.Fatal("the ability activated without returning a land")
	}
	ravB2MustActivate(t, g, me.ID, caller, 0, game.ActivateAbilityParams{Targets: cardTarget(tapped), ReturnIDs: []uuid.UUID{back}})
	if !me.Hand.Contains(back) {
		t.Error("the land paid for the cost did not return to its owner's hand")
	}
	if isTapped(g, tapped) {
		t.Error("the target land was not untapped")
	}
}

func TestSokB1PumpAbilities(t *testing.T) {
	// Akki Drillmaster: haste until end of turn.
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	drill := b12Push(g, me.ID, "Akki Drillmaster", "Creature — Goblin Shaman", sokB1DrillmasterOracle, 2, 2)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	ravB2MustActivate(t, g, me.ID, drill, 0, game.ActivateAbilityParams{Targets: cardTarget(bear)})
	if !isTapped(g, drill) || !hasKeywordOnBattlefield(t, g, bear, "haste") {
		t.Error("Drillmaster did not tap to give its target haste")
	}
	dcCleanup(g)
	if hasKeywordOnBattlefield(t, g, bear, "haste") {
		t.Error("haste outlived the cleanup step")
	}

	// Captive Flame: +1/+0 per activation.
	g = newCatalogGame(t)
	advanceToMain(t, g)
	me = g.Seats[g.Turn.ActiveSeat]
	flame := pushCatalogPermanent(g, me.ID, "Captive Flame", "Enchantment", sokB1CaptiveFlameOracle, false)
	bear = b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	ravB2MustActivate(t, g, me.ID, flame, 0, game.ActivateAbilityParams{Targets: cardTarget(bear)}, "R")
	ravB2MustActivate(t, g, me.ID, flame, 0, game.ActivateAbilityParams{Targets: cardTarget(bear)}, "R")
	if p, tough := effectivePTOf(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("the target is %d/%d, want 4/2 after two activations", p, tough)
	}
	dcCleanup(g)
	if p, _ := effectivePTOf(t, g, bear); p != 2 {
		t.Errorf("the bonus outlived the cleanup step (power %d)", p)
	}
}

// --- Spirit or Arcane cast triggers -------------------------------------------------------

func TestSokB1BriarknitKamiPutsACounterOnASpiritOrArcaneSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Briarknit Kami", "Creature — Spirit", sokB1BriarknitOracle, 3, 3)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("a Bear spell triggered Briarknit Kami")
	}
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "An Arcane Spell", "Instant — Arcane", "", nil)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if n := pr8Counters(g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("the target has %d +1/+1 counters, want 1", n)
	}
}

func TestSokB1DreamcatcherMaySacrificeToDraw(t *testing.T) {
	for _, yes := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		catcher := b12Push(g, me.ID, "Dreamcatcher", "Creature — Spirit", sokB1DreamcatcherOracle, 1, 1)
		advanceToMain(t, g)
		hand := me.Hand.Size()
		castCatalogSpell(t, g, "An Arcane Spell", "Sorcery — Arcane", "", nil)
		answerLatestTriggerPrompt(t, g, me.ID, yes)
		passPriorityAroundTable(t, g)
		if yes {
			if g.Battlefield.Contains(catcher) || !me.Graveyard.Contains(catcher) {
				t.Error("yes: the Dreamcatcher was not sacrificed")
			}
			if me.Hand.Size() != hand+1 {
				t.Errorf("yes: hand %d → %d, want one draw", hand, me.Hand.Size())
			}
			continue
		}
		if !g.Battlefield.Contains(catcher) || me.Hand.Size() != hand {
			t.Error("no: the Dreamcatcher was sacrificed or a card was drawn")
		}
	}
}

// --- the Ghost-Lit cards and channel -----------------------------------------------------

func TestSokB1GhostLitRedeemerGainsLife(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	redeemer := b12Push(g, me.ID, "Ghost-Lit Redeemer", "Creature — Spirit", sokB1RedeemerOracle, 1, 1)
	life := me.Life
	ravB2MustActivate(t, g, me.ID, redeemer, 0, game.ActivateAbilityParams{}, "W")
	if me.Life != life+2 || !isTapped(g, redeemer) {
		t.Errorf("life %d → %d, want +2 and the Redeemer tapped", life, me.Life)
	}

	g = newCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	life = me.Life
	id, err := neoB07Channel(t, g, "Ghost-Lit Redeemer", "Creature — Spirit", sokB1RedeemerOracle, "{C}{W}", 1, game.ActivateAbilityParams{})
	if err != nil {
		t.Fatalf("channel: %v", err)
	}
	if !me.Graveyard.Contains(id) {
		t.Error("the discard is a cost: the card must be in the graveyard")
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Errorf("channel: life %d → %d, want +4", life, me.Life)
	}
}

func TestSokB1GhostLitRaiderDealsDamage(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	raider := b12Push(g, me.ID, "Ghost-Lit Raider", "Creature — Spirit", sokB1RaiderOracle, 2, 1)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	ravB2MustActivate(t, g, me.ID, raider, 0, game.ActivateAbilityParams{Targets: cardTarget(bear)}, "C", "C", "R")
	if g.Battlefield.Contains(bear) {
		t.Error("the 2/2 survived 2 damage")
	}

	g = newCatalogGame(t)
	opp = g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	giant := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 4, 4)
	if _, err := neoB07Channel(t, g, "Ghost-Lit Raider", "Creature — Spirit", sokB1RaiderOracle, "{C}{C}{C}{R}", 1,
		game.ActivateAbilityParams{Targets: cardTarget(giant)}); err != nil {
		t.Fatalf("channel: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(giant) {
		t.Error("the 4/4 survived the channel's 4 damage")
	}
}

func TestSokB1ShinenOfFurysFireChannelGrantsHaste(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if _, err := neoB07Channel(t, g, "Shinen of Fury's Fire", "Creature — Spirit", sokB1FurysFireOracle, "{R}", 0,
		game.ActivateAbilityParams{Targets: cardTarget(bear)}); err != nil {
		t.Fatalf("channel: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasKeywordOnBattlefield(t, g, bear, "haste") {
		t.Error("the target lacks haste")
	}
}

// --- "the number of cards in your hand" ------------------------------------------------------

// sokB1HandOf moves to the active player's main phase and gives them n
// cards in hand.
func sokB1HandOf(t *testing.T, g *game.Game, n int) *game.Player {
	t.Helper()
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	d2dSetHand(me, n)
	return me
}

func TestSokB1InnerFireAddsRedPerCardInHand(t *testing.T) {
	g := newCatalogGame(t)
	me := sokB1HandOf(t, g, 3)
	castCatalogSpell(t, g, "Inner Fire", "Sorcery", sokB1InnerFireOracle, nil)
	passPriorityAroundTable(t, g)
	if n := me.ManaPool.CountColor("R"); n != 3 || me.ManaPool.Total() != 3 {
		t.Errorf("pool holds %d {R} of %d, want 3 of 3", n, me.ManaPool.Total())
	}
}

func TestSokB1InnerCalmOuterStrengthPumpsByHandSize(t *testing.T) {
	g := newCatalogGame(t)
	me := sokB1HandOf(t, g, 3)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Inner Calm, Outer Strength", "Instant — Arcane", sokB1InnerCalmOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("the target is %d/%d, want 5/5 with three cards in hand", p, tough)
	}
	dcCleanup(g)
	if p, tough := effectivePTOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("the bonus outlived the cleanup step (%d/%d)", p, tough)
	}
}

func TestSokB1SpiralingEmbersAndGazeOfAdamaroCountHands(t *testing.T) {
	g := newCatalogGame(t)
	sokB1HandOf(t, g, 3)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	life := opp.Life
	castCatalogSpell(t, g, "Spiraling Embers", "Sorcery — Arcane", sokB1SpiralingEmbersOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("Spiraling Embers: life %d → %d, want 3 damage for your three cards", life, opp.Life)
	}

	g = newCatalogGame(t)
	sokB1HandOf(t, g, 1)
	opp = g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	d2dSetHand(opp, 4)
	life = opp.Life
	castCatalogSpell(t, g, "Gaze of Adamaro", "Instant — Arcane", sokB1GazeOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("Gaze of Adamaro: life %d → %d, want 4 damage for their four cards", life, opp.Life)
	}
}

func TestSokB1PresenceOfTheWiseGainsTwoPerCard(t *testing.T) {
	g := newCatalogGame(t)
	me := sokB1HandOf(t, g, 3)
	life := me.Life
	castCatalogSpell(t, g, "Presence of the Wise", "Sorcery", sokB1PresenceOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Life != life+6 {
		t.Errorf("life %d → %d, want +6 for three cards", life, me.Life)
	}
}

func TestSokB1OneWithNothingDiscardsTheHand(t *testing.T) {
	g := newCatalogGame(t)
	me := sokB1HandOf(t, g, 3)
	grave := me.Graveyard.Size()
	castCatalogSpell(t, g, "One with Nothing", "Instant", sokB1OneWithNothingOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != 0 {
		t.Errorf("hand holds %d cards, want none", me.Hand.Size())
	}
	if me.Graveyard.Size() != grave+4 {
		t.Errorf("graveyard %d → %d, want the three cards and the spell", grave, me.Graveyard.Size())
	}
}

// --- other spells -------------------------------------------------------------------------

func TestSokB1SeekTheHorizonFindsUpToThreeBasics(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var basics []uuid.UUID
	for _, n := range []string{"Forest", "Island", "Swamp", "Plains"} {
		basics = append(basics, pushLibraryCardForTest(me, game.Card{Name: n, TypeLine: "Basic Land — " + n}))
	}
	dual := pushLibraryCardForTest(me, game.Card{Name: "Some Dual", TypeLine: "Land — Forest Island"})
	castCatalogSpell(t, g, "Seek the Horizon", "Sorcery", sokB1SeekOracle, nil)
	passPriorityAroundTable(t, g)
	open := chkB1OpenSearch(g)
	if open == nil {
		t.Fatal("no search prompt")
	}
	beginSearchIfAsked(t, g, open)
	open = chkB1OpenSearch(g)
	if hasID(open.SearchCards, dual) || open.SearchMax != 3 {
		t.Errorf("the search offers %v up to %d, want basic land cards only, up to three", open.SearchCards, open.SearchMax)
	}
	takeSearchPrompts(t, g)
	found := 0
	for _, id := range basics {
		if me.Hand.Contains(id) {
			found++
		}
	}
	if found != 3 {
		t.Errorf("%d basic land cards reached the hand, want 3", found)
	}
}

func TestSokB1DosansOldestChantGainsSixAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	life, hand := me.Life, me.Hand.Size()
	castCatalogSpell(t, g, "Dosan's Oldest Chant", "Sorcery", sokB1DosanOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Life != life+6 || me.Hand.Size() != hand+1 {
		t.Errorf("life %d → %d, hand %d → %d; want +6 and one draw", life, me.Life, hand, me.Hand.Size())
	}
}

func TestSokB1SunderFromWithinDestroysAnArtifactOrLand(t *testing.T) {
	for _, legalType := range []string{"Artifact", "Land"} {
		g := newCatalogGame(t)
		_, opp := seatsForCombat(g)
		legal := b12Permanent(g, opp.ID, "Legal", legalType)
		bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		if err := castCatalogSpellErr(t, g, "Sunder from Within", "Sorcery — Arcane", sokB1SunderOracle, cardTarget(bear)); err == nil {
			t.Error("Sunder from Within targeted a creature")
		}
		castCatalogSpell(t, g, "Sunder from Within", "Sorcery — Arcane", sokB1SunderOracle, cardTarget(legal))
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(legal) || !opp.Graveyard.Contains(legal) {
			t.Errorf("the %s was not destroyed", legalType)
		}
	}
}

func TestSokB1PathOfAngersFlamePumpsYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Path of Anger's Flame", "Instant — Arcane", sokB1PathOfAngersOracle, nil)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, mine); p != 4 || tough != 2 {
		t.Errorf("your creature is %d/%d, want 4/2", p, tough)
	}
	if p, tough := effectivePTOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("an opponent's creature is %d/%d, want 2/2", p, tough)
	}
}
