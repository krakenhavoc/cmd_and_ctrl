package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rune-Brand Juggler — Creature — Human Shaman {B}{R}:
//
//	"When this creature enters, suspect up to one target creature you
//	 control. (A suspected creature has menace and can't block.)
//	 {3}{B}{R}, Sacrifice a suspected creature: Target creature gets
//	 -5/-5 until end of turn."
//
// The enters trigger is the shared "suspect up to one target" clause
// (SuspectEachLegalTarget over UpToOneTargetCreature). The activation's
// cost clause is the Suspected() predicate on a sacrifice spec: the
// Juggler can feed itself when it is the suspected creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8c105948-4b8f-4615-b66a-2bbdb3f3d6dd",
		Name:         "Rune-Brand Juggler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{effects.Targeting(
			effects.WhenThisEnters("Rune-Brand Juggler — suspect up to one target creature you control", effects.SuspectEachLegalTarget),
			effects.UpToOneTargetCreature("up to one target creature you control", effects.YouControl()),
		)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{B}{R}, Sacrifice a suspected creature: Target creature gets -5/-5 until end of turn.",
			Purpose: effects.ForTargets(effects.HarmUntilEOT(0, -5, -5)),
			Cost:    effects.Plus(effects.ManaCost("{3}{B}{R}"), effects.SacrificeASuspectedCreature()),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (effects.BoostUntilEOT{
						Target: t.ID, Power: -5, Toughness: -5,
						Label: "Rune-Brand Juggler — -5/-5 until end of turn",
					}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
