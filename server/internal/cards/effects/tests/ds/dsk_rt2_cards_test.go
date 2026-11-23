package testsds

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dsk_rt2_cards_test.go — Duskmourn: House of Horror, the second R–T
// batch (#4466). Each test plays the real catalog entry and asserts what
// its text says.

const (
	dskRT2PlaymateOracle    = "d0732ebf-91ad-4e3c-a09a-f410d20587fc"
	dskRT2CultistOracle     = "8b6d37a7-a88e-45dd-b45e-642a44caaf05"
	dskRT2MothOracle        = "a7fea839-d7d7-48b4-aeaf-133767532e32"
	dskRT2CentipedeOracle   = "dcfa0476-e88f-43b4-8011-d78a3de97f2a"
	dskRT2DollOracle        = "934e2634-4b1c-4d92-b4d6-9241620f3770"
	dskRT2SawOracle         = "d88bfa98-6869-4d64-a9d2-07a308a67fe2"
	dskRT2SwarmweaverOracle = "50120e90-2fdf-43c1-93e2-2230ee4e63fc"
	dskRT2TrappedOracle     = "20f2fc0e-b1d4-4f80-9c7a-a9fb25fef7b8"
	dskRT2RescueOracle      = "406247ee-2b86-4068-b5fa-65cdc83a82f4"
	dskRT2TyvarOracle       = "5a242eab-204c-4411-860f-c0f3582a8f38"
	dskRT2RideOracle        = "aed1f0cd-8df8-415f-afd6-ccfd12234334"
	dskRT2BalloonOracle     = "82ea519f-0e14-43d3-bf46-a1c5352e9600"
)

func TestDskRT2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		dskRT2PlaymateOracle:    "Ragged Playmate",
		dskRT2CultistOracle:     "Resurrected Cultist",
		dskRT2MothOracle:        "Rite of the Moth",
		dskRT2CentipedeOracle:   "Spineseeker Centipede",
		dskRT2DollOracle:        "Splitskin Doll",
		dskRT2SawOracle:         "Saw",
		dskRT2SwarmweaverOracle: "The Swarmweaver",
		dskRT2TrappedOracle:     "Trapped in the Screen",
		dskRT2RescueOracle:      "Shardmage's Rescue",
		dskRT2TyvarOracle:       "Tyvar, the Pummeler",
		dskRT2RideOracle:        "The Rollercrusher Ride",
		dskRT2BalloonOracle:     "The Jolly Balloon Man",
	} {
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

// --- Ragged Playmate -------------------------------------------------------------------

func TestDskRT2RaggedPlaymateMakesASmallCreatureUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	playmate := effectstest.B12Push(g, me.ID, "Ragged Playmate", "Artifact Creature — Toy", dskRT2PlaymateOracle, 2, 2)
	small := effectstest.B12Creature(g, me.ID, "Small Bear", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, me.ID, "Big Bear", "Creature — Bear", 3, 3)
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.DskRT2Activate(g, me.ID, playmate, 0, "{C}", game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(big)}); err == nil {
		t.Fatal("a creature with power 3 was accepted as the target")
	}
	if err := effectstest.DskRT2Activate(g, me.ID, playmate, 0, "{C}", game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(small)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertRestrictions(t, g, small, game.CantBeBlocked)
	if !effectstest.DskRT2Tapped(g, playmate) {
		t.Error("the Playmate did not tap to pay the cost")
	}
}

// --- Resurrected Cultist ---------------------------------------------------------------

func TestDskRT2ResurrectedCultistReturnsOnlyWithDelirium(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	cultist := effectstest.PushCatalogGraveyardCard(me, "Resurrected Cultist", "Creature — Human Cleric", dskRT2CultistOracle, 4, 1)
	// The Cultist is the creature card; three more types make four.
	effectstest.GyCard(me, "Old Bolt", "Instant", "{R}", "")
	effectstest.GyCard(me, "Old Ritual", "Sorcery", "{B}", "")
	if err := effectstest.DskRT2Activate(g, me.ID, cultist, 0, "{B}{B}{C}{C}", game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated with three card types in the graveyard")
	}
	effectstest.GyCard(me, "Old Forest", "Basic Land — Forest", "", "")
	if err := effectstest.DskRT2Activate(g, me.ID, cultist, 0, "{B}{B}{C}{C}", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate with delirium: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cultist) {
		t.Fatal("the Cultist did not return")
	}
	if got := effectstest.CountersOn(g, cultist, game.CounterFinality); got != 1 {
		t.Errorf("finality counters = %d, want 1", got)
	}
}

// --- Rite of the Moth -------------------------------------------------------------------

func TestDskRT2RiteOfTheMothReturnsACreatureWithAFinalityCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.PushCatalogGraveyardCard(me, "Old Bear", "Creature — Bear", "", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Rite of the Moth", "Sorcery", dskRT2MothOracle, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the creature card did not return")
	}
	if got := effectstest.CountersOn(g, bear, game.CounterFinality); got != 1 {
		t.Errorf("finality counters = %d, want 1", got)
	}
	effectstest.RooncDestroy(t, g, bear)
	if me.Graveyard.Contains(bear) || !g.Exile.Contains(bear) {
		t.Error("the creature with a finality counter went to the graveyard instead of exile")
	}
}

func TestDskRT2RiteOfTheMothHasFlashback(t *testing.T) {
	spec, ok := effects.Lookup(dskRT2MothOracle)
	if !ok {
		t.Fatal("not registered")
	}
	if len(spec.AlternativeCosts) != 1 || spec.AlternativeCosts[0].Key != effects.Flashback("{3}{W}{W}{B}").Key ||
		spec.AlternativeCosts[0].ManaCost != "{3}{W}{W}{B}" {
		t.Errorf("alternative costs %+v, want flashback {3}{W}{W}{B}", spec.AlternativeCosts)
	}
}

// --- Spineseeker Centipede ----------------------------------------------------------------

func TestDskRT2SpineseekerCentipedeFetchesABasicLandToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Bolt", TypeLine: "Instant"},
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
	)
	effectstest.AdvanceToMain(t, g)
	effectstest.DftABCast(t, g, me, game.Card{Name: "Spineseeker Centipede", TypeLine: "Creature — Insect",
		OracleID: dskRT2CentipedeOracle, Power: 2, Toughness: 1})
	effectstest.AnswerSearchNamed(t, g, me.ID, "Forest")
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[1]) {
		t.Error("the basic land did not reach the hand")
	}
	if me.Hand.Contains(ids[0]) {
		t.Error("a nonland card was fetched")
	}
}

func TestDskRT2SpineseekerCentipedeGrowsAndGainsVigilanceWithDelirium(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	centipede := effectstest.B12Push(g, me.ID, "Spineseeker Centipede", "Creature — Insect", dskRT2CentipedeOracle, 2, 1)
	effectstest.DskRT2Delirium(g, me, 3)
	if p, tt := effectstest.PtOf(t, g, centipede); p != 2 || tt != 1 {
		t.Errorf("three card types: %d/%d, want 2/1", p, tt)
	}
	if effectstest.HasEffectiveKeyword(t, g, centipede, "vigilance") {
		t.Error("three card types: vigilance")
	}
	effectstest.DskRT2Delirium(g, me, 4)
	if p, tt := effectstest.PtOf(t, g, centipede); p != 3 || tt != 3 {
		t.Errorf("delirium: %d/%d, want 3/3", p, tt)
	}
	if !effectstest.HasEffectiveKeyword(t, g, centipede, "vigilance") {
		t.Error("delirium: no vigilance")
	}
}

// --- Splitskin Doll ---------------------------------------------------------------------

func TestDskRT2SplitskinDollDrawsWithoutDiscardingBesideASmallCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Creature(g, me.ID, "Small Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	effectstest.DftABCast(t, g, me, game.Card{Name: "Splitskin Doll", TypeLine: "Artifact Creature — Toy",
		OracleID: dskRT2DollOracle, Power: 2, Toughness: 1})
	if effectstest.DiscardChoiceFor(g, me.ID) != nil {
		t.Fatal("asked to discard while controlling another creature with power 2 or less")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn", hand, me.Hand.Size())
	}
}

func TestDskRT2SplitskinDollLootsWithoutAnotherSmallCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	// The Doll itself is a 2/1 and does not count; a 3/3 does not either.
	effectstest.B12Creature(g, me.ID, "Big Bear", "Creature — Bear", 3, 3)
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	effectstest.DftABCast(t, g, me, game.Card{Name: "Splitskin Doll", TypeLine: "Artifact Creature — Toy",
		OracleID: dskRT2DollOracle, Power: 2, Toughness: 1})
	if effectstest.DiscardOwed(g, me.ID) != 1 {
		t.Fatalf("discard owed %d, want 1", effectstest.DiscardOwed(g, me.ID))
	}
	effectstest.AnswerDiscard(t, g, me.ID, me.Hand.Cards[0].InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want a draw and a discard", hand, me.Hand.Size())
	}
}

// --- Saw ------------------------------------------------------------------------------

func TestDskRT2SawPumpsAndSacrificesAnotherPermanentToDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	land := effectstest.PushTypedCard(g, me.ID, "My Land", "Land", "")
	saw := effectstest.SeedEquipment(g, me.ID, "Saw", dskRT2SawOracle)
	effectstest.BrAttach(t, g, saw, bear)
	if p, tt := effectstest.PtOf(t, g, bear); p != 4 || tt != 2 {
		t.Errorf("equipped creature is %d/%d, want 4/2", p, tt)
	}
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if prompt == nil {
		t.Fatal("no sacrifice prompt")
	}
	if prompt.ChooseMin != 0 || !effectstest.HasID(prompt.ChooseCards, land) || effectstest.HasID(prompt.ChooseCards, bear) || effectstest.HasID(prompt.ChooseCards, saw) {
		t.Fatalf("offered %v (min %d): want the land, not the attacker or the Saw, with a floor of zero",
			prompt.ChooseCards, prompt.ChooseMin)
	}
	hand := me.Hand.Size()
	effectstest.AnswerOwnPermanents(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the land was not sacrificed")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn", hand, me.Hand.Size())
	}
}

func TestDskRT2SawDecliningDrawsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	land := effectstest.PushTypedCard(g, me.ID, "My Land", "Land", "")
	saw := effectstest.SeedEquipment(g, me.ID, "Saw", dskRT2SawOracle)
	effectstest.BrAttach(t, g, saw, bear)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.AnswerOwnPermanents(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(land) || me.Hand.Size() != hand {
		t.Error("declining sacrificed something or drew a card")
	}
}

// --- The Swarmweaver ---------------------------------------------------------------------

func TestDskRT2SwarmweaverMakesInsectsAndArmsThemWithDelirium(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	spider := effectstest.B12Creature(g, me.ID, "My Spider", "Creature — Spider", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Insect", "Creature — Insect", 1, 1)
	effectstest.AdvanceToMain(t, g)
	weaver := effectstest.DftABCast(t, g, me, game.Card{Name: "The Swarmweaver", TypeLine: "Legendary Artifact Creature — Scarecrow",
		OracleID: dskRT2SwarmweaverOracle, Power: 2, Toughness: 3})
	var insects []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if effects.IsToken(c) && c.Name == "Insect" && c.Controller == me.ID {
			insects = append(insects, c.InstanceID)
		}
	}
	if len(insects) != 2 {
		t.Fatalf("%d Insect tokens, want 2", len(insects))
	}
	tok, _ := effectstest.BattlefieldCard(g, insects[0])
	if !slices.Equal(tok.Colors, []string{"B", "G"}) {
		t.Errorf("the Insect is %v, want black and green", tok.Colors)
	}
	if p, tt := effectstest.PtOf(t, g, insects[0]); p != 1 || tt != 1 || !effectstest.HasEffectiveKeyword(t, g, insects[0], "flying") {
		t.Errorf("without delirium the Insect is %d/%d, want a 1/1 flyer", p, tt)
	}

	effectstest.DskRT2Delirium(g, me, 4)
	for _, id := range append(insects, spider) {
		want := 2
		if id == spider {
			want = 3
		}
		if p, tt := effectstest.PtOf(t, g, id); p != want || tt != want {
			t.Errorf("an Insect or Spider you control is %d/%d, want %d/%d", p, tt, want, want)
		}
		if !effectstest.HasEffectiveKeyword(t, g, id, "deathtouch") {
			t.Error("an Insect or Spider you control lacks deathtouch")
		}
	}
	for _, id := range []uuid.UUID{bear, weaver, theirs} {
		if effectstest.HasEffectiveKeyword(t, g, id, "deathtouch") {
			t.Error("a creature outside the clause gained deathtouch")
		}
	}
	if p, _ := effectstest.PtOf(t, g, theirs); p != 1 {
		t.Errorf("an opponent's Insect has power %d, want 1", p)
	}
}

// --- Trapped in the Screen -------------------------------------------------------------

func TestDskRT2TrappedInTheScreenExilesUntilItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirLand := effectstest.PushTypedCard(g, opp.ID, "Their Land", "Land", "")
	theirRock := effectstest.PushTypedCard(g, opp.ID, "Their Rock", "Artifact", "{2}")
	trapped := effectstest.CastCatalogSpell(t, g, "Trapped in the Screen", "Enchantment", dskRT2TrappedOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirLand}); err == nil {
		t.Fatal("a land was accepted as the target")
	}
	effectstest.PickCard(t, g, me.ID, theirRock)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirRock) || !g.Exile.Contains(theirRock) {
		t.Fatal("the artifact was not exiled")
	}
	effectstest.RooncDestroy(t, g, trapped)
	g.RunStateChecksForTest()
	// It returns as a new object (CR 400.7), so it is found by name.
	if n := effectstest.NamedOnBattlefield(g, "Their Rock", opp.ID); n != 1 {
		t.Errorf("after the enchantment left, %d copies under its owner's control, want 1", n)
	}
}

func TestDskRT2TrappedInTheScreenHasWardTwo(t *testing.T) {
	spec, ok := effects.Lookup(dskRT2TrappedOracle)
	if !ok {
		t.Fatal("not registered")
	}
	if len(spec.Triggered) < 2 || spec.Triggered[0].Key != "Trapped in the Screen — ward {2}" {
		t.Errorf("first trigger %q, want the ward {2} trigger", spec.Triggered[0].Key)
	}
}

// --- Shardmage's Rescue ------------------------------------------------------------------

func TestDskRT2ShardmagesRescueGrantsHexproofOnlyTheTurnItEntered(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	if _, err := effectstest.Ecl2CastAt(t, g, "Shardmage's Rescue", "Enchantment — Aura", dskRT2RescueOracle, effectstest.B16TargetCard(theirs)); err == nil {
		t.Fatal("enchant creature you control: an opponent's creature was accepted")
	}
	aura, err := effectstest.Ecl2CastAt(t, g, "Shardmage's Rescue", "Enchantment — Aura", dskRT2RescueOracle, effectstest.B16TargetCard(bear))
	if err != nil {
		t.Fatalf("flash: cast in the end step: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(aura) {
		t.Fatal("the Aura did not resolve")
	}
	if p, tt := effectstest.PtOf(t, g, bear); p != 3 || tt != 3 {
		t.Errorf("enchanted creature is %d/%d, want 3/3", p, tt)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("the Aura entered this turn: no hexproof")
	}
	effectstest.EndTurn(t, g)
	if p, tt := effectstest.PtOf(t, g, bear); p != 3 || tt != 3 {
		t.Errorf("next turn the enchanted creature is %d/%d, want 3/3", p, tt)
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("hexproof outlasted the turn the Aura entered")
	}
}

// --- Tyvar, the Pummeler -------------------------------------------------------------------

func TestDskRT2TyvarTapsAnotherCreatureToGainIndestructibleAndTaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tyvar := effectstest.PushCatalogPermanent(g, me.ID, "Tyvar, the Pummeler", "Legendary Creature — Elf Warrior", dskRT2TyvarOracle, true)
	if err := effectstest.DskRT2Activate(g, me.ID, tyvar, 0, "", game.ActivateAbilityParams{TapIDs: []uuid.UUID{tyvar}}); err == nil {
		t.Fatal("Tyvar paid the cost by tapping itself")
	}
	bear := effectstest.PushCatalogPermanent(g, me.ID, "Fresh Bear", "Creature — Bear", "", true)
	if err := effectstest.DskRT2Activate(g, me.ID, tyvar, 0, "", game.ActivateAbilityParams{TapIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("tap a summoning-sick creature for a summoning-sick Tyvar: %v", err)
	}
	if !effectstest.DskRT2Tapped(g, bear) {
		t.Error("the other creature was not tapped as the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, tyvar, "indestructible") {
		t.Error("Tyvar did not gain indestructible")
	}
	if !effectstest.DskRT2Tapped(g, tyvar) {
		t.Error("Tyvar did not tap")
	}
	effectstest.EndTurn(t, g)
	if effectstest.HasEffectiveKeyword(t, g, tyvar, "indestructible") {
		t.Error("indestructible outlasted the turn")
	}
}

func TestDskRT2TyvarPumpsTheTeamByTheGreatestPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	tyvar := effectstest.B12Push(g, me.ID, "Tyvar, the Pummeler", "Legendary Creature — Elf Warrior", dskRT2TyvarOracle, 3, 3)
	big := effectstest.B12Creature(g, me.ID, "Big Bear", "Creature — Bear", 5, 5)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 7, 7)
	if err := effectstest.DskRT2Activate(g, me.ID, tyvar, 1, "{G}{G}{C}{C}{C}", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, tyvar); p != 8 || tt != 8 {
		t.Errorf("Tyvar is %d/%d, want 8/8 (+5/+5)", p, tt)
	}
	if p, tt := effectstest.PtOf(t, g, big); p != 10 || tt != 10 {
		t.Errorf("the Bear is %d/%d, want 10/10", p, tt)
	}
	if p, _ := effectstest.PtOf(t, g, theirs); p != 7 {
		t.Errorf("an opponent's creature has power %d, want 7", p)
	}
	effectstest.EndTurn(t, g)
	if p, _ := effectstest.PtOf(t, g, big); p != 5 {
		t.Errorf("the pump outlasted the turn: power %d", p)
	}
}

// --- The Rollercrusher Ride ----------------------------------------------------------------

func dskRT2RideCard() game.Card {
	return game.Card{Name: "The Rollercrusher Ride", TypeLine: "Legendary Enchantment", OracleID: dskRT2RideOracle, ManaCost: "{X}{2}{R}"}
}

func TestDskRT2RollercrusherRideDealsXToEachOfUpToXCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := effectstest.B12Creature(g, opp.ID, "Giant A", "Creature — Giant", 5, 5)
	b := effectstest.B12Creature(g, opp.ID, "Giant B", "Creature — Giant", 5, 5)
	c := effectstest.B12Creature(g, opp.ID, "Giant C", "Creature — Giant", 5, 5)
	effectstest.RooncCastMain(t, g, dskRT2RideCard(), game.CastSpellParams{XValue: 2})
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{effectstest.RooncRef(a), effectstest.RooncRef(b), effectstest.RooncRef(c)}); err == nil {
		t.Fatal("'up to X' with X=2 accepted three targets")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{effectstest.RooncRef(a), effectstest.RooncRef(b)}); err != nil {
		t.Fatalf("pick two: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DamageMarkedOn(g, a) != 2 || effectstest.DamageMarkedOn(g, b) != 2 || effectstest.DamageMarkedOn(g, c) != 0 {
		t.Errorf("damage %d/%d/%d, want 2/2/0", effectstest.DamageMarkedOn(g, a), effectstest.DamageMarkedOn(g, b), effectstest.DamageMarkedOn(g, c))
	}
}

func TestDskRT2RollercrusherRideDoublesNoncombatDamageWithDelirium(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	giant := effectstest.B12Creature(g, opp.ID, "Giant", "Creature — Giant", 7, 7)
	effectstest.DskRT2Delirium(g, me, 4)
	effectstest.RooncCastMain(t, g, dskRT2RideCard(), game.CastSpellParams{XValue: 1})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, giant)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, giant); got != 2 {
		t.Errorf("the Ride's own noncombat damage with delirium: %d, want 2", got)
	}
	life := me.Life
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerKeyedForEffect(giant, me.ID, 3, game.KeyedTail{}); err != nil {
			t.Fatalf("deal damage: %v", err)
		}
	})
	if me.Life != life-3 {
		t.Errorf("a source an opponent controls was doubled: life %d -> %d", life, me.Life)
	}
}

func TestDskRT2RollercrusherRideDoesNotDoubleWithoutDelirium(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	giant := effectstest.B12Creature(g, opp.ID, "Giant", "Creature — Giant", 7, 7)
	effectstest.DskRT2Delirium(g, me, 3)
	effectstest.RooncCastMain(t, g, dskRT2RideCard(), game.CastSpellParams{XValue: 1})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, giant)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, giant); got != 1 {
		t.Errorf("three card types: %d damage, want 1", got)
	}
}

// --- The Jolly Balloon Man ------------------------------------------------------------------

func TestDskRT2JollyBalloonManMakesAOneOneRedFlyingBalloonCopyThatIsSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	balloonMan := effectstest.B12Push(g, me.ID, "The Jolly Balloon Man", "Legendary Creature — Human Clown", dskRT2BalloonOracle, 1, 4)
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Green Bear", TypeLine: "Creature — Bear", Colors: []string{"G"},
		Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	if err := effectstest.DskRT2Activate(g, me.ID, balloonMan, 0, "{C}", game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(bear)}); err == nil {
		t.Fatal("activated outside a main phase")
	}
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.DskRT2Activate(g, me.ID, balloonMan, 0, "{C}", game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(balloonMan)}); err == nil {
		t.Fatal("the Balloon Man targeted itself")
	}
	if err := effectstest.DskRT2Activate(g, me.ID, balloonMan, 0, "{C}", game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var token uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if effects.IsToken(c) && c.Name == "Green Bear" {
			token = c.InstanceID
		}
	}
	if token == uuid.Nil {
		t.Fatal("no token copy")
	}
	tok, _ := effectstest.BattlefieldCard(g, token)
	if !tok.HasSubtype("Balloon") || !tok.HasSubtype("Bear") {
		t.Errorf("token type line %q, want a Bear Balloon", tok.TypeLine)
	}
	if !slices.Contains(tok.Colors, "R") || !slices.Contains(tok.Colors, "G") {
		t.Errorf("token colours %v, want red in addition to green", tok.Colors)
	}
	if p, tt := effectstest.PtOf(t, g, token); p != 1 || tt != 1 {
		t.Errorf("token is %d/%d, want 1/1", p, tt)
	}
	for _, kw := range []string{"flying", "haste"} {
		if !effectstest.HasEffectiveKeyword(t, g, token, kw) {
			t.Errorf("token lacks %s", kw)
		}
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(token) {
		t.Error("the token was not sacrificed at the end step")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the original was sacrificed")
	}
}
