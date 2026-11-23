package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dimir Keyrune — Artifact {3}:
//
//	"{T}: Add {U} or {B}.
//	 {U}{B}: This artifact becomes a 2/2 blue and black Horror artifact
//	 creature until end of turn and can't be blocked this turn."
//
// Chronatog Totem's animation with a restriction in the same effect: the
// creature type (layer 4; it stays an artifact), the Horror subtype, blue
// and black (layer 5) and the base 2/2 (layer 7b) at one timestamp
// (CR 613.7), pinned to this object so a Keyrune that was flickered in
// response stays an artifact (CR 400.7). "Can't be blocked this turn" is
// part of the same sentence, so it rides the same effect and ends with
// it. The engine supplies summoning sickness (CR 302.6): a Keyrune cast
// this turn can animate but can't attack, and can't tap for mana while it
// is a creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7685ff3f-6d61-4b1d-a381-49bc6274aa01",
		Name:         "Dimir Keyrune",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{U|B}",
			Label:    "Add {U} or {B}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}{B}: This artifact becomes a 2/2 blue and black Horror artifact creature until end of turn and can't be blocked this turn.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate | game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{U}{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := effects.NewContext(g, item)
				return effects.ScopedEffectFor{
					Target: item.SourceCardID,
					Mods: append([]game.Mod{
						game.AddTypesMod("Artifact", "Creature"),
						game.AddSubtypesMod(ctx.TextSubtypes("Horror")...),
						game.SetColorsMod(ctx.TextColors("U", "B")...),
						game.AddRestrictionsMod(game.CantBeBlocked),
					}, game.SetBasePTMods(2, 2)...),
					Duration: effects.DurationUntilEndOfTurn(ctx),
					Label:    "Dimir Keyrune — a 2/2 blue and black Horror artifact creature that can't be blocked until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}
