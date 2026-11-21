package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chupacabra Echo — Creature — Beast Horror Spirit {2}{B}{B}, 3/2 (The
// Lost Caverns of Ixalan, #4470):
//
//	"Fathomless descent — When this creature enters, target creature an
//	 opponent controls gets -X/-X until end of turn, where X is the number
//	 of permanent cards in your graveyard."
//
// Fathomless descent is an ability word (CR 207.2c). Cogwork Wrestler's
// targeted enters trigger, with X counted as the trigger resolves
// (CR 608.2h) over the permanent cards in its controller's graveyard
// (CR 110.4, permanentCardsInGraveyard). With none, X is 0 and the
// creature is unchanged. The target is chosen as the trigger goes on the
// stack (CR 603.3d); with no creature an opponent controls the trigger is
// removed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "68c56db4-c97a-4c1b-a6f6-096392488e21",
		Name:         "Chupacabra Echo",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Chupacabra Echo — target creature an opponent controls gets -X/-X until end of turn",
				chupacabraEchoShrink),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
				effects.ForTargets(effects.HarmUntilEOT(0, 0, 0))),
		},
	})
}

// chupacabraEchoShrink gives the chosen creature -X/-X until end of turn,
// X the permanent cards in the controller's graveyard as it resolves.
func chupacabraEchoShrink(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	x := effects.PermanentCardsInGraveyard(g, ctx.Controller())
	if !ok || x == 0 {
		return nil
	}
	return effects.BoostUntilEOT{Target: id, Power: -x, Toughness: -x, Label: "Chupacabra Echo — -X/-X until end of turn"}.Apply(ctx)
}
