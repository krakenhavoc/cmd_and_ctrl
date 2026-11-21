package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Liberty Prime, Recharged — Legendary Artifact Creature — Robot
// {2}{U}{R}{W}, 8/8:
//
//	"Vigilance, trample, haste
//	 Whenever Liberty Prime attacks or blocks, sacrifice it unless you
//	 pay {E}{E} (two energy counters).
//	 {2}, {T}, Sacrifice an artifact: You get {E}{E} and draw a card."
//
// ADR 0129 §3 (#1995): the attack-or-block trigger is CR 118.12a's
// pay-unless with an energy payment, held in the combat step it was
// asked in; declining, or being short (CR 118.3), sacrifices Liberty
// Prime if it is still the creature that triggered (CR 400.7). The
// activated ability's sacrifice may name Liberty Prime itself, an
// artifact.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b31026fe-faad-49b3-93b0-1324e32bb816",
		Name:            "Liberty Prime, Recharged",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance", "trample", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventAttack, game.EventBlock}, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclared(ev, source) || effects.SelfBlocksOnce(ev, source)
			}, "Liberty Prime — sacrifice it unless you pay {E}{E}",
				effects.SacrificeThisUnlessYouPayEnergy("Liberty Prime", "Liberty Prime", 2)),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice an artifact: You get {E}{E} and draw a card.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeN(1, "an artifact", effects.Artifact())),
			Purpose: game.Purpose{Answers: game.AnswerValue, Energy: 2, Draws: 1},
			Effect:  effects.Do(effects.GetEnergy{N: 2}, effects.DrawCards{N: 1}),
		}},
	})
}
