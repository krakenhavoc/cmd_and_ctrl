package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sunstreak Phoenix — {2}{R}{R} Creature — Phoenix 4/2 (#2586, ADR 0132):
//
//	"Flying
//	 If it's neither day nor night, it becomes day as this creature
//	 enters.
//	 Whenever day becomes night or night becomes day, you may pay
//	 {1}{R}. If you do, return this card from your graveyard to the
//	 battlefield tapped."
//
// The flip trigger watches from the GRAVEYARD, the only zone it can ever
// fire from (Bloodghast's InGraveyard shape). It is a "you may pay" — the
// payment is asked as the ability resolves (MayPay), and the return is
// the payment's consequence. The card may have left the graveyard in the
// meantime (CR 400.7), and then the return does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "94f14f42-ea66-4432-81cc-667ad62a4497",
		Name:            "Sunstreak Phoenix",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		AsEnters:        effects.BecomesDayAsEnters(),
		Triggered: []game.TriggeredAbility{
			effects.InGraveyard(effects.WheneverDayBecomesNightOrNightBecomesDay(
				"Sunstreak Phoenix — you may pay {1}{R} to return it from your graveyard to the battlefield tapped",
				sunstreakPhoenixOffer)),
		},
	})
}

func sunstreakPhoenixOffer(g *game.Game, item *game.StackItem) error {
	if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return effects.MayPay{
		Chooser:  item.Controller,
		Cost:     "{1}{R}",
		Question: "Sunstreak Phoenix — pay {1}{R} to return it to the battlefield tapped?",
		Then:     sunstreakPhoenixPaid,
	}.Apply(effects.NewContext(g, item))
}

// sunstreakPhoenixPaid is the paid half, keyed (#4264).
var sunstreakPhoenixPaid = effects.PayUnlessBranches("pay-unless/sunstreak-phoenix-return",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		if z := ctx.Game.FindCardZoneForEffect(ctx.Item.SourceCardID); z == nil || z.Kind != game.ZoneGraveyard {
			return nil
		}
		return effects.ReturnFromGraveyard{Target: ctx.Item.SourceCardID, Dest: game.ZoneBattlefield, Tapped: true}.Apply(ctx)
	}, nil)
