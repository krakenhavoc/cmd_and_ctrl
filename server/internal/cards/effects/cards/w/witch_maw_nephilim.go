package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Witch-Maw Nephilim — Creature — Nephilim {G}{W}{U}{B}, 1/1 (Guildpact):
//
//	"Whenever you cast a spell, you may put two +1/+1 counters on this
//	 creature.
//	 Whenever this creature attacks, it gains trample until end of turn if
//	 its power is 10 or greater."
//
// The first trigger asks as it resolves and puts the counters on the
// Nephilim if it is still the same permanent on the battlefield. The second
// has an intervening "if" (CR 603.4): its power is read as the creature
// attacks and again as the ability resolves, as it last was if it has left
// (CR 608.2h).
//
// No simplification.
func init() {
	attack := effects.WheneverThisAttacks("Witch-Maw Nephilim — it gains trample until end of turn if its power is 10 or greater",
		func(g *game.Game, item *game.StackItem) error {
			if effects.B43PowerNowOrLastKnown(g, item.SourceCardID) < 10 || !effects.SourceIsStillThisPermanent(g, item) {
				return nil
			}
			return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"trample"}, Label: "Witch-Maw Nephilim — trample"}.
				Apply(effects.NewContext(g, item))
		})
	attackWhen := attack.AppliesTo
	attack.AppliesTo = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return attackWhen(ev, source, lki, g) && source.PowerForComparison() >= 10
	}
	effects.Register(effects.Spec{
		OracleID:     "26edbbb8-331e-42a6-a550-acc7e9efb32c",
		Name:         "Witch-Maw Nephilim",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(nil, "Witch-Maw Nephilim — you may put two +1/+1 counters on this creature",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayChoice{
						Question: "Witch-Maw Nephilim — put two +1/+1 counters on it?",
						Then:     witchMawCounters,
					}.Apply(effects.NewContext(g, item))
				}),
			attack,
		},
	})
}

// witchMawCounters is the "you may" answer, keyed: a table waiting on it is
// a restore point.
var witchMawCounters = effects.MayYes("confirm/witch-maw-nephilim-counters", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	if !effects.OnBattlefield(ctx.Game, ctx.Source()) {
		return nil
	}
	return effects.AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
})
