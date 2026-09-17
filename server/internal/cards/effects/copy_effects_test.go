package effects

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// copy_effects_test.go — the four Clone-class cards, end to end
// through the real cast → resolve → replacement → prompt → entry
// pipeline. What is asserted is deliberately the OBSERVABLE result
// (name, P/T, types, counters, catalog behaviour), because that is
// the only thing that tells "the copy landed" apart from "the copy
// landed on a card nobody can see".

const (
	oracleClone      = "42226b87-0746-4ebf-9fd0-108d508462af"
	oracleMetamorph  = "340bbe8b-e987-4c3e-ab4e-9dee63e57d4f"
	oracleSakashima  = "a7243d25-22a2-4df5-adaf-1f40f5330ec1"
	oracleSparkDoubl = "8dcb35e5-ae44-455f-86e3-4a77d496ff34"
)

// copyPrompt returns the queued copy-target choice, or nil.
func copyPrompt(g *game.Game) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceCopyTarget {
			return c
		}
	}
	return nil
}

// resolveWithCopyChoice passes priority until the copy picker opens,
// answers it with `pick` (uuid.Nil declines), then settles the
// stack. Fails the test if no picker ever appears — every card in
// this file is supposed to ask.
func resolveWithCopyChoice(t *testing.T, g *game.Game, pick uuid.UUID) {
	t.Helper()
	asked := false
	for i := 0; i < 32; i++ {
		if c := copyPrompt(g); c != nil {
			asked = true
			if err := g.ResolveCopyTarget(c.ID, c.Chooser, pick); err != nil {
				t.Fatalf("ResolveCopyTarget: %v", err)
			}
			continue
		}
		if stackFullyEmpty(g) {
			break
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority iter %d: %v", i, err)
		}
	}
	if !asked {
		t.Fatal("no copy-target prompt was ever queued")
	}
}

// battlefieldCard returns the named battlefield card by value.
func copyBattlefieldCard(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c
		}
	}
	t.Fatalf("card %s is not on the battlefield", id)
	return game.Card{}
}

// seedCreature puts a vanilla creature on the battlefield under the
// active seat's control, through the zone-move event so the layer
// listener stamps it.
func seedCopyableCreature(g *game.Game, controller uuid.UUID, name, typeLine string, p, tough int) uuid.UUID {
	id := uuid.New()
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		OracleID:   "oracle-" + name,
		ManaCost:   "{1}{G}",
		Power:      p,
		Toughness:  tough,
		Owner:      controller,
		Controller: controller,
	})
}

// TestCloneEntersAsACopyOfTheChosenCreature is the headline: #335's
// "enters as a 0/0 with nothing to pick" becomes a real picker and a
// real copy.
func TestCloneEntersAsACopyOfTheChosenCreature(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	bears := seedCopyableCreature(g, active.ID, "Grizzly Bears", "Creature — Bear", 2, 2)

	cloneID := castCatalogSpell(t, g, "Clone", "Creature — Shapeshifter", oracleClone, nil)
	resolveWithCopyChoice(t, g, bears)

	got := copyBattlefieldCard(t, g, cloneID)
	if got.Name != "Grizzly Bears" {
		t.Errorf("name = %q, want %q", got.Name, "Grizzly Bears")
	}
	if effectivePower(t, g, cloneID) != 2 || effectiveToughness(t, g, cloneID) != 2 {
		t.Errorf("P/T = %d/%d, want 2/2",
			effectivePower(t, g, cloneID), effectiveToughness(t, g, cloneID))
	}
	if !got.IsCreature() {
		t.Error("clone is not a creature")
	}
	if !got.IsCopy() {
		t.Error("clone does not report as a copy")
	}
}

// TestCloneOffersOnlyCreatures — the candidate list is the card's
// own restriction, and it is what the client renders.
func TestCloneOffersOnlyCreatures(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	bears := seedCopyableCreature(g, active.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Sol Ring",
		TypeLine:   "Artifact",
		OracleID:   "oracle-sol-ring",
		Owner:      active.ID,
		Controller: active.ID,
	})

	castCatalogSpell(t, g, "Clone", "Creature — Shapeshifter", oracleClone, nil)
	for i := 0; i < 8 && copyPrompt(g) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	pc := copyPrompt(g)
	if pc == nil {
		t.Fatal("no copy prompt")
	}
	if len(pc.CopyOptions) != 1 || pc.CopyOptions[0] != bears {
		t.Errorf("CopyOptions = %v, want exactly [%s] (the artifact %s is not a legal choice)",
			pc.CopyOptions, bears, rock)
	}
}

// TestCloneDecliningLeavesItAsItsPrintedSelf — "you may" means the
// decline is a real answer, and the permanent enters as the 0/0
// Clone actually is.
//
// It then SURVIVES, which the printed card does not: the CR 704.5f
// state-based action deliberately skips creatures with printed
// toughness 0 and no counters, because that is also the engine's
// placeholder convention for cards with unparseable stats (see
// stateBasedActionsLocked). Pre-existing, engine-wide, and out of
// scope here — pinned so that the day the convention changes, this
// test says so rather than a player noticing.
func TestCloneDecliningLeavesItAsItsPrintedSelf(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	seedCopyableCreature(g, active.ID, "Grizzly Bears", "Creature — Bear", 2, 2)

	cloneID := castCatalogSpell(t, g, "Clone", "Creature — Shapeshifter", oracleClone, nil)
	resolveWithCopyChoice(t, g, uuid.Nil)

	got := copyBattlefieldCard(t, g, cloneID)
	if got.Name != "Clone" {
		t.Errorf("name = %q, want %q — declining copies nothing", got.Name, "Clone")
	}
	if got.IsCopy() {
		t.Error("a declined Clone reports as a copy")
	}
	if got.CurrentToughness() != 0 {
		t.Errorf("toughness = %d, want the printed 0", got.CurrentToughness())
	}
}

// TestCloneCopiesPrintedValuesNotLayeredOnes is CR 707.2 through the
// real pipeline: the copied creature is wearing +1/+1 counters, and
// the clone is not.
func TestCloneCopiesPrintedValuesNotLayeredOnes(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	bears := seedCopyableCreature(g, active.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	if err := g.AddCounter(bears, "+1/+1", 2); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}

	cloneID := castCatalogSpell(t, g, "Clone", "Creature — Shapeshifter", oracleClone, nil)
	resolveWithCopyChoice(t, g, bears)

	got := copyBattlefieldCard(t, g, cloneID)
	if got.CurrentPower() != 2 || got.CurrentToughness() != 2 {
		t.Errorf("clone is %d/%d, want 2/2 — counters are not copiable values",
			got.CurrentPower(), got.CurrentToughness())
	}
}

// TestCloneOfACatalogCardInheritsItsBehaviour is the reason the copy
// carries the oracle ID rather than only a Characteristic: every
// catalog hook keys on CatalogKey, so a Clone of Lord of Atlantis
// IS a Lord of Atlantis and pumps the other Merfolk. A copy that
// only rewrote name and P/T would look identical on the wire and do
// nothing.
func TestCloneOfACatalogCardInheritsItsBehaviour(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	lord := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Lord of Atlantis",
		TypeLine:   "Creature — Merfolk",
		OracleID:   lordOfAtlantisOracle,
		Power:      2,
		Toughness:  2,
		Owner:      active.ID,
		Controller: active.ID,
	})
	merfolk := seedCopyableCreature(g, active.ID, "Merfolk of the Pearl Trident",
		"Creature — Merfolk", 1, 1)

	// One lord: the vanilla Merfolk is 2/2.
	if got := effectivePower(t, g, merfolk); got != 2 {
		t.Fatalf("baseline Merfolk power = %d, want 2 (1 printed + one lord)", got)
	}

	cloneID := castCatalogSpell(t, g, "Clone", "Creature — Shapeshifter", oracleClone, nil)
	resolveWithCopyChoice(t, g, lord)

	if got := effectivePower(t, g, merfolk); got != 3 {
		t.Errorf("Merfolk power = %d, want 3 — the Clone of Lord of Atlantis should pump it too", got)
	}
	if got := effectivePower(t, g, cloneID); got != 3 {
		t.Errorf("clone power = %d, want 3 (printed 2 + the ORIGINAL lord's pump)", got)
	}
}

// TestSakashimaKeepsItsOwnNameAndStaysLegendary — the name exception
// is what keeps a copy of your own commander alive alongside it.
func TestSakashimaKeepsItsOwnNameAndStaysLegendary(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	target := seedCopyableCreature(g, active.ID, "Grizzly Bears", "Creature — Bear", 2, 2)

	id := castCatalogSpell(t, g, "Sakashima the Impostor",
		"Legendary Creature — Human Rogue", oracleSakashima, nil)
	resolveWithCopyChoice(t, g, target)

	got := copyBattlefieldCard(t, g, id)
	if got.Name != "Sakashima the Impostor" {
		t.Errorf("name = %q, want it to keep its own", got.Name)
	}
	if got.CurrentPower() != 2 || got.CurrentToughness() != 2 {
		t.Errorf("P/T = %d/%d, want the copied 2/2", got.CurrentPower(), got.CurrentToughness())
	}
	if !hasCopySupertype(got, "Legendary") {
		t.Errorf("type line = %q, want it legendary in addition", got.TypeLine)
	}
}

func TestSakashimaCommanderCopyKeepsCommandTowerBlue(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	target := seedCopyableCreature(g, active.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	cmdr := game.NewCommander("Sakashima the Impostor", active.ID)
	cmdr.OracleID = oracleSakashima
	cmdr.TypeLine = "Legendary Creature — Human Rogue"
	cmdr.ManaCost = "{2}{U}{U}"
	cmdr.ColorIdentity = []string{"U"}
	cmdr.Power, cmdr.Toughness = 3, 1
	active.Command.PushTop(cmdr)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 4; i++ {
		active.ManaPool.AddMana(game.ManaToken{Color: "U"})
	}
	if err := g.CastSpell(active.ID, cmdr.InstanceID, game.CastSpellParams{FromZone: "command", Strict: true}); err != nil {
		t.Fatalf("cast commander: %v", err)
	}
	resolveWithCopyChoice(t, g, target)
	if got := copyBattlefieldCard(t, g, cmdr.InstanceID); !got.IsCopy() || got.CurrentPower() != 2 {
		t.Fatal("Sakashima did not enter as a copy of the green creature")
	}
	tower := pushCatalogPermanent(g, active.ID, "Command Tower", "Land", "0895c9b7-ae7d-4bb3-af17-3b75deb50a25", false)
	blueCost, err := game.ParseCost("{U}")
	if err != nil {
		t.Fatal(err)
	}
	if plan, ok := g.AutoTapForCost(active.ID, blueCost, 0); !ok || !reflect.DeepEqual(plan, []uuid.UUID{tower}) {
		t.Errorf("auto-tap for blue = %v, %v; want Command Tower", plan, ok)
	}
	if err := g.ActivateManaAbility(active.ID, tower, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	pick := riderLatestManaPick(g, active.ID)
	if pick == nil || !reflect.DeepEqual(pick.ColorOptions, []string{"U"}) {
		t.Fatalf("Command Tower options with copied commander = %+v, want [U]", pick)
	}
	if err := g.ResolveManaChoice(pick.ID, active.ID, "U"); err != nil {
		t.Fatal(err)
	}
}

// TestSparkDoubleIsNotLegendaryAndGetsAnExtraCounter — three
// exceptions, two of which land somewhere other than the copiable
// values.
func TestSparkDoubleIsNotLegendaryAndGetsAnExtraCounter(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	commander := seedCopyableCreature(g, active.ID, "Kenrith, the Returned King",
		"Legendary Creature — Human Noble", 5, 5)

	id := castCatalogSpell(t, g, "Spark Double", "Creature — Illusion", oracleSparkDoubl, nil)
	resolveWithCopyChoice(t, g, commander)

	got := copyBattlefieldCard(t, g, id)
	if got.Name != "Kenrith, the Returned King" {
		t.Errorf("name = %q, want the copied one", got.Name)
	}
	if hasCopySupertype(got, "Legendary") {
		t.Errorf("type line = %q, want the legendary supertype dropped", got.TypeLine)
	}
	if got.Counters["+1/+1"] != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got.Counters["+1/+1"])
	}
	if got.CurrentPower() != 6 || got.CurrentToughness() != 6 {
		t.Errorf("P/T = %d/%d, want 6/6 (5/5 plus the counter)",
			got.CurrentPower(), got.CurrentToughness())
	}
	// Both permanents are on the board. The CR 704.5j legend rule is
	// not implemented yet, so this cannot fail today — it is here so
	// that the day it lands, the card that exists to dodge it is
	// already proved to dodge it.
	if !g.Battlefield.Contains(commander) {
		t.Error("the copied commander is gone")
	}
}

// TestSparkDoubleOnlyOffersYourOwnPermanents — the candidate filter
// is per-card, and this one is "a creature or planeswalker YOU
// control".
func TestSparkDoubleOnlyOffersYourOwnPermanents(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opponent := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := seedCopyableCreature(g, active.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	theirs := seedCopyableCreature(g, opponent.ID, "Runeclaw Bear", "Creature — Bear", 2, 2)

	castCatalogSpell(t, g, "Spark Double", "Creature — Illusion", oracleSparkDoubl, nil)
	for i := 0; i < 8 && copyPrompt(g) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	pc := copyPrompt(g)
	if pc == nil {
		t.Fatal("no copy prompt")
	}
	if len(pc.CopyOptions) != 1 || pc.CopyOptions[0] != mine {
		t.Errorf("CopyOptions = %v, want exactly [%s]; %s belongs to an opponent",
			pc.CopyOptions, mine, theirs)
	}
}

// TestPhyrexianMetamorphCopiesAnArtifactAndStaysOne — the except
// clause ADDS a type rather than replacing the line, so a copied
// artifact is still just an artifact and a copied creature becomes
// an artifact creature.
func TestPhyrexianMetamorphCopiesAnArtifactAndStaysOne(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	creature := seedCopyableCreature(g, active.ID, "Grizzly Bears", "Creature — Bear", 2, 2)

	id := castCatalogSpell(t, g, "Phyrexian Metamorph",
		"Artifact Creature — Phyrexian Shapeshifter", oracleMetamorph, nil)
	resolveWithCopyChoice(t, g, creature)

	got := copyBattlefieldCard(t, g, id)
	if !got.IsArtifact() {
		t.Errorf("type line = %q, want artifact in addition", got.TypeLine)
	}
	if !got.IsCreature() {
		t.Errorf("type line = %q, want it still a creature", got.TypeLine)
	}
	if got.Name != "Grizzly Bears" {
		t.Errorf("name = %q, want the copied one", got.Name)
	}
}

// TestPhyrexianMetamorphOffersArtifactsToo — the wider candidate set
// is the reason to play it over Clone.
func TestPhyrexianMetamorphOffersArtifactsToo(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Sol Ring",
		TypeLine:   "Artifact",
		OracleID:   "oracle-sol-ring",
		Owner:      active.ID,
		Controller: active.ID,
	})

	id := castCatalogSpell(t, g, "Phyrexian Metamorph",
		"Artifact Creature — Phyrexian Shapeshifter", oracleMetamorph, nil)
	resolveWithCopyChoice(t, g, rock)

	got := copyBattlefieldCard(t, g, id)
	if got.Name != "Sol Ring" {
		t.Errorf("name = %q, want %q", got.Name, "Sol Ring")
	}
	if got.IsCreature() {
		t.Errorf("type line = %q — copying an artifact does not make it a creature", got.TypeLine)
	}
}

// TestCopyChoiceSurvivesACR616OrderingPrompt — when a second
// replacement also applies to the same entry, the affected player
// orders them first (CR 616) and the copy picker has to come AFTER
// that, not be fired blind. Firing a copy selector blind would mark
// it applied and drop the copy silently, which is the one failure
// mode a player could not diagnose.
func TestCopyChoiceSurvivesACR616OrderingPrompt(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	bears := seedCopyableCreature(g, active.ID, "Grizzly Bears", "Creature — Bear", 2, 2)

	// A Kismet-shaped second replacement on the same entry.
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(game.ReplacementEffect{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) bool {
				return ev.Kind == game.RepEventMove && ev.NewZone == game.ZoneBattlefield
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.EntersTapped = true
				return nil
			},
			Label: "test enters tapped",
		})
	})

	cloneID := castCatalogSpell(t, g, "Clone", "Creature — Shapeshifter", oracleClone, nil)

	// Drain to the ordering prompt and answer it.
	var order *game.PendingChoice
	for i := 0; i < 8; i++ {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceReplacementOrder {
				order = c
			}
		}
		if order != nil {
			break
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if order == nil {
		t.Fatal("no CR 616 ordering prompt")
	}
	if err := g.ResolveReplacementOrder(order.ID, order.Chooser, order.ReplacementEffectIDs); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}

	pc := copyPrompt(g)
	if pc == nil {
		t.Fatal("the copy picker was never offered after the ordering prompt")
	}
	if err := g.ResolveCopyTarget(pc.ID, pc.Chooser, bears); err != nil {
		t.Fatalf("ResolveCopyTarget: %v", err)
	}

	got := copyBattlefieldCard(t, g, cloneID)
	if got.Name != "Grizzly Bears" {
		t.Errorf("name = %q, want the copy to have landed", got.Name)
	}
	if !got.Tapped {
		t.Error("the other replacement's enters-tapped was lost")
	}
}

// hasCopySupertype reports whether the card's EFFECTIVE type line
// carries the supertype.
func hasCopySupertype(c game.Card, s string) bool {
	for _, got := range c.Effective().Supertypes {
		if got == s {
			return true
		}
	}
	return false
}
