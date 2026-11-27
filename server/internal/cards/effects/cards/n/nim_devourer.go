package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nim Devourer — Creature — Zombie {3}{B}{B}, 4/1 (Mirrodin, #5606):
//
//	"This creature gets +1/+0 for each artifact you control.
//	 {B}{B}: Return this card from your graveyard to the battlefield, then
//	 sacrifice a creature. Activate only during your upkeep."
//
// The pump is Benalish Honor Guard's per-count static over the artifacts its
// controller controls. The graveyard ability is Necrosavant's window with
// Cauldron Familiar's return; the sacrifice is Stitcher's Apprentice's
// "sacrifice a creature" chosen as the ability resolves, so the Devourer
// itself may be the one given up. A Devourer that has left the graveyard
// since is not returned, and the sacrifice still happens (CR 608.2c).
//
// No simplification.
func init() {
	pump := effects.PumpSelfCreatureOrAttachedPer(1, 0, func(g *game.Game, source *game.Card) int {
		return effects.CountPermanentsYouControl(g, source, func(c *game.Card) bool { return c.IsArtifact() })
	})
	pump.Label = "This creature gets +1/+0 for each artifact you control."
	effects.Register(effects.Spec{
		OracleID:     "89075fbe-cd8a-4054-abab-fcaa2aedddfb",
		Name:         "Nim Devourer",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{pump},
		Activated: []effects.ActivatedAbility{{
			Label:     "{B}{B}: Return this card from your graveyard to the battlefield, then sacrifice a creature. Activate only during your upkeep.",
			Purpose:   game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:      effects.ManaCost("{B}{B}"),
			Zones:     []game.ZoneKind{game.ZoneGraveyard},
			Condition: effects.DuringYourUpkeep(),
			Effect: effects.Do(
				effects.Instruction(func(ctx *effects.Context) error {
					return effects.ReturnThisFromGraveyardToBattlefield(ctx.Game, ctx.Item)
				}),
				effects.Instruction(func(ctx *effects.Context) error {
					ctx.Game.PlayerSacrificesForEffect(ctx.Item.SourceCardID, ctx.Controller(), effects.SacrificeSpec("a creature", effects.Creature()),
						"Nim Devourer — sacrifice a creature")
					return nil
				}),
			),
		}},
	})
}
