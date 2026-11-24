package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glissa, the Traitor — Legendary Creature — Phyrexian Zombie Elf {B}{G}{G},
// 3/3 (Secret Lair Drop):
//
//	"First strike, deathtouch
//	 Whenever a creature an opponent controls dies, you may return target
//	 artifact card from your graveyard to your hand."
//
// The condition is Yahenni's: the dying creature's controller as it last existed
// on the battlefield is not Glissa's controller. The "you may" and the target
// are Midnight Scavengers' (the target is chosen as the ability goes on the
// stack, the question asked as it resolves, CR 603.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cafacba9-195c-44ea-ba80-3e489274759c",
		Name:            "Glissa, the Traitor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike", "deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := effects.DiedCreature(ev, g)
				return ok && effects.LeftUnderControlOf(ev, dead) != source.Controller
			}, "Glissa, the Traitor — you may return target artifact card from your graveyard to your hand", effects.ReturnTargetedCardToHand),
				effects.TargetCardInGraveyard("target artifact card from your graveyard", effects.Artifact(), effects.YouOwn())),
				"Glissa, the Traitor — return the artifact card to your hand?"),
		},
	})
}
