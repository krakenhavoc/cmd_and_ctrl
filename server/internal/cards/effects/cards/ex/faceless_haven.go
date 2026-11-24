package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Faceless Haven — Snow Land (Kaldheim, #5505):
//
//	"{T}: Add {C}.
//	 {S}{S}{S}: This land becomes a 4/3 creature with vigilance and all
//	 creature types until end of turn. It's still a land. ({S} can be paid
//	 with one mana from a snow source.)"
//
// Mutavault's manland, paid in snow mana (CR 107.4h: only mana from a snow
// source pays {S}). The animation is a layered effect on the land for the
// turn (CR 613): it keeps its land type and gains Creature, every
// creature type, vigilance and a base 4/3.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f74107d5-fb4a-464b-9251-42b84d91775d",
		Name:         "Faceless Haven",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{S}{S}{S}: This land becomes a 4/3 creature with vigilance and all creature types until end of turn. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.ManaCost("{S}{S}{S}"),
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
						game.SetBasePowerMod(4),
						game.SetBaseToughnessMod(3),
						game.AddKeywordsMod("vigilance"),
					},
					Duration: effects.DurationUntilEndOfTurn(ctx),
					Label:    "Faceless Haven — a 4/3 creature with vigilance and all creature types until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}
