package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// modes_test.go — S20 sub-PR 4: modal cards resolve the chosen
// option only, and the option's target clause gates the cast.

const (
	rakdosCharmOracle    = "5e62b51d-faec-4aa0-9504-cf2c282d08ea"
	izzetCharmOracle     = "a07698f6-5ad5-49a3-9da2-f82d407f5cd7"
	austereCommandOracle = "09cc8709-fe10-472a-b05c-e89f3523018d"
)

// castModal seeds a modal card into the active seat's hand and casts
// it with the chosen modes + targets.
func castModal(t *testing.T, g *game.Game, name, typeLine, oracle string, modes []int, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{Modes: modes, Targets: targets}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

func pushCostedPermanentForTest(g *game.Game, owner uuid.UUID, name, typeLine, manaCost string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: manaCost,
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	})
	return id
}

func TestModalCatalogSpecsAreWired(t *testing.T) {
	for _, oracle := range []string{rakdosCharmOracle, izzetCharmOracle, austereCommandOracle} {
		if game.ModeSpecFor(oracle) == nil {
			t.Errorf("%s: no ModeSpec through the catalog hook", oracle)
		}
		if game.TargetSpecFor(oracle) != nil {
			t.Errorf("%s: modal cards carry targets on their modes, not the card", oracle)
		}
	}
	if ms := game.ModeSpecFor(austereCommandOracle); ms.Min != 2 || ms.Max != 2 || len(ms.Options) != 4 {
		t.Errorf("Austere Command spec = %+v, want choose two of four", ms)
	}
}

// --- Rakdos Charm ------------------------------------------------

func TestRakdosCharmExilesTargetGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := pushGraveyardCardForTest(opp, "Dead A")
	b := pushGraveyardCardForTest(opp, "Dead B")
	mine := pushGraveyardCardForTest(g.Seats[0], "Mine")

	castModal(t, g, "Rakdos Charm", "Instant", rakdosCharmOracle, []int{0},
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)

	if opp.Graveyard.Size() != 0 {
		t.Errorf("opponent graveyard should be empty, has %d", opp.Graveyard.Size())
	}
	if !g.Exile.Contains(a) || !g.Exile.Contains(b) {
		t.Errorf("graveyard cards should be in exile")
	}
	if !g.Seats[0].Graveyard.Contains(mine) {
		t.Errorf("caster's own graveyard must be untouched")
	}
}

func TestRakdosCharmDestroysTargetArtifact(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	rock := pushCostedPermanentForTest(g, opp.ID, "Sol Ring", "Artifact", "{1}")
	bear := pushCostedPermanentForTest(g, opp.ID, "Bear", "Creature — Bear", "{1}{G}")

	castModal(t, g, "Rakdos Charm", "Instant", rakdosCharmOracle, []int{1},
		[]game.TargetRef{{Kind: game.TargetCard, ID: rock}})
	passPriorityAroundTable(t, g)

	if !opp.Graveyard.Contains(rock) {
		t.Errorf("artifact should be destroyed")
	}
	if !g.Battlefield.Contains(bear) {
		t.Errorf("creature must survive the artifact mode")
	}
}

func TestRakdosCharmModeTargetIsGated(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushCostedPermanentForTest(g, opp.ID, "Bear", "Creature — Bear", "{1}{G}")
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Rakdos Charm", TypeLine: "Instant",
		OracleID: rakdosCharmOracle, Owner: me.ID, Controller: me.ID})
	err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{1},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	if err != game.ErrIllegalTarget {
		t.Fatalf("artifact mode aimed at a creature: %v, want ErrIllegalTarget", err)
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{0, 2},
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != game.ErrInvalidParam {
		t.Fatalf("choose one with two modes: %v, want ErrInvalidParam", err)
	}
}

func TestRakdosCharmEachCreatureDamagesController(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCostedPermanentForTest(g, opp.ID, "A", "Creature — Bear", "{1}{G}")
	pushCostedPermanentForTest(g, opp.ID, "B", "Creature — Bear", "{1}{G}")
	pushCostedPermanentForTest(g, me.ID, "C", "Creature — Bear", "{1}{G}")
	pushCostedPermanentForTest(g, opp.ID, "Rock", "Artifact", "{1}")
	meBefore, oppBefore := me.Life, opp.Life

	castModal(t, g, "Rakdos Charm", "Instant", rakdosCharmOracle, []int{2}, nil)
	passPriorityAroundTable(t, g)

	if opp.Life != oppBefore-2 {
		t.Errorf("opponent with two creatures: %d -> %d, want -2", oppBefore, opp.Life)
	}
	if me.Life != meBefore-1 {
		t.Errorf("caster with one creature: %d -> %d, want -1", meBefore, me.Life)
	}
}

// --- Izzet Charm -------------------------------------------------

func TestIzzetCharmCounterUnlessPays(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	// Opponent's Lightning Bolt at me.
	bolt := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant",
		OracleID: lightningBoltOracle, Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, bolt, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err != nil {
		t.Fatal(err)
	}
	before := me.Life
	castModal(t, g, "Izzet Charm", "Instant", izzetCharmOracle, []int{0},
		[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}})

	// Resolve the Charm: the Bolt's controller gets the pay prompt.
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatalf("no pay_unless prompt for the Bolt's controller")
	}
	answerPayUnless(t, g, opp.ID, false)
	if g.StackMeta[bolt] != nil {
		t.Errorf("declined: Bolt should be countered")
	}
	if !opp.Graveyard.Contains(bolt) {
		t.Errorf("countered Bolt should be in its owner's graveyard")
	}
	passPriorityAroundTable(t, g)
	if me.Life != before {
		t.Errorf("countered Bolt dealt damage: %d -> %d", before, me.Life)
	}
}

func TestIzzetCharmCannotTargetCreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	// A creature spell of mine on the stack (sorcery-speed, empty stack).
	bear := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear",
		Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	charm := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: charm, Name: "Izzet Charm", TypeLine: "Instant",
		OracleID: izzetCharmOracle, Owner: opp.ID, Controller: opp.ID})
	err := g.CastSpell(opp.ID, charm, game.CastSpellParams{Modes: []int{0},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	if err != game.ErrIllegalTarget {
		t.Fatalf("noncreature-spell mode at a creature spell: %v, want ErrIllegalTarget", err)
	}
}

func TestIzzetCharmDealsTwoToCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := pushCostedPermanentForTest(g, opp.ID, "Bear", "Creature — Bear", "{1}{G}")
	castModal(t, g, "Izzet Charm", "Instant", izzetCharmOracle, []int{1},
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bear) {
		t.Errorf("2/2 hit for 2 should die to the SBA")
	}
}

// TestIzzetCharmDrawTwoDiscardTwo — #338 stale-simplification fix.
// Mode 2 used to discard at RANDOM, behind a note saying the "you
// choose" picker was deferred; the picker (DiscardChoiceForEffect,
// via lootOne) had shipped and Faithless Looting was already using
// it. The controller now picks, so the two discards are PENDING
// after resolution rather than already in the graveyard.
func TestIzzetCharmDrawTwoDiscardTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	handBefore, libBefore, gyBefore := me.Hand.Size(), me.Library.Size(), me.Graveyard.Size()
	castModal(t, g, "Izzet Charm", "Instant", izzetCharmOracle, []int{2}, nil)
	passPriorityAroundTable(t, g)

	// Both cards drawn; nothing discarded yet — the controller owes
	// the choice. castModal seeds the Charm then casts it (net 0 on
	// hand), so hand is +2 and the graveyard holds only the Charm.
	if me.Hand.Size() != handBefore+2 {
		t.Errorf("hand %d -> %d, want +2 (drawn, not yet discarded)",
			handBefore, me.Hand.Size())
	}
	if me.Library.Size() != libBefore-2 {
		t.Errorf("library %d -> %d, want -2", libBefore, me.Library.Size())
	}
	if me.Graveyard.Size() != gyBefore+1 {
		t.Errorf("graveyard %d -> %d, want +1 (the Charm only)",
			gyBefore, me.Graveyard.Size())
	}
	if discardOwed(g, me.ID) != 2 {
		t.Fatalf("the mode owes a 2-card discard prompt: got %d — the "+
			"controller should be choosing, not discarding at random",
			discardOwed(g, me.ID))
	}

	// The controller names their two discards.
	answerDiscard(t, g, me.ID,
		me.Hand.Cards[0].InstanceID,
		me.Hand.Cards[1].InstanceID,
	)
	if me.Hand.Size() != handBefore {
		t.Errorf("post-selection hand %d, want %d", me.Hand.Size(), handBefore)
	}
	if me.Graveyard.Size() != gyBefore+3 {
		t.Errorf("post-selection graveyard %d -> %d, want +3 (charm + two discards)",
			gyBefore, me.Graveyard.Size())
	}
}

// --- Austere Command ---------------------------------------------

func TestAustereCommandChooseTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := pushCostedPermanentForTest(g, opp.ID, "Sol Ring", "Artifact", "{1}")
	aura := pushCostedPermanentForTest(g, opp.ID, "Rancor", "Enchantment — Aura", "{G}")
	small := pushCostedPermanentForTest(g, opp.ID, "Bear", "Creature — Bear", "{1}{G}")
	big := pushCostedPermanentForTest(g, me.ID, "Dreadmaw", "Creature — Dinosaur", "{4}{G}{G}")

	// Artifacts + big creatures.
	castModal(t, g, "Austere Command", "Sorcery", austereCommandOracle, []int{0, 3}, nil)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(rock) {
		t.Errorf("artifact should be destroyed")
	}
	if g.Battlefield.Contains(big) {
		t.Errorf("MV 6 creature should be destroyed")
	}
	if !g.Battlefield.Contains(aura) || !g.Battlefield.Contains(small) {
		t.Errorf("unchosen modes must not fire: aura on bf %v, small on bf %v",
			g.Battlefield.Contains(aura), g.Battlefield.Contains(small))
	}
}

func TestAustereCommandRejectsOneMode(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Austere Command", TypeLine: "Sorcery",
		OracleID: austereCommandOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{2}}); err != game.ErrInvalidParam {
		t.Fatalf("choose two with one mode: %v, want ErrInvalidParam", err)
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{2, 2}}); err != game.ErrInvalidParam {
		t.Fatalf("duplicate mode: %v, want ErrInvalidParam", err)
	}
}
