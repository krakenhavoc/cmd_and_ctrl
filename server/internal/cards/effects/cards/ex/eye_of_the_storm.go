package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eye of the Storm — Enchantment {5}{U}{U} (Ravnica: City of Guilds):
//
//	"Whenever a player casts an instant or sorcery card, exile it. Then that player
//	 copies each instant or sorcery card exiled with this enchantment. For each
//	 copy, the player may cast the copy without paying its mana cost."
//
// One trigger over every player's casts of an instant or sorcery card (a copy of
// a card that is cast counts, since it is a card, CR 707.12). As it resolves the
// spell is exiled from the stack and linked to this enchantment (CR 607.2a), then
// the caster copies EVERY instant and sorcery card exiled with the enchantment, the
// spell just exiled and the earlier ones, and is offered a free cast of each copy
// during the resolution (CR 608.2g). A copy that is cast triggers it again
// (the cast copy is exiled and every linked card is copied again), which is the
// card; the caster stops by declining, and a copy not cast ceases to exist
// (CR 704.5e). An Eye that has left the battlefield links nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "01f8db27-1191-488e-b0d5-172821613c15",
		Name:         "Eye of the Storm",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.CardID == uuid.Nil {
					return false
				}
				spell, ok := g.LookupCardForEffect(ev.CardID)
				return ok && (spell.IsInstant() || spell.IsSorcery())
			}, "Eye of the Storm — exile it, then that player copies each instant or sorcery card exiled with this enchantment and may cast the copies without paying their mana costs",
				effects.OldB579EyeOfTheStormResolve),
		},
	})
}
