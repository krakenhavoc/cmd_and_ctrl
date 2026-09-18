package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// activation_conditions_test.go — #743, the catalog half of ADR 0020's
// activation-condition addendum: the condition builders in
// activation_conditions.go and every card that declares one. The
// engine half (nothing paid, no resolution re-check, the error order
// against sorcery speed) is in game/activation_condition_test.go.
//
// Each card test runs the ability with its condition false and asserts
// the refusal paid nothing, then makes the condition true and asserts
// the printed effect.

const (
	acTectonicEdgeOracle       = "4927150d-7ff6-4232-b20e-d2ea245ac710"
	acWeatheredWayfarerOracle  = "1093fdb6-bd4f-46a8-92e3-5aa46b52bb4f"
	acBondersEnclaveOracle     = "f33ce38a-34ec-4b65-a0fc-160484a02007"
	acCrypticCavesOracle       = "b2eb7a64-a307-4a78-a25d-63fb3ae1e237"
	acEndlessAtlasOracle       = "cb923c4d-1d2e-4039-90bb-72d1981b9738"
	acKelpieGuideOracle        = "8fc1e8b1-3afc-4e9d-a3ae-7bd9bcdcb465"
	acBarbarianRingOracle      = "eeb9377b-72c1-4214-9a66-0f55577c17d1"
	acCephalidColiseumOracle   = "c733873e-77db-471f-8061-139db24f7e7c"
	acRivendellOracle          = "2550099d-b3e2-4eb6-9f36-0fc412828ca6"
	acSeaGateWreckageOracle    = "91f34686-cb96-49c0-b4a7-49dd1fd076e2"
	acMinasTirithOracle        = "7b0d7e62-0287-454a-8702-b0bfa7b41245"
	acIdolOfOblivionOracle     = "c232d974-9c65-4bce-b045-e5a7cc0c62e0"
	acSpeakerOfTheHeavens      = "e4bacf1b-0a45-4c73-857d-dffadb7e6fa5"
	acLuminarchAscensionOracle = "90076bf5-aa9a-4a6e-9035-9aa97fd5561e"
	acSliverHiveOracle         = "e7286688-ffbe-4d25-ad55-27990f005368"
	acInventorsFairOracle      = "91d4a5fe-fd6d-4b14-a63f-61b4d0ecd9c4"
	acLeechriddenSwampOracle   = "d83c86c1-126d-49e9-9b13-9e55784c49c5"
	acMistveilPlainsOracle     = "bb5c1817-ac22-4779-9005-251bc354f181"
	acLagomosOracle            = "12c98c1e-7e02-473d-b494-ad8a0731fe82"
	acHallOfOraclesOracle      = "0bb896ba-e15e-43a9-9120-e674d7ba003c"
	acSeedcoreOracle           = "249fdd3e-376c-4ec2-a612-4353e0e61ee2"
	acGoroGoroOracle           = "b80df711-af51-4f9a-9a3f-f7a4cddc9c2c"
	acLilypadVillageOracle     = "5bb06e6f-e3af-4caa-b66d-77248ad46b61"
	acArchOfOrazcaOracle       = "3bb518ff-399b-4ce7-b9ad-a1d563dd7792"
)

// TestActivationConditionCardsAreRegistered pins each card's name and
// which of its activated abilities carries the condition, so a later
// edit that drops the gate — the stronger-than-printed failure (#259)
// — fails here rather than in a game.
func TestActivationConditionCardsAreRegistered(t *testing.T) {
	want := []struct {
		oracle, name string
		gated        int // index of the gated ability
	}{
		{acTectonicEdgeOracle, "Tectonic Edge", 0},
		{acWeatheredWayfarerOracle, "Weathered Wayfarer", 0},
		{acBondersEnclaveOracle, "Bonders' Enclave", 0},
		{acCrypticCavesOracle, "Cryptic Caves", 0},
		{acEndlessAtlasOracle, "Endless Atlas", 0},
		{acKelpieGuideOracle, "Kelpie Guide", 1},
		{acBarbarianRingOracle, "Barbarian Ring", 0},
		{acCephalidColiseumOracle, "Cephalid Coliseum", 0},
		{acRivendellOracle, "Rivendell", 0},
		{acSeaGateWreckageOracle, "Sea Gate Wreckage", 0},
		{acMinasTirithOracle, "Minas Tirith", 0},
		{acIdolOfOblivionOracle, "Idol of Oblivion", 0},
		{acSpeakerOfTheHeavens, "Speaker of the Heavens", 0},
		{acLuminarchAscensionOracle, "Luminarch Ascension", 0},
		{acSliverHiveOracle, "Sliver Hive", 0},
		{acInventorsFairOracle, "Inventors' Fair", 0},
		{acLeechriddenSwampOracle, "Leechridden Swamp", 0},
		{acMistveilPlainsOracle, "Mistveil Plains", 0},
		{acLagomosOracle, "Lagomos, Hand of Hatred", 0},
		{acHallOfOraclesOracle, "Hall of Oracles", 0},
		{acSeedcoreOracle, "The Seedcore", 0},
		{acGoroGoroOracle, "Goro-Goro, Disciple of Ryusei", 1},
		{acLilypadVillageOracle, "Lilypad Village", 0},
		{acArchOfOrazcaOracle, "Arch of Orazca", 0},
		{b36SanctumOfEternityOracle, "Sanctum of Eternity", 0},
	}
	for _, w := range want {
		spec, ok := Lookup(w.oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", w.name, w.oracle)
			continue
		}
		if spec.Name != w.name {
			t.Errorf("oracle %s registered as %q, want %q", w.oracle, spec.Name, w.name)
		}
		for i, ab := range spec.Activated {
			if gated := ab.Condition != nil; gated != (i == w.gated) {
				t.Errorf("%s ability %d: condition declared = %v, want %v", w.name, i, gated, i == w.gated)
			}
		}
	}
}

// --- helpers ---------------------------------------------------------

// acRefused activates and requires ErrConditionNotMet with the source
// untapped and the stack untouched.
func acRefused(t *testing.T, g *game.Game, player, card uuid.UUID, idx int, params game.ActivateAbilityParams) {
	t.Helper()
	stackBefore, metaBefore := g.Stack.Size(), len(g.StackMeta)
	err := g.ActivateCatalogAbility(player, card, idx, params)
	if !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("ability %d: err = %v, want ErrConditionNotMet", idx, err)
	}
	if onBattlefield(g, card) && b16Tapped(t, g, card) {
		t.Error("a refused activation tapped the source")
	}
	if g.Stack.Size() != stackBefore || len(g.StackMeta) != metaBefore {
		t.Error("a refused activation put something on the stack")
	}
}

// acLands seeds n vanilla lands under owner.
func acLands(g *game.Game, owner uuid.UUID, n int, name string) []uuid.UUID {
	out := make([]uuid.UUID, n)
	for i := range out {
		out[i] = seedManaLand(g, owner, name, "Land", "C")
	}
	return out
}

// acPermanent seeds a permanent with the given type line and colours,
// able to attack and tap.
func acPermanent(g *game.Game, owner uuid.UUID, name, typeLine string, power, toughness int, colors ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Colors: colors,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// acPlayerRef is a one-player target list.
func acPlayerRef(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetPlayer, ID: id}}
}

// acSettleDecliningCommandZone passes priority until the stack is
// empty, declining each CR 903.9 command-zone prompt on the way, so
// several stacked bounces can resolve in turn.
func acSettleDecliningCommandZone(t *testing.T, g *game.Game, owner uuid.UUID) {
	t.Helper()
	for i := 0; i < 32; i++ {
		declined := false
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceOptionalReplacement && c.Chooser == owner {
				b21DeclineCommandZone(t, g, owner)
				declined = true
				break
			}
		}
		if declined {
			continue
		}
		if stackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("stack did not settle")
}

func acGraveyardCards(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		pushGraveyardCardForTest(p, "Graveyard Filler")
	}
}

// --- the builders ------------------------------------------------------

// DuringYourTurn opens every step of your own turn — not a sorcery
// window — and none of an opponent's. Sanctum of Eternity is the card.
func TestSanctumOfEternityActivatesAnyTimeDuringYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	sanctum := pushCatalogPermanent(g, me.ID, "Sanctum of Eternity", "Land", b36SanctumOfEternityOracle, false)
	first := b36Commander(g, me.ID, "First Commander")
	second := b36Commander(g, me.ID, "Second Commander")
	untap := func() { g.WithWriteLock(func() { _ = g.UntapTargetForEffect(sanctum) }) }

	// Your combat: declare attackers is not a sorcery window.
	advanceTo(t, g, game.StepDeclareAttackers)
	b06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, sanctum, 0, game.ActivateAbilityParams{Targets: cardRefs(first)}); err != nil {
		t.Fatalf("during your combat: %v", err)
	}
	// With that activation still on the stack — responding on your own
	// turn.
	untap()
	b06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, sanctum, 0, game.ActivateAbilityParams{Targets: cardRefs(second)}); err != nil {
		t.Fatalf("with an item on the stack: %v", err)
	}
	acSettleDecliningCommandZone(t, g, me.ID)
	if !me.Hand.Contains(first) || !me.Hand.Contains(second) {
		t.Fatal("both bounces resolve to hand")
	}

	// Your end step.
	third := b36Commander(g, me.ID, "Third Commander")
	advanceTo(t, g, game.StepEnd)
	untap()
	b06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, sanctum, 0, game.ActivateAbilityParams{Targets: cardRefs(third)}); err != nil {
		t.Fatalf("during your end step: %v", err)
	}
	acSettleDecliningCommandZone(t, g, me.ID)

	// An opponent's main phase: refused, nothing paid.
	fourth := b36Commander(g, me.ID, "Fourth Commander")
	next := (seat + 1) % len(g.Seats)
	advanceToMainOf(t, g, next)
	untap()
	b06AddMana(me, "C", "C")
	acRefused(t, g, me.ID, sanctum, 0, game.ActivateAbilityParams{Targets: cardRefs(fourth)})
	if len(me.ManaPool) != 2 {
		t.Errorf("pool = %v, want the {C}{C} still floating", me.ManaPool)
	}
}

// OpponentControlsAtLeast reads each opponent on their own: three
// opponents on two lands apiece are not "an opponent with four".
func TestTectonicEdgeNeedsOneOpponentWithFourLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var opps []*game.Player
	for _, p := range g.Seats {
		if p.ID != me.ID {
			opps = append(opps, p)
		}
	}
	edge := pushCatalogPermanent(g, me.ID, "Tectonic Edge", "Land", acTectonicEdgeOracle, false)
	for _, opp := range opps {
		acLands(g, opp.ID, 2, "Wastes")
	}
	coffers := seedManaLand(g, opps[0].ID, "Cabal Coffers", "Land", "C")
	advanceToMain(t, g)
	b06AddMana(me, "C")

	acRefused(t, g, me.ID, edge, 0, game.ActivateAbilityParams{Targets: cardRefs(coffers)})
	if !onBattlefield(g, edge) || len(me.ManaPool) != 1 {
		t.Fatal("a refused Tectonic Edge was sacrificed or spent mana")
	}

	// A fourth land for the first opponent opens it.
	fourth := seedManaLand(g, opps[0].ID, "Wastes", "Land", "C")
	if err := g.ActivateCatalogAbility(me.ID, edge, 0, game.ActivateAbilityParams{Targets: cardRefs(coffers)}); err != nil {
		t.Fatalf("activate with an opponent on four lands: %v", err)
	}
	if onBattlefield(g, edge) {
		t.Error("Tectonic Edge is sacrificed as a cost")
	}
	// CR 602.1b: the condition is not re-checked at resolution — the
	// opponent dropping to three lands in response changes nothing.
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(fourth) })
	passPriorityAroundTable(t, g)
	if onBattlefield(g, coffers) {
		t.Error("the Coffers should be destroyed although the opponent is now below four lands")
	}
}

// OpponentControlsMore compares each opponent with you.
func TestWeatheredWayfarerNeedsAnOpponentAhead(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	wayfarer := pushCatalogPermanent(g, me.ID, "Weathered Wayfarer", "Creature — Human Nomad Cleric", acWeatheredWayfarerOracle, false)
	acLands(g, me.ID, 2, "Plains")
	acLands(g, opp.ID, 2, "Island")
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Library Land", TypeLine: "Land", Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	b06AddMana(me, "W")

	acRefused(t, g, me.ID, wayfarer, 0, game.ActivateAbilityParams{})
	acLands(g, opp.ID, 1, "Island")
	handBefore := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, wayfarer, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate with an opponent ahead on lands: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != handBefore+1 {
		t.Errorf("hand %d → %d, want the land card", handBefore, me.Hand.Size())
	}
}

// ControlsAtLeast over current power: a +1/+1 counter counts.
func TestBondersEnclaveNeedsPowerFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	enclave := pushCatalogPermanent(g, me.ID, "Bonders' Enclave", "Land", acBondersEnclaveOracle, false)
	bear := acPermanent(g, me.ID, "Big Bear", "Creature — Bear", 3, 3)
	advanceToMain(t, g)
	b06AddMana(me, "C", "C", "C")

	acRefused(t, g, me.ID, enclave, 0, game.ActivateAbilityParams{})
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterPlusOne, 1) })
	handBefore := me.Hand.Size()
	b16Activate(t, g, me.ID, enclave, 0, game.ActivateAbilityParams{})
	if me.Hand.Size() != handBefore+1 {
		t.Errorf("hand %d → %d, want one card drawn", handBefore, me.Hand.Size())
	}
}

func TestCrypticCavesNeedsFiveLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	caves := pushCatalogPermanent(g, me.ID, "Cryptic Caves", "Land", acCrypticCavesOracle, false)
	acLands(g, me.ID, 3, "Wastes")
	advanceToMain(t, g)
	b06AddMana(me, "C")

	acRefused(t, g, me.ID, caves, 0, game.ActivateAbilityParams{})
	acLands(g, me.ID, 1, "Wastes")
	handBefore := me.Hand.Size()
	b16Activate(t, g, me.ID, caves, 0, game.ActivateAbilityParams{})
	if me.Hand.Size() != handBefore+1 || onBattlefield(g, caves) {
		t.Error("five lands (the Caves among them): sacrifice the Caves and draw a card")
	}
}

func TestEndlessAtlasNeedsThreeLandsWithOneName(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	atlas := pushCatalogPermanent(g, me.ID, "Endless Atlas", "Artifact", acEndlessAtlasOracle, false)
	acLands(g, me.ID, 2, "Island")
	acLands(g, me.ID, 1, "Snow-Covered Island")
	advanceToMain(t, g)
	b06AddMana(me, "C", "C")

	acRefused(t, g, me.ID, atlas, 0, game.ActivateAbilityParams{})
	acLands(g, me.ID, 1, "Island")
	handBefore := me.Hand.Size()
	b16Activate(t, g, me.ID, atlas, 0, game.ActivateAbilityParams{})
	if me.Hand.Size() != handBefore+1 {
		t.Errorf("hand %d → %d, want one card drawn", handBefore, me.Hand.Size())
	}
}

func TestKelpieGuideTapsOnlyWithEightLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	guide := pushCatalogPermanent(g, me.ID, "Kelpie Guide", "Creature — Beast", acKelpieGuideOracle, false)
	lands := acLands(g, me.ID, 7, "Island")
	victim := acPermanent(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)

	// The ungated untap works at any land count.
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(lands[0]) })
	b16Activate(t, g, me.ID, guide, 0, game.ActivateAbilityParams{Targets: cardRefs(lands[0])})
	if b16Tapped(t, g, lands[0]) {
		t.Error("the untap ability should untap the Island")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(guide) })

	acRefused(t, g, me.ID, guide, 1, game.ActivateAbilityParams{Targets: cardRefs(victim)})
	acLands(g, me.ID, 1, "Island")
	b16Activate(t, g, me.ID, guide, 1, game.ActivateAbilityParams{Targets: cardRefs(victim)})
	if !b16Tapped(t, g, victim) {
		t.Error("eight lands: the Guide taps the target")
	}
}

// Threshold is GraveyardAtLeast(7).
func TestBarbarianRingThreshold(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ring := pushCatalogPermanent(g, me.ID, "Barbarian Ring", "Land", acBarbarianRingOracle, false)
	acGraveyardCards(me, 6)
	advanceToMain(t, g)
	b06AddMana(me, "R")

	acRefused(t, g, me.ID, ring, 0, game.ActivateAbilityParams{Targets: acPlayerRef(opp.ID)})
	acGraveyardCards(me, 1)
	before := opp.Life
	b16Activate(t, g, me.ID, ring, 0, game.ActivateAbilityParams{Targets: acPlayerRef(opp.ID)})
	if opp.Life != before-2 {
		t.Errorf("opponent life %d → %d, want 2 damage", before, opp.Life)
	}
}

func TestCephalidColiseumThresholdWheelsThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	coliseum := pushCatalogPermanent(g, me.ID, "Cephalid Coliseum", "Land", acCephalidColiseumOracle, false)
	acGraveyardCards(me, 6)
	advanceToMain(t, g)
	b06AddMana(me, "U")

	acRefused(t, g, me.ID, coliseum, 0, game.ActivateAbilityParams{Targets: acPlayerRef(opp.ID)})
	acGraveyardCards(me, 1)
	handBefore := opp.Hand.Size()
	b16Activate(t, g, me.ID, coliseum, 0, game.ActivateAbilityParams{Targets: acPlayerRef(opp.ID)})
	if opp.Hand.Size() != handBefore+3 {
		t.Errorf("target hand %d → %d, want three drawn", handBefore, opp.Hand.Size())
	}
	if got := discardOwed(g, opp.ID); got != 3 {
		t.Errorf("discard owed = %d, want three", got)
	}
}

func TestRivendellScriesOnlyWithALegendaryCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rivendell := pushCatalogPermanent(g, me.ID, "Rivendell", "Legendary Land", acRivendellOracle, false)
	advanceToMain(t, g)
	b06AddMana(me, "C", "U")

	acRefused(t, g, me.ID, rivendell, 0, game.ActivateAbilityParams{})
	acPermanent(g, me.ID, "Elrond", "Legendary Creature — Elf Noble", 3, 3)
	b16Activate(t, g, me.ID, rivendell, 0, game.ActivateAbilityParams{})
	if scryChoiceFor(g, me.ID) == nil {
		t.Error("with a legendary creature: a scry 2 prompt")
	}
}

func TestSeaGateWreckageNeedsAnEmptyHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	wreckage := pushCatalogPermanent(g, me.ID, "Sea Gate Wreckage", "Land", acSeaGateWreckageOracle, false)
	advanceToMain(t, g)
	b06AddMana(me, "C", "C", "C")

	acRefused(t, g, me.ID, wreckage, 0, game.ActivateAbilityParams{})
	for me.Hand.Size() > 0 {
		me.Graveyard.PushTop(me.Hand.Cards[0])
		me.Hand.Cards = me.Hand.Cards[1:]
	}
	b16Activate(t, g, me.ID, wreckage, 0, game.ActivateAbilityParams{})
	if me.Hand.Size() != 1 {
		t.Errorf("hand = %d, want the one card drawn", me.Hand.Size())
	}
}

// Two DIFFERENT attackers this turn.
func TestMinasTirithNeedsTwoAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	tirith := pushCatalogPermanent(g, me.ID, "Minas Tirith", "Legendary Land", acMinasTirithOracle, false)
	a := acPermanent(g, me.ID, "Soldier A", "Creature — Soldier", 2, 2)
	b := acPermanent(g, me.ID, "Soldier B", "Creature — Soldier", 2, 2)

	// #859: an attack declaration is announced at its lock-in, so the
	// condition — which reads this turn's attack events — is not met
	// until the whole declaration is in.
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	b06AddMana(me, "C", "W")
	acRefused(t, g, me.ID, tirith, 0, game.ActivateAbilityParams{})
	if err := g.DeclareAttacker(b, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	handBefore := me.Hand.Size()
	b16Activate(t, g, me.ID, tirith, 0, game.ActivateAbilityParams{})
	if me.Hand.Size() != handBefore+1 {
		t.Errorf("hand %d → %d, want one card drawn", handBefore, me.Hand.Size())
	}
}

func TestIdolOfOblivionDrawsAfterATokenThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	idol := pushCatalogPermanent(g, me.ID, "Idol of Oblivion", "Artifact", acIdolOfOblivionOracle, false)
	advanceToMain(t, g)

	acRefused(t, g, me.ID, idol, 0, game.ActivateAbilityParams{})
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("1/1 colorless Soldier"), 1) })
	handBefore := me.Hand.Size()
	b16Activate(t, g, me.ID, idol, 0, game.ActivateAbilityParams{})
	if me.Hand.Size() != handBefore+1 {
		t.Errorf("hand %d → %d, want one card drawn", handBefore, me.Hand.Size())
	}

	// The Eldrazi ability has no condition.
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(idol) })
	b06AddMana(me, "C", "C", "C", "C", "C", "C", "C", "C")
	b16Activate(t, g, me.ID, idol, 1, game.ActivateAbilityParams{})
	if len(battlefieldIDsNamed(g, "Eldrazi")) != 1 {
		t.Error("{8}, {T}, sacrifice: a 10/10 Eldrazi")
	}
}

// The card that prints both instructions: each refuses on its own.
func TestSpeakerOfTheHeavensNeedsBothGates(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	speaker := pushCatalogPermanent(g, me.ID, "Speaker of the Heavens", "Creature — Human Cleric", acSpeakerOfTheHeavens, false)
	me.Life = game.StartingLife + 6
	advanceToMain(t, g)

	acRefused(t, g, me.ID, speaker, 0, game.ActivateAbilityParams{})
	me.Life = game.StartingLife + 7
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.ActivateCatalogAbility(me.ID, speaker, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrSorcerySpeedRequired) {
		t.Fatalf("in combat: err = %v, want ErrSorcerySpeedRequired", err)
	}
	advanceTo(t, g, game.StepPostcombatMain)
	b16Activate(t, g, me.ID, speaker, 0, game.ActivateAbilityParams{})
	if len(battlefieldIDsNamed(g, "Angel")) != 1 {
		t.Error("seven above the starting total, at sorcery speed: a 4/4 Angel")
	}
}

func TestLuminarchAscensionQuestAndAngels(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	ascension := pushCatalogPermanent(g, me.ID, "Luminarch Ascension", "Enchantment", acLuminarchAscensionOracle, false)

	// An opponent's end step with no life lost: you may add a counter.
	next := (seat + 1) % len(g.Seats)
	advanceToEndStepOf(t, g, next)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, ascension, "quest"); got != 1 {
		t.Fatalf("quest counters = %d, want 1", got)
	}

	b06AddMana(me, "C", "W")
	acRefused(t, g, me.ID, ascension, 0, game.ActivateAbilityParams{})
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ascension, "quest", 3) })
	b16Activate(t, g, me.ID, ascension, 0, game.ActivateAbilityParams{})
	if len(battlefieldIDsNamed(g, "Angel")) != 1 {
		t.Error("four quest counters: a 4/4 Angel, at instant speed on an opponent's turn")
	}
}

func TestLuminarchAscensionSkipsTheCounterAfterLifeLoss(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	ascension := pushCatalogPermanent(g, me.ID, "Luminarch Ascension", "Enchantment", acLuminarchAscensionOracle, false)
	next := (seat + 1) % len(g.Seats)
	advanceToMainOf(t, g, next)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -1) })
	advanceToEndStepOf(t, g, next)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, ascension, "quest"); got != 0 {
		t.Errorf("quest counters = %d, want 0 after losing life", got)
	}
}

func TestSliverHiveMakesSliversOnlyWithASliver(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hive := pushCatalogPermanent(g, me.ID, "Sliver Hive", "Land", acSliverHiveOracle, false)
	advanceToMain(t, g)
	b06AddMana(me, "C", "C", "C", "C", "C")

	acRefused(t, g, me.ID, hive, 0, game.ActivateAbilityParams{})
	acPermanent(g, me.ID, "Muscle Sliver", "Creature — Sliver", 1, 1, "G")
	b16Activate(t, g, me.ID, hive, 0, game.ActivateAbilityParams{})
	if len(battlefieldIDsNamed(g, "Sliver")) != 1 {
		t.Error("with a Sliver: a 1/1 Sliver token")
	}
}

func TestInventorsFairMetalcraft(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	fair := pushCatalogPermanent(g, me.ID, "Inventors' Fair", "Legendary Land", acInventorsFairOracle, false)
	acPermanent(g, me.ID, "Ornament A", "Artifact", 0, 0)
	acPermanent(g, me.ID, "Ornament B", "Artifact", 0, 0)
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Library Artifact", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	b06AddMana(me, "C", "C", "C", "C")

	acRefused(t, g, me.ID, fair, 0, game.ActivateAbilityParams{})
	acPermanent(g, me.ID, "Ornament C", "Artifact", 0, 0)
	handBefore := me.Hand.Size()
	b16Activate(t, g, me.ID, fair, 0, game.ActivateAbilityParams{})
	if me.Hand.Size() != handBefore+1 {
		t.Errorf("hand %d → %d, want the artifact card", handBefore, me.Hand.Size())
	}
}

func TestInventorsFairUpkeepLifeNeedsThreeArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	fair := pushCatalogPermanent(g, me.ID, "Inventors' Fair", "Legendary Land", acInventorsFairOracle, false)
	acPermanent(g, me.ID, "Ornament A", "Artifact", 0, 0)
	acPermanent(g, me.ID, "Ornament B", "Artifact", 0, 0)
	// Two artifacts: the intervening-if keeps the trigger off the stack.
	before := me.Life
	advanceToUpkeepOf(t, g, seat)
	if len(g.PendingTriggers) != 0 || triggerOnStack(g, fair) != nil {
		t.Error("two artifacts: no upkeep trigger")
	}
	passPriorityAroundTable(t, g)
	if me.Life != before {
		t.Errorf("life %d → %d with two artifacts", before, me.Life)
	}
	// Three: gain 1 at the next upkeep.
	acPermanent(g, me.ID, "Ornament C", "Artifact", 0, 0)
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	advanceToUpkeepOf(t, g, seat)
	passPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("life %d → %d, want +1 with three artifacts", before, me.Life)
	}
}

func TestLeechriddenSwampDrainsWithTwoBlackPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	swamp := pushCatalogPermanent(g, me.ID, "Leechridden Swamp", "Land — Swamp", acLeechriddenSwampOracle, false)
	acPermanent(g, me.ID, "Black Knight", "Creature — Knight", 2, 2, "B")
	advanceToMain(t, g)
	b06AddMana(me, "B")

	acRefused(t, g, me.ID, swamp, 0, game.ActivateAbilityParams{})
	acPermanent(g, me.ID, "Dark Ritualist", "Creature — Human", 1, 1, "B")
	var before []int
	for _, p := range g.Seats {
		before = append(before, p.Life)
	}
	b16Activate(t, g, me.ID, swamp, 0, game.ActivateAbilityParams{})
	for i, p := range g.Seats {
		want := before[i] - 1
		if p.ID == me.ID {
			want = before[i]
		}
		if p.Life != want {
			t.Errorf("seat %d life %d → %d, want %d", i, before[i], p.Life, want)
		}
	}
}

func TestMistveilPlainsTucksWithTwoWhitePermanents(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	plains := pushCatalogPermanent(g, me.ID, "Mistveil Plains", "Land — Plains", acMistveilPlainsOracle, false)
	acPermanent(g, me.ID, "White Knight", "Creature — Knight", 2, 2, "W")
	card := pushGraveyardCardForTest(me, "Swords to Plowshares")
	advanceToMain(t, g)
	b06AddMana(me, "W")

	acRefused(t, g, me.ID, plains, 0, game.ActivateAbilityParams{Targets: cardRefs(card)})
	acPermanent(g, me.ID, "Serra Angel", "Creature — Angel", 4, 4, "W")
	b16Activate(t, g, me.ID, plains, 0, game.ActivateAbilityParams{Targets: cardRefs(card)})
	if me.Library.Size() == 0 || me.Library.Cards[0].InstanceID != card {
		t.Error("the card goes to the bottom of the library")
	}
}

func TestLagomosTutorsAfterFiveDeaths(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lagomos := pushCatalogPermanent(g, me.ID, "Lagomos, Hand of Hatred", "Legendary Creature — Human Shaman", acLagomosOracle, false)
	advanceToMain(t, g)
	kill := func() {
		id := acPermanent(g, me.ID, "Doomed", "Creature — Test", 1, 1)
		g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(id) })
	}
	for i := 0; i < 4; i++ {
		kill()
	}
	acRefused(t, g, me.ID, lagomos, 0, game.ActivateAbilityParams{})
	kill()
	if err := g.ActivateCatalogAbility(me.ID, lagomos, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("after five deaths: %v", err)
	}
}

func TestLagomosMakesAnElementalThatLeavesAtEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Lagomos, Hand of Hatred", "Legendary Creature — Human Shaman", acLagomosOracle, false)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if len(battlefieldIDsNamed(g, "Elemental")) != 1 {
		t.Fatal("beginning of combat on your turn: a 2/1 Elemental")
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if len(battlefieldIDsNamed(g, "Elemental")) != 0 {
		t.Error("the Elemental is sacrificed at the next end step")
	}
}

func TestHallOfOraclesNeedsAnInstantOrSorceryThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	hall := pushCatalogPermanent(g, me.ID, "Hall of Oracles", "Land", acHallOfOraclesOracle, false)
	bear := acPermanent(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)

	acRefused(t, g, me.ID, hall, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)})
	b06AddMana(me, "R")
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, acPlayerRef(opp.ID))
	passPriorityAroundTable(t, g)
	b16Activate(t, g, me.ID, hall, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)})
	if got := counterCount(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
}

func TestSeedcorePumpsOnlyWhenCorrupted(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	seedcore := pushCatalogPermanent(g, me.ID, "The Seedcore", "Land — Sphere", acSeedcoreOracle, false)
	mite := acPermanent(g, me.ID, "Mite", "Artifact Creature — Phyrexian Mite", 1, 1)
	g.WithWriteLock(func() { _ = g.AddPlayerCounterForEffect(opp.ID, game.CounterPoison, 2) })
	advanceToMain(t, g)

	acRefused(t, g, me.ID, seedcore, 0, game.ActivateAbilityParams{Targets: cardRefs(mite)})
	g.WithWriteLock(func() { _ = g.AddPlayerCounterForEffect(opp.ID, game.CounterPoison, 1) })
	b16Activate(t, g, me.ID, seedcore, 0, game.ActivateAbilityParams{Targets: cardRefs(mite)})
	if c, _ := battlefieldCard(g, mite); c.CurrentPower() != 3 || c.CurrentToughness() != 2 {
		t.Errorf("mite is %d/%d, want 3/2", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestGoroGoroDragonNeedsAnAttackingModifiedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	goro := pushCatalogPermanent(g, me.ID, "Goro-Goro, Disciple of Ryusei", "Legendary Creature — Goblin Samurai", acGoroGoroOracle, false)
	samurai := acPermanent(g, me.ID, "Samurai", "Creature — Human Samurai", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(samurai, game.CounterPlusOne, 1) })
	advanceToMain(t, g)
	b06AddMana(me, "C", "C", "C", "R", "R")

	// Modified but not attacking.
	acRefused(t, g, me.ID, goro, 1, game.ActivateAbilityParams{})
	declareAttack(t, g, opp.ID, samurai)
	b06AddMana(me, "C", "C", "C", "R", "R")
	b16Activate(t, g, me.ID, goro, 1, game.ActivateAbilityParams{})
	if len(battlefieldIDsNamed(g, "Dragon Spirit")) != 1 {
		t.Error("with an attacking modified creature: a 5/5 Dragon Spirit")
	}
}

// The printed card allows the Dragon ability in the end of combat
// step, and the engine does too: creatures leave combat as that step
// ENDS (CR 511.3, #785). This is the caveat that came off with it.
func TestGoroGoroDragonWorksInTheEndOfCombatStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	goro := pushCatalogPermanent(g, me.ID, "Goro-Goro, Disciple of Ryusei", "Legendary Creature — Goblin Samurai", acGoroGoroOracle, false)
	samurai := acPermanent(g, me.ID, "Samurai", "Creature — Human Samurai", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(samurai, game.CounterPlusOne, 1) })
	advanceToMain(t, g)

	declareAttack(t, g, opp.ID, samurai)
	advanceTo(t, g, game.StepEndCombat)
	b06AddMana(me, "C", "C", "C", "R", "R")
	b16Activate(t, g, me.ID, goro, 1, game.ActivateAbilityParams{})
	if len(battlefieldIDsNamed(g, "Dragon Spirit")) != 1 {
		t.Error("the Dragon ability found no attacking creature in the end of combat step")
	}
}

func TestLilypadVillageSurveilsAfterAFrogEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	village := pushCatalogPermanent(g, me.ID, "Lilypad Village", "Land", acLilypadVillageOracle, false)
	advanceToMain(t, g)
	b06AddMana(me, "U")

	acRefused(t, g, me.ID, village, 0, game.ActivateAbilityParams{})
	// A Rat token entering under your control opens it.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("1/1 black Rat"), 1) })
	b16Activate(t, g, me.ID, village, 0, game.ActivateAbilityParams{})
	if surveilChoiceFor(g, me.ID) == nil {
		t.Error("after a Rat entered under your control: a surveil 2 prompt")
	}
}

// The condition is about the creature as it entered (#743 review). A
// Soldier that enters and only later becomes every creature type — a
// Maskwood Nexus arriving afterwards — is a Frog now, but it did not
// enter as one, so the surveil stays shut. An opponent's Rat entering
// is not one that entered under your control.
func TestLilypadVillageReadsTypesAsTheCreatureEntered(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	village := pushCatalogPermanent(g, me.ID, "Lilypad Village", "Land", acLilypadVillageOracle, false)
	advanceToMain(t, g)
	b06AddMana(me, "U")

	var soldier uuid.UUID
	g.WithWriteLock(func() {
		ids, err := g.CreateTokensForEffect(me.ID, TokenCard("1/1 white Soldier"), 1, game.TokenEntryOptions{})
		if err != nil || len(ids) != 1 {
			t.Fatalf("create Soldier: %v %v", ids, err)
		}
		soldier = ids[0]
		_ = g.CreateTokenForEffect(opp.ID, TokenCard("1/1 black Rat"), 1)
	})
	acRefused(t, g, me.ID, village, 0, game.ActivateAbilityParams{})

	pushNexusFor(g, me.ID)
	if !layeredCard(t, g, soldier).HasSubtype("Frog") {
		t.Fatal("setup: the Soldier is not a Frog under Maskwood Nexus")
	}
	acRefused(t, g, me.ID, village, 0, game.ActivateAbilityParams{})
}

// A Rat token that entered and has already died still entered this
// turn: the tally recorded its types as it came in, so the token
// ceasing to exist (CR 704.5d) does not take the answer with it.
func TestLilypadVillageCountsARatTokenThatHasSinceDied(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	village := pushCatalogPermanent(g, me.ID, "Lilypad Village", "Land", acLilypadVillageOracle, false)
	advanceToMain(t, g)
	b06AddMana(me, "U")

	g.WithWriteLock(func() {
		ids, err := g.CreateTokensForEffect(me.ID, TokenCard("1/1 black Rat"), 1, game.TokenEntryOptions{})
		if err != nil || len(ids) != 1 {
			t.Fatalf("create Rat: %v %v", ids, err)
		}
		if err := g.SacrificePermanentForEffect(ids[0]); err != nil {
			t.Fatal(err)
		}
	})
	if len(battlefieldIDsNamed(g, "Rat")) != 0 {
		t.Fatal("setup: the Rat token is still on the battlefield")
	}
	b16Activate(t, g, me.ID, village, 0, game.ActivateAbilityParams{})
	if surveilChoiceFor(g, me.ID) == nil {
		t.Error("after a Rat token entered and died: a surveil 2 prompt")
	}
}

// The city's blessing, read live: ten permanents including the Arch.
func TestArchOfOrazcaDrawsWithTheCitysBlessing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	arch := pushCatalogPermanent(g, me.ID, "Arch of Orazca", "Land", acArchOfOrazcaOracle, false)
	acLands(g, me.ID, 8, "Wastes")
	advanceToMain(t, g)
	b06AddMana(me, "C", "C", "C", "C", "C")

	acRefused(t, g, me.ID, arch, 0, game.ActivateAbilityParams{})
	acLands(g, me.ID, 1, "Wastes")
	handBefore := me.Hand.Size()
	b16Activate(t, g, me.ID, arch, 0, game.ActivateAbilityParams{})
	if me.Hand.Size() != handBefore+1 {
		t.Errorf("hand %d → %d, want one card drawn", handBefore, me.Hand.Size())
	}
}
