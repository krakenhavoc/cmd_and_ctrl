package game

import "github.com/google/uuid"

// upkeep_pay_unless.go — "at the beginning of your upkeep, pay <cost>
// or <consequence>", and the rule #997 is about: while that question
// is unanswered, the table does not leave the step that asked it.
//
// THE BUG THIS FILE IS. `pay_unless` is the single kind ADR 0018 §6
// lets the table walk past, and the reason it gives is entirely
// Rhystic Study's: the trigger has resolved and left the stack, the
// question is addressed to a DIFFERENT player, and neither answer
// reads the step. Stasis ("sacrifice Stasis unless you pay {U}"),
// Pact of Negation ("pay {3}{U}{U}. If you don't, you lose the game")
// and every cumulative upkeep borrow the same prompt, and none of
// that survives the move. The question is the ACTIVE player's own,
// asked during their OWN upkeep, and what hangs on the answer is
// whether a permanent is on the battlefield for the rest of the turn
// — or whether the player is still in the game. A table that takes
// the turn while the prompt is open has walked past a decision its
// own turn structure depends on: CR 117.3 does not let priority pass
// on with a required action outstanding, and CR 500.4 does not let
// the step end until it has been taken.
//
// WHY IT IS NOT A FLAG. #567 answered the same question for
// cumulative upkeep with a per-card boolean (`PayUnless.Blocking`,
// riding `PendingChoice.ForceBlocks`), and the evidence against that
// shape is this issue: Stasis and Pact of Negation shipped after it,
// with the same sentence printed on them, and neither author found
// the switch. #951 had already drawn the conclusion for the other
// narrowing — a card says what its text says and the ENGINE works
// out what that means for the cursor — so this is the same move for
// the second reason a pay-unless blocks, and `ForceBlocks` is gone.
//
// So the prompt records WHEN IT IS OWED (PendingChoice.OwedInStep,
// read off the cursor at queue time) and the gate works the rest out
// for itself (Game.ChoicePromptBlocksTable, choice_gate.go). The
// anchor is the step the prompt was RAISED in rather than the literal
// upkeep, which costs nothing and is the honest rule: an end-step
// pay-or-else is owed before the end step ends for the same reason.
//
// THE HALT CANNOT OUTLIVE ITS STEP, which is what keeps it from being
// a wedge — the same property #951's live stack read buys. The block
// is a live comparison against the cursor, so a game that reaches
// another step some other way (an undo to before the trigger, a
// restore, an admin walking the cursor by hand) takes the halt with
// it and the table plays on. Nothing can be stuck behind a question
// about a step the game is no longer in.
//
// SCOPE. This is "pay or else, asked of the payer, about the step
// they are standing in". A pay-unless whose consequence is a draw
// (Rhystic Study), a Treasure (Smothering Tithe) or a token (Kazuul)
// is asked of somebody else about nothing the cursor cares about,
// keeps ADR 0018 §6's latitude, and plays exactly as it did.

// UpkeepPayUnlessPrompt is the whole of "at the beginning of your
// upkeep, pay <Cost> unless <consequence>".
//
// One struct rather than five positional arguments, matching the
// engine's other prompt specs (CounterUnlessPaidPrompt,
// ChooseCardsPrompt), and ONE DOOR rather than one per card: every
// card with this text goes through QueueUpkeepPayUnlessForEffect so
// the halt cannot be forgotten by the next one — which is exactly how
// Stasis and Pact of Negation were wrong (#997).
type UpkeepPayUnlessPrompt struct {
	// Chooser is the player asked to pay — the permanent's
	// controller on every printed card of this family, asked during
	// their own upkeep.
	Chooser uuid.UUID

	// Source attributes the prompt to a card for the client: the
	// permanent being sacrificed, the Pact that is owed.
	Source uuid.UUID

	// Cost is the printed payment ("{U}", "{3}{U}{U}", a cumulative
	// upkeep's cost repeated once per age counter), parsed by
	// ParseCost and auto-tapped for like any other pay-unless.
	Cost string

	// Question is the dialog header.
	Question string

	// OnDecline is the "or else": sacrifice the permanent, lose the
	// game. It runs on "no", and on a "yes" the chooser cannot fund,
	// exactly as it does for an ordinary pay-unless.
	OnDecline func(g *Game) error
}

// QueueUpkeepPayUnlessForEffect queues the CR 118.12 prompt and
// anchors it to the step the game is standing in, so the table cannot
// leave that step until it is answered (CR 117.3, CR 500.4).
//
// The ability that asked has finished resolving; what is left is the
// payer's decision, which arrives as an ordinary resolve_choice. A
// chooser who is no longer seated is not asked and the consequence
// runs immediately — the pay-unless contract, unchanged.
//
// Caller must hold g.mu. Added in #997.
func (g *Game) QueueUpkeepPayUnlessForEffect(p UpkeepPayUnlessPrompt) error {
	return g.queuePayUnlessLocked(p.Chooser, p.Source, p.Cost, p.Question,
		p.OnDecline, g.currentTurnStepLocked(), uuid.Nil, nil)
}

// currentTurnStepLocked freezes the cursor as an anchor a later read
// can compare itself against. Caller must hold g.mu.
func (g *Game) currentTurnStepLocked() TurnStep {
	return TurnStep{Turn: g.Turn.Seq, Step: g.Turn.Step}
}

// choiceOwedBeforeTheStepEnds reports whether `c` is a prompt the
// table owes before it may leave the step that asked it — the second
// live half of Game.ChoicePromptBlocksTable (choice_gate.go), beside
// choiceGuardsALiveStackItem.
//
// Live, not a snapshot of the answer taken when the prompt was
// queued: the cursor is compared every time the gate is asked, so a
// game that has reached another step some other way is not held by a
// question about a step it has left. A halt that could only be lifted
// by answering is how a prompt nobody can answer wedges a table; this
// one cannot outlive what it is about.
func (g *Game) choiceOwedBeforeTheStepEnds(c *PendingChoice) bool {
	if g == nil || c == nil {
		return false
	}
	return c.OwedInStep.IsCurrent(g.Turn)
}
