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
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBeginUpkeep},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Bitterbloom Bearer — lose 1 life, create a Faerie",
					func(g *game.Game, item *game.StackItem) error {
						if err := g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -1); err != nil {
							return err
						}
						return CreateToken{
							Controller: item.Controller,
							Template:   b18BlueBlackFaerieToken(),
							N:          1,
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
