package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Clown Car — Artifact — Vehicle {X}, 1/1 (Unfinity; The Zeta Set,
// #5461):
//
//	"When this Vehicle enters, roll X six-sided dice. For each odd
//	 result, create a 1/1 white Clown Robot artifact creature token. For
//	 each even result, put a +1/+1 counter on this Vehicle.
//	 Crew 2"
//
// X is the value announced for the spell (CR 107.3m), read off the
// permanent as the trigger is built (Farmer Cotton's shape) and carried
// on the item, so a Clown Car put onto the battlefield without being
// cast rolls nothing. The dice go through the one dice door (RollDice),
// so a replacement of the roll and "whenever you roll one or more dice"
// see them; the tail sorts the kept results into odd and even. The
// counters go on this Vehicle only while it is still the permanent that
// entered; the tokens are made either way. Crew 2 is the ordinary crew
// ability.
//
// Not XMatters: cast for X=0 it is still a 1/1 Vehicle, a real move.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e0a6947f-3d52-45e0-8fd7-0bc07cc8c784",
		Name:         "Clown Car",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: Self,
			Key:       clownCarLabel,
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, clownCarLabel)
				item.Params.Amount = source.CastX()
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				x := item.Params.Amount
				if x <= 0 {
					return nil
				}
				return RollDice{Sides: 6, N: x, Then: clownCarResults}.Apply(NewContext(g, item))
			},
		}},
		Activated: []ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Clown Car"),
		}},
	})
}

const clownCarLabel = "Clown Car — roll X six-sided dice: a Clown Robot for each odd result, a +1/+1 counter for each even one"

// clownCarResults is the rest of the trigger, told the kept d6 results: a
// Clown Robot for each odd result, then a +1/+1 counter on this Vehicle
// for each even one. The key is an on-disk identity.
var clownCarResults = DiceTail("dice/clown-car-results", func(ctx *Context, rolls []int) error {
	odd, even := 0, 0
	for _, r := range rolls {
		if r%2 == 1 {
			odd++
		} else {
			even++
		}
	}
	if odd > 0 {
		if err := (CreateToken{Template: TokenCard("1/1 white Clown Robot artifact"), N: odd}).Apply(ctx); err != nil {
			return err
		}
	}
	if even == 0 {
		return nil
	}
	return CounterOnThis{Kind: game.CounterPlusOne, N: even}.Apply(ctx)
})
