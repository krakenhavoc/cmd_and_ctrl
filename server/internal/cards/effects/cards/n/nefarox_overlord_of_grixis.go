package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nefarox, Overlord of Grixis — Legendary Creature — Demon {4}{B}{B}, 5/5
// (Magic 2013):
//
//	"Flying
//	 Exalted (Whenever a creature you control attacks alone, that creature gets
//	 +1/+1 until end of turn.)
//	 Whenever Nefarox attacks alone, defending player sacrifices a creature of
//	 their choice."
//
// Battlegrace Angel's exalted keyword, and Thraximundar's edict on an attack
// that is a lone declaration (CR 506.5): the defending player is the player
// attacked, or the controller of the planeswalker or battle attacked
// (CR 506.3d); they choose the creature as the trigger resolves, and a player
// with none sacrifices nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0632f0f9-8d4f-4e94-8b53-1275b7f21a57",
		Name:            "Nefarox, Overlord of Grixis",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", game.KeywordExalted},
		Triggered: []game.TriggeredAbility{
			effects.TriggerEdict(effects.On(game.EventAttack,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return ev.Kind == game.EventAttack && ev.CardID == source.InstanceID && game.AttackedAlone(g)
				},
				"Nefarox — defending player sacrifices a creature of their choice",
				func(g *game.Game, item *game.StackItem) error {
					victim := effects.DefendingPlayerOfAttack(g, item.Trigger.Event)
					if victim == uuid.Nil || g.PlayerByIDForEffect(victim) == nil {
						return nil
					}
					g.PlayerSacrificesForEffect(item.SourceCardID, victim, effects.SacrificeSpec("a creature", effects.Creature()),
						"Nefarox — sacrifice a creature")
					return nil
				})),
		},
	})
}
