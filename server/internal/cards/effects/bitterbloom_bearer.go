package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bitterbloom Bearer — Creature — Faerie Rogue {B}{B}, 1/1 (EDHREC
// rank 1988):
//
//	"Flash
//	 Flying
//	 At the beginning of your upkeep, you lose 1 life and create a
//	 1/1 blue and black Faerie creature token with flying."
//
// Bitterblossom on a body: the same upkeep trigger, life loss first
// and then the token. The token is a blue-and-black Faerie, not
// Bitterblossom's black Faerie Rogue, so it has its own template.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5effb651-f9e0-4c14-8192-bd0e132b8d5d",
		Name:            "Bitterbloom Bearer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Bitterbloom Bearer — lose 1 life, create a Faerie", func(g *game.Game, item *game.StackItem) error {
				if err := g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -1); err != nil {
					return err
				}
				return CreateToken{
					Controller: item.Controller,
					Template:   TokenCard("1/1 blue and black Faerie with flying"),
					N:          1,
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
