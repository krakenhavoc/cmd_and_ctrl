package effectstest

import (
	"errors"
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
func NewCatalogGame(t *testing.T) *game.Game {
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
	// #692 / CR 103.8c: at four seats the starting player DOES draw on
	// turn 1 — only a two-player game skips it (CR 103.8a). Walk the
	// cursor onto that draw step here so every test's baseline hand
	// size is taken after the turn-based draw. Otherwise the first
	// AdvanceStep inside a test body silently adds a card to seat 0's
	// hand and every "hand +1" assertion in this package is off by one
	// for reasons that have nothing to do with the card under test.
	//
	// The draw step is the EARLIEST cursor position where that is true,
	// which is why the harness stops here rather than walking on to the
	// main phase: upkeep-step and "beginning of precombat main" triggers
	// stay observable, and sorcery-speed restrictions are still in force.
	// The one trap it leaves: g.DrawCard(seat 0) is a deliberate no-op
	// during the active player's own draw step (see Game.DrawCard), so a
	// test that wants a manual draw for the active seat must walk the
	// cursor off this step first.
	for i := 0; g.Turn.Step != game.StepDraw; i++ {
		if i >= 8 {
			t.Fatalf("never reached the turn-1 draw step (stuck at %v)", g.Turn.Step)
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep toward the turn-1 draw step: %v", err)
		}
	}
	return g
}

// castCatalogSpell seeds a card with the given name, type line, and
// oracle ID into the active seat's hand, advances to the precombat
// main phase, then calls CastSpell with the supplied targets. Pass
// priority around afterwards to resolve.
func CastCatalogSpell(t *testing.T, g *game.Game, name, typeLine, oracleID string, targets []game.TargetRef) uuid.UUID {
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

// castCatalogSpellErr is castCatalogSpell for the casts that are
// SUPPOSED to be refused: same setup, but the announce error comes
// back instead of failing the test. The #764 per-clause gate is the
// first thing that needed it — "this pair fits the wrong slots" is
// now an announce refusal rather than a resolution no-op, and that
// is the assertion.
func CastCatalogSpellErr(t *testing.T, g *game.Game, name, typeLine, oracleID string, targets []game.TargetRef) error {
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
	return g.CastSpell(active.ID, id, game.CastSpellParams{Targets: targets})
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
// it (answerLatestTriggerPrompt) then call this again. Since #730
// the engine enforces that stop rather than trusting the helper to
// observe it, so the refusal is the signal to return.
func PassPriorityAroundTable(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 32; i++ {
		if StackFullyEmpty(g) {
			return
		}
		// #1501: nobody can pass while priority is parked for a block
		// declaration; the defenders finish first, as a table would.
		FinishBlockDeclarations(t, g)
		if StackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			if errors.Is(err, game.ErrChoicePending) {
				return
			}
			t.Fatalf("PassPriority iter %d: %v", i, err)
		}
	}
	t.Fatalf("stack did not empty after 32 priority passes")
}

// lockInBlocks completes the block declaration the way play does:
// priority passes around the table, and the defender's pass completes
// their declaration (#1279; the wrap completes anyone left), which is
// the lock-in
// (CR 509.1 / 509.2a, #830) — the point where the engine announces
// the final assignment and harvests the "becomes blocked" / "blocks"
// triggers off it. Nothing is announced by DeclareBlocker itself, so
// a test that asserts on those triggers has to come through here.
//
// Stops as soon as the lock-in produced something (an item on the
// stack or a trigger prompt) or the step advanced — a declaration
// with no triggers wraps on through to combat damage, exactly as an
// unblocked combat does.
func LockInBlocks(t *testing.T, g *game.Game) {
	t.Helper()
	LockInCombatDeclaration(t, g)
}

// lockInAttacks is lockInBlocks for the ATTACK declaration (#859,
// CR 508.1 / 508.2): DeclareAttacker stages the attack and announces
// nothing either, so a test that asserts on "whenever ~ attacks"
// triggers has to reach the lock-in — the first priority boundary of
// the declare-attackers step — the way play does.
func LockInAttacks(t *testing.T, g *game.Game) {
	t.Helper()
	LockInCombatDeclaration(t, g)
}

// lockInCombatDeclaration is the loop both of the above are: priority
// passes around the table until the wrap locks the step's staged
// declaration in. One implementation because it is one boundary —
// PassPriority commits whichever declaration is staged.
//
// #1501: in a declare-blockers step parked for the block declaration
// nobody holds priority, so every defender still declaring finishes
// first (finish_blocks) — that IS the lock-in — and the passes carry
// on from the active player, who receives priority once the last
// declaration is complete.
func LockInCombatDeclaration(t *testing.T, g *game.Game) {
	t.Helper()
	step := g.Turn.Step
	if FinishBlockDeclarations(t, g) && !StackFullyEmpty(g) {
		return
	}
	for i := 0; i < len(g.Seats)+1; i++ {
		if err := g.PassPriority(); err != nil {
			if errors.Is(err, game.ErrChoicePending) {
				return
			}
			t.Fatalf("PassPriority iter %d: %v", i, err)
		}
		if !StackFullyEmpty(g) || g.Turn.Step != step {
			return
		}
	}
}

// finishBlockDeclarations sends finish_blocks for every defender still
// declaring blockers while priority is parked for them (#1501), as a
// table would click "Done blocking", and reports whether it finished
// any. A no-op in every other state.
func FinishBlockDeclarations(t *testing.T, g *game.Game) bool {
	t.Helper()
	if g.Turn.Step != game.StepDeclareBlockers || g.Turn.PriorityHolder != game.NoPriority {
		return false
	}
	finished := false
	for _, s := range g.Seats {
		if s == nil || g.BlockDeclarationStatusOf(s.ID) != game.BlockDeclarationPending {
			continue
		}
		if err := g.FinishBlocks(s.ID); err != nil {
			if errors.Is(err, game.ErrChoicePending) {
				return finished
			}
			t.Fatalf("FinishBlocks for %s: %v", s.Name, err)
		}
		finished = true
	}
	return finished
}

// stackFullyEmpty reports whether nothing is on or headed for the
// stack: Game.Stack, StackMeta, and PendingTriggers are all empty.
func StackFullyEmpty(g *game.Game) bool {
	return g.Stack.Size() == 0 && len(g.StackMeta) == 0 && len(g.PendingTriggers) == 0
}

// triggerOnStack returns the triggered-ability item in StackMeta
// sourced from cardID, or nil. Tests use it to assert a trigger is
// actually waiting on the stack (with its label and targets)
// before priority passes resolve it.
func TriggerOnStack(g *game.Game, cardID uuid.UUID) *game.StackItem {
	for _, item := range g.StackMeta {
		if item != nil && item.Kind == game.StackItemTriggered && item.SourceCardID == cardID {
			return item
		}
	}
	return nil
}

func PushCreatureToBattlefieldForTest(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
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

// TestNonCatalogSpellStaysSandbox proves a spell without a
// catalog-matching OracleID resolves to graveyard with no effect —
// the Cockatrice-style manual fallback is intact. Canary for
// future regressions that might accidentally activate auto-resolve
// on non-catalog cards.
// --- Tutors + recursion + ETB (sub-PR 6) -----------------------

// pushLibraryCardForTest pushes a specific Card onto the BOTTOM of
// a player's library, and returns its instance ID so the caller can
// name it later.
//
// Placement no longer decides which card a search takes — since S22
// the searcher does, via a PendingChoiceSearchLibrary prompt (see
// search_chooser_test.go's answerSearchByID). Bottom placement is
// kept only because it keeps the seeded card out of the way of draw
// steps.
func PushLibraryCardForTest(p *game.Player, c game.Card) uuid.UUID {
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
// pile. Both Regrowth and Eternal Witness take a real target now
// (S20 for the Witness, #338 for Regrowth), so push order no longer
// decides which card comes back — the test names the one it wants.
func PushGraveyardCardForTest(p *game.Player, name string) uuid.UUID {
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

// answerLatestTriggerPrompt answers the most-recently-queued
// PendingChoiceTriggerPrompt addressed to chooserID with `apply`. An
// optional trigger asks its "you may" as it resolves (CR 603.5, #2494),
// so when no such prompt is open yet this passes priority until the
// trigger that asks it resolves (awaitTriggerPrompt). Fails the test
// loudly when none comes — a missing prompt usually means the harvester
// didn't fire (oracle ID typo, AppliesTo predicate wrong, or the
// trigger declared without an OptionalPrompt). Added in S19 sub-PR 3.
//
// A targeted or modal optional trigger chooses its targets and modes as
// it goes on the stack (CR 603.3c/d), before the question. A test that
// says "yes" picks them first; one that says "no" may leave them to
// this helper, which answers them with the first legal choice, since a
// declined trigger does nothing whatever it pointed at.
func AnswerLatestTriggerPrompt(t *testing.T, g *game.Game, chooserID uuid.UUID, apply bool) {
	t.Helper()
	c := awaitTriggerPromptAnswering(t, g, chooserID, !apply)
	if err := g.ResolveTriggerPrompt(c.ID, chooserID, apply); err != nil {
		t.Fatalf("ResolveTriggerPrompt: %v", err)
	}
}

// awaitTriggerPrompt returns the latest trigger_prompt addressed to
// chooserID, passing priority until the optional trigger that asks it
// resolves when none is open yet (CR 603.5, #2494). Another open
// prompt stops the walk: the test has to answer it first.
func AwaitTriggerPrompt(t *testing.T, g *game.Game, chooserID uuid.UUID) *game.PendingChoice {
	t.Helper()
	return awaitTriggerPromptAnswering(t, g, chooserID, false)
}

// awaitTriggerPromptAnswering is awaitTriggerPrompt that, with
// pickAny, answers the chooser's open target or mode pick with the
// first legal choice on the way.
func awaitTriggerPromptAnswering(t *testing.T, g *game.Game, chooserID uuid.UUID, pickAny bool) *game.PendingChoice {
	t.Helper()
	for i := 0; i < 32; i++ {
		if c := LatestTriggerPrompt(g, chooserID); c != nil {
			return c
		}
		if p := LatestPickTarget(g, chooserID); p != nil {
			if !pickAny {
				t.Fatalf("a target pick is open for %s: a trigger's targets are chosen as it goes on the stack (CR 603.3d), so pick them before answering its \"you may\" (CR 603.5)", chooserID)
			}
			var refs []game.TargetRef
			for _, id := range p.PickTargetCards {
				if len(refs) < p.PickTargetMin {
					refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: id})
				}
			}
			for _, id := range p.PickTargetPlayers {
				if len(refs) < p.PickTargetMin {
					refs = append(refs, game.TargetRef{Kind: game.TargetPlayer, ID: id})
				}
			}
			if err := g.ResolvePickTargets(p.ID, chooserID, refs); err != nil {
				t.Fatalf("ResolvePickTargets (any, to decline): %v", err)
			}
			continue
		}
		if m := LatestChoiceOfKind(g, game.PendingChoiceModePick); m != nil && m.Chooser == chooserID && pickAny {
			n := m.ModeMin
			if n < 1 {
				n = 1
			}
			if err := g.ResolveModePick(m.ID, chooserID, append([]int(nil), m.ModeOptionIndex[:n]...)); err != nil {
				t.Fatalf("ResolveModePick (any, to decline): %v", err)
			}
			continue
		}
		if StackFullyEmpty(g) {
			break
		}
		FinishBlockDeclarations(t, g)
		if len(g.PendingTriggers) > 0 {
			// A trigger harvested under the test's own write lock waits
			// for the boundary a table action would have run; run it, so
			// the trigger goes on the stack in this step (CR 603.3).
			g.RunStateChecksForTest()
		}
		if c := LatestTriggerPrompt(g, chooserID); c != nil {
			return c
		}
		if err := g.PassPriority(); err != nil {
			if errors.Is(err, game.ErrChoicePending) {
				break
			}
			t.Fatalf("PassPriority iter %d: %v", i, err)
		}
	}
	t.Fatalf("no PendingChoiceTriggerPrompt addressed to %s", chooserID)
	return nil
}

// owedMays counts the CR 603.5 "you may"s owed by chooser: open
// trigger_prompts, plus optional triggers queued or on the stack that
// will ask chooser as they resolve (#2494). A test that counted
// prompts right after the event counts the same thing through this.
func OwedMays(g *game.Game, chooser uuid.UUID) int {
	n := 0
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt && c.Chooser == chooser {
			n++
		}
	}
	items := append([]*game.StackItem(nil), g.PendingTriggers...)
	for _, it := range g.StackMeta {
		items = append(items, it)
	}
	for _, it := range items {
		if it == nil || it.Optional == nil {
			continue
		}
		who := it.Optional.Chooser
		if who == uuid.Nil {
			who = it.Controller
		}
		if who == chooser {
			n++
		}
	}
	return n
}

// latestTriggerPrompt returns the most-recently-queued
// PendingChoiceTriggerPrompt addressed to chooserID, or nil. It does not
// pass priority: an optional trigger's prompt is open only while the
// trigger resolves (awaitTriggerPrompt walks there).
func LatestTriggerPrompt(g *game.Game, chooserID uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt && c.Chooser == chooserID {
			return c
		}
	}
	return nil
}

// --- S15 sub-PR 2 mana abilities ------------------------------

// pushArtifactToBattlefieldForTest puts a specific instance on the
// shared battlefield with the named oracle_id so the catalog lookup
// fires. Distinct from pushCreatureToBattlefieldForTest so we can
// use a non-creature TypeLine (Sol Ring / Arcane Signet / Birds of
// Paradise all have different shapes).
func PushArtifactToBattlefieldForTest(g *game.Game, owner uuid.UUID, name, typeLine, oracleID string) uuid.UUID {
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

// --- S19 sub-PR 4: dies-triggers (LTB) -------------------------

// pushDiesCreatureForTest seeds a catalog creature onto the
// battlefield with its real OracleID + type line so the LTB
// harvester can match its dies-trigger. Returns the instance ID.
func PushDiesCreatureForTest(g *game.Game, owner uuid.UUID, name, oracleID, typeLine string, power, toughness int) uuid.UUID {
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

// --- S19 sub-PR 5: upkeep triggers (begin_upkeep) --------------

// pushPermanentForTest seeds a non-creature catalog permanent (an
// enchantment, here) onto the battlefield with its real OracleID so
// the harvester can match its upkeep trigger.
func PushPermanentForTest(g *game.Game, owner uuid.UUID, name, oracleID, typeLine string) uuid.UUID {
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
func FindBattlefieldByName(g *game.Game, name string) uuid.UUID {
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
func AdvanceToUpkeepOf(t *testing.T, g *game.Game, seat int) {
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

// --- b5-instants: Counterflux, Devious Cover-Up, Discombobulate,
// Blue Sun's Zenith, Finale of Revelation --------------------------

const (
	CounterfluxOracle        = "8c983da8-1436-4c06-af9c-91b72cab48c1"
	DeviousCoverUpOracle     = "feb221fb-59bf-4671-a53f-1bbe8e9c2ca9"
	DiscombobulateOracle     = "b58d7a20-4bcd-4c33-8cda-955362525f48"
	BlueSunsZenithOracle     = "613a41b8-0b4f-4995-bf1e-ca41f96e6438"
	FinaleOfRevelationOracle = "755bd5d8-67f1-4f24-a4e8-d98edf2f2e03"
	B5ShockOracle            = "a9d288b8-cdc1-4e55-a0c9-d6edfc95e65d"
)
