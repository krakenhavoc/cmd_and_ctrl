package effects

import "github.com/google/uuid"

// s82_old_b421_helpers.go — S82 OLD batch 421 (Return to Ravnica, #5566).

// OldB421RevealUntilLandAndMill is "[player] reveals cards from the top of their
// library until they reveal a land card, then puts those cards into their
// graveyard" (Destroy the Evidence): the reveal, then the whole run.
func OldB421RevealUntilLandAndMill(ctx *Context, player uuid.UUID) error {
	return b14MillUntilLand(ctx, player)
}
