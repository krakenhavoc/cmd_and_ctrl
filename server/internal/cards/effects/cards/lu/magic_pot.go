package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Magic Pot — Artifact Creature — Goblin Construct {3}, 1/4 (Final
// Fantasy, #4462):
//
//	"When this creature dies, create a Treasure token.
//	 {2}, {T}: Exile target card from a graveyard."
//
// Common Crook's death trigger and Soul-Guide Lantern's graveyard exile
// (ExileFirstTarget): any graveyard, any card, and nothing if the card
// left in response (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "caed9999-74b8-43e3-a319-fa505c04c71c",
		Name:         "Magic Pot",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Magic Pot — create a Treasure token", effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Exile target card from a graveyard.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Targets: effects.TargetCardInGraveyard("target card from a graveyard"),
			Effect:  effects.ExileFirstTarget,
		}},
	})
}
