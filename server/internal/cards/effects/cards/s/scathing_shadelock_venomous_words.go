package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scathing Shadelock // Venomous Words — a preparation card (CR 722, ADR
// 0090), Secrets of Strixhaven (#4456):
//
//	Scathing Shadelock {4}{B}, Creature — Snake Warlock 4/6:
//	"At the beginning of your first main phase, this creature becomes
//	 prepared. (While it's prepared, you may cast a copy of its spell. Doing so
//	 unprepares it.)"
//	Venomous Words {B}, Sorcery:
//	"Target creature you control gets +2/+0 and gains deathtouch until end of
//	 turn."
//
// The trigger is the beginning of the precombat main phase of its
// controller's turn (AtYourPrecombatMain); BecomePrepared does nothing if the
// creature left or is already prepared. The prepare spell is Interjection's
// shape (socB27TargetGetsAndGains).
//
// No simplification.
func init() {
	const id = "2796c3d9-9e56-40f2-8398-7a131c4657ff"
	effects.Register(effects.Spec{
		OracleID:     id,
		Name:         "Scathing Shadelock",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourPrecombatMain("Scathing Shadelock — becomes prepared",
				func(g *game.Game, item *game.StackItem) error {
					return effects.BecomePrepared{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
				}),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     id + "#1",
		Name:         "Venomous Words",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 0, "deathtouch")),
		OnResolve:    effects.SocB27TargetGetsAndGains(2, 0, "deathtouch", "Venomous Words"),
	})
}
