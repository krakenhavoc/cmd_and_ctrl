package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Molder Beast — Creature — Beast {4}{G}, 5/3 (Commander Legends):
//
//	"Trample
//	 Whenever an artifact is put into a graveyard from the battlefield, this
//	 creature gets +2/+0 until end of turn."
//
// Disciple of the Vault's condition: any artifact, anyone's, token or card,
// as it went from the battlefield to a graveyard. One trigger per artifact,
// each +2/+0 on top of the last (CR 603.2c), and nothing if the Beast itself
// has gone by resolution.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2b732c6a-f09c-4ac0-ae3a-1d8820b69aba",
		Name:            "Molder Beast",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B23ArtifactPutIntoGraveyardFromBattlefield(ev, g)
			}, "Molder Beast — gets +2/+0 until end of turn", func(g *game.Game, item *game.StackItem) error {
				return effects.BoostUntilEOT{Target: item.SourceCardID, Power: 2, Label: "Molder Beast — +2/+0 until end of turn"}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
