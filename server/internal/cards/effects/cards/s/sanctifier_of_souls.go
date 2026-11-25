package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sanctifier of Souls — {3}{W} Creature — Human Cleric 2/3 (Eldritch Moon,
// S82 OLD, #5541):
//
//	"Whenever another creature you control enters, this creature gets +1/+1
//	 until end of turn.
//	 {2}{W}, Exile a creature card from your graveyard: Create a 1/1 white
//	 Spirit creature token with flying."
//
// Beast-Kin Ranger's trigger with +1/+1, and Moorland Haunt's exile-a-creature
// -card cost around the shared white Spirit.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "806fc324-a17a-4e65-a9b6-d739c471c7f8",
		Name:         "Sanctifier of Souls",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverAnotherCreatureEntersUnderYourControl("Sanctifier of Souls — +1/+1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.BoostUntilEOT{Target: ctx.Source(), Power: 1, Toughness: 1, Label: "Sanctifier of Souls — +1/+1"}.Apply(ctx)
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}, Exile a creature card from your graveyard: Create a 1/1 white Spirit creature token with flying.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1, CreatureTokens: 1},
			Cost:    effects.Plus(effects.ManaCost("{2}{W}"), effects.ExileFromGraveyard(1, "a creature card", effects.MatchCreature)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{Controller: item.Controller, Template: effects.B28WhiteSpiritFlyingToken(), N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
