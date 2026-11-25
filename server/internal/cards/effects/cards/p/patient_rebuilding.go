package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Patient Rebuilding — Enchantment {3}{U}{U} (Core Set 2019, #5527):
//
//	"At the beginning of your upkeep, target opponent mills three cards,
//	 then you draw a card for each land card put into their graveyard this
//	 way."
//
// The Mouth of Sauron's targeted mill with a keyed tail (a mill can pause on
// a replacement prompt): the tail is told the cards that LANDED in the
// graveyard (CR 400.7, #893), so a card Rest in Peace exiled instead is not
// counted. The land cards among them are counted as they were milled and the
// Rebuilding's controller draws that many.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "df9fe645-6f7b-44a0-ab69-b68cb7e525cd",
		Name:         "Patient Rebuilding",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.AtYourUpkeep("Patient Rebuilding — target opponent mills three cards, then you draw a card for each land card put into their graveyard this way",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						t, ok := effects.FirstLegalTarget(ctx)
						if !ok || t.Kind != game.TargetPlayer {
							return nil
						}
						return g.MillToZoneKeyedForEffect(t.ID, 3, game.ZoneGraveyard, game.MillUntil{}, ctx.ItemTail(patientRebuildingDraw))
					}),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}

// patientRebuildingDraw is the tail: a card for each land card the mill put
// into the graveyard. An on-disk key.
var patientRebuildingDraw = effects.CardTail("then/patient-rebuilding-draw", func(ctx *effects.Context, o game.TailOutcome) error {
	lands := 0
	for _, id := range o.IDs {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.IsLand() {
			lands++
		}
	}
	if lands == 0 {
		return nil
	}
	return effects.DrawCards{Player: ctx.Controller(), N: lands}.Apply(ctx)
})
