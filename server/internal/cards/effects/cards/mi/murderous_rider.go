package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Murderous Rider // Swift End — an Adventure card (CR 715):
//
//	Murderous Rider — Creature — Zombie Knight {1}{B}{B}, 2/3
//	  "Lifelink
//	   When this creature dies, put it on the bottom of its owner's
//	   library."
//	Swift End — Instant — Adventure {1}{B}{B}
//	  "Destroy target creature or planeswalker. You lose 2 life. (Then
//	   exile this card. You may cast the creature later from exile.)"
//
// The dies trigger moves the card only while it is still the graveyard
// object that died (CR 400.7), through the tuck route, so a
// commander's owner is offered the command zone (CR 903.9) and a token,
// which ceased to exist (CR 111.7), is not found. Swift End is
// Dreadbore's destroy plus the caster's life loss.
//
// No simplification.
const murderousRiderOracleID = "1080c5b5-6651-4c6a-93e6-099fbe389e26"

func init() {
	effects.Register(effects.Spec{
		OracleID:        murderousRiderOracleID,
		Name:            "Murderous Rider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Murderous Rider — put it on the bottom of its owner's library",
				effects.PutDiedThisOnBottomOfLibrary),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     murderousRiderOracleID + "#1",
		Name:         "Swift End",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetPermanent("target creature or planeswalker",
			effects.Or(effects.Creature(), effects.Planeswalker())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.DestroyChosenPermanent(ctx.Game, item); err != nil {
				return err
			}
			return ctx.Game.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -2)
		},
	})
}
