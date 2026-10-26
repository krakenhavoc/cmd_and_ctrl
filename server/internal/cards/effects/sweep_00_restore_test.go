package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sweep_00_restore_test.go — #4253's closure sweep, batch 00 (abdel_adrian
// through caustic_bronco): the prompts that carried a Go closure are keyed
// now, so a table waiting on any of them is a restore point. Each test opens
// the prompt, takes a STRICT restore through JSON (restoreRoundTrip with
// strict, which fails on a table that is not a restore point) and answers on
// the restored table.

// sweep00PickKey returns the open choose_cards prompt of `chooser`, failing
// the test when it names a key other than `want`.
func sweep00PickKey(t *testing.T, g *game.Game, chooser uuid.UUID, want string) *game.PendingChoice {
	t.Helper()
	c := latestChooseCardsFor(g, chooser)
	if c == nil {
		t.Fatalf("no choose-cards prompt for %s: %+v", chooser, g.PendingChoices)
	}
	if c.PickThen != want {
		t.Fatalf("the prompt names key %q, want %q", c.PickThen, want)
	}
	return c
}

// sweep00Graveyard puts n filler cards into a player's graveyard.
func sweep00Graveyard(g *game.Game, owner uuid.UUID, n int) {
	for i := 0; i < n; i++ {
		b38Graveyard(g, owner, "Filler", "Sorcery", "{1}", 0, 0)
	}
}

// Bolster with nothing hanging off it: the tie-break is a keyed pick whose
// counter count is the carried number, shared by every such bolster.
func TestSweep00BolsterTieIsARestorePoint(t *testing.T) {
	g, me, _ := flipMain(t)
	a := pushVanillaCreature(g, me.ID, "Tie A", 1, 2)
	b := pushVanillaCreature(g, me.ID, "Tie B", 5, 2)
	tall := pushVanillaCreature(g, me.ID, "Tall", 1, 5)
	g.WithWriteLock(func() {
		item := &game.StackItem{Kind: game.StackItemSpell, Controller: me.ID, Owner: me.ID}
		if err := (Bolster{N: 2}).Apply(NewContext(g, item)); err != nil {
			t.Fatal(err)
		}
	})
	sweep00PickKey(t, g, me.ID, "choose-cards/bolster")

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := sweep00PickKey(t, g, me.ID, "choose-cards/bolster")
	if len(c.ChooseCards) != 2 {
		t.Errorf("the restored tie offers %d creatures, want the two tied", len(c.ChooseCards))
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{b}); err != nil {
		t.Fatalf("ResolveChooseCards on the restored table: %v", err)
	}
	if bolsterCounters(t, g, b) != 2 || bolsterCounters(t, g, a) != 0 || bolsterCounters(t, g, tall) != 0 {
		t.Errorf("counters a=%d b=%d tall=%d on the restored table, want only b with 2",
			bolsterCounters(t, g, a), bolsterCounters(t, g, b), bolsterCounters(t, g, tall))
	}
}

// Teval (and Blossoming Tortoise, which shares the pick): the land returns
// tapped on the restored table, the milled land among the choices.
func TestSweep00TevalLandPickIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	teval := b21Push(g, me.ID, "Teval, the Balanced Scale", "Legendary Creature — Spirit Dragon", b25TevalOracle, 4, 4, "B", "G", "U")
	milled := b10LibraryTop(me, "Milled Swamp", "Basic Land — Swamp", "", 0, 0)
	b10LibraryTop(me, "Filler A", "Instant", "{R}", 0, 0)
	b10LibraryTop(me, "Filler B", "Instant", "{R}", 0, 0)
	declareAttack(t, g, opp.ID, teval)
	passPriorityAroundTable(t, g)
	sweep00PickKey(t, g, me.ID, "choose-cards/mill-three-return-land-tapped")

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := sweep00PickKey(t, g, me.ID, "choose-cards/mill-three-return-land-tapped")
	if !hasID(c.ChooseCards, milled) {
		t.Fatal("the restored pick does not offer the land the trigger milled")
	}
	answerChooseCards(t, g, me.ID, milled)
	if !g.Battlefield.Contains(milled) || !b25Tapped(t, g, milled) {
		t.Error("the milled land did not return tapped on the restored table")
	}
}

// Exhume (the walk Breach the Multiverse shares in shape): the first
// player's pick is answered, the table is restored with the second still
// open, and both creatures arrive together after the second answer. The
// question and the cards chosen so far ride the second prompt as data.
func TestSweep00ExhumeWalkIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b38Graveyard(g, me.ID, "My Bear", "Creature — Bear", "{1}{G}", 2, 2)
	theirs := b38Graveyard(g, opp.ID, "Their Bear", "Creature — Bear", "{1}{G}", 2, 2)
	castCatalogSpell(t, g, "Exhume", "Sorcery", b38ExhumeOracle, nil)
	passPriorityAroundTable(t, g)

	// Restored with the first prompt open.
	sweep00PickKey(t, g, me.ID, "choose-cards/each-player-reanimates-one")
	g = restoreRoundTrip(t, g, true)
	me, opp = g.Seats[0], g.Seats[1]
	first := sweep00PickKey(t, g, me.ID, "choose-cards/each-player-reanimates-one")
	if err := g.ResolveChooseCards(first.ID, me.ID, []uuid.UUID{mine}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if g.Battlefield.Contains(mine) {
		t.Fatal("nothing enters until every player has chosen")
	}

	// Restored again with the second open and the first player's pick carried.
	second := sweep00PickKey(t, g, opp.ID, "choose-cards/each-player-reanimates-one")
	if second.Reason != first.Reason {
		t.Errorf("the second question %q is not the first's %q", second.Reason, first.Reason)
	}
	g = restoreRoundTrip(t, g, true)
	opp = g.Seats[1]
	second = sweep00PickKey(t, g, opp.ID, "choose-cards/each-player-reanimates-one")
	if err := g.ResolveChooseCards(second.ID, opp.ID, []uuid.UUID{theirs}); err != nil {
		t.Fatalf("ResolveChooseCards on the restored table: %v", err)
	}
	if !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(theirs) {
		t.Error("both chosen creatures arrive together on the restored table")
	}
}

// Abdel Adrian: the any-number pick carries the asking item as data, so the
// restored pick exiles the picks until he leaves and makes a Soldier for each.
func TestSweep00AbdelAdrianPickIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := pushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)
	castCatalogSpell(t, g, "Abdel Adrian, Gorion's Ward", "Legendary Creature — Human Warrior", mcAbdelOracle, nil)
	passPriorityAroundTable(t, g)
	sweep00PickKey(t, g, me.ID, "choose-cards/abdel-adrian-exile")

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[g.Turn.ActiveSeat]
	c := sweep00PickKey(t, g, me.ID, "choose-cards/abdel-adrian-exile")
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{a, b}); err != nil {
		t.Fatalf("ResolveChooseCards on the restored table: %v", err)
	}
	if !g.Exile.Contains(a) || !g.Exile.Contains(b) {
		t.Fatal("the picks were not exiled on the restored table")
	}
	if n := countBattlefieldNamed(g, me.ID, "Soldier"); n != 2 {
		t.Errorf("Soldiers = %d on the restored table, want 2", n)
	}
}

// Caparocti Sunborn: "exactly two" is a registered set rule, so the
// restored prompt still refuses one permanent, and tapping two discovers 3.
func TestSweep00CaparoctiTapTwoIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	capa := pushCatalogPermanent(g, me.ID, "Caparocti Sunborn", "Legendary Creature — Human Soldier", caparoctiOracle, false)
	a := b16Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b16Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	hit := discoverLibrary(me, 3)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(capa, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	sweep00PickKey(t, g, me.ID, "choose-cards/caparocti-sunborn-tap-two")

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := sweep00PickKey(t, g, me.ID, "choose-cards/caparocti-sunborn-tap-two")
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{a}); err == nil {
		t.Error("the restored prompt accepted one permanent: the set rule did not survive")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{a, b}); err != nil {
		t.Fatalf("ResolveChooseCards on the restored table: %v", err)
	}
	if !b16Tapped(t, g, a) || !b16Tapped(t, g, b) {
		t.Error("the two creatures were not tapped on the restored table")
	}
	expectDiscovered(t, g, me.ID, hit)
}

// Ardenn, Intrepid Archaeologist: the host the Equipment attaches to is the
// carry and its kind the word.
func TestSweep00ArdennAttachPickIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := sqActive(g)
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Ardenn, Intrepid Archaeologist", "Legendary Creature — Kor Scout", sqArdennOracle, false)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	e1 := dcSeed(g, me.ID, "Sword One", "Artifact — Equipment", "", 0, 0)
	advanceTo(t, g, game.StepBeginCombat)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	sweep00PickKey(t, g, me.ID, "choose-cards/ardenn-intrepid-archaeologist-attach")

	g = restoreRoundTrip(t, g, true)
	me, _ = sqActive(g)
	c := sweep00PickKey(t, g, me.ID, "choose-cards/ardenn-intrepid-archaeologist-attach")
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{e1}); err != nil {
		t.Fatalf("ResolveChooseCards on the restored table: %v", err)
	}
	passPriorityAroundTable(t, g)
	if card := sqCard(t, g, e1); card.AttachedTo.ID != bear {
		t.Error("the Equipment was not attached to the host on the restored table")
	}
}

// Apocalypse Demon: the upkeep choice is a keyed option pick, and the
// creature to sacrifice is the engine's own sacrifice prompt, plain data.
func TestSweep00ApocalypseDemonChoiceIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	demon := pushCatalogPermanent(g, me.ID, "Apocalypse Demon", "Creature — Demon", apocalypseDemonOracle, false)
	fodder := pushVanillaCreature(g, me.ID, "Fodder", 1, 1)
	sweep00Graveyard(g, me.ID, 3) // its size is the Demon's power and toughness
	g.WithWriteLock(func() {
		item := &game.StackItem{Kind: game.StackItemTriggered, Controller: me.ID, Owner: me.ID, SourceCardID: demon}
		if err := apocalypseDemonUpkeep(g, item); err != nil {
			t.Fatal(err)
		}
	})
	if c := latestOptionPickFor(g, me.ID); c == nil || c.PickThen != "option-pick/apocalypse-demon-sacrifice-or-tap" {
		t.Fatalf("the upkeep choice is %+v, want the keyed option pick", c)
	}

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	answerOptionPick(t, g, me.ID, 1)
	// The sacrifice is asked next, as plain data too.
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	var sac *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceSacrifice && c.Chooser == me.ID {
			sac = c
		}
	}
	if sac == nil {
		t.Fatalf("no sacrifice prompt after choosing to sacrifice: %+v", g.PendingChoices)
	}
	if hasID(sac.SacrificeOptions, demon) || !hasID(sac.SacrificeOptions, fodder) {
		t.Errorf("sacrifice options %v: want the other creature, never the Demon", sac.SacrificeOptions)
	}
}

// Apocalypse Demon, the other branch: tapping needs only the controller and
// the Demon, both on the restored prompt.
func TestSweep00ApocalypseDemonTapBranchOnARestoredTable(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	demon := pushCatalogPermanent(g, me.ID, "Apocalypse Demon", "Creature — Demon", apocalypseDemonOracle, false)
	pushVanillaCreature(g, me.ID, "Fodder", 1, 1)
	sweep00Graveyard(g, me.ID, 3) // its size is the Demon's power and toughness
	g.WithWriteLock(func() {
		item := &game.StackItem{Kind: game.StackItemTriggered, Controller: me.ID, Owner: me.ID, SourceCardID: demon}
		if err := apocalypseDemonUpkeep(g, item); err != nil {
			t.Fatal(err)
		}
	})
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	answerOptionPick(t, g, me.ID, 0)
	if !b16Tapped(t, g, demon) {
		t.Error("the Demon did not tap on the restored table")
	}
}

// Breeches, the Blastmaker: "sacrifice an artifact" is a keyed permanents
// pick carrying the spell, so the table is a restore point while it is open,
// and the coin flip follows on the restored table.
func TestSweep00BreechesSacrificePickIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	g.SetRNGKeyForTest(breechesRNGKey())
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanentWithID(g, breechesSourceID, me.ID, "Breeches, the Blastmaker",
		"Legendary Creature — Goblin Pirate", breechesBlastmakerOracle, false)
	relic := pushCatalogPermanent(g, me.ID, "Worn Relic", "Artifact", "", false)
	castCatalogSpell(t, g, "Opening Gambit", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	castSpellWithCost(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, "{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	if latestChooseCardsFor(g, me.ID) == nil {
		t.Fatalf("no sacrifice prompt: %+v", g.PendingChoices)
	}

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	sac := latestChooseCardsFor(g, me.ID)
	if sac == nil || !hasID(sac.ChooseCards, relic) {
		t.Fatalf("the restored table's sacrifice prompt is %+v, want the relic on offer", sac)
	}
	if err := g.ResolveChooseCards(sac.ID, me.ID, []uuid.UUID{relic}); err != nil {
		t.Fatalf("ResolveChooseCards on the restored table: %v", err)
	}
	if g.Battlefield.Contains(relic) {
		t.Fatal("the artifact was not sacrificed on the restored table")
	}
	if latestChoiceOfKind(g, game.PendingChoiceCoinCall) == nil {
		t.Errorf("the sacrifice is not followed by the coin flip on the restored table: %+v", g.PendingChoices)
	}
}
