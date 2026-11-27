package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rimehorn Aurochs — Snow Creature — Aurochs {4}{G}, 3/3 (Coldsnap, S82 OLD b560):
//
//	"Trample
//	 Whenever this creature attacks, it gets +1/+0 until end of turn for each
//	 other attacking Aurochs.
//	 {2}{S}: Target creature blocks target creature this turn if able."
//
// The first target must block the second if it can (Hunt Down's requirement,
// CR 509.1c); {S} is paid with snow mana only.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b760b1db-764b-4b40-8a84-691b5c9c1dbc",
		Name:            "Rimehorn Aurochs",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.ThisAttacked,
				"Rimehorn Aurochs — +1/+0 until end of turn for each other attacking Aurochs",
				effects.OldB560PumpPerOtherAttackingAurochs("Rimehorn Aurochs — +1/+0 for each other attacking Aurochs")),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "{2}{S}: Target creature blocks target creature this turn if able.",
			Cost:  effects.ManaCost("{2}{S}"),
			Targets: effects.Clauses(
				effects.TargetCreature("target creature"),
				effects.TargetCreature("target creature"),
			),
			Effect: effects.OldB560TargetBlocksTarget("Rimehorn Aurochs — blocks the other target creature if able"),
		}},
	})
}
