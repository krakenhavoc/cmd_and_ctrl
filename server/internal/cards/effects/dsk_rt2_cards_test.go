package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

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

// dskRT2Delirium puts the first n of four card types into p's graveyard
// (four is delirium, three is not) and drops the layer cache, as a zone
// move would.
func dskRT2Delirium(g *game.Game, p *game.Player, n int) {
	g.WithWriteLock(func() {
		dskHL2SeedTypes(p, n)
		g.BumpLayerVersionForTest()
	})
}

// dskRT2Activate pays `mana` and activates ability idx of a permanent.
func dskRT2Activate(g *game.Game, who, id uuid.UUID, idx int, mana string, params game.ActivateAbilityParams) error {
	if mana != "" {
		if err := g.AddManaForEffect(who, uuid.Nil, mana); err != nil {
			return err
		}
	}
	params.Strict = true
	return g.ActivateCatalogAbility(who, id, idx, params)
}

// dskRT2Tapped reports whether a battlefield permanent is tapped.
func dskRT2Tapped(g *game.Game, id uuid.UUID) bool {
	c, ok := battlefieldCard(g, id)
	return ok && c.Tapped
}

// --- Ragged Playmate -------------------------------------------------------------------

func TestDskRT2RaggedPlaymateMakesASmallCreatureUnblockable(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	playmate := b12Push(g, me.ID, "Ragged Playmate", "Artifact Creature — Toy", dskRT2PlaymateOracle, 2, 2)
	small := b12Creature(g, me.ID, "Small Bear", "Creature — Bear", 2, 2)
	big := b12Creature(g, me.ID, "Big Bear", "Creature — Bear", 3, 3)
	advanceToMain(t, g)
	if err := dskRT2Activate(g, me.ID, playmate, 0, "{C}", game.ActivateAbilityParams{Targets: b16TargetCard(big)}); err == nil {
		t.Fatal("a creature with power 3 was accepted as the target")
	}
	if err := dskRT2Activate(g, me.ID, playmate, 0, "{C}", game.ActivateAbilityParams{Targets: b16TargetCard(small)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	assertRestrictions(t, g, small, game.CantBeBlocked)
	if !dskRT2Tapped(g, playmate) {
		t.Error("the Playmate did not tap to pay the cost")
	}
}

// --- Resurrected Cultist ---------------------------------------------------------------

func TestDskRT2ResurrectedCultistReturnsOnlyWithDelirium(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	cultist := pushCatalogGraveyardCard(me, "Resurrected Cultist", "Creature — Human Cleric", dskRT2CultistOracle, 4, 1)
	// The Cultist is the creature card; three more types make four.
	gyCard(me, "Old Bolt", "Instant", "{R}", "")
	gyCard(me, "Old Ritual", "Sorcery", "{B}", "")
	if err := dskRT2Activate(g, me.ID, cultist, 0, "{B}{B}{C}{C}", game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated with three card types in the graveyard")
	}
	gyCard(me, "Old Forest", "Basic Land — Forest", "", "")
	if err := dskRT2Activate(g, me.ID, cultist, 0, "{B}{B}{C}{C}", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate with delirium: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cultist) {
		t.Fatal("the Cultist did not return")
	}
	if got := countersOn(g, cultist, game.CounterFinality); got != 1 {
		t.Errorf("finality counters = %d, want 1", got)
	}
}

// --- Rite of the Moth -------------------------------------------------------------------

func TestDskRT2RiteOfTheMothReturnsACreatureWithAFinalityCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := pushCatalogGraveyardCard(me, "Old Bear", "Creature — Bear", "", 2, 2)
	castCatalogSpell(t, g, "Rite of the Moth", "Sorcery", dskRT2MothOracle, b16TargetCard(bear))
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the creature card did not return")
	}
	if got := countersOn(g, bear, game.CounterFinality); got != 1 {
		t.Errorf("finality counters = %d, want 1", got)
	}
	rooncDestroy(t, g, bear)
	if me.Graveyard.Contains(bear) || !g.Exile.Contains(bear) {
		t.Error("the creature with a finality counter went to the graveyard instead of exile")
	}
}

func TestDskRT2RiteOfTheMothHasFlashback(t *testing.T) {
	spec, ok := Lookup(dskRT2MothOracle)
	if !ok {
		t.Fatal("not registered")
	}
	if len(spec.AlternativeCosts) != 1 || spec.AlternativeCosts[0].Key != Flashback("{3}{W}{W}{B}").Key ||
		spec.AlternativeCosts[0].ManaCost != "{3}{W}{W}{B}" {
		t.Errorf("alternative costs %+v, want flashback {3}{W}{W}{B}", spec.AlternativeCosts)
	}
}

// --- Spineseeker Centipede ----------------------------------------------------------------

func TestDskRT2SpineseekerCentipedeFetchesABasicLandToHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := seedSearchLibrary(me,
		game.Card{Name: "Bolt", TypeLine: "Instant"},
		searchTestLand("Forest", "Basic Land — Forest"),
	)
	advanceToMain(t, g)
	dftABCast(t, g, me, game.Card{Name: "Spineseeker Centipede", TypeLine: "Creature — Insect",
		OracleID: dskRT2CentipedeOracle, Power: 2, Toughness: 1})
	answerSearchNamed(t, g, me.ID, "Forest")
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[1]) {
		t.Error("the basic land did not reach the hand")
	}
	if me.Hand.Contains(ids[0]) {
		t.Error("a nonland card was fetched")
	}
}

func TestDskRT2SpineseekerCentipedeGrowsAndGainsVigilanceWithDelirium(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	centipede := b12Push(g, me.ID, "Spineseeker Centipede", "Creature — Insect", dskRT2CentipedeOracle, 2, 1)
	dskRT2Delirium(g, me, 3)
	if p, tt := ptOf(t, g, centipede); p != 2 || tt != 1 {
		t.Errorf("three card types: %d/%d, want 2/1", p, tt)
	}
	if hasEffectiveKeyword(t, g, centipede, "vigilance") {
		t.Error("three card types: vigilance")
	}
	dskRT2Delirium(g, me, 4)
	if p, tt := ptOf(t, g, centipede); p != 3 || tt != 3 {
		t.Errorf("delirium: %d/%d, want 3/3", p, tt)
	}
	if !hasEffectiveKeyword(t, g, centipede, "vigilance") {
		t.Error("delirium: no vigilance")
	}
}

// --- Splitskin Doll ---------------------------------------------------------------------

func TestDskRT2SplitskinDollDrawsWithoutDiscardingBesideASmallCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Creature(g, me.ID, "Small Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	hand := me.Hand.Size()
	dftABCast(t, g, me, game.Card{Name: "Splitskin Doll", TypeLine: "Artifact Creature — Toy",
		OracleID: dskRT2DollOracle, Power: 2, Toughness: 1})
	if discardChoiceFor(g, me.ID) != nil {
		t.Fatal("asked to discard while controlling another creature with power 2 or less")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn", hand, me.Hand.Size())
	}
}

func TestDskRT2SplitskinDollLootsWithoutAnotherSmallCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	// The Doll itself is a 2/1 and does not count; a 3/3 does not either.
	b12Creature(g, me.ID, "Big Bear", "Creature — Bear", 3, 3)
	advanceToMain(t, g)
	hand := me.Hand.Size()
	dftABCast(t, g, me, game.Card{Name: "Splitskin Doll", TypeLine: "Artifact Creature — Toy",
		OracleID: dskRT2DollOracle, Power: 2, Toughness: 1})
	if discardOwed(g, me.ID) != 1 {
		t.Fatalf("discard owed %d, want 1", discardOwed(g, me.ID))
	}
	answerDiscard(t, g, me.ID, me.Hand.Cards[0].InstanceID)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want a draw and a discard", hand, me.Hand.Size())
	}
}

// --- Saw ------------------------------------------------------------------------------

func TestDskRT2SawPumpsAndSacrificesAnotherPermanentToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	land := pushTypedCard(g, me.ID, "My Land", "Land", "")
	saw := seedEquipment(g, me.ID, "Saw", dskRT2SawOracle)
	brAttach(t, g, saw, bear)
	if p, tt := ptOf(t, g, bear); p != 4 || tt != 2 {
		t.Errorf("equipped creature is %d/%d, want 4/2", p, tt)
	}
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	prompt := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if prompt == nil {
		t.Fatal("no sacrifice prompt")
	}
	if prompt.ChooseMin != 0 || !hasID(prompt.ChooseCards, land) || hasID(prompt.ChooseCards, bear) || hasID(prompt.ChooseCards, saw) {
		t.Fatalf("offered %v (min %d): want the land, not the attacker or the Saw, with a floor of zero",
			prompt.ChooseCards, prompt.ChooseMin)
	}
	hand := me.Hand.Size()
	answerOwnPermanents(t, g, me.ID, land)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the land was not sacrificed")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn", hand, me.Hand.Size())
	}
}

func TestDskRT2SawDecliningDrawsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	land := pushTypedCard(g, me.ID, "My Land", "Land", "")
	saw := seedEquipment(g, me.ID, "Saw", dskRT2SawOracle)
	brAttach(t, g, saw, bear)
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	answerOwnPermanents(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(land) || me.Hand.Size() != hand {
		t.Error("declining sacrificed something or drew a card")
	}
}

// --- The Swarmweaver ---------------------------------------------------------------------

func TestDskRT2SwarmweaverMakesInsectsAndArmsThemWithDelirium(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	spider := b12Creature(g, me.ID, "My Spider", "Creature — Spider", 2, 2)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Insect", "Creature — Insect", 1, 1)
	advanceToMain(t, g)
	weaver := dftABCast(t, g, me, game.Card{Name: "The Swarmweaver", TypeLine: "Legendary Artifact Creature — Scarecrow",
		OracleID: dskRT2SwarmweaverOracle, Power: 2, Toughness: 3})
	var insects []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if IsToken(c) && c.Name == "Insect" && c.Controller == me.ID {
			insects = append(insects, c.InstanceID)
		}
	}
	if len(insects) != 2 {
		t.Fatalf("%d Insect tokens, want 2", len(insects))
	}
	tok, _ := battlefieldCard(g, insects[0])
	if !slices.Equal(tok.Colors, []string{"B", "G"}) {
		t.Errorf("the Insect is %v, want black and green", tok.Colors)
	}
	if p, tt := ptOf(t, g, insects[0]); p != 1 || tt != 1 || !hasEffectiveKeyword(t, g, insects[0], "flying") {
		t.Errorf("without delirium the Insect is %d/%d, want a 1/1 flyer", p, tt)
	}

	dskRT2Delirium(g, me, 4)
	for _, id := range append(insects, spider) {
		want := 2
		if id == spider {
			want = 3
		}
		if p, tt := ptOf(t, g, id); p != want || tt != want {
			t.Errorf("an Insect or Spider you control is %d/%d, want %d/%d", p, tt, want, want)
		}
		if !hasEffectiveKeyword(t, g, id, "deathtouch") {
			t.Error("an Insect or Spider you control lacks deathtouch")
		}
	}
	for _, id := range []uuid.UUID{bear, weaver, theirs} {
		if hasEffectiveKeyword(t, g, id, "deathtouch") {
			t.Error("a creature outside the clause gained deathtouch")
		}
	}
	if p, _ := ptOf(t, g, theirs); p != 1 {
		t.Errorf("an opponent's Insect has power %d, want 1", p)
	}
}

// --- Trapped in the Screen -------------------------------------------------------------

func TestDskRT2TrappedInTheScreenExilesUntilItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirLand := pushTypedCard(g, opp.ID, "Their Land", "Land", "")
	theirRock := pushTypedCard(g, opp.ID, "Their Rock", "Artifact", "{2}")
	trapped := castCatalogSpell(t, g, "Trapped in the Screen", "Enchantment", dskRT2TrappedOracle, nil)
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirLand}); err == nil {
		t.Fatal("a land was accepted as the target")
	}
	pickCard(t, g, me.ID, theirRock)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirRock) || !g.Exile.Contains(theirRock) {
		t.Fatal("the artifact was not exiled")
	}
	rooncDestroy(t, g, trapped)
	g.RunStateChecksForTest()
	// It returns as a new object (CR 400.7), so it is found by name.
	if n := namedOnBattlefield(g, "Their Rock", opp.ID); n != 1 {
		t.Errorf("after the enchantment left, %d copies under its owner's control, want 1", n)
	}
}

func TestDskRT2TrappedInTheScreenHasWardTwo(t *testing.T) {
	spec, ok := Lookup(dskRT2TrappedOracle)
	if !ok {
		t.Fatal("not registered")
	}
	if len(spec.Triggered) < 2 || spec.Triggered[0].Key != "Trapped in the Screen — ward {2}" {
		t.Errorf("first trigger %q, want the ward {2} trigger", spec.Triggered[0].Key)
	}
}

// --- Shardmage's Rescue ------------------------------------------------------------------

func TestDskRT2ShardmagesRescueGrantsHexproofOnlyTheTurnItEntered(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepEnd)
	if _, err := ecl2CastAt(t, g, "Shardmage's Rescue", "Enchantment — Aura", dskRT2RescueOracle, b16TargetCard(theirs)); err == nil {
		t.Fatal("enchant creature you control: an opponent's creature was accepted")
	}
	aura, err := ecl2CastAt(t, g, "Shardmage's Rescue", "Enchantment — Aura", dskRT2RescueOracle, b16TargetCard(bear))
	if err != nil {
		t.Fatalf("flash: cast in the end step: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(aura) {
		t.Fatal("the Aura did not resolve")
	}
	if p, tt := ptOf(t, g, bear); p != 3 || tt != 3 {
		t.Errorf("enchanted creature is %d/%d, want 3/3", p, tt)
	}
	if !hasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("the Aura entered this turn: no hexproof")
	}
	endTurn(t, g)
	if p, tt := ptOf(t, g, bear); p != 3 || tt != 3 {
		t.Errorf("next turn the enchanted creature is %d/%d, want 3/3", p, tt)
	}
	if hasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("hexproof outlasted the turn the Aura entered")
	}
}

// --- Tyvar, the Pummeler -------------------------------------------------------------------

func TestDskRT2TyvarTapsAnotherCreatureToGainIndestructibleAndTaps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tyvar := pushCatalogPermanent(g, me.ID, "Tyvar, the Pummeler", "Legendary Creature — Elf Warrior", dskRT2TyvarOracle, true)
	if err := dskRT2Activate(g, me.ID, tyvar, 0, "", game.ActivateAbilityParams{TapIDs: []uuid.UUID{tyvar}}); err == nil {
		t.Fatal("Tyvar paid the cost by tapping itself")
	}
	bear := pushCatalogPermanent(g, me.ID, "Fresh Bear", "Creature — Bear", "", true)
	if err := dskRT2Activate(g, me.ID, tyvar, 0, "", game.ActivateAbilityParams{TapIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("tap a summoning-sick creature for a summoning-sick Tyvar: %v", err)
	}
	if !dskRT2Tapped(g, bear) {
		t.Error("the other creature was not tapped as the cost")
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, tyvar, "indestructible") {
		t.Error("Tyvar did not gain indestructible")
	}
	if !dskRT2Tapped(g, tyvar) {
		t.Error("Tyvar did not tap")
	}
	endTurn(t, g)
	if hasEffectiveKeyword(t, g, tyvar, "indestructible") {
		t.Error("indestructible outlasted the turn")
	}
}

func TestDskRT2TyvarPumpsTheTeamByTheGreatestPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	tyvar := b12Push(g, me.ID, "Tyvar, the Pummeler", "Legendary Creature — Elf Warrior", dskRT2TyvarOracle, 3, 3)
	big := b12Creature(g, me.ID, "Big Bear", "Creature — Bear", 5, 5)
	theirs := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 7, 7)
	if err := dskRT2Activate(g, me.ID, tyvar, 1, "{G}{G}{C}{C}{C}", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, tyvar); p != 8 || tt != 8 {
		t.Errorf("Tyvar is %d/%d, want 8/8 (+5/+5)", p, tt)
	}
	if p, tt := ptOf(t, g, big); p != 10 || tt != 10 {
		t.Errorf("the Bear is %d/%d, want 10/10", p, tt)
	}
	if p, _ := ptOf(t, g, theirs); p != 7 {
		t.Errorf("an opponent's creature has power %d, want 7", p)
	}
	endTurn(t, g)
	if p, _ := ptOf(t, g, big); p != 5 {
		t.Errorf("the pump outlasted the turn: power %d", p)
	}
}

// --- The Rollercrusher Ride ----------------------------------------------------------------

func dskRT2RideCard() game.Card {
	return game.Card{Name: "The Rollercrusher Ride", TypeLine: "Legendary Enchantment", OracleID: dskRT2RideOracle, ManaCost: "{X}{2}{R}"}
}

func TestDskRT2RollercrusherRideDealsXToEachOfUpToXCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := b12Creature(g, opp.ID, "Giant A", "Creature — Giant", 5, 5)
	b := b12Creature(g, opp.ID, "Giant B", "Creature — Giant", 5, 5)
	c := b12Creature(g, opp.ID, "Giant C", "Creature — Giant", 5, 5)
	rooncCastMain(t, g, dskRT2RideCard(), game.CastSpellParams{XValue: 2})
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{rooncRef(a), rooncRef(b), rooncRef(c)}); err == nil {
		t.Fatal("'up to X' with X=2 accepted three targets")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{rooncRef(a), rooncRef(b)}); err != nil {
		t.Fatalf("pick two: %v", err)
	}
	passPriorityAroundTable(t, g)
	if damageMarkedOn(g, a) != 2 || damageMarkedOn(g, b) != 2 || damageMarkedOn(g, c) != 0 {
		t.Errorf("damage %d/%d/%d, want 2/2/0", damageMarkedOn(g, a), damageMarkedOn(g, b), damageMarkedOn(g, c))
	}
}

func TestDskRT2RollercrusherRideDoublesNoncombatDamageWithDelirium(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	giant := b12Creature(g, opp.ID, "Giant", "Creature — Giant", 7, 7)
	dskRT2Delirium(g, me, 4)
	rooncCastMain(t, g, dskRT2RideCard(), game.CastSpellParams{XValue: 1})
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, giant)
	passPriorityAroundTable(t, g)
	if got := damageMarkedOn(g, giant); got != 2 {
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	giant := b12Creature(g, opp.ID, "Giant", "Creature — Giant", 7, 7)
	dskRT2Delirium(g, me, 3)
	rooncCastMain(t, g, dskRT2RideCard(), game.CastSpellParams{XValue: 1})
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, giant)
	passPriorityAroundTable(t, g)
	if got := damageMarkedOn(g, giant); got != 1 {
		t.Errorf("three card types: %d damage, want 1", got)
	}
}

// --- The Jolly Balloon Man ------------------------------------------------------------------

func TestDskRT2JollyBalloonManMakesAOneOneRedFlyingBalloonCopyThatIsSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	balloonMan := b12Push(g, me.ID, "The Jolly Balloon Man", "Legendary Creature — Human Clown", dskRT2BalloonOracle, 1, 4)
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Green Bear", TypeLine: "Creature — Bear", Colors: []string{"G"},
		Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	if err := dskRT2Activate(g, me.ID, balloonMan, 0, "{C}", game.ActivateAbilityParams{Targets: b16TargetCard(bear)}); err == nil {
		t.Fatal("activated outside a main phase")
	}
	advanceToMain(t, g)
	if err := dskRT2Activate(g, me.ID, balloonMan, 0, "{C}", game.ActivateAbilityParams{Targets: b16TargetCard(balloonMan)}); err == nil {
		t.Fatal("the Balloon Man targeted itself")
	}
	if err := dskRT2Activate(g, me.ID, balloonMan, 0, "{C}", game.ActivateAbilityParams{Targets: b16TargetCard(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	var token uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if IsToken(c) && c.Name == "Green Bear" {
			token = c.InstanceID
		}
	}
	if token == uuid.Nil {
		t.Fatal("no token copy")
	}
	tok, _ := battlefieldCard(g, token)
	if !tok.HasSubtype("Balloon") || !tok.HasSubtype("Bear") {
		t.Errorf("token type line %q, want a Bear Balloon", tok.TypeLine)
	}
	if !slices.Contains(tok.Colors, "R") || !slices.Contains(tok.Colors, "G") {
		t.Errorf("token colours %v, want red in addition to green", tok.Colors)
	}
	if p, tt := ptOf(t, g, token); p != 1 || tt != 1 {
		t.Errorf("token is %d/%d, want 1/1", p, tt)
	}
	for _, kw := range []string{"flying", "haste"} {
		if !hasEffectiveKeyword(t, g, token, kw) {
			t.Errorf("token lacks %s", kw)
		}
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(token) {
		t.Error("the token was not sacrificed at the end step")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the original was sacrificed")
	}
}
