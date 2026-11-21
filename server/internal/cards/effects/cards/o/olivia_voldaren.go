package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Olivia Voldaren — Legendary Creature — Vampire {2}{B}{R}, 3/3
// (Innistrad Remastered, S82 GAP, #5948):
//
//	"Flying
//	 {1}{R}: Olivia Voldaren deals 1 damage to another target creature.
//	 That creature becomes a Vampire in addition to its other types. Put a
//	 +1/+1 counter on Olivia Voldaren.
//	 {3}{B}{B}: Gain control of target Vampire for as long as you control
//	 Olivia Voldaren."
//
// The ping is the damage, then the type change (a Layer 4 effect with no
// duration, pinned to the creature as it is now, CR 400.7), then the
// counter; Olivia needs to be on the battlefield for the counter only,
// and the counter is put on her even when the damage was prevented
// (CR 608.2c). The steal lasts for as long as the activating player
// controls this Olivia: if she leaves, or changes hands, the creature
// goes back (CR 611.2b), and a new Olivia that arrived while the ability
// waited is not the one it names.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "dd27d0f2-fe1e-4b63-9ba0-6264ca3c70f8",
		Name:            "Olivia Voldaren",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}{R}: Olivia Voldaren deals 1 damage to another target creature. That creature becomes a Vampire in addition to its other types. Put a +1/+1 counter on Olivia Voldaren.",
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
				Cost:    effects.ManaCost("{1}{R}"),
				Targets: effects.Another(effects.TargetCreature("another target creature")),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					target, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					if err := (effects.DealDamage{Source: item.SourceCardID, Target: target, Amount: 1}).Apply(ctx); err != nil {
						return err
					}
					if effects.OnBattlefield(g, target) {
						if err := (effects.ScopedEffectFor{
							Target:   target,
							Mods:     []game.Mod{game.AddSubtypesMod(ctx.TextSubtypes("Vampire")...)},
							Duration: g.PinnedTo(game.IndefiniteDuration(), target),
							Label:    "Olivia Voldaren — it becomes a Vampire in addition to its other types",
						}).Apply(ctx); err != nil {
							return err
						}
					}
					return effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
				},
			},
			{
				Label:   "{3}{B}{B}: Gain control of target Vampire for as long as you control Olivia Voldaren.",
				Cost:    effects.ManaCost("{3}{B}{B}"),
				Targets: effects.TargetCreature("target Vampire", effects.HasSubtype("Vampire")),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					target := effects.FirstLegalBattlefieldTarget(ctx)
					if target == uuid.Nil {
						return nil
					}
					d, ok := effects.DurationWhileYouControlSource(ctx, ctx.Source(), ctx.Controller())
					if !ok {
						return nil
					}
					return effects.GainControl{Target: target, Controller: ctx.Controller(), Duration: d,
						Label: "Olivia Voldaren — control while you control Olivia Voldaren"}.Apply(ctx)
				},
			},
		},
	})
}
