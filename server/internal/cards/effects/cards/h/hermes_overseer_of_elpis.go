package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// b33HermesScryLabel is the stack label of Hermes' attack trigger —
// the "one or more" dedup keys on it.
const b33HermesScryLabel = "Hermes, Overseer of Elpis — scry 2"

// Hermes, Overseer of Elpis — Legendary Creature — Elder Wizard
// {3}{U}, 2/4 (EDHREC rank 3518):
//
//	"Whenever you cast a noncreature spell, create a 1/1 blue Bird
//	 creature token with flying and vigilance.
//	 Whenever you attack with one or more Birds, scry 2."
//
// A Bird for every cantrip, and a scry when the flock attacks. The
// cast trigger is b10NoncreatureSpellCastByYou — the spell's type is
// read off the stack, and it fires on the cast, so a countered spell
// still makes a Bird. The attack trigger is "one or more": the
// engine emits one EventAttack per attacker, so the first Bird
// declared fires it and the rest of that combat's Birds are declined
// as later events of the same batch (OncePerBatch; see docs/adding-cards.md). Effective subtypes, so a changeling attacking is a Bird.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "63e2cba7-ec1a-44bf-b915-d2ae75851430",
		Name:         "Hermes, Overseer of Elpis",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Noncreature(), "Hermes, Overseer of Elpis — create a 1/1 blue Bird with flying and vigilance", effects.Do(effects.CreateToken{Template: effects.B33BlueBirdVigilanceToken(), N: 1})),
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B33BirdYouControlAttacked(ev, source, g)
			}, b33HermesScryLabel, effects.Do(effects.Scry{N: 2}))),
		},
	})
}
