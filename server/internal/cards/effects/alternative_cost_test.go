package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// alternative_cost_test.go — S22, the card side. The engine-side
// mechanism is pinned in server/internal/game/alternative_cost_test.go;
// these tests are about the three cards that shipped without their
// headline mode, plus the one new card the mechanism unlocked.

const (
	washAwayOracle    = "a4630da0-fe9b-4ead-9621-eac4b7825c35"
	mulldrifterOracle = "24d0f5e7-0d9e-4b76-900e-a7274e80312d"
)

// castWithAltCost is castCatalogSpell with an alternative cost
// claimed. Seeds the card in the active seat's hand, walks to a main
// phase, and announces.
func castWithAltCost(
	t *testing.T,
	g *game.Game,
	name, typeLine, oracleID, altCost string,
) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		OracleID:   oracleID,
		Owner:      active.ID,
		Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{
		AlternativeCost: altCost,
	}); err != nil {
		t.Fatalf("CastSpell %s (%s): %v", name, altCost, err)
	}
	return id
}

// Every alternative cost declared in the catalog has to be reachable
// by the key the client will send. A typo here is a cast that gets
// rejected in a real game and nowhere else.
func TestAlternativeCostsAreWired(t *testing.T) {
	for _, tc := range []struct{ name, oracle, key, cost string }{
		{"Vandalblast", vandalblastOracle, "overload", "{4}{R}"},
		{"Cyclonic Rift", cyclonicRiftOracle, "overload", "{6}{U}"},
		{"Slithermuse", slithermuseOracle, "evoke", "{3}{U}"},
		{"Mulldrifter", mulldrifterOracle, "evoke", "{2}{U}"},
		{"Wash Away", washAwayOracle, "cleave", "{1}{U}{U}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ac := game.AlternativeCostByKey(tc.oracle, tc.key)
			if ac == nil {
				t.Fatalf("%s offers no %q cost", tc.name, tc.key)
			}
			if ac.ManaCost != tc.cost {
				t.Errorf("%s %s cost = %q, want %q", tc.name, tc.key, ac.ManaCost, tc.cost)
			}
			if ac.Label == "" {
				t.Errorf("%s %s has no label for the picker", tc.name, tc.key)
			}
		})
	}
}

// --- Overload ----------------------------------------------------

// The headline mode: an overloaded Rift bounces every nonland
// permanent its caster doesn't control, and nothing they do.
func TestCyclonicRiftOverloadBouncesTheirBoardOnly(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[g.Turn.ActiveSeat]
	victim := g.Seats[1]
	if caster.ID == victim.ID {
		victim = g.Seats[2]
	}
	theirs := seedPermanentFor(g, victim.ID, "Grizzly Bears", "Creature — Bear")
	theirRock := seedPermanentFor(g, victim.ID, "Sol Ring", "Artifact")
	theirLand := seedPermanentFor(g, victim.ID, "Island", "Basic Land — Island")
	mine := seedPermanentFor(g, caster.ID, "My Bear", "Creature — Bear")

	castWithAltCost(t, g, "Cyclonic Rift", "Instant", cyclonicRiftOracle, "overload")
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(theirs) || g.Battlefield.Contains(theirRock) {
		t.Error("overloaded Rift left an opponent's nonland permanent on the battlefield")
	}
	if !g.Battlefield.Contains(theirLand) {
		t.Error("overloaded Rift bounced a land — the clause is nonland permanents")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("overloaded Rift is one-sided; it must not touch its caster's board")
	}
}

// The printed mode still works, and still bounces exactly one thing.
func TestCyclonicRiftHardCastStillBouncesOne(t *testing.T) {
	g := newCatalogGame(t)
	victim := g.Seats[1]
	one := seedPermanentFor(g, victim.ID, "Grizzly Bears", "Creature — Bear")
	two := seedPermanentFor(g, victim.ID, "Sol Ring", "Artifact")

	castCatalogSpell(t, g, "Cyclonic Rift", "Instant", cyclonicRiftOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: one}})
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(one) {
		t.Error("targeted permanent was not bounced")
	}
	if !g.Battlefield.Contains(two) {
		t.Error("a hard-cast Rift swept the board — the overload branch leaked")
	}
}

func TestVandalblastOverloadDestroysTheirArtifactsOnly(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[g.Turn.ActiveSeat]
	victim := g.Seats[1]
	if caster.ID == victim.ID {
		victim = g.Seats[2]
	}
	theirs := seedPermanentFor(g, victim.ID, "Sol Ring", "Artifact")
	theirBear := seedPermanentFor(g, victim.ID, "Grizzly Bears", "Creature — Bear")
	mine := seedPermanentFor(g, caster.ID, "My Signet", "Artifact")

	castWithAltCost(t, g, "Vandalblast", "Sorcery", vandalblastOracle, "overload")
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(theirs) {
		t.Error("overloaded Vandalblast spared an opponent's artifact")
	}
	if !g.Battlefield.Contains(theirBear) {
		t.Error("overloaded Vandalblast destroyed a non-artifact")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("overloaded Vandalblast is one-sided; your own artifacts survive")
	}
}

// An overloaded spell has no target clause left, so sending one is a
// rejection — the client would be paying a cost it doesn't
// understand.
func TestOverloadRejectsTargets(t *testing.T) {
	g := newCatalogGame(t)
	victim := g.Seats[1]
	id := seedPermanentFor(g, victim.ID, "Sol Ring", "Artifact")
	active := g.Seats[g.Turn.ActiveSeat]
	spellID := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: spellID, Name: "Vandalblast", TypeLine: "Sorcery",
		OracleID: vandalblastOracle, Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	err := g.CastSpell(active.ID, spellID, game.CastSpellParams{
		AlternativeCost: "overload",
		Targets:         []game.TargetRef{{Kind: game.TargetCard, ID: id}},
	})
	if err == nil {
		t.Fatal("overload with a target should be rejected")
	}
	if !active.Hand.Contains(spellID) {
		t.Error("rejected cast left the spell out of hand")
	}
}

// --- Evoke -------------------------------------------------------

// Evoked Slithermuse enters, is sacrificed by the evoke trigger, and
// the leaves-the-battlefield trigger then draws the hand-size
// difference. That chain is the entire card.
func TestSlithermuseEvokeEntersAndDiesAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[g.Turn.ActiveSeat]
	var opp *game.Player
	for _, p := range g.Seats {
		if p.ID != caster.ID {
			opp = p
			break
		}
	}
	if opp == nil {
		t.Fatal("no opponent seated")
	}
	// Put the opponent three cards up on the caster.
	for i := 0; i < 3; i++ {
		opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "filler",
			Owner: opp.ID, Controller: opp.ID})
	}

	id := castWithAltCost(t, g, "Slithermuse", "Creature — Elemental",
		slithermuseOracle, "evoke")
	// Measured with the spell already on the stack, so the walk to a
	// main phase (and any draw for turn it passed through) is behind
	// us. Drawing the difference lands the caster's hand exactly level
	// with the opponent's.
	diff := len(opp.Hand.Cards) - len(caster.Hand.Cards)
	if diff <= 0 {
		t.Fatalf("test setup: opponent must be ahead, diff = %d", diff)
	}
	want := len(caster.Hand.Cards) + diff

	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(id) {
		t.Error("evoked creature is still on the battlefield")
	}
	if !caster.Graveyard.Contains(id) {
		t.Error("evoked creature did not reach its owner's graveyard")
	}
	if len(caster.Hand.Cards) != want {
		t.Errorf("hand %d, want %d (the leave-trigger drew the difference)",
			len(caster.Hand.Cards), want)
	}
}

// Hard-cast, nothing sacrifices it: the creature stays, and the deck
// keeps the option of blinking it later.
func TestSlithermuseHardCastStaysOnTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	id := castCatalogSpell(t, g, "Slithermuse", "Creature — Elemental", slithermuseOracle, nil)
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(id) {
		t.Error("a hard-cast Slithermuse sacrificed itself — the evoke trigger leaked")
	}
}

// answerAnyTriggerOrderPrompt resolves an outstanding CR 603.3b
// ordering prompt for `chooser`, keeping the queue's current order.
// Returns false when there is no prompt waiting. Bounded priority
// passes first, because the prompt only appears once the drain runs.
func answerAnyTriggerOrderPrompt(t *testing.T, g *game.Game, chooser uuid.UUID) bool {
	t.Helper()
	var prompt *game.PendingChoice
	for i := 0; i < 8 && prompt == nil; i++ {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceTriggerOrder && c.Chooser == chooser {
				prompt = c
			}
		}
		if prompt != nil {
			break
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if prompt == nil {
		return false
	}
	// Answer with exactly the set the prompt covers, in the order it
	// offered — ResolveTriggerOrder rejects anything else.
	order := append([]uuid.UUID(nil), prompt.TriggerOrderIDs...)
	if err := g.ResolveTriggerOrder(prompt.ID, chooser, order); err != nil {
		t.Fatalf("ResolveTriggerOrder: %v", err)
	}
	return true
}

// Mulldrifter is the canonical evoke card, and the one where the
// ordering claim is checkable: the creature really ENTERS, so its ETB
// draw trigger fires alongside the evoke sacrifice. An implementation
// that sacrificed inside resolution would draw nothing at all.
//
// Two triggers, one controller, different abilities — so the engine
// asks for an order (CR 603.3b) before either reaches the stack. That
// prompt is correct behaviour, not an obstacle: the draw happens
// whichever order they are stacked in, which is the point.
func TestMulldrifterEvokeStillDrawsTwo(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[g.Turn.ActiveSeat]

	id := castWithAltCost(t, g, "Mulldrifter", "Creature — Elemental",
		mulldrifterOracle, "evoke")
	// Measured with the spell already on the stack, so the walk to a
	// main phase — and any draw for turn along the way — is behind us
	// and the only cards left to arrive are the trigger's two.
	before := len(caster.Hand.Cards)
	if !answerAnyTriggerOrderPrompt(t, g, caster.ID) {
		t.Fatal("evoked Mulldrifter should queue two triggers and ask for an order")
	}
	passPriorityAroundTable(t, g)

	if got := len(caster.Hand.Cards); got != before+2 {
		t.Errorf("hand %d, want %d (evoked Mulldrifter still draws two)", got, before+2)
	}
	if g.Battlefield.Contains(id) {
		t.Error("evoked Mulldrifter is still on the battlefield")
	}
	if !caster.Graveyard.Contains(id) {
		t.Error("evoked Mulldrifter did not reach its owner's graveyard")
	}
}

// --- Cleave ------------------------------------------------------

// Hard-cast Wash Away can only counter a spell that wasn't cast from
// its owner's hand; cleaved it counters anything.
func TestWashAwayClauseWidensUnderCleave(t *testing.T) {
	g := newCatalogGame(t)
	// Seat 0 is active; seat 1 casts nothing, so put the victim
	// spell on the stack from the active seat's own hand and let a
	// later seat answer it. Simplest arrangement that produces a
	// hand-cast spell on the stack.
	active := g.Seats[g.Turn.ActiveSeat]
	victim := castCatalogSpell(t, g, "Bear Spell", "Creature — Bear", "", nil)

	washID := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: washID, Name: "Wash Away", TypeLine: "Instant",
		OracleID: washAwayOracle, Owner: active.ID, Controller: active.ID,
	})

	// The printed clause excludes it — the bear was cast from hand.
	if err := g.CastSpell(active.ID, washID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err == nil {
		t.Fatal("hard-cast Wash Away must not counter a spell cast from hand")
	}
	// Cleaved, it can.
	if err := g.CastSpell(active.ID, washID, game.CastSpellParams{
		AlternativeCost: "cleave",
		Targets:         []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("cleaved Wash Away: %v", err)
	}
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(victim) {
		t.Error("countered spell resolved onto the battlefield anyway")
	}
}

// A commander cast is exactly what the printed clause is aimed at:
// it wasn't cast from its owner's hand, so {U} answers it.
func TestWashAwayHardCastAnswersACommanderCast(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	cmdID := uuid.New()
	active.Command.PushTop(game.Card{
		InstanceID: cmdID, Name: "Some Commander", TypeLine: "Legendary Creature — Human",
		Owner: active.ID, Controller: active.ID, IsCommander: true,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, cmdID, game.CastSpellParams{FromZone: "command"}); err != nil {
		t.Fatalf("commander cast: %v", err)
	}

	washID := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: washID, Name: "Wash Away", TypeLine: "Instant",
		OracleID: washAwayOracle, Owner: active.ID, Controller: active.ID,
	})
	if err := g.CastSpell(active.ID, washID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: cmdID}},
	}); err != nil {
		t.Fatalf("hard-cast Wash Away on a commander cast: %v", err)
	}
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(cmdID) {
		t.Error("the commander resolved despite being countered")
	}
}
