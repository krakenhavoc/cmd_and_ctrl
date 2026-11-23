package testscm

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestCloneEntersAsACopyOfTheChosenCreature is the headline: #335's
// "enters as a 0/0 with nothing to pick" becomes a real picker and a
// real copy.
func TestCloneEntersAsACopyOfTheChosenCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	bears := effectstest.SeedCopyableCreature(g, active.ID, "Grizzly Bears", "Creature — Bear", 2, 2)

	cloneID := effectstest.CastCatalogSpell(t, g, "Clone", "Creature — Shapeshifter", effectstest.OracleClone, nil)
	effectstest.ResolveWithCopyChoice(t, g, bears)

	got := effectstest.CopyBattlefieldCard(t, g, cloneID)
	if got.Name != "Grizzly Bears" {
		t.Errorf("name = %q, want %q", got.Name, "Grizzly Bears")
	}
	if effectstest.EffectivePower(t, g, cloneID) != 2 || effectstest.EffectiveToughness(t, g, cloneID) != 2 {
		t.Errorf("P/T = %d/%d, want 2/2",
			effectstest.EffectivePower(t, g, cloneID), effectstest.EffectiveToughness(t, g, cloneID))
	}
	if !got.IsCreature() {
		t.Error("clone is not a creature")
	}
	if !got.IsCopy() {
		t.Error("clone does not report as a copy")
	}
}

// TestCloneOfACommanderIsNotACommander — #1363, CR 903.3. The Clone
// permanent is its own physical card; entering as a copy of a
// commander's copiable values (CR 707.2) does not make it that
// commander, because the designation is not one of the values CR
// 707.2 lists. PrintedValues (copy.go) carries no IsCommander field at
// all, and applyCopy/setPrintedValues never touch the Clone's own
// flag, so this is a straight-line pin of "already correct by
// construction" rather than a regression this PR introduced — added
// alongside the spell-copy fix because the issue asked every copy
// path to be checked the same way.
func TestCloneOfACommanderIsNotACommander(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	// Deliberately NOT "Legendary Creature": a Clone that copies a
	// legendary commander's name would also trip the CR 704.5j legend
	// rule, which is a real and separate interaction this test is not
	// about — isolating IsCommander from it keeps the assertions below
	// about that one flag rather than about which legendary the player
	// chooses to keep.
	cmdr := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID:  uuid.New(),
		Name:        "Fixture Commander",
		TypeLine:    "Creature — Test",
		Power:       2,
		Toughness:   2,
		Owner:       active.ID,
		Controller:  active.ID,
		IsCommander: true,
	})

	cloneID := effectstest.CastCatalogSpell(t, g, "Clone", "Creature — Shapeshifter", effectstest.OracleClone, nil)
	effectstest.ResolveWithCopyChoice(t, g, cmdr)

	got := effectstest.CopyBattlefieldCard(t, g, cloneID)
	if got.Name != "Fixture Commander" {
		t.Fatalf("name = %q, want %q — the copy did not land", got.Name, "Fixture Commander")
	}
	if got.IsCommander {
		t.Error("Clone entering as a copy of a commander carries IsCommander")
	}

	// Kill the Clone and confirm nothing asks to send it to the
	// command zone — it is Clone's owner's card, not the commander's.
	if err := g.DestroyPermanentForEffect(cloneID); err != nil {
		t.Fatalf("DestroyPermanentForEffect: %v", err)
	}
	if n := len(g.PendingChoices); n != 0 {
		t.Errorf("%d prompts queued destroying the Clone, want none: %+v", n, g.PendingChoices)
	}
	if n := active.CommanderDamage[cloneID]; n != 0 {
		t.Errorf("CommanderDamage[cloneID] = %d, want 0 — the Clone is not a commander source", n)
	}
}

// TestCloneOffersOnlyCreatures — the candidate list is the card's
// own restriction, and it is what the client renders.
func TestCloneOffersOnlyCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	bears := effectstest.SeedCopyableCreature(g, active.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Sol Ring",
		TypeLine:   "Artifact",
		OracleID:   "oracle-sol-ring",
		Owner:      active.ID,
		Controller: active.ID,
	})

	effectstest.CastCatalogSpell(t, g, "Clone", "Creature — Shapeshifter", effectstest.OracleClone, nil)
	for i := 0; i < 8 && effectstest.CopyPrompt(g) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	pc := effectstest.CopyPrompt(g)
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
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedCopyableCreature(g, active.ID, "Grizzly Bears", "Creature — Bear", 2, 2)

	cloneID := effectstest.CastCatalogSpell(t, g, "Clone", "Creature — Shapeshifter", effectstest.OracleClone, nil)
	effectstest.ResolveWithCopyChoice(t, g, uuid.Nil)

	got := effectstest.CopyBattlefieldCard(t, g, cloneID)
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
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	bears := effectstest.SeedCopyableCreature(g, active.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	if err := g.AddCounter(bears, "+1/+1", 2); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}

	cloneID := effectstest.CastCatalogSpell(t, g, "Clone", "Creature — Shapeshifter", effectstest.OracleClone, nil)
	effectstest.ResolveWithCopyChoice(t, g, bears)

	got := effectstest.CopyBattlefieldCard(t, g, cloneID)
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
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	lord := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Lord of Atlantis",
		TypeLine:   "Creature — Merfolk",
		OracleID:   effectstest.LordOfAtlantisOracle,
		Power:      2,
		Toughness:  2,
		Owner:      active.ID,
		Controller: active.ID,
	})
	merfolk := effectstest.SeedCopyableCreature(g, active.ID, "Merfolk of the Pearl Trident",
		"Creature — Merfolk", 1, 1)

	// One lord: the vanilla Merfolk is 2/2.
	if got := effectstest.EffectivePower(t, g, merfolk); got != 2 {
		t.Fatalf("baseline Merfolk power = %d, want 2 (1 printed + one lord)", got)
	}

	cloneID := effectstest.CastCatalogSpell(t, g, "Clone", "Creature — Shapeshifter", effectstest.OracleClone, nil)
	effectstest.ResolveWithCopyChoice(t, g, lord)

	if got := effectstest.EffectivePower(t, g, merfolk); got != 3 {
		t.Errorf("Merfolk power = %d, want 3 — the Clone of Lord of Atlantis should pump it too", got)
	}
	if got := effectstest.EffectivePower(t, g, cloneID); got != 3 {
		t.Errorf("clone power = %d, want 3 (printed 2 + the ORIGINAL lord's pump)", got)
	}
}

// TestSakashimaKeepsItsOwnNameAndStaysLegendary — the name exception
// is what keeps a copy of your own commander alive alongside it.
func TestSakashimaKeepsItsOwnNameAndStaysLegendary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	target := effectstest.SeedCopyableCreature(g, active.ID, "Grizzly Bears", "Creature — Bear", 2, 2)

	id := effectstest.CastCatalogSpell(t, g, "Sakashima the Impostor",
		"Legendary Creature — Human Rogue", effectstest.OracleSakashima, nil)
	effectstest.ResolveWithCopyChoice(t, g, target)

	got := effectstest.CopyBattlefieldCard(t, g, id)
	if got.Name != "Sakashima the Impostor" {
		t.Errorf("name = %q, want it to keep its own", got.Name)
	}
	if got.CurrentPower() != 2 || got.CurrentToughness() != 2 {
		t.Errorf("P/T = %d/%d, want the copied 2/2", got.CurrentPower(), got.CurrentToughness())
	}
	if !effectstest.HasCopySupertype(got, "Legendary") {
		t.Errorf("type line = %q, want it legendary in addition", got.TypeLine)
	}
}

func TestSakashimaCommanderCopyKeepsCommandTowerBlue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	target := effectstest.SeedCopyableCreature(g, active.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	cmdr := game.NewCommander("Sakashima the Impostor", active.ID)
	cmdr.OracleID = effectstest.OracleSakashima
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
	effectstest.ResolveWithCopyChoice(t, g, target)
	if got := effectstest.CopyBattlefieldCard(t, g, cmdr.InstanceID); !got.IsCopy() || got.CurrentPower() != 2 {
		t.Fatal("Sakashima did not enter as a copy of the green creature")
	}
	tower := effectstest.PushCatalogPermanent(g, active.ID, "Command Tower", "Land", "0895c9b7-ae7d-4bb3-af17-3b75deb50a25", false)
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
	pick := effectstest.RiderLatestManaPick(g, active.ID)
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
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	commander := effectstest.SeedCopyableCreature(g, active.ID, "Kenrith, the Returned King",
		"Legendary Creature — Human Noble", 5, 5)

	id := effectstest.CastCatalogSpell(t, g, "Spark Double", "Creature — Illusion", effectstest.OracleSparkDoubl, nil)
	effectstest.ResolveWithCopyChoice(t, g, commander)

	got := effectstest.CopyBattlefieldCard(t, g, id)
	if got.Name != "Kenrith, the Returned King" {
		t.Errorf("name = %q, want the copied one", got.Name)
	}
	if effectstest.HasCopySupertype(got, "Legendary") {
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
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opponent := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.SeedCopyableCreature(g, active.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	theirs := effectstest.SeedCopyableCreature(g, opponent.ID, "Runeclaw Bear", "Creature — Bear", 2, 2)

	effectstest.CastCatalogSpell(t, g, "Spark Double", "Creature — Illusion", effectstest.OracleSparkDoubl, nil)
	for i := 0; i < 8 && effectstest.CopyPrompt(g) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	pc := effectstest.CopyPrompt(g)
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
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	creature := effectstest.SeedCopyableCreature(g, active.ID, "Grizzly Bears", "Creature — Bear", 2, 2)

	id := effectstest.CastCatalogSpell(t, g, "Phyrexian Metamorph",
		"Artifact Creature — Phyrexian Shapeshifter", effectstest.OracleMetamorph, nil)
	effectstest.ResolveWithCopyChoice(t, g, creature)

	got := effectstest.CopyBattlefieldCard(t, g, id)
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
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Sol Ring",
		TypeLine:   "Artifact",
		OracleID:   "oracle-sol-ring",
		Owner:      active.ID,
		Controller: active.ID,
	})

	id := effectstest.CastCatalogSpell(t, g, "Phyrexian Metamorph",
		"Artifact Creature — Phyrexian Shapeshifter", effectstest.OracleMetamorph, nil)
	effectstest.ResolveWithCopyChoice(t, g, rock)

	got := effectstest.CopyBattlefieldCard(t, g, id)
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
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	bears := effectstest.SeedCopyableCreature(g, active.ID, "Grizzly Bears", "Creature — Bear", 2, 2)

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

	cloneID := effectstest.CastCatalogSpell(t, g, "Clone", "Creature — Shapeshifter", effectstest.OracleClone, nil)

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

	pc := effectstest.CopyPrompt(g)
	if pc == nil {
		t.Fatal("the copy picker was never offered after the ordering prompt")
	}
	if err := g.ResolveCopyTarget(pc.ID, pc.Chooser, bears); err != nil {
		t.Fatalf("ResolveCopyTarget: %v", err)
	}

	got := effectstest.CopyBattlefieldCard(t, g, cloneID)
	if got.Name != "Grizzly Bears" {
		t.Errorf("name = %q, want the copy to have landed", got.Name)
	}
	if !got.Tapped {
		t.Error("the other replacement's enters-tapped was lost")
	}
}
