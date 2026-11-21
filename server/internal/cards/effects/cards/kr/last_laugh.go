package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Last Laugh — Enchantment {2}{B}{B}:
//
//	"Whenever a permanent other than this enchantment is put into a
//	 graveyard from the battlefield, this enchantment deals 1 damage to
//	 each creature and each player.
//	 When no creatures are on the battlefield, sacrifice this
//	 enchantment."
//
// ADR 0107 §1 (#1858).
//
//   - The damage trigger is a leaves-the-battlefield trigger (CR 603.6c,
//     603.10a): any permanent, anyone's, a token included — a token is
//     put into a graveyard before it ceases to exist. Each one is its own
//     trigger, so a wipe of five creatures deals 5 to everything.
//   - The sacrifice is a CR 603.8 state trigger over every creature on
//     the battlefield. Casting Last Laugh onto an empty board triggers it
//     at once.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5facb256-b993-43f8-b971-b3a59a7434bf",
		Name:         "Last Laugh",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.NewZone == game.ZoneGraveyard && ev.CardID != source.InstanceID
			}, "Last Laugh — 1 damage to each creature and each player", func(g *game.Game, item *game.StackItem) error {
				return effects.B23DamageEachCreatureAndEachPlayer(effects.NewContext(g, item), 1)
			}),
			effects.WhenThereAreNo(effects.QueryType("creature"), "Last Laugh — sacrifice it", effects.SacrificeThisIfStillOnBattlefield),
		},
	})
}
