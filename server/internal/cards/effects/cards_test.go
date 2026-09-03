package effects

import (
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cards_test.go exercises each registered catalog card end-to-end:
// seed a game, cast the card targeting an appropriate victim, pass
// priority around, assert the resulting state + graveyard. Grouped
// by family to keep the per-card helpers small.

// newCatalogGame returns a fully-started 4-player game (four seats
// so target-rotation tests have room). 20-card libraries give room
// for mill / discard / draw primitives without bottoming out.
func newCatalogGame(t *testing.T) *game.Game {
	t.Helper()
	g := game.NewGame()
	for i := 0; i < 4; i++ {
		deck := make([]game.Card, 20)
		for j := range deck {
			deck[j] = game.NewCard("basic-filler", uuid.Nil)
		}
		if _, err := g.AddPlayer(string(rune('A'+i)), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(11, 22))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	return g
}

// castCatalogSpell seeds a card with the given name, type line, and
// oracle ID into the active seat's hand, advances to the precombat
// main phase, then calls CastSpell with the supplied targets. Pass
// priority around afterwards to resolve.
func castCatalogSpell(t *testing.T, g *game.Game, name, typeLine, oracleID string, targets []game.TargetRef) uuid.UUID {
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
	// Advance to main phase if needed.
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{Targets: targets}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// passPriorityAroundTable passes priority until the stack is empty
// — no spell cards on Game.Stack, no ability items in StackMeta,
// and no triggers waiting in PendingTriggers to be drained onto it
// (bounded at 32 to catch runaway loops). Triggered abilities
// harvested while a spell resolves land on the stack at the next
// priority boundary and need their own trip around the table, so
// a single call settles "cast creature → ETB trigger → effect".
//
// Stops short of a pending trigger *prompt*: an optional trigger
// waits for its yes/no before anything reaches the stack. Answer
// it (answerLatestTriggerPrompt) then call this again.
func passPriorityAroundTable(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 32; i++ {
		if stackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority iter %d: %v", i, err)
		}
	}
	t.Fatalf("stack did not empty after 32 priority passes")
}

// stackFullyEmpty reports whether nothing is on or headed for the
// stack: Game.Stack, StackMeta, and PendingTriggers are all empty.
func stackFullyEmpty(g *game.Game) bool {
	return g.Stack.Size() == 0 && len(g.StackMeta) == 0 && len(g.PendingTriggers) == 0
}

// triggerOnStack returns the triggered-ability item in StackMeta
// sourced from cardID, or nil. Tests use it to assert a trigger is
// actually waiting on the stack (with its label and targets)
// before priority passes resolve it.
func triggerOnStack(g *game.Game, cardID uuid.UUID) *game.StackItem {
	for _, item := range g.StackMeta {
		if item != nil && item.Kind == game.StackItemTriggered && item.SourceCardID == cardID {
			return item
		}
	}
	return nil
}

func pushCreatureToBattlefieldForTest(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   "Creature — Test",
		Power:      2,
		Toughness:  2,
		Owner:      owner,
		Controller: owner,
	})
	return id
}

// --- Direct damage ---------------------------------------------

func TestLightningBoltDamagesPlayer(t *testing.T) {
	g := newCatalogGame(t)
	target := g.Seats[1]
	beforeLife := target.Life

	castCatalogSpell(t, g, "Lightning Bolt", "Instant",
		"4457ed35-7c10-48c8-9776-456485fdf070",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: target.ID}},
	)
	passPriorityAroundTable(t, g)

	if got := target.Life; got != beforeLife-3 {
		t.Errorf("target life: got %d, want %d", got, beforeLife-3)
	}
}

func TestShockDamagesPlayer(t *testing.T) {
	g := newCatalogGame(t)
	target := g.Seats[1]
	beforeLife := target.Life

	castCatalogSpell(t, g, "Shock", "Instant",
		"a9d288b8-cdc1-4e55-a0c9-d6edfc95e65d",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: target.ID}},
	)
	passPriorityAroundTable(t, g)

	if got := target.Life; got != beforeLife-2 {
		t.Errorf("target life: got %d, want %d", got, beforeLife-2)
	}
}

func TestLightningHelixDamagesAndGainsLife(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	target := g.Seats[1]
	beforeTargetLife := target.Life
	beforeCasterLife := caster.Life

	castCatalogSpell(t, g, "Lightning Helix", "Instant",
		"800c258a-cfc4-4a54-a667-065ea8dea69e",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: target.ID}},
	)
	passPriorityAroundTable(t, g)

	if got := target.Life; got != beforeTargetLife-3 {
		t.Errorf("target life: got %d, want %d", got, beforeTargetLife-3)
	}
	if got := caster.Life; got != beforeCasterLife+3 {
		t.Errorf("caster life: got %d, want %d", got, beforeCasterLife+3)
	}
}

func TestPyroclasmDamagesEachCreature(t *testing.T) {
	g := newCatalogGame(t)
	c1 := pushCreatureToBattlefieldForTest(g, g.Seats[0].ID, "Creature A")
	c2 := pushCreatureToBattlefieldForTest(g, g.Seats[1].ID, "Creature B")

	castCatalogSpell(t, g, "Pyroclasm", "Sorcery",
		"e4bcd4ea-e7cd-4471-8f3b-18bb51d3d70c",
		nil,
	)
	passPriorityAroundTable(t, g)

	// Each 2/2 took 2 damage → lethal → SBA routes to graveyard.
	if g.Battlefield.Contains(c1) {
		t.Errorf("creature A should have been destroyed")
	}
	if g.Battlefield.Contains(c2) {
		t.Errorf("creature B should have been destroyed")
	}
}

// --- Mass removal ----------------------------------------------

func TestWrathOfGodDestroysAllCreatures(t *testing.T) {
	g := newCatalogGame(t)
	c1 := pushCreatureToBattlefieldForTest(g, g.Seats[0].ID, "Mine")
	c2 := pushCreatureToBattlefieldForTest(g, g.Seats[1].ID, "Theirs")

	castCatalogSpell(t, g, "Wrath of God", "Sorcery",
		"34515b16-c9a4-4f98-8c77-416a7a523407",
		nil,
	)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(c1) || g.Battlefield.Contains(c2) {
		t.Errorf("Wrath did not clear the battlefield")
	}
}

func TestDamnationDestroysAllCreatures(t *testing.T) {
	g := newCatalogGame(t)
	c1 := pushCreatureToBattlefieldForTest(g, g.Seats[0].ID, "One")

	castCatalogSpell(t, g, "Damnation", "Sorcery",
		"d57a8f0b-7989-4db5-8756-6f2690097252",
		nil,
	)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(c1) {
		t.Errorf("Damnation did not destroy creature")
	}
}

func TestDayOfJudgmentDestroysAllCreatures(t *testing.T) {
	g := newCatalogGame(t)
	c1 := pushCreatureToBattlefieldForTest(g, g.Seats[0].ID, "One")

	castCatalogSpell(t, g, "Day of Judgment", "Sorcery",
		"d057289d-5e28-43d5-8ff3-4a1bc723477d",
		nil,
	)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(c1) {
		t.Errorf("Day of Judgment did not destroy creature")
	}
}

// --- Counter magic ---------------------------------------------

// TestCounterspellCountersLightningBolt is the multi-spell scenario
// that proves the counter chain works: opponent casts Bolt; caster
// casts Counterspell targeting it; priority wraps and resolves
// Counterspell first (top of stack), which pulls Bolt off the stack
// into opponent's graveyard without damage.
func TestCounterspellCountersLightningBolt(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	opponent := g.Seats[1]
	beforeLife := caster.Life

	// Advance to main phase so the active seat can cast a sorcery /
	// non-instant. Lightning Bolt is an instant, so opponent can
	// cast it at instant speed during the active player's priority
	// — but castCatalogSpell seeds into the active seat's hand. For
	// this test we'll skip that helper and place the Bolt manually
	// in opponent's hand + cast it.

	// Put us in main phase.
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}

	// Seed + cast opponent's Lightning Bolt. Since opponent isn't
	// the active seat, they'd normally need priority — which they
	// can hold as an instant. But CastSpell applies the sorcery-
	// speed gate to non-instants only, and Bolt is an Instant, so
	// this works.
	boltID := uuid.New()
	opponent.Hand.PushTop(game.Card{
		InstanceID: boltID,
		Name:       "Lightning Bolt",
		TypeLine:   "Instant",
		OracleID:   "4457ed35-7c10-48c8-9776-456485fdf070",
		Owner:      opponent.ID,
		Controller: opponent.ID,
	})
	if err := g.CastSpell(opponent.ID, boltID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: caster.ID}},
	}); err != nil {
		t.Fatalf("Opponent CastSpell Bolt: %v", err)
	}

	// Active seat casts Counterspell targeting the Bolt stack item.
	counterID := uuid.New()
	caster.Hand.PushTop(game.Card{
		InstanceID: counterID,
		Name:       "Counterspell",
		TypeLine:   "Instant",
		OracleID:   "cc187110-1148-4090-bbb8-e205694a39f5",
		Owner:      caster.ID,
		Controller: caster.ID,
	})
	if err := g.CastSpell(caster.ID, counterID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: boltID}},
	}); err != nil {
		t.Fatalf("Caster CastSpell Counterspell: %v", err)
	}

	// Pass priority around — Counterspell resolves first (top of
	// stack), pulling Bolt to opponent's graveyard. Then Bolt would
	// try to resolve but is no longer on the stack.
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(boltID) || g.Stack.Contains(boltID) {
		t.Errorf("Bolt did not leave the stack")
	}
	if !opponent.Graveyard.Contains(boltID) {
		t.Errorf("countered Bolt not in opponent graveyard")
	}
	if !caster.Graveyard.Contains(counterID) {
		t.Errorf("Counterspell not in caster graveyard")
	}
	// Life total unchanged — Bolt never dealt damage.
	if caster.Life != beforeLife {
		t.Errorf("caster life changed: got %d, want %d", caster.Life, beforeLife)
	}
}

func TestNegateCountersSpell(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	opponent := g.Seats[1]

	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}

	// Opponent casts something (use Shock — any spell targeting
	// Negate tests the counter independent of the noncreature
	// predicate that Negate doesn't enforce in S14).
	shockID := uuid.New()
	opponent.Hand.PushTop(game.Card{
		InstanceID: shockID,
		Name:       "Shock",
		TypeLine:   "Instant",
		OracleID:   "a9d288b8-cdc1-4e55-a0c9-d6edfc95e65d",
		Owner:      opponent.ID,
		Controller: opponent.ID,
	})
	if err := g.CastSpell(opponent.ID, shockID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: caster.ID}},
	}); err != nil {
		t.Fatalf("Opponent Shock: %v", err)
	}

	// Caster Negates.
	negateID := uuid.New()
	caster.Hand.PushTop(game.Card{
		InstanceID: negateID,
		Name:       "Negate",
		TypeLine:   "Instant",
		OracleID:   "3407fe41-fdd3-4119-8f70-4bc4590a379f",
		Owner:      caster.ID,
		Controller: caster.ID,
	})
	if err := g.CastSpell(caster.ID, negateID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: shockID}},
	}); err != nil {
		t.Fatalf("Caster Negate: %v", err)
	}

	passPriorityAroundTable(t, g)

	if !opponent.Graveyard.Contains(shockID) {
		t.Errorf("countered Shock not in opponent graveyard")
	}
}

// TestSwanSongCountersAndCreatesBirdToken proves the two-step
// composition (CounterTarget → CreateToken) + the "token under the
// countered spell's controller" routing.
func TestSwanSongCountersAndCreatesBirdToken(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	opponent := g.Seats[1]

	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}

	// Opponent casts Shock (a valid Swan Song target: instants are
	// enchantment/instant/sorcery).
	shockID := uuid.New()
	opponent.Hand.PushTop(game.Card{
		InstanceID: shockID,
		Name:       "Shock",
		TypeLine:   "Instant",
		OracleID:   "a9d288b8-cdc1-4e55-a0c9-d6edfc95e65d",
		Owner:      opponent.ID,
		Controller: opponent.ID,
	})
	if err := g.CastSpell(opponent.ID, shockID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: caster.ID}},
	}); err != nil {
		t.Fatalf("Opponent Shock: %v", err)
	}

	// Caster Swan Songs.
	swanID := uuid.New()
	caster.Hand.PushTop(game.Card{
		InstanceID: swanID,
		Name:       "Swan Song",
		TypeLine:   "Instant",
		OracleID:   "8ddfc283-c9b4-41a5-af88-cf0068e986cc",
		Owner:      caster.ID,
		Controller: caster.ID,
	})

	bfBefore := len(g.Battlefield.Cards)

	if err := g.CastSpell(caster.ID, swanID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: shockID}},
	}); err != nil {
		t.Fatalf("Caster Swan Song: %v", err)
	}
	passPriorityAroundTable(t, g)

	// Shock countered into opponent's graveyard.
	if !opponent.Graveyard.Contains(shockID) {
		t.Errorf("countered Shock not in opponent graveyard")
	}
	// One new token on the battlefield under opponent (spell
	// controller) control.
	bfDelta := len(g.Battlefield.Cards) - bfBefore
	if bfDelta != 1 {
		t.Errorf("battlefield delta: got %d, want 1 (Bird token)", bfDelta)
	}
	var tokenCard game.Card
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Bird" {
			tokenCard = c
			found = true
		}
	}
	if !found {
		t.Fatalf("bird token not on battlefield")
	}
	if tokenCard.Controller != opponent.ID {
		t.Errorf("token controller: got %v, want %v (countered spell's controller)", tokenCard.Controller, opponent.ID)
	}
	if tokenCard.Power != 2 || tokenCard.Toughness != 2 {
		t.Errorf("token stats: got %d/%d, want 2/2", tokenCard.Power, tokenCard.Toughness)
	}
}

// --- Opt-in canary ---------------------------------------------

// --- Draw + mill + discard (sub-PR 5) --------------------------

func TestDivinationDrawsTwo(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	handBefore := caster.Hand.Size()

	castCatalogSpell(t, g, "Divination", "Sorcery",
		"273b339c-964b-4a18-8eb5-ceb8abcdfd9e",
		nil,
	)
	passPriorityAroundTable(t, g)

	if got := caster.Hand.Size() - handBefore; got != 2 {
		t.Errorf("hand delta: got %d, want 2 (Divination spell is now in graveyard, net +2)", got)
	}
}

func TestHarmonizeDrawsThree(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	handBefore := caster.Hand.Size()

	castCatalogSpell(t, g, "Harmonize", "Sorcery",
		"7eff84f1-f772-497a-b350-bbc93d0230f7",
		nil,
	)
	passPriorityAroundTable(t, g)

	if got := caster.Hand.Size() - handBefore; got != 3 {
		t.Errorf("hand delta: got %d, want 3", got)
	}
}

func TestSignInBloodDrawsAndDrains(t *testing.T) {
	g := newCatalogGame(t)
	target := g.Seats[1]
	targetHand := target.Hand.Size()
	targetLife := target.Life

	castCatalogSpell(t, g, "Sign in Blood", "Sorcery",
		"c6207f6a-a624-4754-88f5-dbe700c841ff",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: target.ID}},
	)
	passPriorityAroundTable(t, g)

	if got := target.Hand.Size() - targetHand; got != 2 {
		t.Errorf("target hand delta: got %d, want 2", got)
	}
	if target.Life != targetLife-2 {
		t.Errorf("target life: got %d, want %d", target.Life, targetLife-2)
	}
}

func TestGlimpseMillsTen(t *testing.T) {
	g := newCatalogGame(t)
	target := g.Seats[1]
	libBefore := target.Library.Size()
	gyBefore := target.Graveyard.Size()

	castCatalogSpell(t, g, "Glimpse the Unthinkable", "Sorcery",
		"552f0163-a19d-4671-888f-044fc0354875",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: target.ID}},
	)
	passPriorityAroundTable(t, g)

	if got := libBefore - target.Library.Size(); got != 10 {
		t.Errorf("library delta: got %d, want 10", got)
	}
	if got := target.Graveyard.Size() - gyBefore; got != 10 {
		t.Errorf("graveyard delta: got %d, want 10", got)
	}
}

func TestMindRotQueuesDiscardChoice(t *testing.T) {
	g := newCatalogGame(t)
	target := g.Seats[1]
	// Seed extra hand cards so the discard-2 choice has options.
	pushHandCard(g, target)
	pushHandCard(g, target)
	handBefore := target.Hand.Size()
	gyBefore := target.Graveyard.Size()

	castCatalogSpell(t, g, "Mind Rot", "Sorcery",
		"ad44cf74-b717-48fb-9fa2-77512024d76a",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: target.ID}},
	)
	passPriorityAroundTable(t, g)

	// The discard is deferred until the target picks — no hand /
	// graveyard movement yet.
	if got := target.Hand.Size(); got != handBefore {
		t.Errorf("hand size changed pre-choice: got %d, want %d", got, handBefore)
	}
	if got := target.Graveyard.Size() - gyBefore; got != 0 {
		t.Errorf("graveyard changed pre-choice: got delta %d, want 0", got)
	}
	if g.DiscardPending[target.ID] != 2 {
		t.Fatalf("DiscardPending[target]: got %d, want 2", g.DiscardPending[target.ID])
	}

	// Simulate the target submitting their picks.
	picks := []uuid.UUID{
		target.Hand.Cards[0].InstanceID,
		target.Hand.Cards[1].InstanceID,
	}
	if err := g.DiscardSelection(target.ID, picks); err != nil {
		t.Fatalf("DiscardSelection: %v", err)
	}

	if got := handBefore - target.Hand.Size(); got != 2 {
		t.Errorf("post-selection hand delta: got %d, want 2", got)
	}
	if got := target.Graveyard.Size() - gyBefore; got != 2 {
		t.Errorf("post-selection graveyard delta: got %d, want 2", got)
	}
	if _, still := g.DiscardPending[target.ID]; still {
		t.Errorf("DiscardPending not cleared after selection")
	}
}

func TestThoughtseizeQueuesCasterChoice(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	target := g.Seats[1]
	handBefore := target.Hand.Size()
	gyBefore := target.Graveyard.Size()
	casterLife := caster.Life

	castCatalogSpell(t, g, "Thoughtseize", "Sorcery",
		"edd8d1e8-be43-4c38-bb3a-83081fbaf0b5",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: target.ID}},
	)
	passPriorityAroundTable(t, g)

	// Life loss fires at resolution; the discard is deferred.
	if caster.Life != casterLife-2 {
		t.Errorf("caster life: got %d, want %d", caster.Life, casterLife-2)
	}
	// Hand is revealed to the caster.
	for _, c := range target.Hand.Cards {
		if !c.IsKnownTo(caster.ID) {
			t.Errorf("caster not a knower of %v after Thoughtseize reveal", c.InstanceID)
		}
	}
	// Pre-pick: nothing moved.
	if got := target.Hand.Size(); got != handBefore {
		t.Errorf("hand size changed before resolve_choice: got %d, want %d", got, handBefore)
	}
	if got := target.Graveyard.Size() - gyBefore; got != 0 {
		t.Errorf("graveyard changed before resolve_choice: got delta %d, want 0", got)
	}
	// A PendingChoice should exist addressed to the caster, sourced
	// from the target.
	if len(g.PendingChoices) != 1 {
		t.Fatalf("PendingChoices len: got %d, want 1", len(g.PendingChoices))
	}
	choice := g.PendingChoices[0]
	if choice.Chooser != caster.ID {
		t.Errorf("choice.Chooser: got %v, want %v (caster)", choice.Chooser, caster.ID)
	}
	if choice.FromPlayer != target.ID {
		t.Errorf("choice.FromPlayer: got %v, want %v (target)", choice.FromPlayer, target.ID)
	}
	if choice.Count != 1 {
		t.Errorf("choice.Count: got %d, want 1", choice.Count)
	}

	// Caster picks one of the target's hand cards and submits.
	pick := target.Hand.Cards[0].InstanceID
	if err := g.ResolvePendingChoice(choice.ID, caster.ID, []uuid.UUID{pick}); err != nil {
		t.Fatalf("ResolvePendingChoice: %v", err)
	}

	if target.Hand.Contains(pick) {
		t.Errorf("picked card still in target hand")
	}
	if !target.Graveyard.Contains(pick) {
		t.Errorf("picked card not in target graveyard")
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("PendingChoices not drained after resolve: len=%d", len(g.PendingChoices))
	}
}

// --- Targeted removal ------------------------------------------

func TestSwordsToPlowsharesExilesAndGainsLife(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	opponent := g.Seats[1]
	// Put a 4/4 on opponent's battlefield.
	creatureID := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: creatureID,
		Name:       "Big Thing",
		TypeLine:   "Creature — Giant",
		Power:      4,
		Toughness:  4,
		Owner:      opponent.ID,
		Controller: opponent.ID,
	})
	oppLifeBefore := opponent.Life

	castCatalogSpell(t, g, "Swords to Plowshares", "Instant",
		"b1544f21-7e98-461b-aed5-e748b0168c52",
		[]game.TargetRef{{Kind: game.TargetCard, ID: creatureID}},
	)
	passPriorityAroundTable(t, g)

	if !g.Exile.Contains(creatureID) {
		t.Errorf("creature not in exile")
	}
	if opponent.Life != oppLifeBefore+4 {
		t.Errorf("opponent life: got %d, want %d", opponent.Life, oppLifeBefore+4)
	}
	_ = caster
}

func TestPathToExileExilesAndFetchesBasic(t *testing.T) {
	g := newCatalogGame(t)
	opponent := g.Seats[1]
	creatureID := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: creatureID,
		Name:       "Beater",
		TypeLine:   "Creature — Beast",
		Power:      3,
		Toughness:  3,
		Owner:      opponent.ID,
		Controller: opponent.ID,
	})
	// Seed a Forest in opponent's library so Path's search has
	// something to find.
	forestID := uuid.New()
	opponent.Library.PushTop(game.Card{
		InstanceID: forestID,
		Name:       "Forest",
		TypeLine:   "Basic Land — Forest",
		Owner:      opponent.ID,
		Controller: opponent.ID,
	})

	castCatalogSpell(t, g, "Path to Exile", "Instant",
		"d683d985-9888-4d21-8b5f-69e69ce4a03b",
		[]game.TargetRef{{Kind: game.TargetCard, ID: creatureID}},
	)
	passPriorityAroundTable(t, g)

	if !g.Exile.Contains(creatureID) {
		t.Errorf("creature not exiled")
	}
	if !g.Battlefield.Contains(forestID) {
		t.Errorf("fetched Forest not on battlefield")
	}
}

func TestUnsummonBouncesToHand(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	creatureID := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: creatureID,
		Name:       "Bouncable",
		TypeLine:   "Creature — Bird",
		Power:      2,
		Toughness:  2,
		Owner:      owner.ID,
		Controller: owner.ID,
	})

	castCatalogSpell(t, g, "Unsummon", "Instant",
		"837182db-1bf3-4a2c-bd01-1af9d9873561",
		[]game.TargetRef{{Kind: game.TargetCard, ID: creatureID}},
	)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(creatureID) {
		t.Errorf("creature still on battlefield")
	}
	if !owner.Hand.Contains(creatureID) {
		t.Errorf("creature not in owner's hand")
	}
}

// TestNonCatalogSpellStaysSandbox proves a spell without a
// catalog-matching OracleID resolves to graveyard with no effect —
// the Cockatrice-style manual fallback is intact. Canary for
// future regressions that might accidentally activate auto-resolve
// on non-catalog cards.
// --- Tutors + recursion + ETB (sub-PR 6) -----------------------

// pushLibraryCardForTest pushes a specific Card onto the BOTTOM of
// a player's library. SearchLibrary iterates bottom-to-top and picks
// the first predicate match, so bottom placement guarantees the
// seeded card wins for accept-all predicates (Demonic / Vampiric
// Tutor). For tutors with a selective predicate (Cultivate,
// Solemn Simulacrum) the filler library is IsBasicLand-negative,
// so bottom placement still wins deterministically.
func pushLibraryCardForTest(p *game.Player, c game.Card) uuid.UUID {
	if c.InstanceID == uuid.Nil {
		c.InstanceID = uuid.New()
	}
	if c.Owner == uuid.Nil {
		c.Owner = p.ID
	}
	if c.Controller == uuid.Nil {
		c.Controller = p.ID
	}
	p.Library.PushBottom(c)
	return c.InstanceID
}

// pushGraveyardCardForTest seeds a card into a player's graveyard
// pile. Regrowth / Eternal Witness sandbox-pick the TOP (most
// recently pushed) card, so push order matters in the tests.
func pushGraveyardCardForTest(p *game.Player, name string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   "Creature — Test",
		Power:      1,
		Toughness:  1,
		Owner:      p.ID,
		Controller: p.ID,
	})
	return id
}

func TestDemonicTutorSearchesIntoHand(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	// Seed a unique needle on top of the library so we can assert
	// it was the card moved into hand.
	needle := pushLibraryCardForTest(caster, game.Card{Name: "Needle"})
	libBefore := caster.Library.Size()
	handBefore := caster.Hand.Size()

	castCatalogSpell(t, g, "Demonic Tutor", "Sorcery",
		"82004860-e589-4e38-8d61-8c0210e4ea39",
		nil,
	)
	passPriorityAroundTable(t, g)

	if !caster.Hand.Contains(needle) {
		t.Errorf("needle card not tutored into hand")
	}
	if caster.Library.Size() != libBefore-1 {
		t.Errorf("library size delta: got %d, want -1", caster.Library.Size()-libBefore)
	}
	// Net +1 hand (Demonic Tutor itself went to graveyard on resolve).
	if got := caster.Hand.Size() - handBefore; got != 1 {
		t.Errorf("hand delta: got %d, want 1", got)
	}
}

func TestVampiricTutorSearchesAndLosesLife(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	needle := pushLibraryCardForTest(caster, game.Card{Name: "Vamp Needle"})
	lifeBefore := caster.Life

	castCatalogSpell(t, g, "Vampiric Tutor", "Instant",
		"ededbdae-d9dc-4206-9335-d7158f2d7700",
		nil,
	)
	passPriorityAroundTable(t, g)

	if !caster.Hand.Contains(needle) {
		t.Errorf("needle not tutored into hand")
	}
	if caster.Life != lifeBefore-2 {
		t.Errorf("caster life: got %d, want %d", caster.Life, lifeBefore-2)
	}
}

func TestCultivateFetchesOneToFieldOneToHand(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	// Seed two basics on top (plus a non-basic decoy beneath).
	forest1 := pushLibraryCardForTest(caster, game.Card{
		Name: "Forest", TypeLine: "Basic Land — Forest",
	})
	forest2 := pushLibraryCardForTest(caster, game.Card{
		Name: "Forest", TypeLine: "Basic Land — Forest",
	})

	castCatalogSpell(t, g, "Cultivate", "Sorcery",
		"8b755881-a72d-4e21-a369-d2924eb4585a",
		nil,
	)
	passPriorityAroundTable(t, g)

	// Two basics moved out of the library.
	onField := g.Battlefield.Contains(forest1) || g.Battlefield.Contains(forest2)
	inHand := caster.Hand.Contains(forest1) || caster.Hand.Contains(forest2)
	if !onField {
		t.Errorf("no Forest reached the battlefield")
	}
	if !inHand {
		t.Errorf("no Forest reached the hand")
	}
	// Exactly one of each — no double-land, no self-destruct.
	if g.Battlefield.Contains(forest1) && g.Battlefield.Contains(forest2) {
		t.Errorf("both forests on battlefield (expected one each to bf + hand)")
	}
	if caster.Hand.Contains(forest1) && caster.Hand.Contains(forest2) {
		t.Errorf("both forests in hand (expected one each to bf + hand)")
	}
}

func TestRegrowthReturnsTopOfGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	// Older card at bottom, target on top.
	pushGraveyardCardForTest(caster, "Older Body")
	top := pushGraveyardCardForTest(caster, "Latest Body")

	castCatalogSpell(t, g, "Regrowth", "Sorcery",
		"e6e4a8bd-5c40-4654-8de1-0da9afed90fd",
		nil,
	)
	passPriorityAroundTable(t, g)

	if !caster.Hand.Contains(top) {
		t.Errorf("top-of-graveyard card not returned to hand")
	}
	if caster.Graveyard.Contains(top) {
		t.Errorf("returned card still in graveyard")
	}
}

// TestEternalWitnessETBReturnsTopGraveyard casts E-Witness, resolves
// it, accepts the S19 OptionalPrompt (CR 603.4 "you may"), and
// expects the top-of-graveyard sandbox auto-pick to land in hand.
//
// S14 ran the OnETB hook inline during the resolve mutation; S19
// sub-PR 3 routes the same auto-pick through the listener path with
// a yes/no prompt for the controller. The test now answers yes via
// resolve_choice.
func TestEternalWitnessETBReturnsTopGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	pushGraveyardCardForTest(caster, "Older Body")
	top := pushGraveyardCardForTest(caster, "Latest Body")

	witnessID := castCatalogSpell(t, g, "Eternal Witness", "Creature — Human Shaman",
		"30b24e8e-3b0e-4d8e-90f3-f66eb7c1858c",
		nil,
	)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, caster.ID, true)

	// "Yes" puts the trigger on the stack with the auto-picked
	// target; nothing moves until it resolves.
	item := triggerOnStack(g, witnessID)
	if item == nil {
		t.Fatalf("E-Witness trigger not on the stack after accepting the prompt")
	}
	if len(item.Targets) != 1 || item.Targets[0].ID != top {
		t.Errorf("trigger targets = %+v, want the top-of-graveyard card %s", item.Targets, top)
	}
	if caster.Hand.Contains(top) {
		t.Fatalf("card returned before the trigger resolved")
	}
	passPriorityAroundTable(t, g)

	if !caster.Hand.Contains(top) {
		t.Errorf("E-Witness ETB did not return top-of-graveyard card")
	}
}

// TestSolemnSimulacrumETBFetchesBasic casts the Golem, resolves it,
// accepts the S19 OptionalPrompt, and expects the Forest auto-pick
// to land tapped on the battlefield.
func TestSolemnSimulacrumETBFetchesBasic(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	forestID := pushLibraryCardForTest(caster, game.Card{
		Name: "Forest", TypeLine: "Basic Land — Forest",
	})

	solemnID := castCatalogSpell(t, g, "Solemn Simulacrum", "Artifact Creature — Golem",
		"00c0543c-2a1f-4425-8283-4062d74a1637",
		nil,
	)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, caster.ID, true)
	if triggerOnStack(g, solemnID) == nil {
		t.Fatalf("Solemn ETB trigger not on the stack after accepting the prompt")
	}
	if g.Battlefield.Contains(forestID) {
		t.Fatalf("Forest fetched before the trigger resolved")
	}
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(forestID) {
		t.Errorf("fetched Forest not on battlefield")
	}
}

// answerLatestTriggerPrompt finds the most-recently-queued
// PendingChoiceTriggerPrompt addressed to chooserID and submits the
// given apply answer. Fails the test loudly when no such prompt is
// pending — a missing prompt usually means the harvester didn't
// fire (oracle ID typo, AppliesTo predicate wrong, or the trigger
// declared without an OptionalPrompt). Added in S19 sub-PR 3.
func answerLatestTriggerPrompt(t *testing.T, g *game.Game, chooserID uuid.UUID, apply bool) {
	t.Helper()
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c == nil || c.Kind != game.PendingChoiceTriggerPrompt {
			continue
		}
		if c.Chooser != chooserID {
			continue
		}
		if err := g.ResolveTriggerPrompt(c.ID, chooserID, apply); err != nil {
			t.Fatalf("ResolveTriggerPrompt: %v", err)
		}
		return
	}
	t.Fatalf("no PendingChoiceTriggerPrompt addressed to %s", chooserID)
}

// --- S19 sub-PR 3: ETB-trigger cards ---------------------------

// TestMulldrifterETBDrawsTwo casts Mulldrifter and expects the
// mandatory ETB trigger to go on the stack when the creature
// resolves, then draw 2 cards when the trigger itself resolves.
// No prompt (mandatory trigger); passPriorityAroundTable settles
// both the spell and the trigger it queues.
func TestMulldrifterETBDrawsTwo(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	handBefore := caster.Hand.Size()

	mullID := castCatalogSpell(t, g, "Mulldrifter", "Creature — Elemental",
		"24d0f5e7-0d9e-4b76-900e-a7274e80312d",
		nil,
	)
	// Resolve the creature spell only: the ETB trigger must be on
	// the stack, and the draw must not have happened yet.
	for i := 0; i < 8 && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !g.Battlefield.Contains(mullID) {
		t.Fatalf("Mulldrifter did not resolve to the battlefield")
	}
	item := triggerOnStack(g, mullID)
	if item == nil {
		t.Fatalf("Mulldrifter ETB trigger not on the stack after the creature resolved")
	}
	if item.Controller != caster.ID {
		t.Errorf("trigger controller = %s, want caster %s", item.Controller, caster.ID)
	}
	if got := caster.Hand.Size() - handBefore; got != 0 {
		t.Fatalf("hand changed before the trigger resolved: delta %d", got)
	}
	passPriorityAroundTable(t, g)

	if got := caster.Hand.Size() - handBefore; got != 2 {
		t.Errorf("hand delta after Mulldrifter ETB: got %d, want 2", got)
	}
	// Mandatory trigger leaves no prompt behind.
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt {
			t.Errorf("unexpected trigger prompt for mandatory Mulldrifter ETB")
		}
	}
}

// TestReclamationSageETBDestroysOpponentArtifact pre-seeds an
// opponent-controlled artifact, casts Reclamation Sage, accepts the
// optional ETB prompt, and verifies the auto-pick destroyed the
// artifact.
func TestReclamationSageETBDestroysOpponentArtifact(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	opp := g.Seats[1]

	artifactID := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: artifactID,
		Name:       "Test Sword",
		TypeLine:   "Artifact",
		Owner:      opp.ID,
		Controller: opp.ID,
	})

	sageID := castCatalogSpell(t, g, "Reclamation Sage", "Creature — Elf Shaman",
		"032ec6e2-6cc3-4a97-9cc7-3233f5e11904",
		nil,
	)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, caster.ID, true)

	item := triggerOnStack(g, sageID)
	if item == nil {
		t.Fatalf("Reclamation Sage trigger not on the stack after accepting the prompt")
	}
	if len(item.Targets) != 1 || item.Targets[0].Kind != game.TargetCard || item.Targets[0].ID != artifactID {
		t.Errorf("trigger targets = %+v, want the opponent's artifact %s", item.Targets, artifactID)
	}
	if !g.Battlefield.Contains(artifactID) {
		t.Fatalf("artifact destroyed before the trigger resolved")
	}
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(artifactID) {
		t.Errorf("Reclamation Sage ETB did not destroy the opponent's artifact")
	}
	if !opp.Graveyard.Contains(artifactID) {
		t.Errorf("destroyed artifact did not land in opponent's graveyard")
	}
}

// TestReclamationSageETBDeclineSparesTarget exercises the optional
// path: caster declines the prompt → effect does not fire → the
// opponent's artifact stays on the battlefield.
func TestReclamationSageETBDeclineSparesTarget(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	opp := g.Seats[1]

	artifactID := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: artifactID,
		Name:       "Test Sword",
		TypeLine:   "Artifact",
		Owner:      opp.ID,
		Controller: opp.ID,
	})

	castCatalogSpell(t, g, "Reclamation Sage", "Creature — Elf Shaman",
		"032ec6e2-6cc3-4a97-9cc7-3233f5e11904",
		nil,
	)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, caster.ID, false)

	if !g.Battlefield.Contains(artifactID) {
		t.Errorf("declined Reclamation Sage prompt still destroyed the artifact")
	}
}

// latestTriggerPrompt returns the most-recently-queued
// PendingChoiceTriggerPrompt addressed to chooserID, or nil. Used by
// the no-legal-target regression tests below to inspect the
// NoLegalTarget flag the harvester stamps at queue time.
func latestTriggerPrompt(g *game.Game, chooserID uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt && c.Chooser == chooserID {
			return c
		}
	}
	return nil
}

// TestReclamationSageETBNoLegalTargetWhenOnlyOwnArtifact pins the
// behaviour that confused a manual tester: the S19 sandbox
// auto-picker (pickFirstOpponentNonland) only targets OPPONENT
// permanents, so when the only artifact in play is the caster's own,
// Reclamation Sage's optional trigger has no legal target. The prompt
// still queues (CR 603.4), but it's flagged NoLegalTarget so the
// client can warn the chooser, and answering "Yes" spares the
// caster's own artifact. Real target selection lands with S20.
func TestReclamationSageETBNoLegalTargetWhenOnlyOwnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]

	// The ONLY artifact in play belongs to the caster — no opponent
	// artifact exists.
	ownArtifactID := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: ownArtifactID,
		Name:       "Test Sol Ring",
		TypeLine:   "Artifact",
		Owner:      caster.ID,
		Controller: caster.ID,
	})

	castCatalogSpell(t, g, "Reclamation Sage", "Creature — Elf Shaman",
		"032ec6e2-6cc3-4a97-9cc7-3233f5e11904",
		nil,
	)
	passPriorityAroundTable(t, g)

	prompt := latestTriggerPrompt(g, caster.ID)
	if prompt == nil {
		t.Fatalf("expected a Reclamation Sage trigger prompt for the caster")
	}
	if !prompt.NoLegalTarget {
		t.Errorf("expected NoLegalTarget=true when the only artifact is the caster's own")
	}

	answerLatestTriggerPrompt(t, g, caster.ID, true)

	if !g.Battlefield.Contains(ownArtifactID) {
		t.Errorf("Reclamation Sage destroyed the caster's OWN artifact — auto-picker must target opponents only")
	}
}

// TestEternalWitnessETBNoLegalTargetEmptyGraveyard pins the other
// half of the same confusion: Eternal Witness returns from the
// caster's own graveyard, so an empty graveyard means no legal
// target. The prompt is flagged NoLegalTarget and "Yes" is a no-op.
func TestEternalWitnessETBNoLegalTargetEmptyGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	if caster.Graveyard.Size() != 0 {
		t.Fatalf("precondition: caster graveyard should start empty, got %d", caster.Graveyard.Size())
	}

	castCatalogSpell(t, g, "Eternal Witness", "Creature — Human Shaman",
		"30b24e8e-3b0e-4d8e-90f3-f66eb7c1858c",
		nil,
	)
	passPriorityAroundTable(t, g)

	prompt := latestTriggerPrompt(g, caster.ID)
	if prompt == nil {
		t.Fatalf("expected an Eternal Witness trigger prompt for the caster")
	}
	if !prompt.NoLegalTarget {
		t.Errorf("expected NoLegalTarget=true with an empty graveyard")
	}

	handBefore := caster.Hand.Size()
	answerLatestTriggerPrompt(t, g, caster.ID, true)
	if caster.Hand.Size() != handBefore {
		t.Errorf("Eternal Witness returned a card from an empty graveyard")
	}
}

// TestReclamationSageETBHasLegalTargetFlagFalseWithOpponentArtifact
// is the positive control: with a genuine opponent artifact present,
// the prompt is NOT flagged NoLegalTarget.
func TestReclamationSageETBHasLegalTargetFlagFalseWithOpponentArtifact(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	opp := g.Seats[1]

	g.Battlefield.PushTop(game.Card{
		InstanceID: uuid.New(),
		Name:       "Test Sword",
		TypeLine:   "Artifact",
		Owner:      opp.ID,
		Controller: opp.ID,
	})

	castCatalogSpell(t, g, "Reclamation Sage", "Creature — Elf Shaman",
		"032ec6e2-6cc3-4a97-9cc7-3233f5e11904",
		nil,
	)
	passPriorityAroundTable(t, g)

	prompt := latestTriggerPrompt(g, caster.ID)
	if prompt == nil {
		t.Fatalf("expected a Reclamation Sage trigger prompt for the caster")
	}
	if prompt.NoLegalTarget {
		t.Errorf("expected NoLegalTarget=false when an opponent artifact is in play")
	}
}

// TestAcidicSlimeETBDestroysOpponentLand seeds an opponent's land,
// casts Acidic Slime, and expects the mandatory ETB to destroy it
// without prompting (no "you may" gate on Acidic Slime).
func TestAcidicSlimeETBDestroysOpponentLand(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]

	landID := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: landID,
		Name:       "Mountain",
		TypeLine:   "Basic Land — Mountain",
		Owner:      opp.ID,
		Controller: opp.ID,
	})

	castCatalogSpell(t, g, "Acidic Slime", "Creature — Ooze",
		"21f45043-5419-4019-8b6c-e5294bd5f549",
		nil,
	)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(landID) {
		t.Errorf("Acidic Slime ETB did not destroy the opponent's land")
	}
	if !opp.Graveyard.Contains(landID) {
		t.Errorf("destroyed land did not land in opponent's graveyard")
	}
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt {
			t.Errorf("Acidic Slime is mandatory — no trigger prompt should queue")
		}
	}
}

// --- Vanilla permanents ----------------------------------------

// TestSolRingResolvesToBattlefield just proves that registering a
// vanilla spec (no OnResolve / OnETB) still routes the card through
// the normal battlefield landing — the catalog's presence shouldn't
// block resolution.
func TestSolRingResolvesToBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	solID := castCatalogSpell(t, g, "Sol Ring", "Artifact",
		"6ad8011d-3471-4369-9d68-b264cc027487",
		nil,
	)
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(solID) {
		t.Errorf("Sol Ring did not reach battlefield")
	}
}

func TestArcaneSignetResolvesToBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	id := castCatalogSpell(t, g, "Arcane Signet", "Artifact",
		"0bc7f093-bef0-4f1a-852c-4b75ebf54838",
		nil,
	)
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(id) {
		t.Errorf("Arcane Signet did not reach battlefield")
	}
}

func TestBirdsOfParadiseResolvesToBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	id := castCatalogSpell(t, g, "Birds of Paradise", "Creature — Bird",
		"d3a0b660-358c-41bd-9cd2-41fbf3491b1a",
		nil,
	)
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(id) {
		t.Errorf("Birds of Paradise did not reach battlefield")
	}
}

// --- Planeswalker ----------------------------------------------

// TestWanderingEmperorEntersWithStartingLoyalty exercises the
// StartingLoyalty stamp: the ETB hook writes "loyalty" counters
// equal to the Spec's value before the next SBA pass.
func TestWanderingEmperorEntersWithStartingLoyalty(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	pwID := castCatalogSpell(t, g, "The Wandering Emperor", "Legendary Planeswalker — Wanderer",
		"0c7f18d5-36cb-4bc6-a358-443b97666215",
		nil,
	)
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(pwID) {
		t.Fatalf("planeswalker did not reach battlefield")
	}
	var card *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == pwID {
			card = &g.Battlefield.Cards[i]
			break
		}
	}
	if card == nil {
		t.Fatalf("planeswalker card not found on battlefield")
	}
	if got := card.Counters["loyalty"]; got != 3 {
		t.Errorf("loyalty counters: got %d, want 3", got)
	}
	_ = caster
}

// --- S15 sub-PR 2 mana abilities ------------------------------

// pushArtifactToBattlefieldForTest puts a specific instance on the
// shared battlefield with the named oracle_id so the catalog lookup
// fires. Distinct from pushCreatureToBattlefieldForTest so we can
// use a non-creature TypeLine (Sol Ring / Arcane Signet / Birds of
// Paradise all have different shapes).
func pushArtifactToBattlefieldForTest(g *game.Game, owner uuid.UUID, name, typeLine, oracleID string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		OracleID:   oracleID,
		Owner:      owner,
		Controller: owner,
	})
	return id
}

func TestSolRingTapAddsTwoColorless(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	solID := pushArtifactToBattlefieldForTest(g, caster.ID, "Sol Ring", "Artifact",
		"6ad8011d-3471-4369-9d68-b264cc027487",
	)

	if err := g.ActivateManaAbility(caster.ID, solID, 0); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if len(caster.ManaPool) != 2 {
		t.Fatalf("mana pool: got %d tokens, want 2", len(caster.ManaPool))
	}
	for i, tok := range caster.ManaPool {
		if tok.Color != "C" {
			t.Errorf("token[%d]: got color %q, want C", i, tok.Color)
		}
		if tok.Source != solID {
			t.Errorf("token[%d]: got source %v, want %v", i, tok.Source, solID)
		}
	}
	// Sol Ring should now be tapped (tap cost).
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == solID && !c.Tapped {
			t.Errorf("Sol Ring should be tapped after activation")
		}
	}
	// Re-activating while tapped should fail.
	if err := g.ActivateManaAbility(caster.ID, solID, 0); err == nil {
		t.Errorf("second activation on tapped Sol Ring should have returned an error")
	}
	// No PendingChoice — no pipe in the produced string.
	if len(g.PendingChoices) != 0 {
		t.Errorf("unexpected PendingChoice: %+v", g.PendingChoices)
	}
}

func TestBirdsOfParadiseTapQueuesMagicChoice(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	birdID := pushArtifactToBattlefieldForTest(g, caster.ID, "Birds of Paradise", "Creature — Bird",
		"d3a0b660-358c-41bd-9cd2-41fbf3491b1a",
	)

	if err := g.ActivateManaAbility(caster.ID, birdID, 0); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	// No mana lands in the pool yet — deferred to resolve_choice.
	if len(caster.ManaPool) != 0 {
		t.Errorf("mana pool should be empty pre-choice: got %+v", caster.ManaPool)
	}
	// One PendingChoice addressed to the controller with all five
	// colors in the option set (no commander identity present in
	// the test harness, so the filter falls through unchanged).
	if len(g.PendingChoices) != 1 {
		t.Fatalf("PendingChoices: got %d, want 1", len(g.PendingChoices))
	}
	choice := g.PendingChoices[0]
	if choice.Kind != game.PendingChoiceMana {
		t.Errorf("choice.Kind: got %q, want mana_pick", choice.Kind)
	}
	if choice.Chooser != caster.ID {
		t.Errorf("choice.Chooser: got %v, want %v", choice.Chooser, caster.ID)
	}
	if len(choice.ColorOptions) != 5 {
		t.Errorf("choice.ColorOptions: got %v, want 5 entries", choice.ColorOptions)
	}

	// Caster picks blue. Mana lands, queue drains.
	if err := g.ResolveManaChoice(choice.ID, caster.ID, "U"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if len(caster.ManaPool) != 1 || caster.ManaPool[0].Color != "U" {
		t.Errorf("post-resolve mana pool: got %+v, want 1×U", caster.ManaPool)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("queue not drained: %+v", g.PendingChoices)
	}
}

func TestArcaneSignetNarrowsToCommanderIdentity(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	// Seed a Bant commander (white/blue/green identity derived from
	// ManaCost via distinctColorsInManaCost).
	caster.Command.PushTop(game.Card{
		InstanceID:  uuid.New(),
		Name:        "Test Bant Commander",
		TypeLine:    "Legendary Creature",
		ManaCost:    "{1}{G}{W}{U}",
		Owner:       caster.ID,
		Controller:  caster.ID,
		IsCommander: true,
	})
	signetID := pushArtifactToBattlefieldForTest(g, caster.ID, "Arcane Signet", "Artifact",
		"0bc7f093-bef0-4f1a-852c-4b75ebf54838",
	)

	if err := g.ActivateManaAbility(caster.ID, signetID, 0); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if len(g.PendingChoices) != 1 {
		t.Fatalf("PendingChoices: got %d, want 1", len(g.PendingChoices))
	}
	choice := g.PendingChoices[0]
	// Options must include W, U, G and NOT B, R.
	want := map[string]bool{"W": true, "U": true, "G": true}
	got := map[string]bool{}
	for _, c := range choice.ColorOptions {
		got[c] = true
	}
	for color := range want {
		if !got[color] {
			t.Errorf("ColorOptions missing %q (Bant identity): %v", color, choice.ColorOptions)
		}
	}
	for _, bad := range []string{"B", "R"} {
		if got[bad] {
			t.Errorf("ColorOptions contains %q — should be narrowed out of Bant identity", bad)
		}
	}
}

// TestBasicLandSyntheticAbilityAddsOneMana exercises the fallback
// path: a card with no catalog entry but a basic-land subtype on
// its TypeLine gets a synthetic "{T}: Add {G}" ability derived
// server-side. Proves basics work without any catalog registration.
func TestBasicLandSyntheticAbilityAddsOneMana(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	forestID := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: forestID,
		Name:       "Forest",
		TypeLine:   "Basic Land — Forest",
		Owner:      caster.ID,
		Controller: caster.ID,
	})

	if err := g.ActivateManaAbility(caster.ID, forestID, 0); err != nil {
		t.Fatalf("ActivateManaAbility on Forest: %v", err)
	}
	if len(caster.ManaPool) != 1 || caster.ManaPool[0].Color != "G" {
		t.Errorf("Forest tap: got %+v, want 1×G", caster.ManaPool)
	}
}

// TestStepAdvanceEmptiesAllManaPools proves the CR 106.4 empty-on-
// step-boundary hook fires from runStepEntryHooksLocked.
func TestStepAdvanceEmptiesAllManaPools(t *testing.T) {
	g := newCatalogGame(t)
	// Seed floating mana into every seat's pool.
	for _, p := range g.Seats {
		p.ManaPool.AddMana(game.ManaToken{Color: "R"}, game.ManaToken{Color: "C"})
	}
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	for _, p := range g.Seats {
		if len(p.ManaPool) != 0 {
			t.Errorf("seat %s pool not emptied after step advance: %+v", p.Name, p.ManaPool)
		}
	}
}

func TestNonCatalogSpellStaysSandbox(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	target := g.Seats[1]
	beforeLife := target.Life

	spellID := castCatalogSpell(t, g, "Random Spell", "Instant",
		"", // empty oracle ID → cannot match the catalog
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: target.ID}},
	)
	passPriorityAroundTable(t, g)

	if target.Life != beforeLife {
		t.Errorf("non-catalog spell applied damage: target life %d, want %d", target.Life, beforeLife)
	}
	if !caster.Graveyard.Contains(spellID) {
		t.Errorf("non-catalog spell did not route to graveyard")
	}
}

// --- S19 sub-PR 4: dies-triggers (LTB) -------------------------

// pushDiesCreatureForTest seeds a catalog creature onto the
// battlefield with its real OracleID + type line so the LTB
// harvester can match its dies-trigger. Returns the instance ID.
func pushDiesCreatureForTest(g *game.Game, owner uuid.UUID, name, oracleID, typeLine string, power, toughness int) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		OracleID:   oracleID,
		TypeLine:   typeLine,
		Power:      power,
		Toughness:  toughness,
		Owner:      owner,
		Controller: owner,
	})
	return id
}

// TestFiligreeFamiliarDiesDrawsCard kills a Filigree Familiar with
// Damnation and expects the mandatory dies-trigger to draw a card.
func TestFiligreeFamiliarDiesDrawsCard(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	pushDiesCreatureForTest(g, caster.ID, "Filigree Familiar",
		"b544f690-e4bf-4a5b-984d-9256518fd574", "Artifact Creature — Fox", 2, 2)
	handBefore := caster.Hand.Size()

	castCatalogSpell(t, g, "Damnation", "Sorcery",
		"d57a8f0b-7989-4db5-8756-6f2690097252", nil)
	passPriorityAroundTable(t, g)

	if got := caster.Hand.Size() - handBefore; got != 1 {
		t.Errorf("Filigree Familiar dies-draw: hand delta %d, want 1", got)
	}
}

// TestDoomedTravelerDiesCreatesSpirit kills a Doomed Traveler and
// expects one Spirit token under the controller.
func TestDoomedTravelerDiesCreatesSpirit(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	pushDiesCreatureForTest(g, caster.ID, "Doomed Traveler",
		"a30907c0-fbde-4fd3-a8c7-f304305fcea7", "Creature — Human Soldier", 1, 1)

	castCatalogSpell(t, g, "Damnation", "Sorcery",
		"d57a8f0b-7989-4db5-8756-6f2690097252", nil)
	passPriorityAroundTable(t, g)

	spirits := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Spirit" && c.Controller == caster.ID {
			spirits++
		}
	}
	if spirits != 1 {
		t.Errorf("Doomed Traveler dies: got %d Spirit tokens, want 1", spirits)
	}
}

// TestWurmcoilEngineDiesCreatesTwoWurms kills a Wurmcoil Engine and
// expects two Phyrexian Wurm tokens under the controller.
func TestWurmcoilEngineDiesCreatesTwoWurms(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	pushDiesCreatureForTest(g, caster.ID, "Wurmcoil Engine",
		"d1a60f44-7696-49ee-91fb-cab5b3102962", "Artifact Creature — Phyrexian Wurm", 6, 6)

	castCatalogSpell(t, g, "Damnation", "Sorcery",
		"d57a8f0b-7989-4db5-8756-6f2690097252", nil)
	passPriorityAroundTable(t, g)

	wurms := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Phyrexian Wurm" && c.Controller == caster.ID {
			wurms++
		}
	}
	if wurms != 2 {
		t.Errorf("Wurmcoil Engine dies: got %d Wurm tokens, want 2", wurms)
	}
}

// TestSolemnSimulacrumDiesOptionalDraw exercises the optional dies
// half: the prompt queues, and "Yes" draws a card.
func TestSolemnSimulacrumDiesOptionalDraw(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	pushDiesCreatureForTest(g, caster.ID, "Solemn Simulacrum",
		"00c0543c-2a1f-4425-8283-4062d74a1637", "Artifact Creature — Golem", 2, 2)
	handBefore := caster.Hand.Size()

	castCatalogSpell(t, g, "Damnation", "Sorcery",
		"d57a8f0b-7989-4db5-8756-6f2690097252", nil)
	passPriorityAroundTable(t, g)

	answerLatestTriggerPrompt(t, g, caster.ID, true)
	if got := caster.Hand.Size() - handBefore; got != 0 {
		t.Fatalf("Solemn dies-draw fired before the trigger resolved: hand delta %d", got)
	}
	passPriorityAroundTable(t, g)
	if got := caster.Hand.Size() - handBefore; got != 1 {
		t.Errorf("Solemn dies-draw on Yes: hand delta %d, want 1", got)
	}
}

// TestFiligreeFamiliarBounceDoesNotDraw is the gating regression:
// bouncing the Familiar (Unsummon) is an LTB but NOT a death, so
// cardDied must suppress the dies-draw. Verified via library size
// (a draw would shrink it).
func TestFiligreeFamiliarBounceDoesNotDraw(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	famID := pushDiesCreatureForTest(g, caster.ID, "Filigree Familiar",
		"b544f690-e4bf-4a5b-984d-9256518fd574", "Artifact Creature — Fox", 2, 2)
	libBefore := caster.Library.Size()

	castCatalogSpell(t, g, "Unsummon", "Instant",
		"837182db-1bf3-4a2c-bd01-1af9d9873561",
		[]game.TargetRef{{Kind: game.TargetCard, ID: famID}})
	passPriorityAroundTable(t, g)

	if !caster.Hand.Contains(famID) {
		t.Errorf("Unsummon did not return Filigree Familiar to hand")
	}
	if caster.Library.Size() != libBefore {
		t.Errorf("bounced Familiar drew a card (lib %d -> %d) — dies-trigger mis-fired on a non-death LTB",
			libBefore, caster.Library.Size())
	}
}

// --- S19 sub-PR 5: upkeep triggers (begin_upkeep) --------------

// pushPermanentForTest seeds a non-creature catalog permanent (an
// enchantment, here) onto the battlefield with its real OracleID so
// the harvester can match its upkeep trigger.
func pushPermanentForTest(g *game.Game, owner uuid.UUID, name, oracleID, typeLine string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		OracleID:   oracleID,
		TypeLine:   typeLine,
		Owner:      owner,
		Controller: owner,
	})
	return id
}

// findBattlefieldByName returns the instance ID of the first
// battlefield card with the given name, or uuid.Nil.
func findBattlefieldByName(g *game.Game, name string) uuid.UUID {
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			return c.InstanceID
		}
	}
	return uuid.Nil
}

// advanceToUpkeepOf walks the turn engine forward until the given
// seat is the active player at its upkeep step — the point at which
// EventBeginUpkeep fires and "your upkeep" triggers resolve. Fatals
// if the cursor never lands there within a turn-cycle bound.
func advanceToUpkeepOf(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	for i := 0; i < 300; i++ {
		if g.Turn.Step == game.StepUpkeep && g.Turn.ActiveSeat == seat {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep toward upkeep of seat %d: %v", seat, err)
		}
	}
	t.Fatalf("never reached upkeep of seat %d", seat)
}

// TestPhyrexianArenaUpkeepLosesLifeDraws verifies the mandatory
// "your upkeep" trigger fires when its controller's upkeep begins.
func TestPhyrexianArenaUpkeepLosesLifeDraws(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	pushPermanentForTest(g, owner.ID, "Phyrexian Arena",
		"ee579a32-a048-4335-b966-231ba731cdea", "Enchantment")
	lifeBefore := owner.Life
	handBefore := owner.Hand.Size()

	advanceToUpkeepOf(t, g, 1)
	// The trigger is on the stack at upkeep; nothing has happened
	// yet. Priority passes resolve it.
	if arenaID := findBattlefieldByName(g, "Phyrexian Arena"); triggerOnStack(g, arenaID) == nil {
		t.Fatalf("Phyrexian Arena trigger not on the stack at the controller's upkeep")
	}
	if owner.Life != lifeBefore || owner.Hand.Size() != handBefore {
		t.Fatalf("Phyrexian Arena applied before its trigger resolved")
	}
	passPriorityAroundTable(t, g)

	if owner.Life != lifeBefore-1 {
		t.Errorf("Phyrexian Arena upkeep: life %d -> %d, want -1", lifeBefore, owner.Life)
	}
	if owner.Hand.Size() != handBefore+1 {
		t.Errorf("Phyrexian Arena upkeep: hand delta %d, want +1", owner.Hand.Size()-handBefore)
	}
}

// TestBitterblossomUpkeepLosesLifeMakesFaerie checks the life cost
// plus the Faerie Rogue token.
func TestBitterblossomUpkeepLosesLifeMakesFaerie(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	pushPermanentForTest(g, owner.ID, "Bitterblossom",
		"fb868840-09fa-49b1-85cb-b08ad065e972", "Kindred Enchantment — Faerie")
	lifeBefore := owner.Life

	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)

	if owner.Life != lifeBefore-1 {
		t.Errorf("Bitterblossom upkeep: life %d -> %d, want -1", lifeBefore, owner.Life)
	}
	faeries := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Faerie Rogue" && c.Controller == owner.ID {
			faeries++
		}
	}
	if faeries != 1 {
		t.Errorf("Bitterblossom upkeep: got %d Faerie tokens, want 1", faeries)
	}
}

// TestSulfuricVortexUpkeepDamagesController checks the 2-damage
// upkeep trigger.
func TestSulfuricVortexUpkeepDamagesController(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	pushPermanentForTest(g, owner.ID, "Sulfuric Vortex",
		"7652f328-e142-494b-a869-772ced10c26a", "Enchantment")
	lifeBefore := owner.Life

	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)

	if owner.Life != lifeBefore-2 {
		t.Errorf("Sulfuric Vortex upkeep: life %d -> %d, want -2", lifeBefore, owner.Life)
	}
}

// TestAwakeningZoneUpkeepMakesSpawn checks the token-only upkeep
// trigger.
func TestAwakeningZoneUpkeepMakesSpawn(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	pushPermanentForTest(g, owner.ID, "Awakening Zone",
		"f955bc96-d602-4142-a9a2-87009cc7028c", "Enchantment")

	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)

	spawns := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Eldrazi Spawn" && c.Controller == owner.ID {
			spawns++
		}
	}
	if spawns != 1 {
		t.Errorf("Awakening Zone upkeep: got %d Eldrazi Spawn tokens, want 1", spawns)
	}
}

// TestUpkeepTriggerGatedToControllersUpkeep is the gating
// regression: a Phyrexian Arena controlled by seat 2 must NOT fire
// on seat 1's upkeep (seat 1's upkeep is reached before seat 2's
// turn, so seat 2's life should be untouched).
func TestUpkeepTriggerGatedToControllersUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	other := g.Seats[2]
	pushPermanentForTest(g, other.ID, "Phyrexian Arena",
		"ee579a32-a048-4335-b966-231ba731cdea", "Enchantment")
	lifeBefore := other.Life

	advanceToUpkeepOf(t, g, 1)
	if !stackFullyEmpty(g) {
		t.Fatalf("something reached the stack at seat 1's upkeep — seat 2's Arena should not have triggered")
	}
	passPriorityAroundTable(t, g)

	if other.Life != lifeBefore {
		t.Errorf("seat 2's Phyrexian Arena fired on seat 1's upkeep (life %d -> %d) — not gated to its own upkeep",
			lifeBefore, other.Life)
	}
}

// --- S19: triggers use the stack ---------------------------------
//
// The cases below exist because the sub-PR 3-5 cards originally
// applied their effect inline from Build. They pin the two things
// that behaviour denied players: a window to respond to a trigger,
// and the CR 608.2b re-check when the response removes the target.

// TestTriggerOnStackCanBeCountered: Mulldrifter's ETB trigger sits
// on the stack after the creature resolves; countering the ability
// (Stifle-style) means no cards are drawn.
func TestTriggerOnStackCanBeCountered(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	handBefore := caster.Hand.Size()

	mullID := castCatalogSpell(t, g, "Mulldrifter", "Creature — Elemental",
		"24d0f5e7-0d9e-4b76-900e-a7274e80312d", nil)
	for i := 0; i < 8 && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	item := triggerOnStack(g, mullID)
	if item == nil {
		t.Fatalf("Mulldrifter ETB trigger not on the stack")
	}
	if err := g.CounterAbility(item.ID); err != nil {
		t.Fatalf("CounterAbility: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := caster.Hand.Size() - handBefore; got != 0 {
		t.Errorf("countered Mulldrifter trigger still drew: hand delta %d, want 0", got)
	}
}

// TestTriggerFizzlesWhenTargetRemovedInResponse: Reclamation Sage's
// trigger targets the opponent's artifact when it goes on the
// stack. Exiling that artifact in response leaves the trigger with
// no legal target — it's countered by game rules (EventFizzle) and
// nothing else is destroyed.
func TestTriggerFizzlesWhenTargetRemovedInResponse(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	opp := g.Seats[1]

	firstID := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: firstID, Name: "First Sword", TypeLine: "Artifact",
		Owner: opp.ID, Controller: opp.ID,
	})
	secondID := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: secondID, Name: "Second Sword", TypeLine: "Artifact",
		Owner: opp.ID, Controller: opp.ID,
	})

	sageID := castCatalogSpell(t, g, "Reclamation Sage", "Creature — Elf Shaman",
		"032ec6e2-6cc3-4a97-9cc7-3233f5e11904", nil)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, caster.ID, true)

	item := triggerOnStack(g, sageID)
	if item == nil {
		t.Fatalf("Reclamation Sage trigger not on the stack")
	}
	if len(item.Targets) != 1 || item.Targets[0].ID != firstID {
		t.Fatalf("trigger targets = %+v, want the first artifact %s", item.Targets, firstID)
	}

	// "In response": the targeted artifact leaves the game entirely
	// (the existence check only fails when no tracked zone holds
	// the card — exile alone would still count as existing).
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == firstID {
				g.Battlefield.Cards = append(g.Battlefield.Cards[:i], g.Battlefield.Cards[i+1:]...)
				break
			}
		}
	})
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(secondID) {
		t.Errorf("fizzled trigger destroyed a different artifact — targets must not be re-picked at resolution")
	}
	var fizzled bool
	for _, ev := range g.Events {
		if ev.Kind == game.EventFizzle && ev.Source == sageID {
			fizzled = true
		}
	}
	if !fizzled {
		t.Errorf("no EventFizzle for the Reclamation Sage trigger whose target left")
	}
}

// TestDiesTriggersFromOneWrathStackAPNAP: two dies-trigger creatures
// under different controllers die to one Damnation. Both triggers
// reach the stack (APNAP: the active player's first, so the
// non-active player's resolves first) and both effects land only
// after resolution.
func TestDiesTriggersFromOneWrathStackAPNAP(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[0]
	other := g.Seats[2]
	travelerID := pushDiesCreatureForTest(g, active.ID, "Doomed Traveler",
		"a30907c0-fbde-4fd3-a8c7-f304305fcea7", "Creature — Human Soldier", 1, 1)
	familiarID := pushDiesCreatureForTest(g, other.ID, "Filigree Familiar",
		"b544f690-e4bf-4a5b-984d-9256518fd574", "Artifact Creature — Fox", 2, 2)
	otherHandBefore := other.Hand.Size()

	castCatalogSpell(t, g, "Damnation", "Sorcery",
		"d57a8f0b-7989-4db5-8756-6f2690097252", nil)
	for i := 0; i < 8 && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}

	travelerTrig := triggerOnStack(g, travelerID)
	familiarTrig := triggerOnStack(g, familiarID)
	if travelerTrig == nil || familiarTrig == nil {
		t.Fatalf("both dies triggers should be on the stack: traveler=%v familiar=%v", travelerTrig != nil, familiarTrig != nil)
	}
	// APNAP: active player's trigger is placed first (lower Seq) and
	// therefore resolves last.
	if !(travelerTrig.Seq < familiarTrig.Seq) {
		t.Errorf("APNAP order: active player's trigger Seq %d should be below the non-active player's %d",
			travelerTrig.Seq, familiarTrig.Seq)
	}
	if other.Hand.Size() != otherHandBefore {
		t.Fatalf("Filigree Familiar drew before its trigger resolved")
	}
	passPriorityAroundTable(t, g)

	if got := other.Hand.Size() - otherHandBefore; got != 1 {
		t.Errorf("Filigree Familiar dies-draw: hand delta %d, want 1", got)
	}
	spirits := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Spirit" && c.Controller == active.ID {
			spirits++
		}
	}
	if spirits != 1 {
		t.Errorf("Doomed Traveler dies: got %d Spirit tokens, want 1", spirits)
	}
}
