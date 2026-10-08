package effects

import (
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sami_c_test.go — the Sami Whammy deck's lands and removal artifacts
// (slice sami-c, tracker #2190).

const (
	samiConduitPylonsOracle    = "37f924e1-7c25-4f06-88bb-054693a21e5a"
	samiCrystalGrottoOracle    = "f15fb0cc-8e96-4f03-94d0-b51410415afd"
	samiEiganjoOracle          = "7edb3d15-4f70-4ebe-8c5e-caf6a225076d"
	samiFomoriVaultOracle      = "622a53bd-d894-422c-a606-7126041afa02"
	samiMonumentalHengeOracle  = "c48df45c-3513-4d56-aed6-30c2f3a759cd"
	samiNeedlevergePathOracle  = "a9b8d020-4d72-4934-8942-df29ef19fc1d"
	samiSokenzanOracle         = "c5ee72d5-3a9e-4fe5-8802-3286ee612055"
	samiSpireOfIndustryOracle  = "55a519b4-61cb-448a-875b-4d6dbe00580f"
	samiSunbillowVergeOracle   = "a202276b-1f1b-4277-95ee-26877a204f5e"
	samiBlitzballOracle        = "75719f87-ce1a-40d6-ac5a-5ebf0c936c84"
	samiBoommobileOracle       = "cb3edb86-ddd5-461c-98b5-ffb4df437a39"
	samiDragonsparkOracle      = "a8acb0b0-7265-42f6-8143-62d0c02d5a28"
	samiMemorialVaultOracle    = "10dda10f-2549-47e5-9a8b-fb414cf38d3f"
	samiTestLibraryCardTypeFmt = "Sorcery"
)

func TestSamiCCardsAreRegistered(t *testing.T) {
	want := map[string]string{
		samiConduitPylonsOracle:   "Conduit Pylons",
		samiCrystalGrottoOracle:   "Crystal Grotto",
		samiEiganjoOracle:         "Eiganjo, Seat of the Empire",
		samiFomoriVaultOracle:     "Fomori Vault",
		samiMonumentalHengeOracle: "Monumental Henge",
		samiNeedlevergePathOracle: "Needleverge Pathway",
		samiSokenzanOracle:        "Sokenzan, Crucible of Defiance",
		samiSpireOfIndustryOracle: "Spire of Industry",
		samiSunbillowVergeOracle:  "Sunbillow Verge",
		samiBlitzballOracle:       "Blitzball",
		samiBoommobileOracle:      "Boommobile",
		samiDragonsparkOracle:     "Dragonspark Reactor",
		samiMemorialVaultOracle:   "Memorial Vault",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
		if spec.Completeness != CompletenessFull {
			t.Errorf("%s completeness = %v, want full", name, spec.Completeness)
		}
	}
	if _, ok := Lookup(game.CatalogKeyForFace(samiNeedlevergePathOracle, 1)); !ok {
		t.Error("Pillarverge Pathway (the back face) is not registered")
	}
}

// samiLibraryTopFirst seeds a player's library so the first entry is
// the top card, with the given type lines.
func samiLibraryTopFirst(p *game.Player, cards ...game.Card) []uuid.UUID {
	ids := make([]uuid.UUID, len(cards))
	for i := range cards {
		ids[i] = uuid.New()
	}
	for i := len(cards) - 1; i >= 0; i-- {
		c := cards[i]
		c.InstanceID = ids[i]
		c.Owner, c.Controller = p.ID, p.ID
		if c.TypeLine == "" {
			c.TypeLine = samiTestLibraryCardTypeFmt
		}
		p.Library.PushTop(c)
	}
	return ids
}

func samiMainPhase(t *testing.T, g *game.Game) {
	t.Helper()
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
}

// --- the lands -----------------------------------------------------

func TestConduitPylonsEntersUntappedAndSurveils(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "Top", "Next")

	id := playLandFromHand(t, g, "Conduit Pylons", samiConduitPylonsOracle)
	if c, _ := battlefieldCard(g, id); c.Tapped {
		t.Error("Conduit Pylons entered tapped")
	}
	if triggerOnStack(g, id) == nil {
		t.Fatal("the surveil trigger did not go on the stack")
	}
	passPriorityAroundTable(t, g)
	if latestChoiceOfKindFor(g, game.PendingChoiceSurveil, me.ID) == nil {
		t.Error("resolving the trigger queued no surveil prompt")
	}
}

func TestConduitPylonsFilterNeedsOneMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := seedPermanentWithOracle(g, me.ID, "Conduit Pylons", "Land — Desert", samiConduitPylonsOracle)

	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("{T}: Add {C}: %v", err)
	}
	if got := poolColors(me); !reflect.DeepEqual(got, []string{"C"}) {
		t.Errorf("pool = %v, want {C}", got)
	}

	other := seedPermanentWithOracle(g, me.ID, "Conduit Pylons", "Land — Desert", samiConduitPylonsOracle)
	me.ManaPool = nil
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateManaAbility(me.ID, other, 1, game.ManaAbilityParams{Colors: []string{"U"}}); err != nil {
		t.Fatalf("{1},{T}: any colour: %v", err)
	}
	if got := poolColors(me); !reflect.DeepEqual(got, []string{"U"}) {
		t.Errorf("pool after the filter = %v, want {U} (the {C} paid for it)", got)
	}
}

func TestCrystalGrottoEntersUntappedAndScries(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "Top", "Next")

	id := playLandFromHand(t, g, "Crystal Grotto", samiCrystalGrottoOracle)
	if c, _ := battlefieldCard(g, id); c.Tapped {
		t.Error("Crystal Grotto entered tapped")
	}
	if triggerOnStack(g, id) == nil {
		t.Fatal("the scry trigger did not go on the stack")
	}
	passPriorityAroundTable(t, g)
	if scryChoiceFor(g, me.ID) == nil {
		t.Error("resolving the trigger queued no scry prompt")
	}
}

func TestSpireOfIndustryNeedsAnArtifactAndCostsOneLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	spire := seedPermanentWithOracle(g, me.ID, "Spire of Industry", "Land", samiSpireOfIndustryOracle)
	life := me.Life

	if err := g.ActivateManaAbility(me.ID, spire, 1, game.ManaAbilityParams{Colors: []string{"G"}}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("any colour with no artifact: err = %v, want ErrConditionNotMet", err)
	}
	if c, _ := battlefieldCard(g, spire); c.Tapped {
		t.Error("a refused activation tapped the land")
	}
	if me.Life != life {
		t.Errorf("a refused activation cost life: %d → %d", life, me.Life)
	}

	seedPermanentWithOracle(g, me.ID, "Trinket", "Artifact", "")
	if err := g.ActivateManaAbility(me.ID, spire, 1, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("any colour with an artifact: %v", err)
	}
	if got := poolColors(me); !reflect.DeepEqual(got, []string{"G"}) {
		t.Errorf("pool = %v, want {G}", got)
	}
	if me.Life != life-1 {
		t.Errorf("life %d → %d, want one paid", life, me.Life)
	}
}

func TestSunbillowVergeRedNeedsAMountainOrAPlains(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	verge := seedPermanentWithOracle(g, me.ID, "Sunbillow Verge", "Land", samiSunbillowVergeOracle)
	seedManaLand(g, me.ID, "Swamp", "Basic Land — Swamp", "B")

	if err := g.ActivateManaAbility(me.ID, verge, 1, game.ManaAbilityParams{}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("{R} beside a Swamp: err = %v, want ErrConditionNotMet", err)
	}
	if err := g.ActivateManaAbility(me.ID, verge, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("unconditional {W}: %v", err)
	}
	if got := poolColors(me); !reflect.DeepEqual(got, []string{"W"}) {
		t.Errorf("pool = %v, want {W}", got)
	}

	me.ManaPool = nil
	verge2 := seedPermanentWithOracle(g, me.ID, "Sunbillow Verge", "Land", samiSunbillowVergeOracle)
	seedManaLand(g, me.ID, "Plains", "Basic Land — Plains", "W")
	if err := g.ActivateManaAbility(me.ID, verge2, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("{R} beside a Plains: %v", err)
	}
	if got := poolColors(me); !reflect.DeepEqual(got, []string{"R"}) {
		t.Errorf("pool = %v, want {R}", got)
	}
}

func TestNeedlevergePathwayFrontTapsForRed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := seedPermanentWithOracle(g, me.ID, "Needleverge Pathway", "Land", samiNeedlevergePathOracle)
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("{T}: Add {R}: %v", err)
	}
	if got := poolColors(me); !reflect.DeepEqual(got, []string{"R"}) {
		t.Errorf("pool = %v, want {R}", got)
	}
}

func TestNeedlevergePathwayBackFaceTapsForWhite(t *testing.T) {
	g, me, id := handWithMDFC(t, func(owner uuid.UUID) game.Card {
		return mdfcCard(owner, samiNeedlevergePathOracle, game.LayoutModalDFC,
			game.Face{Name: "Needleverge Pathway", TypeLine: "Land"},
			game.Face{Name: "Pillarverge Pathway", TypeLine: "Land"},
		)
	})
	me.LandDropsPerTurn++
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("play Pillarverge Pathway: %v", err)
	}
	if c, ok := battlefieldCard(g, id); !ok || c.Tapped {
		t.Fatal("the back face did not enter untapped")
	}
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("back face {T}: %v", err)
	}
	if got := poolColors(me); !reflect.DeepEqual(got, []string{"W"}) {
		t.Errorf("pool = %v, want {W}", got)
	}
}

func TestNeedlevergePathwayFrontFacePlayedFromHandTapsForRed(t *testing.T) {
	g, me, id := handWithMDFC(t, func(owner uuid.UUID) game.Card {
		return mdfcCard(owner, samiNeedlevergePathOracle, game.LayoutModalDFC,
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
	if got := poolColors(me); !reflect.DeepEqual(got, []string{"R"}) {
		t.Errorf("pool = %v, want {R}", got)
	}
}

func TestMonumentalHengeEntersTappedUnlessYouControlAPlains(t *testing.T) {
	g := newCatalogGame(t)
	id := playLandFromHand(t, g, "Monumental Henge", samiMonumentalHengeOracle)
	if c, _ := battlefieldCard(g, id); !c.Tapped {
		t.Error("Henge entered untapped with no Plains")
	}
	if n := tapEventsFor(g, id); n != 0 {
		t.Errorf("%d tap events: it should enter tapped, not be tapped", n)
	}

	g2 := newCatalogGame(t)
	me := g2.Seats[g2.Turn.ActiveSeat]
	seedManaLand(g2, me.ID, "Plains", "Basic Land — Plains", "W")
	id2 := playLandFromHand(t, g2, "Monumental Henge", samiMonumentalHengeOracle)
	if c, _ := battlefieldCard(g2, id2); c.Tapped {
		t.Error("Henge entered tapped with a Plains out")
	}
}

func TestMonumentalHengeFindsAHistoricCardAndBottomsTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	henge := seedPermanentWithOracle(g, me.ID, "Monumental Henge", "Land", samiMonumentalHengeOracle)
	ids := samiLibraryTopFirst(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Shock", TypeLine: "Instant"},
		game.Card{Name: "Wurm", TypeLine: "Artifact Creature — Wurm"},
		game.Card{Name: "Lone Hero", TypeLine: "Legendary Creature — Human"},
		game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Sixth", TypeLine: "Instant"},
	)
	samiMainPhase(t, g)
	libBefore := me.Library.Size()
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{W}{W}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, henge, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	pick := latestChooseCardsFor(g, me.ID)
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	henge := seedPermanentWithOracle(g, me.ID, "Monumental Henge", "Land", samiMonumentalHengeOracle)
	ids := samiLibraryTopFirst(me,
		game.Card{Name: "Wurm", TypeLine: "Artifact Creature — Wurm"},
		game.Card{Name: "B"}, game.Card{Name: "C"}, game.Card{Name: "D"}, game.Card{Name: "E"},
		game.Card{Name: "F"},
	)
	samiMainPhase(t, g)
	handBefore := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, henge, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	vault := seedPermanentWithOracle(g, me.ID, "Fomori Vault", "Land", samiFomoriVaultOracle)
	seedPermanentWithOracle(g, me.ID, "Trinket A", "Artifact", "")
	seedPermanentWithOracle(g, me.ID, "Trinket B", "Artifact", "")
	ids := samiLibraryTopFirst(me, game.Card{Name: "One"}, game.Card{Name: "Two"}, game.Card{Name: "Three"})
	samiMainPhase(t, g)
	discard := pushCatalogHandCard(me, "Pitch", "Sorcery", "")
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, vault, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{discard}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !me.Graveyard.Contains(discard) {
		t.Error("the discard cost did not discard")
	}
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	vault := seedPermanentWithOracle(g, me.ID, "Fomori Vault", "Land", samiFomoriVaultOracle)
	ids := samiLibraryTopFirst(me, game.Card{Name: "One"}, game.Card{Name: "Two"})
	samiMainPhase(t, g)
	discard := pushCatalogHandCard(me, "Pitch", "Sorcery", "")
	if err := g.ActivateCatalogAbility(me.ID, vault, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{discard}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if latestChooseCardsFor(g, me.ID) != nil {
		t.Error("X = 0 still opened a prompt")
	}
	if top := me.Library.Cards[me.Library.Size()-1].InstanceID; top != ids[0] {
		t.Error("the library changed")
	}
}

// --- the channel lands ---------------------------------------------

func TestEiganjoChannelHitsAnAttackingCreatureForFour(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	attacker := b12Creature(g, opp.ID, "Attacker", "Creature — Ogre", 4, 4)
	idle := b12Creature(g, opp.ID, "Idle", "Creature — Ogre", 4, 4)
	g.WithWriteLock(func() {
		if c := battlefieldCardFor(g, attacker); c != nil {
			c.AttackingTarget = me.ID
		}
	})
	samiMainPhase(t, g)
	id := pushCatalogHandCard(me, "Eiganjo, Seat of the Empire", "Legendary Land", samiEiganjoOracle)
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
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) {
		t.Error("4 damage did not kill the 4/4 attacker")
	}
	if !g.Battlefield.Contains(idle) {
		t.Error("the idle creature was hit")
	}
}

func TestEiganjoChannelCostsOneLessPerLegendaryCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	attacker := b12Creature(g, opp.ID, "Attacker", "Creature — Ogre", 2, 2)
	g.WithWriteLock(func() {
		if c := battlefieldCardFor(g, attacker); c != nil {
			c.AttackingTarget = me.ID
		}
	})
	samiMainPhase(t, g)
	id := pushCatalogHandCard(me, "Eiganjo, Seat of the Empire", "Legendary Land", samiEiganjoOracle)
	target := []game.TargetRef{{Kind: game.TargetCard, ID: attacker}}

	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{W}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Targets: target, Strict: true})
	if !errors.Is(err, game.ErrInsufficientMana) {
		t.Fatalf("{1}{W} with no legend: err = %v, want insufficient mana", err)
	}

	pushCatalogPermanent(g, me.ID, "Some Legend", "Legendary Creature — Human", "", false)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Targets: target, Strict: true}); err != nil {
		t.Fatalf("{1}{W} with one legend: %v", err)
	}
}

func TestSokenzanChannelMakesTwoHastySpirits(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	samiMainPhase(t, g)
	id := pushCatalogHandCard(me, "Sokenzan, Crucible of Defiance", "Legendary Land", samiSokenzanOracle)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{R}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("channel: %v", err)
	}
	if !me.Graveyard.Contains(id) {
		t.Error("the discard-this cost did not bin Sokenzan")
	}
	passPriorityAroundTable(t, g)

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
		if !effectiveAbilitiesContain(t, g, s, "haste") {
			t.Error("a Spirit token lacks haste")
		}
		if effectivePower(t, g, s) != 1 {
			t.Error("a Spirit token is not 1/1")
		}
	}
}

// --- the artifacts -------------------------------------------------

func TestBlitzballNeedsALegendToConnect(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ball := pushCatalogPermanent(g, me.ID, "Blitzball", "Artifact", samiBlitzballOracle, false)
	plain := b12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	legend := b12Creature(g, me.ID, "Legend", "Legendary Creature — Human", 2, 2)
	oppLegend := b12Creature(g, opp.ID, "Their Legend", "Legendary Creature — Human", 2, 2)
	seedLibrary(me, "A", "B", "C")
	before := me.Hand.Size()

	err := g.ActivateCatalogAbility(me.ID, ball, 0, game.ActivateAbilityParams{})
	if !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("no damage dealt: err = %v, want ErrConditionNotMet", err)
	}
	dealCombatDamage(g, me.ID, plain, opp.ID)
	if err := g.ActivateCatalogAbility(me.ID, ball, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("a nonlegendary creature connected: err = %v, want ErrConditionNotMet", err)
	}
	// An opponent's legend hitting ME is not "an opponent was dealt damage".
	dealCombatDamage(g, opp.ID, oppLegend, me.ID)
	if err := g.ActivateCatalogAbility(me.ID, ball, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("a legend hit me: err = %v, want ErrConditionNotMet", err)
	}
	dealCombatDamage(g, me.ID, legend, opp.ID)
	if err := g.ActivateCatalogAbility(me.ID, ball, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("a legend connected: %v", err)
	}
	if g.Battlefield.Contains(ball) {
		t.Error("Blitzball was not sacrificed")
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+2 {
		t.Errorf("hand %d → %d, want two cards drawn", before, got)
	}
}

func TestBlitzballCountsALegendThatDiedAfterConnecting(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ball := pushCatalogPermanent(g, me.ID, "Blitzball", "Artifact", samiBlitzballOracle, false)
	legend := b12Creature(g, me.ID, "Legend", "Legendary Creature — Human", 2, 2)
	seedLibrary(me, "A", "B")
	dealCombatDamage(g, me.ID, legend, opp.ID)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(legend) })

	if err := g.ActivateCatalogAbility(me.ID, ball, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("the legend traded after connecting: %v", err)
	}
}

func TestBlitzballTapsForAnyColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ball := pushCatalogPermanent(g, me.ID, "Blitzball", "Artifact", samiBlitzballOracle, false)
	if err := g.ActivateManaAbility(me.ID, ball, 0, game.ManaAbilityParams{Colors: []string{"B"}}); err != nil {
		t.Fatalf("mana: %v", err)
	}
	if got := poolColors(me); !reflect.DeepEqual(got, []string{"B"}) {
		t.Errorf("pool = %v, want {B}", got)
	}
}

func TestBoommobileMakesFourActivateOnlyMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "Boommobile", "Artifact — Vehicle", samiBoommobileOracle, nil)
	passPriorityAroundTable(t, g)

	pick := latestChoiceOfKind(g, game.PendingChoiceColor)
	if pick == nil {
		t.Fatal("no colour prompt for the four mana")
	}
	if err := g.ResolveColorChoice(pick.ID, me.ID, "R"); err != nil {
		t.Fatalf("ResolveColorChoice: %v", err)
	}
	if len(me.ManaPool) != 4 {
		t.Fatalf("pool = %d mana, want 4", len(me.ManaPool))
	}
	for _, tok := range me.ManaPool {
		if tok.Color != "R" {
			t.Errorf("mana colour %q, want R (one colour for all four)", tok.Color)
		}
		if !reflect.DeepEqual(tok.Restrictions, []string{game.ManaRestrictActivate}) {
			t.Errorf("restrictions = %v, want activate-only", tok.Restrictions)
		}
	}
}

func TestBoommobileExhaustDealsXAndGrowsOnce(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	car := pushCatalogPermanent(g, me.ID, "Boommobile", "Artifact — Vehicle", samiBoommobileOracle, false)
	samiMainPhase(t, g)
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
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life %d → %d, want 2 damage", life, opp.Life)
	}
	if c, _ := battlefieldCard(g, car); c.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", c.Counters[game.CounterPlusOne])
	}
	err := g.ActivateCatalogAbility(me.ID, car, 0, game.ActivateAbilityParams{XValue: 1, Targets: target})
	if !errors.Is(err, game.ErrAbilityExhausted) {
		t.Fatalf("second exhaust: err = %v, want ErrAbilityExhausted", err)
	}
}

func TestBoommobileCrewsAtTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	car := pushVehicleForTest(g, me.ID, "Boommobile", samiBoommobileOracle, 6, 4)
	weak := pushCrewerForTest(g, me.ID, "Weak", 1)
	strong := pushCrewerForTest(g, me.ID, "Strong", 2)

	if err := g.ActivateCatalogAbility(me.ID, car, 1, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{weak}}); err == nil {
		t.Fatal("crewing with power 1 was accepted for Crew 2")
	}
	if err := g.ActivateCatalogAbility(me.ID, car, 1, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{strong}}); err != nil {
		t.Fatalf("crew with power 2: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCardByID(g, car); !c.IsCreature() {
		t.Error("the crewed Vehicle is not a creature")
	}
}

func TestDragonsparkReactorCountsArtifactsYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	castCatalogSpell(t, g, "Dragonspark Reactor", "Artifact", samiDragonsparkOracle, nil)
	passPriorityAroundTable(t, g)
	var reactor uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.OracleID == samiDragonsparkOracle {
			reactor = c.InstanceID
		}
	}
	if reactor == uuid.Nil {
		t.Fatal("the Reactor did not resolve")
	}
	charge := func() int {
		c, _ := battlefieldCard(g, reactor)
		return c.Counters[game.CounterCharge]
	}
	if charge() != 1 {
		t.Fatalf("charge counters after its own entry = %d, want 1", charge())
	}
	castCatalogSpell(t, g, "Trinket", "Artifact", "", nil)
	passPriorityAroundTable(t, g)
	if charge() != 2 {
		t.Errorf("charge counters after another artifact = %d, want 2", charge())
	}
	castCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if charge() != 2 {
		t.Errorf("a creature added a counter: %d", charge())
	}
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Trinket", TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID,
	})
	passPriorityAroundTable(t, g)
	if charge() != 2 {
		t.Errorf("an opponent's artifact added a counter: %d", charge())
	}
	_ = me
}

func TestDragonsparkReactorSacrificeHitsPlayerAndCreatureForTheCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	reactor := pushCatalogPermanent(g, me.ID, "Dragonspark Reactor", "Artifact", samiDragonsparkOracle, false)
	g.WithWriteLock(func() {
		if c := battlefieldCardFor(g, reactor); c != nil {
			if c.Counters == nil {
				c.Counters = map[string]int{}
			}
			c.Counters[game.CounterCharge] = 3
		}
	})
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 3, 3)
	samiMainPhase(t, g)
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
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d → %d, want 3 (the counters)", life, opp.Life)
	}
	if g.Battlefield.Contains(bear) {
		t.Error("3 damage did not kill the 3/3")
	}
}

func TestDragonsparkReactorCreatureTargetIsOptional(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	reactor := pushCatalogPermanent(g, me.ID, "Dragonspark Reactor", "Artifact", samiDragonsparkOracle, false)
	g.WithWriteLock(func() {
		if c := battlefieldCardFor(g, reactor); c != nil {
			c.Counters = map[string]int{game.CounterCharge: 2}
		}
	})
	samiMainPhase(t, g)
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, reactor, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("player only: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life %d → %d, want 2", life, opp.Life)
	}
}

func TestMemorialVaultExilesOnePlusTheSacrificedManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	vault := pushCatalogPermanent(g, me.ID, "Memorial Vault", "Artifact", samiMemorialVaultOracle, false)
	fodder := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Three-drop", TypeLine: "Artifact", ManaCost: "{3}",
		Owner: me.ID, Controller: me.ID,
	})
	ids := samiLibraryTopFirst(me,
		game.Card{Name: "One"}, game.Card{Name: "Two"}, game.Card{Name: "Three"},
		game.Card{Name: "Four"}, game.Card{Name: "Five"}, game.Card{Name: "Six"},
	)
	samiMainPhase(t, g)
	if err := g.ActivateCatalogAbility(me.ID, vault, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the artifact was not sacrificed")
	}
	passPriorityAroundTable(t, g)
	for i, id := range ids {
		if want := i < 4; g.Exile.Contains(id) != want {
			t.Errorf("card %d exiled = %v, want %v (X = 1 + 3 = 4)", i, !want, want)
		}
	}
}

func TestMemorialVaultRefusesItselfAsTheSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	vault := pushCatalogPermanent(g, me.ID, "Memorial Vault", "Artifact", samiMemorialVaultOracle, false)
	samiMainPhase(t, g)
	if err := g.ActivateCatalogAbility(me.ID, vault, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{vault}}); err == nil {
		t.Fatal("the Vault sacrificed itself to pay for \"another artifact\"")
	}
}
