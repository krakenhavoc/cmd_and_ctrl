package cardsch

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chevill, Bane of Monsters — Legendary Creature — Human Rogue {B}{G}, 1/3 (Ikoria: Lair of Behemoths, #5512):
//
//	"Deathtouch
//	 At the beginning of your upkeep, if your opponents control no permanents
//	 with bounty counters on them, put a bounty counter on target creature or
//	 planeswalker an opponent controls.
//	 Whenever a permanent an opponent controls with a bounty counter on it
//	 dies, you gain 3 life and draw a card."
//
// The upkeep trigger has an intervening "if" (CR 603.4), asked as the upkeep
// begins and again as the trigger resolves; a permanent of any opponent with
// a bounty counter, whoever put it there, switches it off. The death trigger
// reads the counters the permanent had as it died, and who controlled it,
// from last-known information (CR 603.10a); any permanent counts, not only a
// creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "50863075-a30b-4238-8e31-5af32a39d886",
		Name:            "Chevill, Bane of Monsters",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.InterveningIfYou("your opponents control no permanents with bounty counters",
					func(g *game.Game, you uuid.UUID) bool { return !effects.S82OldB113BountyOnOpponentsPermanent(g, you) },
					effects.AtYourUpkeep("Chevill, Bane of Monsters — put a bounty counter on target creature or planeswalker an opponent controls",
						effects.PutACounterOnTheTarget("bounty"))),
				effects.TargetPermanent("target creature or planeswalker an opponent controls",
					effects.And(effects.Or(effects.Creature(), effects.Planeswalker()), effects.OpponentControls())),
			),
			effects.On(game.EventLTB, effects.S82OldB113OpponentsBountyPermanentDied,
				"Chevill, Bane of Monsters — you gain 3 life and draw a card",
				effects.Do(effects.GainLife{Amount: 3}, effects.DrawCards{N: 1})),
		},
	})
}
