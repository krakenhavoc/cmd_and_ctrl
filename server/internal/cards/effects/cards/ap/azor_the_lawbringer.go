package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Azor, the Lawbringer — Legendary Creature — Sphinx {2}{W}{W}{U}{U}, 6/6:
//
//	"Flying
//	 When Azor enters, each opponent can't cast instant or sorcery spells during
//	 that player's next turn.
//	 Whenever Azor attacks, you may pay {X}{W}{U}{U}. If you do, you gain X life
//	 and draw X cards."
//
// The enters trigger grants each opponent still in the game a cast ban over
// instants and sorceries (CastBanRule with an instant-or-sorcery filter) that
// reaches only their next turn (CastBanRule.OnTurnBegun, #5531): casts at any
// other time, including on the turns before it, are untouched, and the effect
// ends with that turn. The attack trigger is Elenda and Azor's MayPayX: X is
// chosen as it resolves and the payment is optional; the life and the cards are
// both X.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "89812139-9fc1-4fc9-b1d0-d07067d62083",
		Name:            "Azor, the Lawbringer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Azor — each opponent can't cast instant or sorcery spells during that player's next turn", azorBan),
			effects.WheneverThisAttacks("Azor — you may pay {X}{W}{U}{U}; if you do, you gain X life and draw X cards", azorAttack),
		},
	})
}

func azorBan(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, opp := range ctx.Opponents() {
		if err := (effects.RestrictCasting{
			Player: opp,
			Rule: game.CastBanRule{
				Kind:        game.CastBanOutright,
				Filter:      game.PermissionFilter{InstantOrSorceryOnly: true},
				OnTurnBegun: g.NextTurnNumberForEffect(opp),
			},
			Label:    "Azor, the Lawbringer — can't cast instant or sorcery spells during your next turn",
			Duration: g.UntilEndOfYourNextTurnDuration(opp),
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

func azorAttack(g *game.Game, item *game.StackItem) error {
	return effects.MayPayX{
		Cost:  "{X}{W}{U}{U}",
		Label: "Azor, the Lawbringer",
		Buys:  "gain X life and draw X cards",
		Goal:  effects.XAsHighAsYouCan,
		Unit:  game.PayAmountCards,
		Then:  azorPaid,
	}.Apply(effects.NewContext(g, item))
}

var azorPaid = effects.PayUnlessBranches("pay-unless/s82-old-b216-azor-gain-and-draw-x",
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		if err := (effects.GainLife{Amount: r.Number}).Apply(ctx); err != nil {
			return err
		}
		return effects.DrawCards{N: r.Number}.Apply(ctx)
	}, nil)
