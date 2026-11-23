package testss82v

import (
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestSamiCCardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.SamiConduitPylonsOracle:   "Conduit Pylons",
		effectstest.SamiCrystalGrottoOracle:   "Crystal Grotto",
		effectstest.SamiEiganjoOracle:         "Eiganjo, Seat of the Empire",
		effectstest.SamiFomoriVaultOracle:     "Fomori Vault",
		effectstest.SamiMonumentalHengeOracle: "Monumental Henge",
		effectstest.SamiNeedlevergePathOracle: "Needleverge Pathway",
		effectstest.SamiSokenzanOracle:        "Sokenzan, Crucible of Defiance",
		effectstest.SamiSpireOfIndustryOracle: "Spire of Industry",
		effectstest.SamiSunbillowVergeOracle:  "Sunbillow Verge",
		effectstest.SamiBlitzballOracle:       "Blitzball",
		effectstest.SamiBoommobileOracle:      "Boommobile",
		effectstest.SamiDragonsparkOracle:     "Dragonspark Reactor",
		effectstest.SamiMemorialVaultOracle:   "Memorial Vault",
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
		if spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s completeness = %v, want full", name, spec.Completeness)
		}
	}
	if _, ok := effects.Lookup(game.CatalogKeyForFace(effectstest.SamiNeedlevergePathOracle, 1)); !ok {
		t.Error("Pillarverge Pathway (the back face) is not registered")
	}
}

// --- the lands -----------------------------------------------------

func TestConduitPylonsEntersUntappedAndSurveils(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "Top", "Next")

	id := effectstest.PlayLandFromHand(t, g, "Conduit Pylons", effectstest.SamiConduitPylonsOracle)
	if c, _ := effectstest.BattlefieldCard(g, id); c.Tapped {
		t.Error("Conduit Pylons entered tapped")
	}
	if effectstest.TriggerOnStack(g, id) == nil {
		t.Fatal("the surveil trigger did not go on the stack")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceSurveil, me.ID) == nil {
		t.Error("resolving the trigger queued no surveil prompt")
	}
}

func TestConduitPylonsFilterNeedsOneMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := effectstest.SeedPermanentWithOracle(g, me.ID, "Conduit Pylons", "Land — Desert", effectstest.SamiConduitPylonsOracle)

	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("{T}: Add {C}: %v", err)
	}
	if got := effectstest.PoolColors(me); !reflect.DeepEqual(got, []string{"C"}) {
		t.Errorf("pool = %v, want {C}", got)
	}

	other := effectstest.SeedPermanentWithOracle(g, me.ID, "Conduit Pylons", "Land — Desert", effectstest.SamiConduitPylonsOracle)
	me.ManaPool = nil
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateManaAbility(me.ID, other, 1, game.ManaAbilityParams{Colors: []string{"U"}}); err != nil {
		t.Fatalf("{1},{T}: any colour: %v", err)
	}
	if got := effectstest.PoolColors(me); !reflect.DeepEqual(got, []string{"U"}) {
		t.Errorf("pool after the filter = %v, want {U} (the {C} paid for it)", got)
	}
}

func TestCrystalGrottoEntersUntappedAndScries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "Top", "Next")

	id := effectstest.PlayLandFromHand(t, g, "Crystal Grotto", effectstest.SamiCrystalGrottoOracle)
	if c, _ := effectstest.BattlefieldCard(g, id); c.Tapped {
		t.Error("Crystal Grotto entered tapped")
	}
	if effectstest.TriggerOnStack(g, id) == nil {
		t.Fatal("the scry trigger did not go on the stack")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ScryChoiceFor(g, me.ID) == nil {
		t.Error("resolving the trigger queued no scry prompt")
	}
}

func TestSpireOfIndustryNeedsAnArtifactAndCostsOneLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spire := effectstest.SeedPermanentWithOracle(g, me.ID, "Spire of Industry", "Land", effectstest.SamiSpireOfIndustryOracle)
	life := me.Life

	if err := g.ActivateManaAbility(me.ID, spire, 1, game.ManaAbilityParams{Colors: []string{"G"}}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("any colour with no artifact: err = %v, want ErrConditionNotMet", err)
	}
	if c, _ := effectstest.BattlefieldCard(g, spire); c.Tapped {
		t.Error("a refused activation tapped the land")
	}
	if me.Life != life {
		t.Errorf("a refused activation cost life: %d → %d", life, me.Life)
	}

	effectstest.SeedPermanentWithOracle(g, me.ID, "Trinket", "Artifact", "")
	if err := g.ActivateManaAbility(me.ID, spire, 1, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("any colour with an artifact: %v", err)
	}
	if got := effectstest.PoolColors(me); !reflect.DeepEqual(got, []string{"G"}) {
		t.Errorf("pool = %v, want {G}", got)
	}
	if me.Life != life-1 {
		t.Errorf("life %d → %d, want one paid", life, me.Life)
	}
}

func TestSunbillowVergeRedNeedsAMountainOrAPlains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	verge := effectstest.SeedPermanentWithOracle(g, me.ID, "Sunbillow Verge", "Land", effectstest.SamiSunbillowVergeOracle)
	effectstest.SeedManaLand(g, me.ID, "Swamp", "Basic Land — Swamp", "B")

	if err := g.ActivateManaAbility(me.ID, verge, 1, game.ManaAbilityParams{}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("{R} beside a Swamp: err = %v, want ErrConditionNotMet", err)
	}
	if err := g.ActivateManaAbility(me.ID, verge, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("unconditional {W}: %v", err)
	}
	if got := effectstest.PoolColors(me); !reflect.DeepEqual(got, []string{"W"}) {
		t.Errorf("pool = %v, want {W}", got)
	}

	me.ManaPool = nil
	verge2 := effectstest.SeedPermanentWithOracle(g, me.ID, "Sunbillow Verge", "Land", effectstest.SamiSunbillowVergeOracle)
	effectstest.SeedManaLand(g, me.ID, "Plains", "Basic Land — Plains", "W")
	if err := g.ActivateManaAbility(me.ID, verge2, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("{R} beside a Plains: %v", err)
	}
	if got := effectstest.PoolColors(me); !reflect.DeepEqual(got, []string{"R"}) {
		t.Errorf("pool = %v, want {R}", got)
	}
}

func TestNeedlevergePathwayFrontTapsForRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.SeedPermanentWithOracle(g, me.ID, "Needleverge Pathway", "Land", effectstest.SamiNeedlevergePathOracle)
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("{T}: Add {R}: %v", err)
	}
	if got := effectstest.PoolColors(me); !reflect.DeepEqual(got, []string{"R"}) {
		t.Errorf("pool = %v, want {R}", got)
	}
}

func TestNeedlevergePathwayBackFaceTapsForWhite(t *testing.T) {
	g, me, id := effectstest.HandWithMDFC(t, func(owner uuid.UUID) game.Card {
		return effectstest.MdfcCard(owner, effectstest.SamiNeedlevergePathOracle, game.LayoutModalDFC,
			game.Face{Name: "Needleverge Pathway", TypeLine: "Land"},
			game.Face{Name: "Pillarverge Pathway", TypeLine: "Land"},
		)
	})
	me.LandDropsPerTurn++
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("play Pillarverge Pathway: %v", err)
	}
	if c, ok := effectstest.BattlefieldCard(g, id); !ok || c.Tapped {
		t.Fatal("the back face did not enter untapped")
	}
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("back face {T}: %v", err)
	}
	if got := effectstest.PoolColors(me); !reflect.DeepEqual(got, []string{"W"}) {
		t.Errorf("pool = %v, want {W}", got)
	}
}

func TestNeedlevergePathwayFrontFacePlayedFromHandTapsForRed(t *testing.T) {
	g, me, id := effectstest.HandWithMDFC(t, func(owner uuid.UUID) game.Card {
		return effectstest.MdfcCard(owner, effectstest.SamiNeedlevergePathOracle, game.LayoutModalDFC,
			game.Face{Name: "Needleverge Pathway", TypeLine: "Land"},
			game.Face{Name: "Pillarverge Pathway", TypeLine: "Land"},
		)
	})
	me.LandDropsPerTurn++
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("play Needleverge Pathway: %v", err)
	}
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("front face {T}: %v", err)
	}
	if got := effectstest.PoolColors(me); !reflect.DeepEqual(got, []string{"R"}) {
		t.Errorf("pool = %v, want {R}", got)
	}
}

func TestMonumentalHengeEntersTappedUnlessYouControlAPlains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.PlayLandFromHand(t, g, "Monumental Henge", effectstest.SamiMonumentalHengeOracle)
	if c, _ := effectstest.BattlefieldCard(g, id); !c.Tapped {
		t.Error("Henge entered untapped with no Plains")
	}
	if n := effectstest.TapEventsFor(g, id); n != 0 {
		t.Errorf("%d tap events: it should enter tapped, not be tapped", n)
	}

	g2 := effectstest.NewCatalogGame(t)
	me := g2.Seats[g2.Turn.ActiveSeat]
	effectstest.SeedManaLand(g2, me.ID, "Plains", "Basic Land — Plains", "W")
	id2 := effectstest.PlayLandFromHand(t, g2, "Monumental Henge", effectstest.SamiMonumentalHengeOracle)
	if c, _ := effectstest.BattlefieldCard(g2, id2); c.Tapped {
		t.Error("Henge entered tapped with a Plains out")
	}
}

func TestMonumentalHengeFindsAHistoricCardAndBottomsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	henge := effectstest.SeedPermanentWithOracle(g, me.ID, "Monumental Henge", "Land", effectstest.SamiMonumentalHengeOracle)
	ids := effectstest.SamiLibraryTopFirst(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Shock", TypeLine: "Instant"},
		game.Card{Name: "Wurm", TypeLine: "Artifact Creature — Wurm"},
		game.Card{Name: "Lone Hero", TypeLine: "Legendary Creature — Human"},
		game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Sixth", TypeLine: "Instant"},
	)
	effectstest.SamiMainPhase(t, g)
	libBefore := me.Library.Size()
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{W}{W}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, henge, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no historic-card prompt")
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Errorf("prompt bounds = %d..%d, want 0..1 (\"you may\", one card)", pick.ChooseMin, pick.ChooseMax)
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range pick.ChooseCards {
		offered[id] = true
	}
	for i, want := range []bool{false, false, true, true, false, false} {
		if offered[ids[i]] != want {
			t.Errorf("card %d offered = %v, want %v (historic = artifact, legendary or Saga)", i, offered[ids[i]], want)
		}
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[3]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(ids[3]) {
		t.Error("the chosen legendary did not reach the hand")
	}
	if me.Library.Size() != libBefore-1 {
		t.Fatalf("library = %d, want %d (a card left it, the rest went under)", me.Library.Size(), libBefore-1)
	}
	if top := me.Library.Cards[me.Library.Size()-1].InstanceID; top != ids[5] {
		t.Error("the sixth card must now be on top; the other four went under it")
	}
}

func TestMonumentalHengeMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	henge := effectstest.SeedPermanentWithOracle(g, me.ID, "Monumental Henge", "Land", effectstest.SamiMonumentalHengeOracle)
	ids := effectstest.SamiLibraryTopFirst(me,
		game.Card{Name: "Wurm", TypeLine: "Artifact Creature — Wurm"},
		game.Card{Name: "B"}, game.Card{Name: "C"}, game.Card{Name: "D"}, game.Card{Name: "E"},
		game.Card{Name: "F"},
	)
	effectstest.SamiMainPhase(t, g)
	handBefore := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, henge, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("declining: %v", err)
	}
	if me.Hand.Size() != handBefore {
		t.Error("declining still took a card")
	}
	if top := me.Library.Cards[me.Library.Size()-1].InstanceID; top != ids[5] {
		t.Error("all five looked-at cards must go under the sixth")
	}
}

func TestFomoriVaultLooksAtAsManyCardsAsYouHaveArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	vault := effectstest.SeedPermanentWithOracle(g, me.ID, "Fomori Vault", "Land", effectstest.SamiFomoriVaultOracle)
	effectstest.SeedPermanentWithOracle(g, me.ID, "Trinket A", "Artifact", "")
	effectstest.SeedPermanentWithOracle(g, me.ID, "Trinket B", "Artifact", "")
	ids := effectstest.SamiLibraryTopFirst(me, game.Card{Name: "One"}, game.Card{Name: "Two"}, game.Card{Name: "Three"})
	effectstest.SamiMainPhase(t, g)
	discard := effectstest.PushCatalogHandCard(me, "Pitch", "Sorcery", "")
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, vault, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{discard}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !me.Graveyard.Contains(discard) {
		t.Error("the discard cost did not discard")
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick prompt")
	}
	if len(pick.ChooseCards) != 2 {
		t.Errorf("%d cards offered, want 2 (two artifacts)", len(pick.ChooseCards))
	}
	if pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Errorf("bounds = %d..%d, want exactly one", pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(ids[1]) {
		t.Error("the chosen card did not reach the hand")
	}
	if top := me.Library.Cards[me.Library.Size()-1].InstanceID; top != ids[2] {
		t.Error("the unlooked-at third card must be on top")
	}
}

func TestFomoriVaultWithNoArtifactsDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	vault := effectstest.SeedPermanentWithOracle(g, me.ID, "Fomori Vault", "Land", effectstest.SamiFomoriVaultOracle)
	ids := effectstest.SamiLibraryTopFirst(me, game.Card{Name: "One"}, game.Card{Name: "Two"})
	effectstest.SamiMainPhase(t, g)
	discard := effectstest.PushCatalogHandCard(me, "Pitch", "Sorcery", "")
	if err := g.ActivateCatalogAbility(me.ID, vault, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{discard}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChooseCardsFor(g, me.ID) != nil {
		t.Error("X = 0 still opened a prompt")
	}
	if top := me.Library.Cards[me.Library.Size()-1].InstanceID; top != ids[0] {
		t.Error("the library changed")
	}
}

// --- the channel lands ---------------------------------------------

func TestEiganjoChannelHitsAnAttackingCreatureForFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	attacker := effectstest.B12Creature(g, opp.ID, "Attacker", "Creature — Ogre", 4, 4)
	idle := effectstest.B12Creature(g, opp.ID, "Idle", "Creature — Ogre", 4, 4)
	g.WithWriteLock(func() {
		if c := effectstest.BattlefieldCardFor(g, attacker); c != nil {
			c.AttackingTarget = me.ID
		}
	})
	effectstest.SamiMainPhase(t, g)
	id := effectstest.PushCatalogHandCard(me, "Eiganjo, Seat of the Empire", "Legendary Land", effectstest.SamiEiganjoOracle)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{W}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: idle}},
	}); err == nil {
		t.Fatal("a creature that is neither attacking nor blocking was accepted as a target")
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: attacker}},
	}); err != nil {
		t.Fatalf("channel: %v", err)
	}
	if !me.Graveyard.Contains(id) {
		t.Error("the discard-this cost did not bin Eiganjo")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) {
		t.Error("4 damage did not kill the 4/4 attacker")
	}
	if !g.Battlefield.Contains(idle) {
		t.Error("the idle creature was hit")
	}
}

func TestEiganjoChannelCostsOneLessPerLegendaryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	attacker := effectstest.B12Creature(g, opp.ID, "Attacker", "Creature — Ogre", 2, 2)
	g.WithWriteLock(func() {
		if c := effectstest.BattlefieldCardFor(g, attacker); c != nil {
			c.AttackingTarget = me.ID
		}
	})
	effectstest.SamiMainPhase(t, g)
	id := effectstest.PushCatalogHandCard(me, "Eiganjo, Seat of the Empire", "Legendary Land", effectstest.SamiEiganjoOracle)
	target := []game.TargetRef{{Kind: game.TargetCard, ID: attacker}}

	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{W}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Targets: target, Strict: true})
	if !errors.Is(err, game.ErrInsufficientMana) {
		t.Fatalf("{1}{W} with no legend: err = %v, want insufficient mana", err)
	}

	effectstest.PushCatalogPermanent(g, me.ID, "Some Legend", "Legendary Creature — Human", "", false)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Targets: target, Strict: true}); err != nil {
		t.Fatalf("{1}{W} with one legend: %v", err)
	}
}

func TestSokenzanChannelMakesTwoHastySpirits(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SamiMainPhase(t, g)
	id := effectstest.PushCatalogHandCard(me, "Sokenzan, Crucible of Defiance", "Legendary Land", effectstest.SamiSokenzanOracle)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{R}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("channel: %v", err)
	}
	if !me.Graveyard.Contains(id) {
		t.Error("the discard-this cost did not bin Sokenzan")
	}
	effectstest.PassPriorityAroundTable(t, g)

	var spirits []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Spirit" {
			spirits = append(spirits, c.InstanceID)
		}
	}
	if len(spirits) != 2 {
		t.Fatalf("%d Spirit tokens, want 2", len(spirits))
	}
	for _, s := range spirits {
		if !effectstest.EffectiveAbilitiesContain(t, g, s, "haste") {
			t.Error("a Spirit token lacks haste")
		}
		if effectstest.EffectivePower(t, g, s) != 1 {
			t.Error("a Spirit token is not 1/1")
		}
	}
}

// --- the artifacts -------------------------------------------------

func TestBlitzballNeedsALegendToConnect(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ball := effectstest.PushCatalogPermanent(g, me.ID, "Blitzball", "Artifact", effectstest.SamiBlitzballOracle, false)
	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	legend := effectstest.B12Creature(g, me.ID, "Legend", "Legendary Creature — Human", 2, 2)
	oppLegend := effectstest.B12Creature(g, opp.ID, "Their Legend", "Legendary Creature — Human", 2, 2)
	effectstest.SeedLibrary(me, "A", "B", "C")
	before := me.Hand.Size()

	err := g.ActivateCatalogAbility(me.ID, ball, 0, game.ActivateAbilityParams{})
	if !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("no damage dealt: err = %v, want ErrConditionNotMet", err)
	}
	effectstest.DealCombatDamage(g, me.ID, plain, opp.ID)
	if err := g.ActivateCatalogAbility(me.ID, ball, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("a nonlegendary creature connected: err = %v, want ErrConditionNotMet", err)
	}
	// An opponent's legend hitting ME is not "an opponent was dealt damage".
	effectstest.DealCombatDamage(g, opp.ID, oppLegend, me.ID)
	if err := g.ActivateCatalogAbility(me.ID, ball, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("a legend hit me: err = %v, want ErrConditionNotMet", err)
	}
	effectstest.DealCombatDamage(g, me.ID, legend, opp.ID)
	if err := g.ActivateCatalogAbility(me.ID, ball, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("a legend connected: %v", err)
	}
	if g.Battlefield.Contains(ball) {
		t.Error("Blitzball was not sacrificed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+2 {
		t.Errorf("hand %d → %d, want two cards drawn", before, got)
	}
}

func TestBlitzballCountsALegendThatDiedAfterConnecting(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ball := effectstest.PushCatalogPermanent(g, me.ID, "Blitzball", "Artifact", effectstest.SamiBlitzballOracle, false)
	legend := effectstest.B12Creature(g, me.ID, "Legend", "Legendary Creature — Human", 2, 2)
	effectstest.SeedLibrary(me, "A", "B")
	effectstest.DealCombatDamage(g, me.ID, legend, opp.ID)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(legend) })

	if err := g.ActivateCatalogAbility(me.ID, ball, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("the legend traded after connecting: %v", err)
	}
}

func TestBlitzballTapsForAnyColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ball := effectstest.PushCatalogPermanent(g, me.ID, "Blitzball", "Artifact", effectstest.SamiBlitzballOracle, false)
	if err := g.ActivateManaAbility(me.ID, ball, 0, game.ManaAbilityParams{Colors: []string{"B"}}); err != nil {
		t.Fatalf("mana: %v", err)
	}
	if got := effectstest.PoolColors(me); !reflect.DeepEqual(got, []string{"B"}) {
		t.Errorf("pool = %v, want {B}", got)
	}
}

func TestBoommobileMakesFourActivateOnlyMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastCatalogSpell(t, g, "Boommobile", "Artifact — Vehicle", effectstest.SamiBoommobileOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	pick := effectstest.LatestChoiceOfKind(g, game.PendingChoiceColor)
	if pick == nil {
		t.Fatal("no colour prompt for the four mana")
	}
	if err := g.ResolveColorChoice(pick.ID, me.ID, "R"); err != nil {
		t.Fatalf("ResolveColorChoice: %v", err)
	}
	if me.ManaPool.Total() != 4 {
		t.Fatalf("pool = %d mana, want 4", me.ManaPool.Total())
	}
	for _, tok := range me.ManaPool.Units() {
		if tok.Color != "R" {
			t.Errorf("mana colour %q, want R (one colour for all four)", tok.Color)
		}
		if !reflect.DeepEqual(tok.Restrictions, []string{game.ManaRestrictActivate}) {
			t.Errorf("restrictions = %v, want activate-only", tok.Restrictions)
		}
	}
}

func TestBoommobileExhaustDealsXAndGrowsOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	car := effectstest.PushCatalogPermanent(g, me.ID, "Boommobile", "Artifact — Vehicle", effectstest.SamiBoommobileOracle, false)
	effectstest.SamiMainPhase(t, g)
	life := opp.Life
	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{C}{R}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, car, 0, game.ActivateAbilityParams{
		XValue: 2, Targets: target, Strict: true,
	}); err != nil {
		t.Fatalf("exhaust X=2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life %d → %d, want 2 damage", life, opp.Life)
	}
	if c, _ := effectstest.BattlefieldCard(g, car); c.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", c.Counters[game.CounterPlusOne])
	}
	err := g.ActivateCatalogAbility(me.ID, car, 0, game.ActivateAbilityParams{XValue: 1, Targets: target})
	if !errors.Is(err, game.ErrAbilityExhausted) {
		t.Fatalf("second exhaust: err = %v, want ErrAbilityExhausted", err)
	}
}

func TestBoommobileCrewsAtTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	car := effectstest.PushVehicleForTest(g, me.ID, "Boommobile", effectstest.SamiBoommobileOracle, 6, 4)
	weak := effectstest.PushCrewerForTest(g, me.ID, "Weak", 1)
	strong := effectstest.PushCrewerForTest(g, me.ID, "Strong", 2)

	if err := g.ActivateCatalogAbility(me.ID, car, 1, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{weak}}); err == nil {
		t.Fatal("crewing with power 1 was accepted for Crew 2")
	}
	if err := g.ActivateCatalogAbility(me.ID, car, 1, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{strong}}); err != nil {
		t.Fatalf("crew with power 2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCardByID(g, car); !c.IsCreature() {
		t.Error("the crewed Vehicle is not a creature")
	}
}

func TestDragonsparkReactorCountsArtifactsYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	effectstest.CastCatalogSpell(t, g, "Dragonspark Reactor", "Artifact", effectstest.SamiDragonsparkOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	var reactor uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.OracleID == effectstest.SamiDragonsparkOracle {
			reactor = c.InstanceID
		}
	}
	if reactor == uuid.Nil {
		t.Fatal("the Reactor did not resolve")
	}
	charge := func() int {
		c, _ := effectstest.BattlefieldCard(g, reactor)
		return c.Counters[game.CounterCharge]
	}
	if charge() != 1 {
		t.Fatalf("charge counters after its own entry = %d, want 1", charge())
	}
	effectstest.CastCatalogSpell(t, g, "Trinket", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if charge() != 2 {
		t.Errorf("charge counters after another artifact = %d, want 2", charge())
	}
	effectstest.CastCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if charge() != 2 {
		t.Errorf("a creature added a counter: %d", charge())
	}
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Trinket", TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.PassPriorityAroundTable(t, g)
	if charge() != 2 {
		t.Errorf("an opponent's artifact added a counter: %d", charge())
	}
	_ = me
}

func TestDragonsparkReactorSacrificeHitsPlayerAndCreatureForTheCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	reactor := effectstest.PushCatalogPermanent(g, me.ID, "Dragonspark Reactor", "Artifact", effectstest.SamiDragonsparkOracle, false)
	g.WithWriteLock(func() {
		if c := effectstest.BattlefieldCardFor(g, reactor); c != nil {
			if c.Counters == nil {
				c.Counters = map[string]int{}
			}
			c.Counters[game.CounterCharge] = 3
		}
	})
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 3, 3)
	effectstest.SamiMainPhase(t, g)
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, reactor, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{
			{Kind: game.TargetPlayer, ID: opp.ID},
			{Kind: game.TargetCard, ID: bear},
		},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(reactor) {
		t.Error("the Reactor was not sacrificed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d → %d, want 3 (the counters)", life, opp.Life)
	}
	if g.Battlefield.Contains(bear) {
		t.Error("3 damage did not kill the 3/3")
	}
}

func TestDragonsparkReactorCreatureTargetIsOptional(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	reactor := effectstest.PushCatalogPermanent(g, me.ID, "Dragonspark Reactor", "Artifact", effectstest.SamiDragonsparkOracle, false)
	g.WithWriteLock(func() {
		if c := effectstest.BattlefieldCardFor(g, reactor); c != nil {
			c.Counters = map[string]int{game.CounterCharge: 2}
		}
	})
	effectstest.SamiMainPhase(t, g)
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, reactor, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("player only: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life %d → %d, want 2", life, opp.Life)
	}
}

func TestMemorialVaultExilesOnePlusTheSacrificedManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	vault := effectstest.PushCatalogPermanent(g, me.ID, "Memorial Vault", "Artifact", effectstest.SamiMemorialVaultOracle, false)
	fodder := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Three-drop", TypeLine: "Artifact", ManaCost: "{3}",
		Owner: me.ID, Controller: me.ID,
	})
	ids := effectstest.SamiLibraryTopFirst(me,
		game.Card{Name: "One"}, game.Card{Name: "Two"}, game.Card{Name: "Three"},
		game.Card{Name: "Four"}, game.Card{Name: "Five"}, game.Card{Name: "Six"},
	)
	effectstest.SamiMainPhase(t, g)
	if err := g.ActivateCatalogAbility(me.ID, vault, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the artifact was not sacrificed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	for i, id := range ids {
		if want := i < 4; g.Exile.Contains(id) != want {
			t.Errorf("card %d exiled = %v, want %v (X = 1 + 3 = 4)", i, !want, want)
		}
	}
}

func TestMemorialVaultRefusesItselfAsTheSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	vault := effectstest.PushCatalogPermanent(g, me.ID, "Memorial Vault", "Artifact", effectstest.SamiMemorialVaultOracle, false)
	effectstest.SamiMainPhase(t, g)
	if err := g.ActivateCatalogAbility(me.ID, vault, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{vault}}); err == nil {
		t.Fatal("the Vault sacrificed itself to pay for \"another artifact\"")
	}
}
