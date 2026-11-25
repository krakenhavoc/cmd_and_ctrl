package effects

import "github.com/google/uuid"

// pile_guess_tally.go — the card side of the turn's pile and guess
// tally (#6175, game/pile_guess_tally.go, CR 700.3).
//
// EVERY effect that makes a player guess, or that groups cards or
// permanents into a pile, calls NotePileOrGuess as it asks. The shared
// pile helpers (PileSplit, the face-down pile family) do it for the
// cards that go through them; a card that builds its piles or its guess
// by hand does it itself, once, before the first prompt. The audit test
// in the effects test tree holds every catalog card whose oracle text
// says "pile" or "guess" to one of those two, so a new one cannot be
// added without a reviewer deciding which it is.
//
// The unit is the SPELL OR ABILITY, not the pile: Hostile Negotiations
// makes two piles and counts once. So the note is keyed by the resolving
// stack item, and a second note of the same item changes nothing.

// NotePileOrGuess records the resolving spell or ability as one that
// caused a player to guess or to group objects into a pile this turn.
// Called with g.mu held, from the resolution path.
//
// The item is the resolving one; a context rebuilt from a prompt's
// asked item (AskedContext) carries that same item, so a continuation
// that does the asking records the spell that started it. With no item
// at all (an as-enters callback) the source permanent stands in, so the
// effect is still counted once.
func NotePileOrGuess(ctx *Context) {
	if ctx == nil || ctx.Game == nil {
		return
	}
	id := uuid.Nil
	if ctx.Item != nil {
		id = ctx.Item.ID
	}
	if id == uuid.Nil {
		id = ctx.Source()
	}
	ctx.Game.NotePileOrGuessForEffect(id)
}
