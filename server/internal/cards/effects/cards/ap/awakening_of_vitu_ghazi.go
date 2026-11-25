package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Awakening of Vitu-Ghazi — Instant {3}{G}{G} (War of the Spark):
//
//	"Put nine +1/+1 counters on target land you control. It becomes a legendary
//	 0/0 Elemental creature with haste named Vitu-Ghazi. It's still a land."
//
// The nine counters go on first, then the land becomes a creature with no
// duration: the effect lasts as long as the land is that object on the
// battlefield (CR 611.2a, 400.7). It is layer 3 (the name, CR 612.8), layer 4 (the
// creature type and Elemental subtype and the legendary supertype, in addition to
// what it was) and layer 7b (the base 0/0), with haste a layer 6 grant; the
// counters make it 9/9 and being a land is kept. A land that is no longer a legal
// target as the spell resolves is not touched (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5c326102-005d-463c-b9f7-14b034b66baa",
		Name:         "Awakening of Vitu-Ghazi",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPump | game.AnswerAnimate},
		Targets:      effects.TargetPermanent("target land you control", effects.Land(), effects.YouControl()),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 9}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.ScopedEffectFor{
					Target: id,
					Mods: []game.Mod{
						game.SetNameMod("Vitu-Ghazi"),
						game.AddTypesMod("Creature"),
						game.AddSubtypesMod(ctx.TextSubtype("Elemental")),
						game.AddSupertypesMod("Legendary"),
						game.SetBasePowerMod(0),
						game.SetBaseToughnessMod(0),
						game.AddKeywordsMod("haste"),
					},
					Duration: ctx.Game.PinnedTo(game.IndefiniteDuration(), id),
					Label:    "Awakening of Vitu-Ghazi — a legendary 0/0 Elemental creature with haste named Vitu-Ghazi",
				}.Apply(ctx)
			}),
		),
	})
}
