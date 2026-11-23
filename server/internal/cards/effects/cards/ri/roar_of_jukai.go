package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Roar of Jukai — Instant — Arcane {2}{G}:
//
//	"If you control a Forest, each blocked creature gets +2/+2 until end of
//	 turn.
//	 Splice onto Arcane—An opponent gains 5 life. (As you cast an Arcane
//	 spell, you may reveal this card from your hand and pay its splice cost.
//	 If you do, add this card's effects to that spell.)"
//
// "Each blocked creature" is every attacking creature that became blocked
// (CR 509.1h), whoever controls it, picked out as the spell resolves
// (CR 611.2c): a creature that was blocked and has since lost every blocker
// is still blocked. The splice cost is an opponent gaining 5 life (#4055,
// ADR 0202's 2026-10-19 amendment): the caster chooses which opponent as
// the spell is cast, as gift's recipient is chosen, and the life is that
// player's gain through the replacement window. An opponent who can't gain
// life can't be chosen (CR 614.17b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d611c594-dd2d-4f4b-a958-1f05f22f7f36",
		Name:         "Roar of Jukai",
		Completeness: effects.CompletenessFull,
		Splice:       effects.SpliceOntoArcaneCost(effects.SpliceOpponentGainsLife(5)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if !effects.ControlsA(ctx.TextSubtype("Forest"))(ctx.Game, ctx.Controller()) {
				return nil
			}
			return effects.BoostUntilEOT{
				Match:     effects.BlockedCreature(),
				Power:     2,
				Toughness: 2,
				Label:     "Roar of Jukai — blocked creatures get +2/+2",
			}.Apply(ctx)
		},
	})
}
