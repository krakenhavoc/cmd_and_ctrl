package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hawkeye, Avenging Archer — Legendary Creature — Human Archer Hero
// {2}{U}{R}, 3/4 (Marvel Super Heroes Commander, #4455):
//
//	"Reach
//	 Whenever a creature an opponent controls dies, if Hawkeye dealt damage
//	 to it this turn, draw a card.
//	 {T}: Hawkeye deals 1 damage to any target."
//
// The draw has an intervening if (CR 603.4): Hawkeye dealt damage to that
// creature this turn, read from the turn's damage record (Zurgo Helmsmasher's
// dealtDamageByThisTurn, a victim and a source each the object it is now or
// last was), checked as it dies. Damage already dealt this turn cannot stop
// having been dealt, so the resolution-time recheck CR 603.4 allows could never
// change the answer and the draw is a plain Do step. The ping is Prodigal
// Sorcerer's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e90d6903-5415-4ddf-88c0-15ddd8614bde",
		Name:            "Hawkeye, Avenging Archer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B18OpponentsCreatureDied(ev, source, g) && effects.DealtDamageByThisTurn(g, source.InstanceID, ev.CardID)
			}, "Hawkeye, Avenging Archer — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Hawkeye deals 1 damage to any target.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(1),
		}},
	})
}
