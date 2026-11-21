package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crystal Grotto — Land:
//
//	"When this land enters, scry 1.
//	 {T}: Add {C}.
//	 {1}, {T}: Add one mana of any color."
//
// Conduit Pylons' twin with scry in place of surveil: an untapped
// land whose ETB trigger goes on the stack, a colourless tap, and a
// {1}, {T} filter into the five-colour pipe.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f15fb0cc-8e96-4f03-94d0-b51410415afd",
		Name:         "Crystal Grotto",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			effects.PainlessColorless(),
			{
				Cost:     effects.ManaAbilityCost{Tap: true, Mana: "{1}"},
				Produced: "{W|U|B|R|G}",
				Label:    "{1}, {T}: Add one mana of any color",
			},
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Crystal Grotto — scry 1", effects.Do(effects.Scry{N: 1})),
		},
	})
}
