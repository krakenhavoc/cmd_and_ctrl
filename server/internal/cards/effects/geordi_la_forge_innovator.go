package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Geordi La Forge, Innovator — Legendary Creature — Human Artificer
// {2}{U}, 2/3 (Star Trek Commander, #4974):
//
//	"Geordi La Forge can't be blocked if you control a Spacecraft.
//	 Whenever Geordi La Forge deals combat damage to a player, draw a
//	 card."
//
// The evasion is a block rule read when blocks are declared
// (CantBeBlockedWhile): while Geordi's controller controls a permanent
// with the Spacecraft subtype (post-layer, through the text so a text
// change follows), no creature can block him. One that stops being so
// after blocks changes nothing (CR 509.1h). The draw is the ordinary
// combat-damage trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "adbef171-3640-4f77-8c92-b50c109ca303",
		Name:         "Geordi La Forge, Innovator",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		BlockRules: []game.BlockRule{CantBeBlockedWhile(OnSelf(), func(g *game.Game, _ uuid.UUID, c game.Card) bool {
			return ControlsA(c.TextSubtype("Spacecraft"))(g, c.Controller)
		})},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Geordi La Forge, Innovator — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
