package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elmar, Ulvenwald Informant — Legendary Creature — Human {1}{R}{G}, 3/2
// (Secret Lair Drop):
//
//	"Haste
//	 Whenever you cast your second spell each turn, untap target creature,
//	 then investigate.
//	 Partner—Friends forever (You can have two commanders if both have this
//	 ability.)"
//
// The condition is Goldbug's (YouCastYourSecondSpellEachTurn: the cast tally is
// bumped before the event, so it fires once, on the cast that takes the turn's
// count to exactly two). The target is chosen as the ability goes on the stack;
// if it is illegal on resolution the whole ability does nothing, the investigate
// included (CR 608.2b).
// Partner—Friends forever is a deck-construction rule read off the oracle text.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5a35201d-e9b1-433a-bb70-3280ceedbcd6",
		Name:            "Elmar, Ulvenwald Informant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Purpose:         game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventCast, effects.YouCastYourSecondSpellEachTurn,
				"Elmar, Ulvenwald Informant — untap target creature, then investigate",
				func(g *game.Game, item *game.StackItem) error {
					if err := effects.UntapTheTarget(g, item); err != nil {
						return err
					}
					return effects.Investigate{N: 1}.Apply(effects.NewContext(g, item))
				}),
				effects.TargetCreature("target creature")),
		},
	})
}
