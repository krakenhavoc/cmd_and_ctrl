package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lingering Phantom — Creature — Spirit {5}{B}, 5/4 (Dominaria, S82 OLD,
// #5530):
//
//	"Whenever you cast a historic spell, you may pay {B}. If you do, return
//	 this card from your graveyard to your hand. (Artifacts, legendaries, and
//	 Sagas are historic.)"
//
// A trigger that works from the graveyard (CR 113.6) on "you cast a historic
// spell" (Jhoira's reading, CR 700.6). The payment is asked as the trigger
// resolves (CR 118.12), and the return happens only if it was paid; a card that
// has left the graveyard in the meantime, or come back as a new object, is not
// returned (CR 400.7).
//
// No simplification.
func init() {
	trigger := effects.On(game.EventCast, effects.YouCastAHistoricSpell,
		"Lingering Phantom — you may pay {B}; if you do, return it from your graveyard to your hand",
		lingeringPhantomOffer)
	trigger.Zones = []game.ZoneKind{game.ZoneGraveyard}
	effects.Register(effects.Spec{
		OracleID:     "76164af1-275e-4d68-9df5-9801768512ef",
		Name:         "Lingering Phantom",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{trigger},
	})
}

// lingeringPhantomPaid is the paid branch: the card returns from the graveyard.
// An on-disk key: never renamed, never reused.
var lingeringPhantomPaid = effects.PayUnlessBranches("pay-unless/s82-old-b212-lingering-phantom-return",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.S82OldB212ReturnThisFromGraveyardToHand(ctx.Game, ctx.Item)
	}, nil)

func lingeringPhantomOffer(g *game.Game, item *game.StackItem) error {
	return effects.MayPay{
		Chooser:  item.Controller,
		Cost:     "{B}",
		Question: "Lingering Phantom — pay {B} to return it from your graveyard to your hand?",
		Then:     lingeringPhantomPaid,
	}.Apply(effects.NewContext(g, item))
}
