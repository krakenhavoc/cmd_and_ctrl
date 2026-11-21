package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Demonic Dread — Sorcery {1}{B}{R} (Alara Reborn, #5586):
//
//	"Cascade (When you cast this spell, exile cards from the top of your
//	 library until you exile a nonland card that costs less. You may cast
//	 it without paying its mana cost. Put the exiled cards on the bottom in
//	 a random order.)
//	 Target creature can't block this turn."
//
// Cascade is the catalog keyword (CR 702.85a). The restriction is Duel
// Tactics' can't-block for the turn, without its damage; a target that
// left in response is skipped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b090f1e5-1f7e-4492-9928-c5911a8ef95b",
		Name:         "Demonic Dread",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose: effects.ForTargets(game.TargetPurpose{
			Slot: 0, UntilEndOfTurn: game.TargetHarm, Pump: &game.Pump{CantBlock: true},
		}),
		Triggered: []game.TriggeredAbility{effects.Cascade()},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.RestrictUntilEOT{
				Target:       id,
				Restrictions: game.CantBlock,
				Label:        "Demonic Dread — can't block this turn",
			}.Apply(ctx)
		},
	})
}
