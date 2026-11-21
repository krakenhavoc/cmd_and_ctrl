package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shadowborn Demon — {3}{B}{B} Creature — Demon 5/6 (Jumpstart 2022, #5948):
//
//	"Flying
//	 When this creature enters, destroy target non-Demon creature.
//	 At the beginning of your upkeep, if there are fewer than six creature
//	 cards in your graveyard, sacrifice a creature."
//
// Reaper from the Abyss's non-Demon target; the upkeep clause is an
// intervening "if" (CR 603.4), checked when the step begins and again as
// the trigger resolves, and the sacrifice is any creature you choose,
// the Demon included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7dacc5aa-ea4d-414e-8324-6a7d24bb7671",
		Name:            "Shadowborn Demon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Shadowborn Demon — destroy target non-Demon creature", effects.DestroyFirstLegalCardTarget),
				effects.TargetCreature("target non-Demon creature", effects.Not(effects.Subtype("Demon")))),
			effects.On(game.EventBeginUpkeep, effects.AllOf(effects.ByYou, shadowbornDemonFewerThanSix),
				"Shadowborn Demon — fewer than six creature cards in your graveyard: sacrifice a creature",
				func(g *game.Game, item *game.StackItem) error {
					if effects.B11CreatureCardsInGraveyard(g, item.Controller) >= 6 {
						return nil
					}
					g.PlayerSacrificesForEffect(item.SourceCardID, item.Controller,
						effects.SacrificeSpec("a creature", effects.Creature()), "Shadowborn Demon — sacrifice a creature")
					return nil
				}),
		},
	})
}

func shadowbornDemonFewerThanSix(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return effects.B11CreatureCardsInGraveyard(g, source.Controller) < 6
}
