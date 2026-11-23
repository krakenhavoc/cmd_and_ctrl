package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Poxwalkers — Creature — Zombie {2}{B}, 3/1 (Warhammer 40,000 Commander,
// S82 GAP batch 188, #5948):
//
//	"Deathtouch
//	 Curse of the Walking Pox — Whenever you cast a spell from anywhere
//	 other than your hand, return this card from your graveyard to the
//	 battlefield tapped."
//
// The trigger works from the graveyard only (CR 113.6k), on any spell its
// controller casts from a zone other than their hand (flashback, exile, the
// command zone, a library top); the card comes back as a new object,
// tapped, under its owner's control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e4275202-4960-4cde-b941-27be8163ad6e",
		Name:            "Poxwalkers",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.InGraveyard(effects.On(game.EventCast, func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
				return effects.YouCast(nil)(ev, source, ch, g) && ev.OldZone != game.ZoneHand
			}, "Poxwalkers — return this card from your graveyard to the battlefield tapped",
				func(g *game.Game, item *game.StackItem) error {
					if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneGraveyard {
						return nil
					}
					_, err := g.ReturnToBattlefieldForEffect(item.SourceCardID, uuid.Nil, true)
					return err
				})),
		},
	})
}
