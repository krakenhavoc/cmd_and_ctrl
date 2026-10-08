package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "94f14f42-ea66-4432-81cc-667ad62a4497",
		Name:            "Sunstreak Phoenix",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		AsEnters:        BecomesDayAsEnters(),
		Triggered: []game.TriggeredAbility{
			InGraveyard(WheneverDayBecomesNightOrNightBecomesDay(
				"Sunstreak Phoenix — you may pay {1}{R} to return it from your graveyard to the battlefield tapped",
				sunstreakPhoenixOffer)),
		},
	})
}

func sunstreakPhoenixOffer(g *game.Game, item *game.StackItem) error {
	if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return MayPay{
		Chooser:  item.Controller,
		Cost:     "{1}{R}",
		Question: "Sunstreak Phoenix — pay {1}{R} to return it to the battlefield tapped?",
		OnPay: func(ctx *Context) error {
			if z := ctx.Game.FindCardZoneForEffect(ctx.Item.SourceCardID); z == nil || z.Kind != game.ZoneGraveyard {
				return nil
			}
			return ReturnFromGraveyard{Target: ctx.Item.SourceCardID, Dest: game.ZoneBattlefield, Tapped: true}.Apply(ctx)
		},
	}.Apply(NewContext(g, item))
}
