package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Magnetic Mine — Artifact {4} (Mirrodin Besieged, S82 OLD b467, #5576):
//
//	"Whenever another artifact is put into a graveyard from the battlefield,
//	 this artifact deals 2 damage to that artifact's controller."
//
// One trigger per other artifact that goes from the battlefield to a graveyard,
// anyone's (a bounce or an exile does not count; the Mine itself never counts,
// "another"). The damaged player is the artifact's controller as it last
// existed on the battlefield (CR 608.2h), fixed as the ability triggers and
// carried on the item. The damage is dealt by the Mine, so it can be
// prevented; a player who has left the game takes none.
//
// No simplification.
func init() {
	const label = "Magnetic Mine — 2 damage to that artifact's controller"
	t := effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return ev.CardID != source.InstanceID && effects.B23ArtifactPutIntoGraveyardFromBattlefield(ev, g)
	}, label, func(g *game.Game, item *game.StackItem) error {
		if g.PlayerByIDForEffect(item.Params.Player) == nil {
			return nil
		}
		return effects.DealDamage{Source: item.SourceCardID, Target: item.Params.Player, Amount: 2}.Apply(effects.NewContext(g, item))
	})
	t.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		if info, ok := g.LastKnownPermanentForEffect(ev.CardID); ok {
			item.Params.Player = info.Controller
		}
		return item
	}
	effects.Register(effects.Spec{
		OracleID:     "bbb50cf5-cd07-4ae8-bfa8-d747e44fb1ef",
		Name:         "Magnetic Mine",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}
