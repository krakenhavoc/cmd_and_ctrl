package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Duskwatch Recruiter // Krallenhorde Howler — {1}{G} Creature — Human
// Warrior Werewolf 2/2 // Creature — Werewolf 3/3 (#4071, ADR 0132
// amendment 2026-10-16):
//
//	Front: "{2}{G}: Look at the top three cards of your library. You may
//	        reveal a creature card from among them and put it into your
//	        hand. Put the rest on the bottom of your library in any order.
//	        At the beginning of each upkeep, if no spells were cast last
//	        turn, transform this creature."
//	Back:  "Creature spells you cast cost {1} less to cast.
//	        At the beginning of each upkeep, if a player cast two or more
//	        spells last turn, transform this creature."
//
// The look is private and the creature taken is revealed; the player
// orders the rest on the bottom (Goblin Ringleader's "any order"). The
// back face's discount is an ordinary cost modifier over your creature
// spells, live only while the Howler is face up.
//
// No simplification.
func init() {
	const oracle = "87c43776-dd98-4e12-9b95-82d2b8f4f1ab"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Duskwatch Recruiter",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}: Look at the top three cards of your library. You may reveal a creature card from among them and put it into your hand. Put the rest on the bottom of your library in any order.",
			Cost:    effects.ManaCost("{2}{G}"),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				player := ctx.Controller()
				return effects.TakeFromLibraryToHand{
					Player:   player,
					Cards:    g.LookAtTopOfLibraryForEffect(player, 3),
					Match:    effects.Creature(),
					Max:      1,
					Optional: true,
					Label:    "Duskwatch Recruiter — you may reveal a creature card from among them and put it into your hand",
					Key:      effects.TakeRevealRestBottomAny,
				}.Apply(ctx)
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.TransformAtUpkeepIfNoSpellsLastTurn("Duskwatch Recruiter"),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Krallenhorde Howler",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Creature spells you cast cost {1} less to cast.", effects.YourSpell(), effects.CreatureSpell()),
		},
		Triggered: []game.TriggeredAbility{
			effects.TransformAtUpkeepIfAPlayerCastTwoSpellsLastTurn("Krallenhorde Howler"),
		},
	})
}
