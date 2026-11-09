package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// theWanderingMinstrelElemental is the printed 2/2 all-colors Elemental.
const theWanderingMinstrelElemental = "2/2 white and blue and black and red and green Elemental"

// The Wandering Minstrel — Legendary Creature — Human Bard {G}{U}, 1/3
// (Final Fantasy, #4462):
//
//	"Lands you control enter untapped.
//	 The Minstrel's Ballad — At the beginning of combat on your turn, if
//	 you control five or more Towns, create a 2/2 Elemental creature token
//	 that's all colors.
//	 {3}{W}{U}{B}{R}{G}: Other creatures you control get +X/+X until end
//	 of turn, where X is the number of Towns you control."
//
// The first ability is Horizon Explorer's replacement: a land you control
// that would enter tapped enters untapped, the controller ordering it
// against the land's own "enters tapped" (the rulings). The second is an
// intervening "if" (CR 603.4), checked as combat begins and again as the
// ability resolves, over the Towns its controller controls (a land with
// the Town subtype, read through the layers); the token is the printed
// Elemental, all five colors. The third counts the Towns once, as it
// resolves (the ruling), and pumps the other creatures its controller
// controls as of then (CR 611.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e6e6e3d2-f2fb-46a1-8798-13a2f2efe8cc",
		Name:         "The Wandering Minstrel",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{b16LandsYouControlEnterUntapped()},
		Triggered: []game.TriggeredAbility{
			InterveningIf(wandersMinstrelFiveTowns{}, On(game.EventStepBegan,
				AllOf(StepBegan(game.StepBeginCombat, true), func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return wanderingMinstrelTowns(g, source.Controller) >= 5
				}),
				"The Wandering Minstrel — create a 2/2 Elemental creature token that's all colors",
				Do(CreateToken{Template: TokenCard(theWanderingMinstrelElemental), N: 1}))),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}{W}{U}{B}{R}{G}: Other creatures you control get +X/+X until end of turn, where X is the number of Towns you control.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{3}{W}{U}{B}{R}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				x := wanderingMinstrelTowns(g, item.Controller)
				return BoostUntilEOT{
					Match: And(Creature(), YouControl(), OtherThan(item.SourceCardID)),
					Power: x, Toughness: x,
					Label: "The Wandering Minstrel — +X/+X",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}

func wanderingMinstrelTowns(g *game.Game, player uuid.UUID) int {
	return countControlled(g, player, MatchLandSubtype("Town"))
}

// wandersMinstrelFiveTowns is the intervening "if you control five or
// more Towns".
type wandersMinstrelFiveTowns struct{}

func (wandersMinstrelFiveTowns) Holds(ctx *Context) bool {
	return wanderingMinstrelTowns(ctx.Game, ctx.Controller()) >= 5
}

func (wandersMinstrelFiveTowns) condition() (game.FootprintStep, bool) {
	return condStep(game.FootprintCondBoard, 0), true
}
