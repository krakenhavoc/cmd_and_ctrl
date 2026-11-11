package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cunning Rhetoric — Enchantment {2}{B} (Outlaws of Thunder Junction
// Commander):
//
//	"Whenever an opponent attacks you and/or one or more planeswalkers you
//	 control, exile the top card of that player's library. You may play that
//	 card for as long as it remains exiled, and you may spend mana as though
//	 it were mana of any color to cast it."
//
// One trigger per attacking opponent per attack declaration (CR 603.2c),
// however many creatures or planeswalkers it names: the batch key is the
// attacker. An attack at a battle you protect is neither you nor a
// planeswalker (CR 508.3e), so it does not trigger. The card is exiled face
// up and played with any-color mana (not any type).
func init() {
	Register(Spec{
		OracleID:     "e7eb7ac8-08ab-419b-8508-987afd0bb93b",
		Name:         "Cunning Rhetoric",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			func() game.TriggeredAbility {
				t := On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return ev.Kind == game.EventAttack && ev.Actor != uuid.Nil && ev.Actor != source.Controller &&
						attacksYouOrYourPlaneswalker(g, ev.Target, source.Controller)
				}, "Cunning Rhetoric — exile the top card of that player's library; you may play it",
					func(g *game.Game, item *game.StackItem) error {
						if item.Trigger == nil {
							return nil
						}
						return ExileTopWithPermission{
							From:        item.Trigger.Event.Actor,
							GrantTo:     item.Controller,
							N:           1,
							AnyColor:    true,
							WhileExiled: true,
						}.Apply(NewContext(g, item))
					})
				t.OncePerBatch = true
				t.BatchKey = func(ev game.Event, _ *game.Card, _ *game.Game) string { return ev.Actor.String() }
				return t
			}(),
		},
	})
}
