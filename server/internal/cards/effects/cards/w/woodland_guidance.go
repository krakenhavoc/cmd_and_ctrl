package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Woodland Guidance — Sorcery {3}{G} (Lorwyn, #5593):
//
//	"Return target card from your graveyard to your hand. Clash with an
//	 opponent. If you win, untap all Forests you control. Exile Woodland
//	 Guidance. (Each clashing player reveals the top card of their library,
//	 then puts that card on their choice of the top or bottom. A player wins
//	 if their card had a greater mana value.)"
//
// The return comes first, then the clash. Its tail untaps the controller's
// Forests on a win and then exiles the spell card whether or not it won, as
// printed (Rise of the Eldrazi's ExileCardForEffect on the resolving card).
// A target that became illegal does nothing, clash and exile included
// (CR 608.2b), and the card is then put into the graveyard as normal.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "09e84398-7782-48d0-8411-9055d253395c",
		Name:         "Woodland Guidance",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      effects.TargetCardInGraveyard("target card from your graveyard", effects.YouOwn()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			if err := (effects.ReturnFromGraveyard{Target: item.Targets[0].ID, Dest: game.ZoneHand}).Apply(ctx); err != nil {
				return err
			}
			return effects.ClashWithAnOpponent{Then: ctx.ItemTail(woodlandGuidanceClashed)}.Apply(ctx)
		},
	})
}

// woodlandGuidanceClashed is "If you win, untap all Forests you control.
// Exile Woodland Guidance." An on-disk key.
var woodlandGuidanceClashed = effects.CardTail("then/woodland-guidance-clashed", func(ctx *effects.Context, o game.TailOutcome) error {
	if effects.WonTheClash(o) {
		if err := effects.B16UntapAllYouControlMatching(ctx, ctx.Controller(), func(c game.Card) bool { return c.IsLand() && c.HasSubtypeWord("Forest") }); err != nil {
			return err
		}
	}
	return ctx.Game.ExileCardForEffect(ctx.Item.SourceCardID)
})
