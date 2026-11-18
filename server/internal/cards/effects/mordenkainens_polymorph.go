package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mordenkainen's Polymorph — Instant {1}{U} (Adventures in the Forgotten
// Realms):
//
//	"Until end of turn, target creature becomes a Dragon with base power and
//	 toughness 4/4 and gains flying. (It loses all other creature types.)"
//
// One record (CR 613.7): the layer-4 creature-type replacement (CR 205.1a),
// the layer-7b base P/T and the layer-6 flying share a timestamp. Counters
// and +N/+N effects still apply on top of the 4/4.
//
// No Purpose: the change is partly a benefit and partly a harm (see Nameless
// Inversion).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5ac8c234-030d-4ba1-bc2f-d05761045278",
		Name:         "Mordenkainen's Polymorph",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			mods := []game.Mod{game.SetCreatureTypesMod(ctx.TextSubtype("Dragon"))}
			mods = append(mods, game.SetBasePTMods(4, 4)...)
			mods = append(mods, game.AddKeywordsMod("flying"))
			return untilEndOfTurn(ctx, item.Targets[0].ID, nil,
				"Mordenkainen's Polymorph — becomes a 4/4 Dragon with flying", mods...)
		},
	})
}
