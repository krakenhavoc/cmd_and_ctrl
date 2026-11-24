package cardsgr

import (
	"errors"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grenzo, Dungeon Warden — Legendary Creature — Goblin Rogue {X}{B}{R}, 2/2
// (Secret Lair Drop):
//
//	"Grenzo enters with X +1/+1 counters on it.
//	 {2}: Put the bottom card of your library into your graveyard. If it's a
//	 creature card with power less than or equal to Grenzo's power, put it
//	 onto the battlefield."
//
// Cellar Door's bottom-card move (index 0 is the bottom); the creature's
// power is its own printed or current power as a card in the graveyard, and
// Grenzo's is read as the ability resolves (last known if he is gone).
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "a374f99e-6443-486c-8026-8eba4c7e6ad7",
		Name:                       "Grenzo, Dungeon Warden",
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.XCounters(game.CounterPlusOne)},
		XMatters:                   true,
		Completeness:               effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: Put the bottom card of your library into your graveyard. If it's a creature card with power less than or equal to Grenzo's power, put it onto the battlefield.",
			Cost:    effects.ManaCost("{2}"),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				p := g.PlayerByIDForEffect(item.Controller)
				if p == nil || p.Library == nil || len(p.Library.Cards) == 0 {
					return nil
				}
				// Index 0 is the bottom (Zone.Bottom).
				bottom := p.Library.Cards[0]
				if err := g.PutIntoGraveyardForEffect(bottom.InstanceID); err != nil {
					return err
				}
				if !bottom.IsCreature() {
					return nil
				}
				if bottom.Power > effects.B29PowerAsItLastStood(g, item.SourceCardID) {
					return nil
				}
				// A replacement that sent the card elsewhere leaves nothing here to return.
				if _, err := g.ReturnToBattlefieldForEffect(bottom.InstanceID, item.Controller, false); err != nil && !errors.Is(err, game.ErrCardNotFound) {
					return err
				}
				return nil
			},
		}},
	})
}
