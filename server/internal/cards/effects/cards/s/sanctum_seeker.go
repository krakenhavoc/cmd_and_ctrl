package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sanctum Seeker — Creature — Vampire Knight {2}{B}{B}, 3/4 (EDHREC
// rank 2320):
//
//	"Whenever a Vampire you control attacks, each opponent loses 1
//	 life and you gain 1 life."
//
// The Vampire deck's attack drain. One trigger PER attacking Vampire
// — the printed text is "a Vampire", not "one or more", and the
// engine's per-creature EventAttack is exactly that — the Seeker
// itself included, since it is a Vampire and the text says "a
// Vampire you control", not "another". Effective subtypes, so a
// changeling counts. Each trigger is a footprinted drain (#3013): every
// opponent loses 1 (life loss, not damage) and the controller gains
// 1 — once, not once per opponent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "afb71560-0fc9-4ea5-9d52-d93c17d72519",
		Name:         "Sanctum Seeker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B21CreatureOfSubtypeYouControlAttacked(ev, source, g, "Vampire")
			}, "Sanctum Seeker — each opponent loses 1 life, you gain 1 life", effects.Do(effects.EachOpponentLosesLife{N: 1}, effects.GainLife{Amount: 1})),
		},
	})
}
