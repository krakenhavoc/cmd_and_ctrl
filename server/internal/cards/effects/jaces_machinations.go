package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jace's Machinations — Instant {2}{U} (Reality Fracture, tracker #2795):
//
//	"Until end of turn, you may activate loyalty abilities of Jace
//	 planeswalkers you control on any player's turn any time you could
//	 cast an instant.
//	 Empower Jace 8."
//
// The window is GrantLoyaltyAbilitiesAtInstantSpeed{Subtype: "Jace"} (ADR
// 0066's 2026-10-09 amendment), stored on the player until end of turn.
// It opens the window only: CR 606.3's one activation per permanent per
// turn still holds. The Jace token the keyword action creates is a Jace
// planeswalker, so it is covered, and the grant is applied first so the
// token empowered in the same resolution already is.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "89bd056d-8f5b-4d0e-b80a-58a984e21100",
		Name:         "Jace's Machinations",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (GrantLoyaltyAbilitiesAtInstantSpeed{
				Subtype: frJaceSubtype,
				Label:   "Jace's Machinations — loyalty abilities of Jace planeswalkers at instant speed",
			}).Apply(ctx); err != nil {
				return err
			}
			return EmpowerJace{N: 8}.Apply(ctx)
		},
	})
}
