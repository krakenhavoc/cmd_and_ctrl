package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blatant Thievery — Sorcery {4}{U}{U}{U}:
//
//	"For each opponent, gain control of target permanent that player
//	 controls."
//
// One REQUIRED target per opponent (CR 601.2c), each bound to its
// player: perOpponentRequiredClauses builds a clause for each seat after
// the caster's, and the engine reads whether a clause is required from
// the table (game/opponent_slot.go, #4128). At a two-player table only
// the first applies; at a four-player table a spell with one opponent
// out of the game needs two picks, not three; and an opponent who
// controls nothing targetable makes the spell uncastable, as printed.
// The pick is bound to its player at resolution too (CR 608.2b), so a
// permanent that changed hands in response is an illegal target even
// when it went to another opponent.
//
// "Target permanent", so a land, an artifact, an enchantment or a
// planeswalker is as good as a creature. The control change has no
// stated duration (CR 611.2a), the Agent of Treachery shape: it lasts
// until the game ends or the permanent stops being the object it was.
// Each pick resolves on its own, so a target that left (or became
// illegal) is skipped and the rest are still taken.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8bb51dd2-cc62-48ea-ae8e-52351073a8c9",
		Name:         "Blatant Thievery",
		Completeness: CompletenessFull,
		Targets:      perOpponentRequiredClauses("permanent"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (GainControl{
					Target:     t.ID,
					Controller: item.Controller,
					Duration:   game.IndefiniteDuration(),
					Label:      "Blatant Thievery — gain control (no stated duration)",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
