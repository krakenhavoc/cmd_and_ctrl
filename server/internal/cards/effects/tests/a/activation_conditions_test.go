package testsa

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
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
		{effectstest.AcTectonicEdgeOracle, "Tectonic Edge", 0},
		{effectstest.AcWeatheredWayfarerOracle, "Weathered Wayfarer", 0},
		{effectstest.AcBondersEnclaveOracle, "Bonders' Enclave", 0},
		{effectstest.AcCrypticCavesOracle, "Cryptic Caves", 0},
		{effectstest.AcEndlessAtlasOracle, "Endless Atlas", 0},
		{effectstest.AcKelpieGuideOracle, "Kelpie Guide", 1},
		{effectstest.AcBarbarianRingOracle, "Barbarian Ring", 0},
		{effectstest.AcCephalidColiseumOracle, "Cephalid Coliseum", 0},
		{effectstest.AcRivendellOracle, "Rivendell", 0},
		{effectstest.AcSeaGateWreckageOracle, "Sea Gate Wreckage", 0},
		{effectstest.AcMinasTirithOracle, "Minas Tirith", 0},
		{effectstest.AcIdolOfOblivionOracle, "Idol of Oblivion", 0},
		{effectstest.AcSpeakerOfTheHeavens, "Speaker of the Heavens", 0},
		{effectstest.AcLuminarchAscensionOracle, "Luminarch Ascension", 0},
		{effectstest.AcSliverHiveOracle, "Sliver Hive", 0},
		{effectstest.AcInventorsFairOracle, "Inventors' Fair", 0},
		{effectstest.AcLeechriddenSwampOracle, "Leechridden Swamp", 0},
		{effectstest.AcMistveilPlainsOracle, "Mistveil Plains", 0},
		{effectstest.AcLagomosOracle, "Lagomos, Hand of Hatred", 0},
		{effectstest.AcHallOfOraclesOracle, "Hall of Oracles", 0},
		{effectstest.AcSeedcoreOracle, "The Seedcore", 0},
		{effectstest.AcGoroGoroOracle, "Goro-Goro, Disciple of Ryusei", 1},
		{effectstest.AcLilypadVillageOracle, "Lilypad Village", 0},
		{effectstest.AcArchOfOrazcaOracle, "Arch of Orazca", 0},
		{effectstest.B36SanctumOfEternityOracle, "Sanctum of Eternity", 0},
	}
	for _, w := range want {
		spec, ok := effects.Lookup(w.oracle)
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

// acSettleDecliningCommandZone passes priority until the stack is
// empty, declining each CR 903.9b command-zone prompt (a bounced
// commander headed for hand) on the way, so several stacked bounces can
// resolve in turn.
func acSettleDecliningCommandZone(t *testing.T, g *game.Game, owner uuid.UUID) {
	t.Helper()
	for i := 0; i < 32; i++ {
		declined := false
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceOptionalReplacement && c.Chooser == owner {
				effectstest.B21DeclineCommandZone(t, g, owner)
				declined = true
				break
			}
		}
		if declined {
			continue
		}
		if effectstest.StackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("stack did not settle")
}

// --- the builders ------------------------------------------------------

// DuringYourTurn opens every step of your own turn — not a sorcery
// window — and none of an opponent's. Sanctum of Eternity is the card.
func TestSanctumOfEternityActivatesAnyTimeDuringYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	sanctum := effectstest.PushCatalogPermanent(g, me.ID, "Sanctum of Eternity", "Land", effectstest.B36SanctumOfEternityOracle, false)
	first := effectstest.B36Commander(g, me.ID, "First Commander")
	second := effectstest.B36Commander(g, me.ID, "Second Commander")
	untap := func() { g.WithWriteLock(func() { _ = g.UntapTargetForEffect(sanctum) }) }

	// Your combat: declare attackers is not a sorcery window.
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	effectstest.B06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, sanctum, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(first)}); err != nil {
		t.Fatalf("during your combat: %v", err)
	}
	// With that activation still on the stack — responding on your own
	// turn.
	untap()
	effectstest.B06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, sanctum, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(second)}); err != nil {
		t.Fatalf("with an item on the stack: %v", err)
	}
	acSettleDecliningCommandZone(t, g, me.ID)
	if !me.Hand.Contains(first) || !me.Hand.Contains(second) {
		t.Fatal("both bounces resolve to hand")
	}

	// Your end step.
	third := effectstest.B36Commander(g, me.ID, "Third Commander")
	effectstest.AdvanceTo(t, g, game.StepEnd)
	untap()
	effectstest.B06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, sanctum, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(third)}); err != nil {
		t.Fatalf("during your end step: %v", err)
	}
	acSettleDecliningCommandZone(t, g, me.ID)

	// An opponent's main phase: refused, nothing paid.
	fourth := effectstest.B36Commander(g, me.ID, "Fourth Commander")
	next := (seat + 1) % len(g.Seats)
	effectstest.AdvanceToMainOf(t, g, next)
	untap()
	effectstest.B06AddMana(me, "C", "C")
	effectstest.AcRefused(t, g, me.ID, sanctum, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(fourth)})
	if me.ManaPool.Total() != 2 {
		t.Errorf("pool = %v, want the {C}{C} still floating", me.ManaPool)
	}
}

// OpponentControlsAtLeast reads each opponent on their own: three
// opponents on two lands apiece are not "an opponent with four".
func TestTectonicEdgeNeedsOneOpponentWithFourLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var opps []*game.Player
	for _, p := range g.Seats {
		if p.ID != me.ID {
			opps = append(opps, p)
		}
	}
	edge := effectstest.PushCatalogPermanent(g, me.ID, "Tectonic Edge", "Land", effectstest.AcTectonicEdgeOracle, false)
	for _, opp := range opps {
		effectstest.AcLands(g, opp.ID, 2, "Wastes")
	}
	coffers := effectstest.SeedManaLand(g, opps[0].ID, "Cabal Coffers", "Land", "C")
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C")

	effectstest.AcRefused(t, g, me.ID, edge, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(coffers)})
	if !effects.OnBattlefield(g, edge) || me.ManaPool.Total() != 1 {
		t.Fatal("a refused Tectonic Edge was sacrificed or spent mana")
	}

	// A fourth land for the first opponent opens it.
	fourth := effectstest.SeedManaLand(g, opps[0].ID, "Wastes", "Land", "C")
	if err := g.ActivateCatalogAbility(me.ID, edge, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(coffers)}); err != nil {
		t.Fatalf("activate with an opponent on four lands: %v", err)
	}
	if effects.OnBattlefield(g, edge) {
		t.Error("Tectonic Edge is sacrificed as a cost")
	}
	// CR 602.1b: the condition is not re-checked at resolution — the
	// opponent dropping to three lands in response changes nothing.
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(fourth) })
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, coffers) {
		t.Error("the Coffers should be destroyed although the opponent is now below four lands")
	}
}

// OpponentControlsMore compares each opponent with you.
func TestWeatheredWayfarerNeedsAnOpponentAhead(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	wayfarer := effectstest.PushCatalogPermanent(g, me.ID, "Weathered Wayfarer", "Creature — Human Nomad Cleric", effectstest.AcWeatheredWayfarerOracle, false)
	effectstest.AcLands(g, me.ID, 2, "Plains")
	effectstest.AcLands(g, opp.ID, 2, "Island")
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Library Land", TypeLine: "Land", Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "W")

	effectstest.AcRefused(t, g, me.ID, wayfarer, 0, game.ActivateAbilityParams{})
	effectstest.AcLands(g, opp.ID, 1, "Island")
	handBefore := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, wayfarer, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate with an opponent ahead on lands: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	if me.Hand.Size() != handBefore+1 {
		t.Errorf("hand %d → %d, want the land card", handBefore, me.Hand.Size())
	}
}

// ControlsAtLeast over current power: a +1/+1 counter counts.
func TestBondersEnclaveNeedsPowerFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	enclave := effectstest.PushCatalogPermanent(g, me.ID, "Bonders' Enclave", "Land", effectstest.AcBondersEnclaveOracle, false)
	bear := effectstest.AcPermanent(g, me.ID, "Big Bear", "Creature — Bear", 3, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C", "C")

	effectstest.AcRefused(t, g, me.ID, enclave, 0, game.ActivateAbilityParams{})
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterPlusOne, 1) })
	handBefore := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, enclave, 0, game.ActivateAbilityParams{})
	if me.Hand.Size() != handBefore+1 {
		t.Errorf("hand %d → %d, want one card drawn", handBefore, me.Hand.Size())
	}
}

func TestCrypticCavesNeedsFiveLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	caves := effectstest.PushCatalogPermanent(g, me.ID, "Cryptic Caves", "Land", effectstest.AcCrypticCavesOracle, false)
	effectstest.AcLands(g, me.ID, 3, "Wastes")
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C")

	effectstest.AcRefused(t, g, me.ID, caves, 0, game.ActivateAbilityParams{})
	effectstest.AcLands(g, me.ID, 1, "Wastes")
	handBefore := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, caves, 0, game.ActivateAbilityParams{})
	if me.Hand.Size() != handBefore+1 || effects.OnBattlefield(g, caves) {
		t.Error("five lands (the Caves among them): sacrifice the Caves and draw a card")
	}
}

func TestEndlessAtlasNeedsThreeLandsWithOneName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	atlas := effectstest.PushCatalogPermanent(g, me.ID, "Endless Atlas", "Artifact", effectstest.AcEndlessAtlasOracle, false)
	effectstest.AcLands(g, me.ID, 2, "Island")
	effectstest.AcLands(g, me.ID, 1, "Snow-Covered Island")
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C")

	effectstest.AcRefused(t, g, me.ID, atlas, 0, game.ActivateAbilityParams{})
	effectstest.AcLands(g, me.ID, 1, "Island")
	handBefore := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, atlas, 0, game.ActivateAbilityParams{})
	if me.Hand.Size() != handBefore+1 {
		t.Errorf("hand %d → %d, want one card drawn", handBefore, me.Hand.Size())
	}
}

func TestKelpieGuideTapsOnlyWithEightLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	guide := effectstest.PushCatalogPermanent(g, me.ID, "Kelpie Guide", "Creature — Beast", effectstest.AcKelpieGuideOracle, false)
	lands := effectstest.AcLands(g, me.ID, 7, "Island")
	victim := effectstest.AcPermanent(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)

	// The ungated untap works at any land count.
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(lands[0]) })
	effectstest.B16Activate(t, g, me.ID, guide, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(lands[0])})
	if effectstest.B16Tapped(t, g, lands[0]) {
		t.Error("the untap ability should untap the Island")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(guide) })

	effectstest.AcRefused(t, g, me.ID, guide, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(victim)})
	effectstest.AcLands(g, me.ID, 1, "Island")
	effectstest.B16Activate(t, g, me.ID, guide, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(victim)})
	if !effectstest.B16Tapped(t, g, victim) {
		t.Error("eight lands: the Guide taps the target")
	}
}

// Threshold is GraveyardAtLeast(7).
func TestBarbarianRingThreshold(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ring := effectstest.PushCatalogPermanent(g, me.ID, "Barbarian Ring", "Land", effectstest.AcBarbarianRingOracle, false)
	effectstest.AcGraveyardCards(me, 6)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "R")

	effectstest.AcRefused(t, g, me.ID, ring, 0, game.ActivateAbilityParams{Targets: effectstest.AcPlayerRef(opp.ID)})
	effectstest.AcGraveyardCards(me, 1)
	before := opp.Life
	effectstest.B16Activate(t, g, me.ID, ring, 0, game.ActivateAbilityParams{Targets: effectstest.AcPlayerRef(opp.ID)})
	if opp.Life != before-2 {
		t.Errorf("opponent life %d → %d, want 2 damage", before, opp.Life)
	}
}

func TestCephalidColiseumThresholdWheelsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	coliseum := effectstest.PushCatalogPermanent(g, me.ID, "Cephalid Coliseum", "Land", effectstest.AcCephalidColiseumOracle, false)
	effectstest.AcGraveyardCards(me, 6)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "U")

	effectstest.AcRefused(t, g, me.ID, coliseum, 0, game.ActivateAbilityParams{Targets: effectstest.AcPlayerRef(opp.ID)})
	effectstest.AcGraveyardCards(me, 1)
	handBefore := opp.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, coliseum, 0, game.ActivateAbilityParams{Targets: effectstest.AcPlayerRef(opp.ID)})
	if opp.Hand.Size() != handBefore+3 {
		t.Errorf("target hand %d → %d, want three drawn", handBefore, opp.Hand.Size())
	}
	if got := effectstest.DiscardOwed(g, opp.ID); got != 3 {
		t.Errorf("discard owed = %d, want three", got)
	}
}

func TestRivendellScriesOnlyWithALegendaryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rivendell := effectstest.PushCatalogPermanent(g, me.ID, "Rivendell", "Legendary Land", effectstest.AcRivendellOracle, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "U")

	effectstest.AcRefused(t, g, me.ID, rivendell, 0, game.ActivateAbilityParams{})
	effectstest.AcPermanent(g, me.ID, "Elrond", "Legendary Creature — Elf Noble", 3, 3)
	effectstest.B16Activate(t, g, me.ID, rivendell, 0, game.ActivateAbilityParams{})
	if effectstest.ScryChoiceFor(g, me.ID) == nil {
		t.Error("with a legendary creature: a scry 2 prompt")
	}
}

func TestSeaGateWreckageNeedsAnEmptyHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	wreckage := effectstest.PushCatalogPermanent(g, me.ID, "Sea Gate Wreckage", "Land", effectstest.AcSeaGateWreckageOracle, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C", "C")

	effectstest.AcRefused(t, g, me.ID, wreckage, 0, game.ActivateAbilityParams{})
	for me.Hand.Size() > 0 {
		me.Graveyard.PushTop(me.Hand.Cards[0])
		me.Hand.Cards = me.Hand.Cards[1:]
	}
	effectstest.B16Activate(t, g, me.ID, wreckage, 0, game.ActivateAbilityParams{})
	if me.Hand.Size() != 1 {
		t.Errorf("hand = %d, want the one card drawn", me.Hand.Size())
	}
}

// Two DIFFERENT attackers this turn.
func TestMinasTirithNeedsTwoAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	tirith := effectstest.PushCatalogPermanent(g, me.ID, "Minas Tirith", "Legendary Land", effectstest.AcMinasTirithOracle, false)
	a := effectstest.AcPermanent(g, me.ID, "Soldier A", "Creature — Soldier", 2, 2)
	b := effectstest.AcPermanent(g, me.ID, "Soldier B", "Creature — Soldier", 2, 2)

	// #859: an attack declaration is announced at its lock-in, so the
	// condition — which reads this turn's attack events — is not met
	// until the whole declaration is in.
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.B06AddMana(me, "C", "W")
	effectstest.AcRefused(t, g, me.ID, tirith, 0, game.ActivateAbilityParams{})
	if err := g.DeclareAttacker(b, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	handBefore := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, tirith, 0, game.ActivateAbilityParams{})
	if me.Hand.Size() != handBefore+1 {
		t.Errorf("hand %d → %d, want one card drawn", handBefore, me.Hand.Size())
	}
}

func TestIdolOfOblivionDrawsAfterATokenThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	idol := effectstest.PushCatalogPermanent(g, me.ID, "Idol of Oblivion", "Artifact", effectstest.AcIdolOfOblivionOracle, false)
	effectstest.AdvanceToMain(t, g)

	effectstest.AcRefused(t, g, me.ID, idol, 0, game.ActivateAbilityParams{})
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("1/1 white Soldier"), 1) })
	handBefore := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, idol, 0, game.ActivateAbilityParams{})
	if me.Hand.Size() != handBefore+1 {
		t.Errorf("hand %d → %d, want one card drawn", handBefore, me.Hand.Size())
	}

	// The Eldrazi ability has no condition.
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(idol) })
	effectstest.B06AddMana(me, "C", "C", "C", "C", "C", "C", "C", "C")
	effectstest.B16Activate(t, g, me.ID, idol, 1, game.ActivateAbilityParams{})
	if len(effectstest.BattlefieldIDsNamed(g, "Eldrazi")) != 1 {
		t.Error("{8}, {T}, sacrifice: a 10/10 Eldrazi")
	}
}

// The card that prints both instructions: each refuses on its own.
func TestSpeakerOfTheHeavensNeedsBothGates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	speaker := effectstest.PushCatalogPermanent(g, me.ID, "Speaker of the Heavens", "Creature — Human Cleric", effectstest.AcSpeakerOfTheHeavens, false)
	me.Life = game.StartingLife + 6
	effectstest.AdvanceToMain(t, g)

	effectstest.AcRefused(t, g, me.ID, speaker, 0, game.ActivateAbilityParams{})
	me.Life = game.StartingLife + 7
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.ActivateCatalogAbility(me.ID, speaker, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrSorcerySpeedRequired) {
		t.Fatalf("in combat: err = %v, want ErrSorcerySpeedRequired", err)
	}
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	effectstest.B16Activate(t, g, me.ID, speaker, 0, game.ActivateAbilityParams{})
	if len(effectstest.BattlefieldIDsNamed(g, "Angel")) != 1 {
		t.Error("seven above the starting total, at sorcery speed: a 4/4 Angel")
	}
}

func TestLuminarchAscensionQuestAndAngels(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	ascension := effectstest.PushCatalogPermanent(g, me.ID, "Luminarch Ascension", "Enchantment", effectstest.AcLuminarchAscensionOracle, false)

	// An opponent's end step with no life lost: you may add a counter.
	next := (seat + 1) % len(g.Seats)
	effectstest.AdvanceToEndStepOf(t, g, next)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, ascension, "quest"); got != 1 {
		t.Fatalf("quest counters = %d, want 1", got)
	}

	effectstest.B06AddMana(me, "C", "W")
	effectstest.AcRefused(t, g, me.ID, ascension, 0, game.ActivateAbilityParams{})
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ascension, "quest", 3) })
	effectstest.B16Activate(t, g, me.ID, ascension, 0, game.ActivateAbilityParams{})
	if len(effectstest.BattlefieldIDsNamed(g, "Angel")) != 1 {
		t.Error("four quest counters: a 4/4 Angel, at instant speed on an opponent's turn")
	}
}

func TestLuminarchAscensionSkipsTheCounterAfterLifeLoss(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	ascension := effectstest.PushCatalogPermanent(g, me.ID, "Luminarch Ascension", "Enchantment", effectstest.AcLuminarchAscensionOracle, false)
	next := (seat + 1) % len(g.Seats)
	effectstest.AdvanceToMainOf(t, g, next)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -1) })
	effectstest.AdvanceToEndStepOf(t, g, next)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, ascension, "quest"); got != 0 {
		t.Errorf("quest counters = %d, want 0 after losing life", got)
	}
}

func TestSliverHiveMakesSliversOnlyWithASliver(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hive := effectstest.PushCatalogPermanent(g, me.ID, "Sliver Hive", "Land", effectstest.AcSliverHiveOracle, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C", "C", "C", "C")

	effectstest.AcRefused(t, g, me.ID, hive, 0, game.ActivateAbilityParams{})
	effectstest.AcPermanent(g, me.ID, "Muscle Sliver", "Creature — Sliver", 1, 1, "G")
	effectstest.B16Activate(t, g, me.ID, hive, 0, game.ActivateAbilityParams{})
	if len(effectstest.BattlefieldIDsNamed(g, "Sliver")) != 1 {
		t.Error("with a Sliver: a 1/1 Sliver token")
	}
}

func TestInventorsFairMetalcraft(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	fair := effectstest.PushCatalogPermanent(g, me.ID, "Inventors' Fair", "Legendary Land", effectstest.AcInventorsFairOracle, false)
	effectstest.AcPermanent(g, me.ID, "Ornament A", "Artifact", 0, 0)
	effectstest.AcPermanent(g, me.ID, "Ornament B", "Artifact", 0, 0)
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Library Artifact", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C", "C", "C")

	effectstest.AcRefused(t, g, me.ID, fair, 0, game.ActivateAbilityParams{})
	effectstest.AcPermanent(g, me.ID, "Ornament C", "Artifact", 0, 0)
	handBefore := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, fair, 0, game.ActivateAbilityParams{})
	effectstest.TakeSearchPrompts(t, g)
	if me.Hand.Size() != handBefore+1 {
		t.Errorf("hand %d → %d, want the artifact card", handBefore, me.Hand.Size())
	}
}

func TestInventorsFairUpkeepLifeNeedsThreeArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	fair := effectstest.PushCatalogPermanent(g, me.ID, "Inventors' Fair", "Legendary Land", effectstest.AcInventorsFairOracle, false)
	effectstest.AcPermanent(g, me.ID, "Ornament A", "Artifact", 0, 0)
	effectstest.AcPermanent(g, me.ID, "Ornament B", "Artifact", 0, 0)
	// Two artifacts: the intervening-if keeps the trigger off the stack.
	before := me.Life
	effectstest.AdvanceToUpkeepOf(t, g, seat)
	if len(g.PendingTriggers) != 0 || effectstest.TriggerOnStack(g, fair) != nil {
		t.Error("two artifacts: no upkeep trigger")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before {
		t.Errorf("life %d → %d with two artifacts", before, me.Life)
	}
	// Three: gain 1 at the next upkeep.
	effectstest.AcPermanent(g, me.ID, "Ornament C", "Artifact", 0, 0)
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	effectstest.AdvanceToUpkeepOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("life %d → %d, want +1 with three artifacts", before, me.Life)
	}
}

func TestLeechriddenSwampDrainsWithTwoBlackPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	swamp := effectstest.PushCatalogPermanent(g, me.ID, "Leechridden Swamp", "Land — Swamp", effectstest.AcLeechriddenSwampOracle, false)
	effectstest.AcPermanent(g, me.ID, "Black Knight", "Creature — Knight", 2, 2, "B")
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "B")

	effectstest.AcRefused(t, g, me.ID, swamp, 0, game.ActivateAbilityParams{})
	effectstest.AcPermanent(g, me.ID, "Dark Ritualist", "Creature — Human", 1, 1, "B")
	var before []int
	for _, p := range g.Seats {
		before = append(before, p.Life)
	}
	effectstest.B16Activate(t, g, me.ID, swamp, 0, game.ActivateAbilityParams{})
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	plains := effectstest.PushCatalogPermanent(g, me.ID, "Mistveil Plains", "Land — Plains", effectstest.AcMistveilPlainsOracle, false)
	effectstest.AcPermanent(g, me.ID, "White Knight", "Creature — Knight", 2, 2, "W")
	card := effectstest.PushGraveyardCardForTest(me, "Swords to Plowshares")
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "W")

	effectstest.AcRefused(t, g, me.ID, plains, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(card)})
	effectstest.AcPermanent(g, me.ID, "Serra Angel", "Creature — Angel", 4, 4, "W")
	effectstest.B16Activate(t, g, me.ID, plains, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(card)})
	if me.Library.Size() == 0 || me.Library.Cards[0].InstanceID != card {
		t.Error("the card goes to the bottom of the library")
	}
}

func TestLagomosTutorsAfterFiveDeaths(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lagomos := effectstest.PushCatalogPermanent(g, me.ID, "Lagomos, Hand of Hatred", "Legendary Creature — Human Shaman", effectstest.AcLagomosOracle, false)
	effectstest.AdvanceToMain(t, g)
	kill := func() {
		id := effectstest.AcPermanent(g, me.ID, "Doomed", "Creature — Test", 1, 1)
		g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(id) })
	}
	for i := 0; i < 4; i++ {
		kill()
	}
	effectstest.AcRefused(t, g, me.ID, lagomos, 0, game.ActivateAbilityParams{})
	kill()
	if err := g.ActivateCatalogAbility(me.ID, lagomos, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("after five deaths: %v", err)
	}
}

func TestLagomosMakesAnElementalThatLeavesAtEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Lagomos, Hand of Hatred", "Legendary Creature — Human Shaman", effectstest.AcLagomosOracle, false)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if len(effectstest.BattlefieldIDsNamed(g, "Elemental")) != 1 {
		t.Fatal("beginning of combat on your turn: a 2/1 Elemental")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if len(effectstest.BattlefieldIDsNamed(g, "Elemental")) != 0 {
		t.Error("the Elemental is sacrificed at the next end step")
	}
}

func TestHallOfOraclesNeedsAnInstantOrSorceryThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	hall := effectstest.PushCatalogPermanent(g, me.ID, "Hall of Oracles", "Land", effectstest.AcHallOfOraclesOracle, false)
	bear := effectstest.AcPermanent(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)

	effectstest.AcRefused(t, g, me.ID, hall, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	effectstest.B06AddMana(me, "R")
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.AcPlayerRef(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B16Activate(t, g, me.ID, hall, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
}

func TestSeedcorePumpsOnlyWhenCorrupted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	seedcore := effectstest.PushCatalogPermanent(g, me.ID, "The Seedcore", "Land — Sphere", effectstest.AcSeedcoreOracle, false)
	mite := effectstest.AcPermanent(g, me.ID, "Mite", "Artifact Creature — Phyrexian Mite", 1, 1)
	g.WithWriteLock(func() { _ = g.AddPlayerCounterForEffect(opp.ID, game.CounterPoison, 2) })
	effectstest.AdvanceToMain(t, g)

	effectstest.AcRefused(t, g, me.ID, seedcore, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(mite)})
	g.WithWriteLock(func() { _ = g.AddPlayerCounterForEffect(opp.ID, game.CounterPoison, 1) })
	effectstest.B16Activate(t, g, me.ID, seedcore, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(mite)})
	if c, _ := effectstest.BattlefieldCard(g, mite); c.CurrentPower() != 3 || c.CurrentToughness() != 2 {
		t.Errorf("mite is %d/%d, want 3/2", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestGoroGoroDragonNeedsAnAttackingModifiedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	goro := effectstest.PushCatalogPermanent(g, me.ID, "Goro-Goro, Disciple of Ryusei", "Legendary Creature — Goblin Samurai", effectstest.AcGoroGoroOracle, false)
	samurai := effectstest.AcPermanent(g, me.ID, "Samurai", "Creature — Human Samurai", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(samurai, game.CounterPlusOne, 1) })
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C", "C", "R", "R")

	// Modified but not attacking.
	effectstest.AcRefused(t, g, me.ID, goro, 1, game.ActivateAbilityParams{})
	effectstest.DeclareAttack(t, g, opp.ID, samurai)
	effectstest.B06AddMana(me, "C", "C", "C", "R", "R")
	effectstest.B16Activate(t, g, me.ID, goro, 1, game.ActivateAbilityParams{})
	if len(effectstest.BattlefieldIDsNamed(g, "Dragon Spirit")) != 1 {
		t.Error("with an attacking modified creature: a 5/5 Dragon Spirit")
	}
}

// The printed card allows the Dragon ability in the end of combat
// step, and the engine does too: creatures leave combat as that step
// ENDS (CR 511.3, #785). This is the caveat that came off with it.
func TestGoroGoroDragonWorksInTheEndOfCombatStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	goro := effectstest.PushCatalogPermanent(g, me.ID, "Goro-Goro, Disciple of Ryusei", "Legendary Creature — Goblin Samurai", effectstest.AcGoroGoroOracle, false)
	samurai := effectstest.AcPermanent(g, me.ID, "Samurai", "Creature — Human Samurai", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(samurai, game.CounterPlusOne, 1) })
	effectstest.AdvanceToMain(t, g)

	effectstest.DeclareAttack(t, g, opp.ID, samurai)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.B06AddMana(me, "C", "C", "C", "R", "R")
	effectstest.B16Activate(t, g, me.ID, goro, 1, game.ActivateAbilityParams{})
	if len(effectstest.BattlefieldIDsNamed(g, "Dragon Spirit")) != 1 {
		t.Error("the Dragon ability found no attacking creature in the end of combat step")
	}
}

func TestLilypadVillageSurveilsAfterAFrogEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	village := effectstest.PushCatalogPermanent(g, me.ID, "Lilypad Village", "Land", effectstest.AcLilypadVillageOracle, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "U")

	effectstest.AcRefused(t, g, me.ID, village, 0, game.ActivateAbilityParams{})
	// A Rat token entering under your control opens it.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("1/1 black Rat"), 1) })
	effectstest.B16Activate(t, g, me.ID, village, 0, game.ActivateAbilityParams{})
	if effectstest.SurveilChoiceFor(g, me.ID) == nil {
		t.Error("after a Rat entered under your control: a surveil 2 prompt")
	}
}

// The condition is about the creature as it entered (#743 review). A
// Soldier that enters and only later becomes every creature type — a
// Maskwood Nexus arriving afterwards — is a Frog now, but it did not
// enter as one, so the surveil stays shut. An opponent's Rat entering
// is not one that entered under your control.
func TestLilypadVillageReadsTypesAsTheCreatureEntered(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	village := effectstest.PushCatalogPermanent(g, me.ID, "Lilypad Village", "Land", effectstest.AcLilypadVillageOracle, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "U")

	var soldier uuid.UUID
	g.WithWriteLock(func() {
		ids, err := g.CreateTokensForEffect(me.ID, effects.TokenCard("1/1 white Soldier"), 1, game.TokenEntryOptions{})
		if err != nil || len(ids) != 1 {
			t.Fatalf("create Soldier: %v %v", ids, err)
		}
		soldier = ids[0]
		_ = g.CreateTokenForEffect(opp.ID, effects.TokenCard("1/1 black Rat"), 1)
	})
	effectstest.AcRefused(t, g, me.ID, village, 0, game.ActivateAbilityParams{})

	effectstest.PushNexusFor(g, me.ID)
	if !effectstest.LayeredCard(t, g, soldier).HasSubtype("Frog") {
		t.Fatal("setup: the Soldier is not a Frog under Maskwood Nexus")
	}
	effectstest.AcRefused(t, g, me.ID, village, 0, game.ActivateAbilityParams{})
}

// A Rat token that entered and has already died still entered this
// turn: the tally recorded its types as it came in, so the token
// ceasing to exist (CR 704.5d) does not take the answer with it.
func TestLilypadVillageCountsARatTokenThatHasSinceDied(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	village := effectstest.PushCatalogPermanent(g, me.ID, "Lilypad Village", "Land", effectstest.AcLilypadVillageOracle, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "U")

	g.WithWriteLock(func() {
		ids, err := g.CreateTokensForEffect(me.ID, effects.TokenCard("1/1 black Rat"), 1, game.TokenEntryOptions{})
		if err != nil || len(ids) != 1 {
			t.Fatalf("create Rat: %v %v", ids, err)
		}
		if err := g.SacrificePermanentForEffect(ids[0]); err != nil {
			t.Fatal(err)
		}
	})
	if len(effectstest.BattlefieldIDsNamed(g, "Rat")) != 0 {
		t.Fatal("setup: the Rat token is still on the battlefield")
	}
	effectstest.B16Activate(t, g, me.ID, village, 0, game.ActivateAbilityParams{})
	if effectstest.SurveilChoiceFor(g, me.ID) == nil {
		t.Error("after a Rat token entered and died: a surveil 2 prompt")
	}
}

// The city's blessing (CR 702.131, #2696): ten permanents including the
// Arch earn it through the Arch's own ascend, and it is kept after the
// board shrinks.
func TestArchOfOrazcaDrawsWithTheCitysBlessing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	arch := effectstest.PushCatalogPermanent(g, me.ID, "Arch of Orazca", "Land", effectstest.AcArchOfOrazcaOracle, false)
	effectstest.AcLands(g, me.ID, 8, "Wastes")
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C", "C", "C", "C")

	g.RunStateChecksForTest()
	effectstest.AcRefused(t, g, me.ID, arch, 0, game.ActivateAbilityParams{})
	effectstest.AcLands(g, me.ID, 1, "Wastes")
	g.RunStateChecksForTest()
	if !me.CitysBlessing {
		t.Fatal("ten permanents with the Arch's ascend did not give the city's blessing")
	}
	// Kept: take the board back under ten and the draw is still open.
	g.WithWriteLock(func() {
		removed := 0
		for i := len(g.Battlefield.Cards) - 1; i >= 0 && removed < 4; i-- {
			if c := g.Battlefield.Cards[i]; c.InstanceID != arch && c.Controller == me.ID {
				_, _ = g.Battlefield.Remove(c.InstanceID)
				removed++
			}
		}
	})
	handBefore := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, arch, 0, game.ActivateAbilityParams{})
	if me.Hand.Size() != handBefore+1 {
		t.Errorf("hand %d → %d, want one card drawn after the board shrank", handBefore, me.Hand.Size())
	}
}
