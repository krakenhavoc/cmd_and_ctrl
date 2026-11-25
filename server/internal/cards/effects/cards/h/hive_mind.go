package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hive Mind — Enchantment {5}{U} (Secret Lair Drop, #5462):
//
//	"Whenever a player casts an instant or sorcery spell, each other player
//	 copies that spell. Each of those players may choose new targets for
//	 their copy."
//
// A cast trigger over every player's instant and sorcery spells. The caster
// is read off the event; each OTHER player still in the game, in turn order
// from the caster, makes a copy they control (CR 707.10) and is offered the
// CR 707.10c re-target prompt for it. The copy is made from the spell on the
// stack, or from its last-known information if it was countered in response
// to the trigger (the spell is named, not targeted, CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f97e405d-4c24-4d4d-9e19-e349c073113c",
		Name:         "Hive Mind",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, hiveMindCastInstantOrSorcery,
				"Hive Mind — each other player copies that spell and may choose new targets for their copy",
				hiveMindCopies),
		},
	})
}

// hiveMindCastInstantOrSorcery is "a player casts an instant or sorcery
// spell".
func hiveMindCastInstantOrSorcery(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor == uuid.Nil {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	return ok && (spell.IsInstant() || spell.IsSorcery())
}

// hiveMindCopies has every other player copy the spell, in turn order from
// the caster.
func hiveMindCopies(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ctx := effects.NewContext(g, item)
	spell, caster := item.Trigger.Event.CardID, item.Trigger.Event.Actor
	n := len(g.Seats)
	start := -1
	for i, p := range g.Seats {
		if p != nil && p.ID == caster {
			start = i
		}
	}
	if start < 0 {
		return nil
	}
	for k := 1; k < n; k++ {
		p := g.Seats[g.SeatInTurnOrderForEffect(start, k)]
		if p == nil || p.Eliminated || p.ID == caster {
			continue
		}
		if err := (effects.CopySpell{StackID: spell, Controller: p.ID, ChooseNewTargets: true, FromLastKnown: true}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
