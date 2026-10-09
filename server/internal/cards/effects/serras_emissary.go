package effects

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Serra's Emissary — Creature — Angel {4}{W}{W}{W}, 7/7:
//
//	"Flying
//	 As this creature enters, choose a card type.
//	 You and creatures you control have protection from the chosen
//	 card type."
//
// The choice is the CR 614.12 option pick the Sieges use, offered over
// game.ChoosableCardTypes and stored on Card.ChosenOption, which already
// has the lifecycle the choice needs: per instance, carried by the
// snapshot and clone, cleared when the permanent leaves (CR 400.7), not
// copiable (#2742).
//
// The protection is resolved from the Emissary, where the answer is,
// rather than from each protected creature: a layer-6 static appends
// game.ProtectionFromCardType of the answer to each creature you
// control, and the player half is the PlayerKeywords placeholder
// game.ProtectionFromTheChosenCardType, which the engine's player walk
// resolves the same way. Unanswered, neither half grants anything.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7a56c6e1-0509-4783-9b29-cf3163977166",
		Name:            "Serra's Emissary",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		AsEnters:        ChooseOptionAsEnters("Serra's Emissary", game.ChoosableCardTypes...),
		PlayerKeywords:  []string{game.ProtectionFromTheChosenCardType},
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return source.ChosenOption != "" && target.Controller == source.Controller && target.IsCreature()
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
				tok := game.ProtectionFromCardType(source.ChosenOption)
				if tok != "" && !slices.Contains(c.Abilities, tok) {
					c.Abilities = append(c.Abilities, tok)
				}
			},
		}},
	})
}
