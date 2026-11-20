package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Paleoloth — Creature — Beast {4}{G}{G}, 5/5 (Conflux, #5587):
//
//	"Whenever another creature you control with power 5 or greater
//	 enters, you may return target creature card from your graveyard to
//	 your hand."
//
// Garruk's Uprising's "another creature you control with power N or
// greater enters" at 5, read on the creature as it enters, counters it
// entered with included. The target is chosen as the trigger goes on
// the stack (CR 603.3d) and the "you may" is asked as it resolves (CR
// 603.5).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "627de3d8-ad5e-4f9b-b83c-476cc98fea9b",
		Name:         "Paleoloth",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(Targeting(
				On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if ev.CardID == source.InstanceID {
						return false
					}
					c, ok := g.LookupCardForEffect(ev.CardID)
					return ok && c.IsCreature() && c.Controller == source.Controller && c.CurrentPower() >= 5
				}, "Paleoloth — return target creature card from your graveyard to your hand",
					returnFirstLegalGraveyardTargetToHand),
				TargetCardInGraveyard("target creature card from your graveyard", Creature(), YouOwn())),
				"Paleoloth — return target creature card from your graveyard to your hand?"),
		},
	})
}
