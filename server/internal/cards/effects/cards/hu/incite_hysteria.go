package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Incite Hysteria — Sorcery {2}{R}:
//
//	"Radiance — Until end of turn, target creature and each other
//	 creature that shares a color with it gain "This creature can't
//	 block.""
//
// The creatures are read as the spell resolves (radianceSet, #5736). They
// GAIN AN ABILITY, so it is a catalog bundle granted until end of turn
// (GrantAbilitiesFor, ADR 0093), not a restriction laid on them: a
// creature that later loses its abilities loses this one too, under
// CR 613.7's timestamps, as it would lose a printed "can't block". The
// set is locked as the spell resolves (CR 611.2c, and the card's ruling:
// "It doesn't matter if those creatures change colors before blockers
// are declared").
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0a992ad5-136f-482d-8390-fdfe745fe764",
		Name:         "Incite Hysteria",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.CantBlockUntilEOT(0, 0, 0)),
		Targets:      effects.TargetCreature("target creature"),
		Grants: []effects.AbilityGrant{{
			Key:    inciteHysteriaCantBlock,
			Static: []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
			Text:   "This creature can't block.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			ids := effects.RadianceSet(ctx, effects.Creature())
			if len(ids) == 0 {
				return nil
			}
			return effects.GrantAbilitiesFor{
				Match: effects.B10OneOf(ids),
				Keys:  []string{inciteHysteriaCantBlock},
				Label: "Incite Hysteria — this creature can't block",
			}.Apply(ctx)
		},
	})
}

// inciteHysteriaCantBlock is the bundle Incite Hysteria grants. An
// on-disk key.
const inciteHysteriaCantBlock = "incite-hysteria/cant-block"
