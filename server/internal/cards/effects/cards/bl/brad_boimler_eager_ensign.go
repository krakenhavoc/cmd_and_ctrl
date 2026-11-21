package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brad Boimler, Eager Ensign — Legendary Creature — Human Officer
// {1}{W}, 2/2 (Star Trek, #4973):
//
//	"Lifelink
//	 Whenever Brad Boimler becomes tapped, until end of turn, if one or
//	 more counters would be put on a permanent you control, that many
//	 plus one of each of those kinds of counters are put on that
//	 permanent instead."
//
// "Becomes tapped" is ThisBecameTapped over both tap events, so an
// attack taps him (and a vigilance attack does not). The resolved
// ability makes a turn-scoped replacement record (game.ModOneMoreCounter,
// ADR 0041 P8): Pir, Imaginative Rascal's +1 for the rest of the turn,
// on permanents the trigger's controller controls, whoever puts the
// counters. It does not depend on Brad staying on the battlefield, and
// tapping him twice makes two records, so two more counters (CR 616:
// each applies once to the event).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "10af9cd9-1700-48f9-97e1-61e239536fef",
		Name:            "Brad Boimler, Eager Ensign",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, effects.ThisBecameTapped,
				"Brad Boimler — this turn, one more of each kind of counter on your permanents",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					ctx.Game.OneMoreCounterThisTurnForEffect(ctx.Item.SourceCardID, ctx.Controller(),
						"Brad Boimler, Eager Ensign — one more counter of each kind")
					return nil
				}))),
		},
	})
}
