package game

// resolution_pay_unless.go — #5793: a pay-unless asked while a spell or
// ability resolves is part of that resolution, so the table waits for
// the answer.
//
// THE RULE. CR 118.12: in "[do something]. If [a player] doesn't,
// [effect]" (and CR 118.12a's "[do something] unless [a player does
// something else]", which means the same), the action "is a cost, paid
// when the spell or ability resolves". CR 608.2g lets the payer activate
// mana abilities for it in the middle of that resolution, and CR 117.3b
// gives the active player priority only after the spell or ability has
// resolved. So no player may act, and the step may not end, while the
// payer is deciding. The rest of the resolution (CR 608.2c's later
// instructions, CR 608.2n's exit) comes after the answer.
//
// THE BUG IT CLOSES. The pay_unless KIND does not stop the table (ADR
// 0018 §6, choice_gate.go), and before this change only two narrowings
// made a prompt of it block: GuardsStackItem (#951) and an OwedInStep the
// asking card set itself (#997: the upkeep pay-or-else, a MayPay in this
// step, an any-player offer). Every other pay-unless — Crystal Shard's
// bounce, Arcum's Whistle's forced attack, Combustion Man's destroy —
// let the resolution close before the payer answered. The spell's later
// instructions ran first, any seat could pass, the step could end, and
// the "unless" consequence landed whenever the payer got round to it.
//
// THE FIX, derived and not declared. A pay_unless queued while a
// resolution is open and stamped as part of it (midResolution,
// resolution_pause.go) is owed in the step it was asked in. That is
// #997's anchor, so the existing gate does the rest: the prompt blocks
// (Game.ChoicePromptBlocksTable), a blocking stamped prompt holds the
// resolution open (pausedResolutionChoiceLocked), and the answer's own
// runStateChecksLocked resumes it. Nothing is new on the snapshot:
// OwedInStep is already written to it. A card says what its text says
// and the engine works out the halt, as #951 and #997 decided; there is
// no card-side switch to forget.
//
// WHY IT CANNOT WEDGE A TABLE. The halt is #997's live comparison
// against the cursor, so it cannot outlive its step; every pay_unless
// has an always-legal answer (decline); and a seat that leaves has its
// prompts dropped or reassigned by the departure table. A prompt asked
// outside a resolution (an answer branch after the resolution closed,
// or a test queuing one by hand) is not stamped and keeps the kind's
// classification.
//
// ADR 0127. The derived anchor is stamped AFTER the prompt is keyed for
// a standing answer (QueueChoiceForEffect), so "Always / Never pay"
// still answers Rhystic Study, Smothering Tithe and the rest: their halt
// comes from being asked mid-resolution, not from words on the card
// that make the answer depend on the board. A step anchor the card sets
// itself (Stasis, cumulative upkeep, echo, Hellkite Charger) is still
// never covered. See ADR 0018's 2026-11-20 amendment and ADR 0127's.

// oweResolutionPayUnlessInStepLocked anchors a pay_unless a resolution
// asks to the step it is asked in (#5793). A prompt that is not a
// pay_unless, was not asked as part of a resolution, or already names
// its step is left as it is. Caller must hold g.mu.
func (g *Game) oweResolutionPayUnlessInStepLocked(c *PendingChoice) {
	if c == nil || c.Kind != PendingChoicePayUnless || !c.midResolution {
		return
	}
	if c.OwedInStep != (TurnStep{}) {
		return
	}
	c.OwedInStep = g.currentTurnStepLocked()
}
