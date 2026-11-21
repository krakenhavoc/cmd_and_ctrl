package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jenny, Generated Anomaly — Legendary Creature — Time Lord Soldier
// {2}{R}{W}, 2/3 (Doctor Who, #5948):
//
//	"Double strike
//	 Whenever Jenny deals combat damage to a player, it explores."
//
// Double strike is a printed keyword. Jenny explores each time she deals
// combat damage to a player, so with double strike she explores in both
// damage steps (CR 510.2, CR 702.4b). A Jenny that has left by the time the
// trigger resolves still reveals the card and gets no counter (CR 701.44c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fbddb6a4-f817-4f3d-bf75-63a98e71954b",
		Name:            "Jenny, Generated Anomaly",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Jenny, Generated Anomaly — it explores", jennyExplores),
		},
	})
}

// jennyExplores makes the trigger's source explore.
func jennyExplores(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	return effects.Explores{Explorer: ctx.Source()}.Apply(ctx)
}
