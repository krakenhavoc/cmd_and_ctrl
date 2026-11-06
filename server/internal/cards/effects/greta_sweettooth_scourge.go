package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	sacrificeAFood := game.AbilityCost{SacrificeOther: sacrificeSpec("a Food", HasSubtype("Food"))}
	Register(Spec{
		OracleID:     "00173df7-a584-410c-af1d-ada9c791056a",
		Name:         "Greta, Sweettooth Scourge",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Greta, Sweettooth Scourge — create a Food token", Do(CreateToken{Template: FoodToken(), N: 1})),
		},
		Activated: []ActivatedAbility{
			{
				Label:        "{G}, Sacrifice a Food: Put a +1/+1 counter on target creature. Activate only as a sorcery.",
				Cost:         Plus(ManaCost("{G}"), sacrificeAFood),
				Targets:      TargetCreature("target creature"),
				SorcerySpeed: true,
				Effect:       putACounterOnTheTarget(game.CounterPlusOne),
			},
			{
				Label:   "{1}{B}, Sacrifice a Food: You draw a card and you lose 1 life.",
				Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
				Cost:    Plus(ManaCost("{1}{B}"), sacrificeAFood),
				Effect:  Do(DrawCards{N: 1}, GainLife{Amount: -1}),
			},
		},
	})
}
