package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fistful of Force — Instant {1}{G} (Lorwyn, #5593):
//
//	"Target creature gets +2/+2 until end of turn. Clash with an opponent.
//	 If you win, that creature gets an additional +2/+2 and gains trample
//	 until end of turn. (Each clashing player reveals the top card of their
//	 library, then puts that card on their choice of the top or bottom. A
//	 player wins if their card had a greater mana value.)"
//
// The first +2/+2 lands as the spell resolves, then the clash; on a win the
// target (carried through the clash's tail) is checked again (CR 608.2b) and
// gets another +2/+2 and trample. A target that became illegal does nothing,
// clash included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "81641cf1-2dcf-4cd3-ab92-7f5e8e20cb84",
		Name:         "Fistful of Force",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Fistful of Force — +2/+2"}).Apply(ctx); err != nil {
				return err
			}
			then := ctx.ItemTail(fistfulOfForceClashed)
			then.Carry = []uuid.UUID{id}
			return effects.ClashWithAnOpponent{Then: then}.Apply(ctx)
		},
	})
}

// fistfulOfForceClashed is "If you win, that creature gets an additional
// +2/+2 and gains trample until end of turn". An on-disk key.
var fistfulOfForceClashed = effects.CardTail("then/fistful-of-force-clashed", func(ctx *effects.Context, o game.TailOutcome) error {
	if !effects.WonTheClash(o) || len(o.Tail.Carry) == 0 || !effects.OnBattlefield(ctx.Game, o.Tail.Carry[0]) {
		return nil
	}
	id := o.Tail.Carry[0]
	if err := (effects.BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Fistful of Force — an additional +2/+2"}).Apply(ctx); err != nil {
		return err
	}
	return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}, Label: "Fistful of Force — trample"}.Apply(ctx)
})
