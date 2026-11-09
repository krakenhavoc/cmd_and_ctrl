package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	registerPunisher(perforatingArtistPunisher, "perforating-artist")
	Register(Spec{
		OracleID:        "a3643f66-dfa2-4275-a093-23e162496354",
		Name:            "Perforating Artist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, AllOf(ByYou, tmcB6YouAttackedThisTurn),
				"Perforating Artist — Raid: each opponent loses 3 life unless they sacrifice a nonland permanent or discard a card",
				Do(If{Cond: avatarB11CondYouAttacked{}, Then: []Applier{
					Instruction(func(ctx *Context) error {
						return perforatingArtistPunisher.start(ctx, 1)
					}),
				}})),
		},
	})
}

// perforatingArtistPunisher is the printed punisher: 3 life a refusal.
var perforatingArtistPunisher = punisherRepeat{Name: "Perforating Artist", Life: 3}
