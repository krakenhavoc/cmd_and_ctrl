package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ahn-Crop Invader — Creature — Zombie Minotaur Warrior {2}{R}, 2/2 (War of the Spark, #5520):
//
//	"During your turn, this creature has first strike.
//	 {1}, Sacrifice another creature: This creature gets +2/+0 until end of
//	 turn."
//
// Blood Burglar's turn-gated keyword (layer 6) and Blood Host's
// sacrifice-another-creature cost; the pump goes to the Invader if it is still
// on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d7d163a7-6f3f-40ab-a47c-ed497df92244",
		Name:         "Ahn-Crop Invader",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return effects.ThisPermanent(target, g, source) && effects.IsActivePlayer(g, source.Controller)
			}, "first strike"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice another creature: This creature gets +2/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeAnotherN(1, "another creature", effects.Creature())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.OnBattlefield(g, item.SourceCardID) {
					return nil
				}
				return effects.BoostUntilEOT{Target: item.SourceCardID, Power: 2, Label: "Ahn-Crop Invader — +2/+0"}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
