package game

import "github.com/google/uuid"

// draw_instruction.go — replacements that rewrite the SIZE of a whole
// "draw N cards" instruction (Quantum Riddler: "if you would draw one or
// more cards, you draw that many cards plus one instead").
//
// CR 121.2 makes the instruction N single draws, and every other draw
// replacement works on a single draw. The ruling on Quantum Riddler says
// its effect applies to the instruction first, before the replacements on
// single draws: a Thought Reflection beside it turns "draw two" into
// three draws, each doubled. So the instruction is asked once, here, and
// the draws that follow open their own windows as ever. A resumed
// instruction (a draw paused on a prompt) was already asked and is not
// asked again (once applied, a replacement can't apply again to the event
// it modified, CR 616.1f).

// drawInstructionSizeLocked returns n plus the extra cards every active
// DrawInstructionExtra effect adds to an instruction of n >= 1 draws by
// playerID. Each effect applies once; its AppliesTo reads the board as it
// stands before any card is drawn.
//
// Caller must hold g.mu.
func (g *Game) drawInstructionSizeLocked(playerID uuid.UUID, n int) int {
	if n <= 0 {
		return n
	}
	ev := &ReplacementEvent{
		Kind: RepEventDraw, Actor: playerID, DrawPlayer: playerID,
		DrawCount: n, DrawInstruction: true,
	}
	for _, a := range g.gatherActiveReplacementsReadOnlyLocked(ev) {
		if a.effect.DrawInstructionExtra > 0 {
			n += a.effect.DrawInstructionExtra
		}
	}
	return n
}
