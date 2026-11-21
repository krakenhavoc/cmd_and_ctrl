package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mandible Justiciar — Artifact Creature — Phyrexian Cleric {1}{W}, 2/1
// (Phyrexia: All Will Be One):
//
//	"Lifelink
//	 Whenever another artifact you control enters, this creature gets +1/+1
//	 until end of turn."
//
// One trigger per artifact entering under your control other than the
// Justiciar itself; the pump goes on the Justiciar only while it is still
// the same object on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ee8f3dbd-6e86-46c2-9393-7da6c3c4b46d",
		Name:            "Mandible Justiciar",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.IsArtifactEnteredUnderYourControl(ev, source, g)
			}, "Mandible Justiciar — gets +1/+1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if ctx.IsNewSourceObject(ctx.Source()) || !effects.OnBattlefield(g, ctx.Source()) {
						return nil
					}
					return effects.BoostUntilEOT{Target: ctx.Source(), Power: 1, Toughness: 1, Label: "Mandible Justiciar — +1/+1"}.Apply(ctx)
				}),
		},
	})
}
