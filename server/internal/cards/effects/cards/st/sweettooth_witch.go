package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sweettooth Witch — Creature — Human Warlock {2}{B}, 3/2 (Wilds of
// Eldraine, #4471):
//
//	"When this creature enters, create a Food token.
//	 {2}, Sacrifice a Food: Target player loses 2 life."
//
// The entry Food is the token table's. The activated cost is Greta,
// Sweettooth Scourge's "Sacrifice a Food" (any Food you control, read by
// its subtype, so a non-Food artifact can't pay it), and the drain is the
// shared TargetPlayerLosesLife. There is no tap symbol, so it can be
// activated while the Witch is summoning sick and more than once a turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "635b854d-de71-4c85-9755-8c97bf3cba36",
		Name:         "Sweettooth Witch",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sweettooth Witch — create a Food token", effects.Do(effects.CreateToken{Template: effects.FoodToken(), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, Sacrifice a Food: Target player loses 2 life.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), game.AbilityCost{SacrificeOther: effects.SacrificeSpec("a Food", effects.HasSubtype("Food"))}),
			Targets: effects.TargetPlayer("target player"),
			Effect:  effects.Do(effects.TargetPlayerLosesLife{N: 2}),
		}},
	})
}
