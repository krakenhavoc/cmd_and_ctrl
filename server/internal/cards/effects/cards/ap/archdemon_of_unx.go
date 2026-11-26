package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Archdemon of Unx — {5}{B}{B} Creature — Demon 6/6 (Shards of Alara, S82 OLD
// b516):
//
//	"Flying, trample
//	 At the beginning of your upkeep, sacrifice a non-Zombie creature, then
//	 create a 2/2 black Zombie creature token."
//
// Lord of the Pit's sacrifice choice (any creature that is not a Zombie, the
// Archdemon itself included), then the Zombie for the controller whether or not
// anything was sacrificed: the "then" is not conditional.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "41e250e6-8e38-4cbc-ae66-98800a5feba8",
		Name:            "Archdemon of Unx",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Archdemon of Unx — sacrifice a non-Zombie creature, then create a 2/2 black Zombie creature token",
				func(g *game.Game, item *game.StackItem) error {
					return g.PlayerSacrificesThenKeyedForEffect(item.SourceCardID, item.Controller,
						effects.SacrificeSpec("a non-Zombie creature", effects.Creature(), effects.Not(effects.Subtype("Zombie"))),
						"Archdemon of Unx — sacrifice a non-Zombie creature", 1,
						archdemonOfUnxZombie, []uuid.UUID{item.Controller})
				}),
		},
	})
}

var archdemonOfUnxZombie = game.RegisterPromptRunThen("prompt-run/old-b516-archdemon-of-unx-zombie", func(g *game.Game, r game.RunDone) error {
	if len(r.Carry) == 0 {
		return nil
	}
	return effects.CreateToken{Template: effects.TokenCard("2/2 black Zombie"), N: 1}.Apply(effects.KeyedContext(g, r.Carry[0], r.Source))
})
