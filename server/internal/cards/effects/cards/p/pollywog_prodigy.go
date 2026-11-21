package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pollywog Prodigy — Creature — Frog Wizard {1}{U}, 1/3:
//
//	"Evolve (Whenever a creature you control enters, if that creature
//	 has greater power or toughness than this creature, put a +1/+1
//	 counter on this creature.)
//	 Whenever an opponent casts a noncreature spell with mana value less
//	 than this creature's power, draw a card."
//
// The comparison is part of the trigger condition, so it is made once,
// as the spell is cast: the spell's mana value where it is (on the
// stack, X counted as announced — ManaValueForEffect, CR 202.3e)
// against the Prodigy's power right then, counters included. There is
// no "if", so nothing is re-checked on resolution (CR 603.4 only
// applies to an intervening "if"). A spell whose cost can't be read is
// not counted. Evolve is the engine's keyword trigger (game/evolve.go,
// #1805), and every +1/+1 counter it adds widens the net.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "99a3cea9-f4ab-4da3-a085-c09dc93fc8cb",
		Name:            "Pollywog Prodigy",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordEvolve},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, PollywogProdigyApplies, "Pollywog Prodigy — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}

// pollywogProdigyApplies is "an opponent casts a noncreature spell with
// mana value less than this creature's power".
func PollywogProdigyApplies(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !effects.ByAnOpponent(ev, source, game.Characteristic{}, g) || ev.CardID == uuid.Nil {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || spell.IsCreature() {
		return false
	}
	mv, ok := g.ManaValueForEffect(spell)
	return ok && mv < source.PowerForComparison()
}
