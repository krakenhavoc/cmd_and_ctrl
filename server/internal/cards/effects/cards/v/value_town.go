package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Value Town // Take a Trip to... — Land — Town, with a Sorcery —
// Adventure half {4}{U}{R} (oracle f4b47eff…, #2176):
//
//	Value Town
//	  "Value Town enters the battlefield tapped.
//	   {T}: Add {U} or {R}."
//	Take a Trip to...
//	  "Draw two cards. This spell deals 2 damage to each opponent.
//	   (Then exile this card. You may play the land later from
//	   exile.)"
//
// A Town land with an Adventure half; see Jidoor for the land-play
// lifecycle. The damage is one damage instance across the table, dealt
// by the spell.
//
// No simplification.
const ValueTownOracleID = "f4b47eff-88b0-49c0-b92a-286d083652f8"

func init() {
	effects.Register(effects.Spec{
		OracleID:     ValueTownOracleID,
		Name:         "Value Town",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{U|R}",
			Label:    "Add {U} or {R}",
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     ValueTownOracleID + "#1",
		Name:         "Take a Trip to...",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := (effects.DrawCards{Player: item.Controller, N: 2}).Apply(ctx); err != nil {
				return err
			}
			return effects.DamageToEachOpponent(ctx.Game, item, 2)
		},
	})
}
