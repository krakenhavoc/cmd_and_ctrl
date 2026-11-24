package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dragonlord Kolaghan — Legendary Creature — Elder Dragon {4}{B}{R}, 6/5
// (Secret Lair Drop, reversible printing):
//
//	"Flying, haste
//	 Other creatures you control have haste.
//	 Whenever an opponent casts a creature or planeswalker spell with the same
//	 name as a card in their graveyard, that player loses 10 life."
//
// The haste grant is Tannuk's. The trigger watches every opponent's cast, reads
// the spell's name as it sits on the stack (the face being cast, CR 201.2) and
// asks whether any card in THAT player's own graveyard has the same name, with
// the types read on the spell. The loser is the caster, as it resolves, and the
// loss is Kolaghan's. A copy of a spell is not cast, so it never triggers.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5ff89859-ddfd-48e5-8423-57b3956173ac",
		Name:            "Dragonlord Kolaghan",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{Others: true, YoursOnly: true}, "haste"),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventCast || ev.Actor == source.Controller || ev.Actor == uuid.Nil {
					return false
				}
				spell, ok := g.LookupCardForEffect(ev.CardID)
				if !ok || !(spell.IsCreature() || spell.HasCardType("planeswalker")) {
					return false
				}
				p := g.PlayerByIDForEffect(ev.Actor)
				if p == nil || p.Eliminated || p.Graveyard == nil {
					return false
				}
				for _, c := range p.Graveyard.Cards {
					if c.Name == spell.Name {
						return true
					}
				}
				return false
			}, "Dragonlord Kolaghan — that player loses 10 life", func(g *game.Game, item *game.StackItem) error {
				return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Trigger.Event.Actor, -10)
			}),
		},
	})
}
