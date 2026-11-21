package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Explorer's Scope — Artifact — Equipment {1}:
//
//	"Whenever equipped creature attacks, look at the top card of your
//	 library. If it's a land card, you may put it onto the battlefield
//	 tapped.
//	 Equip {1}"
//
// An attack trigger on the Equipment, keyed on the creature it is
// attached to (attachedCreatureAttacked, Argentum Armor's condition).
// The look is private to the controller; a land may go onto the
// battlefield tapped through #745's library-to-battlefield move, which
// is a put rather than a land play (CR 305.4). Anything else — a
// nonland card, or a declined land — stays on top of the library, and
// the controller keeps knowing what it is, because they looked.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a563ede9-b92f-4285-88f8-abcbdd017742",
		Name:         "Explorer's Scope",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack,
				func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return effects.AttachedCreatureAttacked(ev, source)
				},
				"Explorer's Scope — look at the top card; you may put a land onto the battlefield tapped",
				func(g *game.Game, item *game.StackItem) error {
					return effects.PutFromLibraryOntoBattlefield{
						Player:   item.Controller,
						Cards:    g.LookAtTopOfLibraryForEffect(item.Controller, 1),
						Match:    effects.Land(),
						Max:      1,
						Optional: true,
						Tapped:   true,
						Label:    "Explorer's Scope — you may put the land onto the battlefield tapped",
					}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{1}"),
		},
	})
}
