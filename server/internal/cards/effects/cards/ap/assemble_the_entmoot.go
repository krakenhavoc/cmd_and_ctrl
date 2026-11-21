package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Assemble the Entmoot — Enchantment {3}{G} (Tales of Middle-earth
// Commander):
//
//	"Creatures you control have reach.
//	 Sacrifice this enchantment: Create three tapped X/X green Treefolk
//	 creature tokens, where X is the amount of life you gained this turn.
//	 Put a reach counter on each of them."
//
// The static is the shared "creatures you control have <keyword>" grant. X is
// the life gained this turn by the ability's controller (the turn's tally,
// not net of life lost), read as the ability resolves, so the Entmoot having
// been sacrificed as a cost does not change it; an X of zero makes tokens
// that die at once (CR 704.5f), as printed. The Treefolk is the printed
// Treefolk token (*/* green) given its size as it is created, and the reach
// counters are on each token as it enters, so none is ever without them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1f1ad44c-1811-40e3-976b-03f19866abd5",
		Name:         "Assemble the Entmoot",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.TribalKeywordGrant(effects.TribeFilter{YoursOnly: true}, "reach")},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this enchantment: Create three tapped X/X green Treefolk creature tokens, where X is the amount of life you gained this turn. Put a reach counter on each of them.",
			Purpose: game.Purpose{Tokens: 3, Answers: game.AnswerValue},
			Cost:    effects.SacrificeThis(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				x := effects.B15LifeGainedThisTurn(g, item.Controller)
				// The printed Treefolk token (*/* green) with its size set as it is
				// created, Arboreal Alliance's shape.
				tmpl := game.Card{Name: "Treefolk", TypeLine: "Token Creature — Treefolk", Colors: []string{"G"}, Power: x, Toughness: x}
				return effects.CreateTokenAdvanced{
					Spec: effects.Token(tmpl).EntersTapped().WithCounters(game.CounterReach, 1),
					N:    3,
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
