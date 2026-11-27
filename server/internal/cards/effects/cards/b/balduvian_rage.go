package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Balduvian Rage — Instant {X}{R} (Coldsnap):
//
//	"Target attacking creature gets +X/+0 until end of turn.
//	 Draw a card at the beginning of the next turn's upkeep."
//
// X is announced with the spell (CR 107.3). The draw is a delayed trigger set
// up as the spell resolves (CR 603.7), and it still happens when the target is
// gone as long as the spell resolved at all (CR 608.2b).
//
// No simplification.
func init() {
	purpose := effects.ForTargets(effects.PowerXUntilEOT(0))
	purpose.Draws = 1
	effects.Register(effects.Spec{
		OracleID:     "1bc7d73b-0f54-4807-b13a-3def0c5e5616",
		Name:         "Balduvian Rage",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Purpose:      purpose,
		Targets:      effects.TargetCreature("target attacking creature", effects.AttackingCreature()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if id, ok := effects.B16FirstLegalTargetCard(ctx); ok {
				if err := (effects.BoostUntilEOT{Target: id, Power: ctx.X(), Label: "Balduvian Rage — +X/+0"}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.DrawACardAtTheNextUpkeep("Balduvian Rage — draw a card").Apply(ctx)
		},
	})
}
