package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zanarkand, Ancient Metropolis // Lasting Fayth — Land — Town, with a
// Sorcery — Adventure half {4}{G}{G} (oracle 5f2b3ea8…, #2176):
//
//	Zanarkand, Ancient Metropolis
//	  "This land enters tapped.
//	   {T}: Add {G}."
//	Lasting Fayth
//	  "Create a 1/1 colorless Hero creature token. Put a +1/+1 counter
//	   on it for each land you control. (Then exile this card. You may
//	   play the land later from exile.)"
//
// A Town land with an Adventure half; see Jidoor for the land-play
// lifecycle. The land count is read as the spell resolves, and the
// counters ride the token's entry (TokenEntryOptions.Counters) rather
// than being placed a beat later — the finished 1/1-plus-counters
// object is what enters, so a Doubling Season still doubles them. The
// token is not a land, so creating it first changes no count.
//
// No simplification.
const zanarkandOracleID = "5f2b3ea8-99ee-47a4-8a1c-4b27478d524c"

func init() {
	Register(Spec{
		OracleID:     zanarkandOracleID,
		Name:         "Zanarkand, Ancient Metropolis",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
	})
	Register(Spec{
		OracleID:     zanarkandOracleID + "#1",
		Name:         "Lasting Fayth",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			lands := b02CountLandsControlledBy(ctx.Game, item.Controller)
			opts := game.TokenEntryOptions{}
			if lands > 0 {
				opts.Counters = map[string]int{"+1/+1": lands}
			}
			_, err := ctx.Game.CreateTokensForEffect(item.Controller, TokenCard("1/1 colorless Hero"), 1, opts)
			return err
		},
	})
}
