package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mutavault — Land (Star Trek Commander, #4974):
//
//	"{T}: Add {C}.
//	 {1}: This land becomes a 2/2 creature with all creature types until
//	 end of turn. It's still a land."
//
// A manland built like Mobilized District: the animation adds the
// Creature type, every creature type (AllCreatureTypesMod, layer 4) and
// a 2/2 base P/T (layer 7b) until end of turn. It does not gain
// changeling (ruling 2008-04-01); the land keeps its colour (colourless)
// and its mana ability. Summoning sickness reads continuous control
// since the turn began, not how long it has been a creature (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6b3cc59a-7ea5-4eb5-9bf9-5a9c07f80e2b",
		Name:         "Mutavault",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}: This land becomes a 2/2 creature with all creature types until end of turn. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.ManaCost("{1}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := effects.NewContext(g, item)
				return effects.ScopedEffectFor{
					Target: item.SourceCardID,
					Mods: []game.Mod{
						game.AddTypesMod("Creature"),
						game.AllCreatureTypesMod(),
						game.SetBasePowerMod(2),
						game.SetBaseToughnessMod(2),
					},
					Duration: effects.DurationUntilEndOfTurn(ctx),
					Label:    "Mutavault — a 2/2 creature with all creature types until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}
