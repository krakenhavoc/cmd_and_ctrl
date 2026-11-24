package game

import "github.com/google/uuid"

// draw_instruction_rewrite.go — "If an opponent would draw two or more
// cards, instead you and that player each draw a card" (Alms Collector,
// #6089).
//
// Like Quantum Riddler's "that many plus one" (draw_instruction.go) this is
// a replacement of the WHOLE draw instruction, not of a single draw: CR
// 121.2 makes "draw two cards" two individual draws, and no single one of
// them can see "two or more". So it is asked once, on the same synthetic
// event (ReplacementEvent.DrawInstruction), before any card is drawn. When
// it applies, the instruction becomes one draw for the drawing player and
// one draw for the effect's controller (CR 614.1a, "instead").
//
// The two draws are performed one at a time in APNAP order (CR 101.4),
// each through its own CR 614 window, by the machinery an amount
// replacement already uses for its descendants (ReplacementEvent.drawPending,
// draw_instead.go): a draw that pauses on a prompt (a dredge offer) holds
// the other behind it, and the instruction's "then" (DrawNThenForEffect)
// runs after both.
//
// The drawing player is the affected player and would order the
// instruction-level replacements (CR 616.1). Two are possible: this rewrite
// and DrawInstructionExtra. Applying the rewrite first never leaves the
// drawing player with fewer cards (a Riddler then adds its one to the one
// left), and once the instruction is down to one draw no other rewrite
// applies (CR 616.1e asks again, and "two or more" is false), so the engine
// makes that order. Of several rewrites that apply to the same instruction,
// the first gathered (oldest on the battlefield) is applied.

// drawInstructionRewriteLocked finds the first active
// DrawInstructionEachDrawOne effect that applies to an instruction of n
// draws by playerID and returns the player it gives the second draw to (its
// controller), or uuid.Nil when none applies.
//
// Caller must hold g.mu.
func (g *Game) drawInstructionRewriteLocked(playerID uuid.UUID, n int) uuid.UUID {
	if n < 2 {
		return uuid.Nil
	}
	ev := &ReplacementEvent{
		Kind: RepEventDraw, Actor: playerID, DrawPlayer: playerID,
		DrawCount: n, DrawInstruction: true,
	}
	for _, a := range g.gatherActiveReplacementsReadOnlyLocked(ev) {
		if !a.effect.DrawInstructionEachDrawOne {
			continue
		}
		var who uuid.UUID
		switch {
		case a.source != nil:
			who = a.source.Controller
		case a.effect.Controller != nil:
			who = a.effect.Controller(ev, g, nil)
		}
		if who != uuid.Nil && who != playerID && g.playerInGameLocked(who) {
			return who
		}
	}
	return uuid.Nil
}

// drawInstructionRunLocked performs a "draw n cards" instruction by
// playerID: the instruction-level replacements first (the rewrite above,
// then DrawInstructionExtra), then the individual draws, then `then`.
//
// Caller must hold g.mu.
func (g *Game) drawInstructionRunLocked(playerID uuid.UUID, n int, then DrawThen) error {
	other := g.drawInstructionRewriteLocked(playerID, n)
	if other == uuid.Nil {
		return g.drawRunLocked(playerID, g.drawInstructionSizeLocked(playerID, n), then)
	}
	counts := map[uuid.UUID]int{
		playerID: g.drawInstructionSizeLocked(playerID, 1),
		other:    g.drawInstructionSizeLocked(other, 1),
	}
	var draws []pendingDraw
	for _, seat := range g.apnapSeatsExceptLocked(uuid.Nil) {
		for i := 0; i < counts[seat]; i++ {
			draws = append(draws, pendingDraw{player: seat, count: 1})
		}
	}
	if len(draws) == 0 {
		return g.runDrawThenLocked(then)
	}
	ev := &ReplacementEvent{
		Kind: RepEventDraw, Actor: playerID, DrawPlayer: draws[0].player,
		DrawCount: 1, drawPending: draws[1:], drawThen: then,
	}
	_, err := g.drawEventLocked(ev)
	return err
}
