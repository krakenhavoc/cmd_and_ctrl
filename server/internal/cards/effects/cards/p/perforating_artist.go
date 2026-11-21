package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Perforating Artist — Creature — Devil {1}{B}{R}, 3/2 (Foundations,
// #4465):
//
//	"Deathtouch
//	 Raid — At the beginning of your end step, if you attacked this turn,
//	 each opponent loses 3 life unless that player sacrifices a nonland
//	 permanent of their choice or discards a card."
//
// Searslicer Goblin's raid end-step trigger (an intervening "if", CR
// 603.4, checked as it triggers and again as it resolves) with Torment of
// Hailfire's punisher asked once: each opponent, in APNAP order, picks
// losing 3 life, sacrificing a nonland permanent or discarding a card, and
// is offered only what they can do.
//
// No simplification.
func init() {
	effects.RegisterPunisher(perforatingArtistPunisher, "perforating-artist")
	effects.Register(effects.Spec{
		OracleID:        "a3643f66-dfa2-4275-a093-23e162496354",
		Name:            "Perforating Artist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginEndStep, effects.AllOf(effects.ByYou, effects.TmcB6YouAttackedThisTurn),
				"Perforating Artist — Raid: each opponent loses 3 life unless they sacrifice a nonland permanent or discard a card",
				effects.Do(effects.If{Cond: effects.AvatarB11CondYouAttacked{}, Then: []effects.Applier{
					effects.Instruction(func(ctx *effects.Context) error {
						return perforatingArtistPunisher.Start(ctx, 1)
					}),
				}})),
		},
	})
}

// perforatingArtistPunisher is the printed punisher: 3 life a refusal.
var perforatingArtistPunisher = effects.PunisherRepeat{Name: "Perforating Artist", Life: 3}
