package cardsto

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Torrent Elemental — {4}{U} Creature — Elemental 3/5 (Fate Reforged):
//
//	"Flying
//	 Whenever this creature attacks, tap all creatures defending player controls.
//	 {3}{B/G}{B/G}: Put this card from exile onto the battlefield tapped.
//	 Activate only as a sorcery."
//
// The defending player is the one attacked, or the controller of the
// planeswalker or battle attacked (CR 506.3d); every creature they control is
// tapped as the trigger resolves. The second ability works from exile (CR
// 113.6): it puts the card onto the battlefield tapped under the activating
// player's control, and does nothing for a card that has left exile and come
// back (CR 400.7).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0335e7db-5e0e-4271-8ae7-437a1d2c16ac",
		Name:            "Torrent Elemental",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Torrent Elemental — tap all creatures defending player controls",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					defender := effects.DefendingPlayerOfAttack(g, item.Trigger.Event)
					if defender == uuid.Nil {
						return nil
					}
					ctx := effects.NewContext(g, item)
					for _, id := range effects.B751CreatureIDs(g, defender) {
						if err := (effects.TapTarget{Target: id}).Apply(ctx.AsGroupMember()); err != nil {
							return err
						}
					}
					return nil
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{3}{B/G}{B/G}: Put this card from exile onto the battlefield tapped. Activate only as a sorcery.",
			Cost:         effects.ManaCost("{3}{B/G}{B/G}"),
			Zones:        []game.ZoneKind{game.ZoneExile},
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				if effects.SourceIsNewObject(g, item) {
					return nil
				}
				if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneExile {
					return nil
				}
				_, err := g.ReturnFromExileToBattlefieldForEffect(item.SourceCardID, item.Controller, true)
				return err
			},
		}},
	})
}
