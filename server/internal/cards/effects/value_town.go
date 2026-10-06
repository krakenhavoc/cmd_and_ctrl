package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
const valueTownOracleID = "f4b47eff-88b0-49c0-b92a-286d083652f8"

func init() {
	Register(Spec{
		OracleID:     valueTownOracleID,
		Name:         "Value Town",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U|R}",
			Label:    "Add {U} or {R}",
		}},
	})
	Register(Spec{
		OracleID:     valueTownOracleID + "#1",
		Name:         "Take a Trip to...",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (DrawCards{Player: item.Controller, N: 2}).Apply(ctx); err != nil {
				return err
			}
			return damageToEachOpponent(ctx.Game, item, 2)
		},
	})
}
