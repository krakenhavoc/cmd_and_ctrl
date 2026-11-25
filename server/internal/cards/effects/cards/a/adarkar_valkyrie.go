package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Adarkar Valkyrie — Snow Creature — Angel {4}{W}{W}, 4/5 (Commander 2018,
// S82 OLD batch 193, #5526):
//
//	"Flying, vigilance
//	 {T}: When target creature other than this creature dies this turn,
//	 return that card to the battlefield under your control."
//
// The activated ability creates a delayed triggered ability as it resolves
// (CR 603.7a), about the target as the object it is then (CR 603.7c): a
// creature that left and came back is a new object and never fires it. The
// delayed trigger is controlled by whoever activated (CR 603.7d), so the card
// returns under the Valkyrie's controller's control whoever owned or
// controlled the creature. It returns "that card" only while it is still in
// the graveyard it died into; a token (not a card), or a card exiled in
// response, is not returned. The default "this turn" duration ends the
// delayed trigger at cleanup (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2932baae-bdf1-455a-9ac4-23ba1f83e7e1",
		Name:            "Adarkar Valkyrie",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: When target creature other than this creature dies this turn, return that card to the battlefield under your control.",
			Cost:    effects.TapCost(),
			Targets: effects.Another(effects.TargetCreature("target creature other than this creature")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				target := effects.FirstLegalBattlefieldTarget(ctx)
				return effects.WhenThatCreatureDiesThisTurn(ctx, target,
					"Adarkar Valkyrie — return that card to the battlefield under your control", adarkarValkyrieReturn)
			},
		}},
	})
}

// adarkarValkyrieReturn is the fired trigger's body: the dead creature's card,
// if it is still in a graveyard, comes back under the trigger's controller's
// control. An on-disk key: never renamed, never reused.
var adarkarValkyrieReturn = game.DelayedBody("old-b193/adarkar-valkyrie-return",
	func(g *game.Game, item *game.StackItem, _ game.EffectParams) error {
		ctx := effects.NewContext(g, item)
		obj, ok := effects.PayloadObject(ctx)
		if !ok {
			return nil
		}
		z := g.FindCardZoneForEffect(obj.ID)
		if z == nil || z.Kind != game.ZoneGraveyard {
			return nil
		}
		return effects.ReturnFromGraveyard{Target: obj.ID, Dest: game.ZoneBattlefield, Controller: ctx.Controller()}.Apply(ctx)
	})
