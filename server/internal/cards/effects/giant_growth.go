package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Giant Growth — Instant for {G}:
//
//	"Target creature gets +3/+3 until end of turn."
//
// The card the S32 turn-scoped continuous-effect registry was built
// to make possible, and the reason a pump spell had never appeared
// in the catalog before: the layer engine recomputed continuous
// effects from the battlefield on every pass, so there was nowhere
// for a +3/+3 to live once the instant that made it hit the
// graveyard. `BoostUntilEOT` puts it in `Game.ScopedEffects`,
// where the recompute finds it until the cleanup step sweeps it
// (CR 514.2). See ADR 0035.
//
// Everything about this card is real:
//
//   - The +3/+3 is a layer 7c modification, so it stacks with an
//     anthem rather than overwriting it, and applies on top of a
//     layer 7b "base P/T becomes N/N" the way the rules say.
//   - It is visible the instant the spell resolves — combat damage,
//     the lethal-damage SBA and the snapshot path all recompute
//     layers before reading power and toughness, so a blocked 2/2
//     pumped mid-combat really does kill the 4/4.
//   - It expires at cleanup, including when cast during the end
//     step: "until end of turn" is the cleanup step, not the end
//     step.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "5748ebf1-24e3-499d-ab7c-c2cebd462a24",
		Name:         "Giant Growth",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return BoostUntilEOT{
				Target:    item.Targets[0].ID,
				Power:     3,
				Toughness: 3,
				Label:     "Giant Growth — +3/+3",
			}.Apply(ctx)
		},
	})
}
