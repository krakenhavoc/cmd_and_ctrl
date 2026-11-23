package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Butch DeLoria, Tunnel Snake — Legendary Creature — Human Rogue {1}{B},
// 2/2 (Fallout, #5948):
//
//	"Menace
//	 Tunnel Snakes Rule! — Whenever Butch DeLoria attacks, it gets +1/+1
//	 until end of turn for each other Rogue and/or Snake you control.
//	 {1}{B}: Put a menace counter on another target creature. It becomes
//	 a Rogue in addition to its other types."
//
// The count is read as the attack trigger resolves: every other
// permanent you control that is a Rogue or a Snake (a Rogue-and-Snake
// counts once; Butch himself, a Rogue, never counts). The ability's
// menace counter is a keyword counter (CR 122.1b); "it becomes a Rogue
// in addition" is a type-adding effect with no end, pinned to that
// object (CR 611.2b, 400.7), put on only if the target is still legal as
// the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "20f92504-d03c-437e-9814-b25ee384b3ba",
		Name:            "Butch DeLoria, Tunnel Snake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Butch DeLoria — gets +1/+1 until end of turn for each other Rogue and/or Snake you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					n := len(effects.MatchingBattlefield(ctx, effects.And(effects.YouControl(), effects.Or(effects.HasSubtype("Rogue"), effects.HasSubtype("Snake")), effects.NotSelf(item.SourceCardID))))
					if n == 0 {
						return nil
					}
					return effects.BoostUntilEOT{Target: item.SourceCardID, Power: n, Toughness: n,
						Label: "Butch DeLoria — +1/+1 for each other Rogue and/or Snake"}.Apply(ctx)
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}: Put a menace counter on another target creature. It becomes a Rogue in addition to its other types.",
			Cost:    effects.ManaCost("{1}{B}"),
			Targets: effects.Another(effects.TargetCreature("another target creature")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				t, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				if err := (effects.AddCounter{Target: t, Kind: game.CounterMenace, N: 1}).Apply(ctx); err != nil {
					return err
				}
				return effects.ScopedEffectFor{
					Target:   t,
					Mods:     []game.Mod{game.AddSubtypesMod(ctx.TextSubtype("Rogue"))},
					Duration: g.PinnedTo(game.IndefiniteDuration(), t),
					Label:    "Butch DeLoria — a Rogue in addition to its other types",
				}.Apply(ctx)
			},
		}},
	})
}
