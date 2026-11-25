package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Crackdown Construct — Artifact Creature — Construct {4}, 2/2 (Aether
// Revolt, S82 older sets, #5536):
//
//	"Whenever you activate an ability of an artifact or creature that isn't
//	 a mana ability, this creature gets +1/+1 until end of turn."
//
// A mana ability never reaches the stack and carries its own event kind
// (CR 605.1a), so watching only EventActivateAbility is "isn't a mana
// ability". The source is read as it is on the battlefield; a loyalty
// ability of a planeswalker that is not also an artifact or creature does
// not count. Crackdown Construct's own abilities count (it is an artifact
// and a creature).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bccbd5ae-4e3f-43ee-a51b-d0ec2c645238",
		Name:         "Crackdown Construct",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventActivateAbility,
				func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
					return effects.ByYou(ev, source, lki, g) &&
						effects.ActivationSourceMatches(g, ev, effects.Or(effects.Artifact(), effects.Creature()))
				},
				"Crackdown Construct — gets +1/+1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.SourceIsStillThisPermanent(g, item) {
						return nil
					}
					return effects.BoostUntilEOT{Target: item.SourceCardID, Power: 1, Toughness: 1, Label: "Crackdown Construct — +1/+1"}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
