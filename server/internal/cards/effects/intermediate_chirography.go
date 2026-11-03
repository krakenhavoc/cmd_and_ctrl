package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Intermediate Chirography — Enchantment — Class {1}{B}:
//
//	"(Gain the next level as a sorcery to add its ability.)
//	 When this Class enters, create a 2/1 white and black Inkling
//	 creature token with flying.
//	 {1}{B}: Level 2
//	 Whenever you lose life for the first time each turn, put a +1/+1
//	 counter on target creature you control.
//	 {2}{B}: Level 3
//	 At the beginning of each end step, if a modified creature died under
//	 your control this turn, create a 2/1 white and black Inkling creature
//	 token with flying. (Equipment, Auras you control, and counters are
//	 modifications.)"
//
// Level 2 reads the turn's life-loss tally, so a loss before the Class
// reached level 2 uses up the turn's first and nothing triggers later that
// turn (the card's rulings). Level 3 is an intervening "if" (CR 603.4)
// over the turn's count of modified creatures that died under its
// controller's control (CR 700.9: a counter of any kind, any Equipment,
// or an Aura its controller controlled, as it last existed on the
// battlefield), checked as each end step begins and again on resolution.
//
// No simplification.
func init() {
	inkling := Do(CreateToken{Template: TokenCard("2/1 white and black Inkling with flying"), N: 1})
	Register(Spec{
		OracleID:     "6c948f71-5c86-450c-8828-df4d8deaf1dd",
		Name:         "Intermediate Chirography",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			LevelUp(2, ManaCost("{1}{B}")),
			LevelUp(3, ManaCost("{2}{B}")),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Intermediate Chirography — create a 2/1 white and black Inkling creature token with flying", inkling),
			AtLevel(2, Targeting(
				WheneverYouLoseLifeForTheFirstTimeEachTurn("Intermediate Chirography — put a +1/+1 counter on target creature you control",
					onePlusOneCounterOnFirstLegalTarget),
				TargetCreature("target creature you control", YouControl()))),
			AtLevel(3, InterveningIf(CondBoard{
				Name: "a modified creature died under your control this turn",
				Test: func(ctx *Context) bool { return ctx.Game.TurnTallyFor(ctx.Controller()).ModifiedCreaturesDied > 0 },
			}, On(game.EventBeginEndStep, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return g.TurnTallyFor(source.Controller).ModifiedCreaturesDied > 0
			}, "Intermediate Chirography — level 3: create a 2/1 white and black Inkling creature token with flying", inkling))),
		},
	})
}
