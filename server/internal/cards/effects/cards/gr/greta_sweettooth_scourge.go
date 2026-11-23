package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Greta, Sweettooth Scourge — Legendary Creature — Human Warrior
// {1}{B}{G}, 3/3 (Wilds of Eldraine, #4471):
//
//	"When Greta enters, create a Food token. (It's an artifact with "{2},
//	 {T}, Sacrifice this token: You gain 3 life.")
//	 {G}, Sacrifice a Food: Put a +1/+1 counter on target creature.
//	 Activate only as a sorcery.
//	 {1}{B}, Sacrifice a Food: You draw a card and you lose 1 life."
//
// The entry trigger makes the catalog's Food token (CR 111.10b). Both
// activated abilities sacrifice a Food as a cost (Gyome, Master Chef's
// cost), any Food, token or not; neither taps Greta, so both work while
// she is summoning sick. The counter ability keeps sorcery timing
// (CR 602.5d) and puts nothing on a target that has become illegal
// (CR 608.2b). The draw is Dread Whispers' body.
//
// No simplification.
func init() {
	sacrificeAFood := game.AbilityCost{SacrificeOther: effects.SacrificeSpec("a Food", effects.HasSubtype("Food"))}
	effects.Register(effects.Spec{
		OracleID:     "00173df7-a584-410c-af1d-ada9c791056a",
		Name:         "Greta, Sweettooth Scourge",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Greta, Sweettooth Scourge — create a Food token", effects.Do(effects.CreateToken{Template: effects.FoodToken(), N: 1})),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:        "{G}, Sacrifice a Food: Put a +1/+1 counter on target creature. Activate only as a sorcery.",
				Cost:         effects.Plus(effects.ManaCost("{G}"), sacrificeAFood),
				Targets:      effects.TargetCreature("target creature"),
				SorcerySpeed: true,
				Effect:       effects.PutACounterOnTheTarget(game.CounterPlusOne),
			},
			{
				Label:   "{1}{B}, Sacrifice a Food: You draw a card and you lose 1 life.",
				Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{1}{B}"), sacrificeAFood),
				Effect:  effects.Do(effects.DrawCards{N: 1}, effects.GainLife{Amount: -1}),
			},
		},
	})
}
