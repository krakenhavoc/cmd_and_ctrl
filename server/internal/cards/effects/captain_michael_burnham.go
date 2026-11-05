package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Captain Michael Burnham — {2}{R} Legendary Creature — Human Officer, 3/3
// (Star Trek, #4973):
//
//	"Trample
//	 Landfall — Whenever a land you control enters, Captain Burnham deals
//	 1 damage to each opponent. If this is the second time this ability
//	 has resolved this turn, put a +1/+1 counter on Captain Burnham."
//
// Zimone, Mystery Unraveler's resolution tally (Game.ResolvedThisTurn,
// keyed on this trigger's stack label): it counts the resolution in
// progress, so the second resolution reads exactly 2, and the third and
// later put no counter. Per object (CR 400.7): a Burnham that left and
// came back this turn starts again. The damage is dealt whether or not
// she is still on the battlefield (as she last existed, CR 608.2h); the
// counter needs her still there as the same object (CounterOnThis).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9a71aa6c-2d5f-468d-b588-252fae37ccbb",
		Name:            "Captain Michael Burnham",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			LandfallPaying(captainBurnhamLandfallLabel, game.LandfallPayoff{DamageEachOpponent: 1}, captainBurnhamLandfall),
		},
	})
}

const captainBurnhamLandfallLabel = "Captain Michael Burnham — landfall: 1 damage to each opponent, and a +1/+1 counter the second time this turn"

func captainBurnhamLandfall(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (DamageEachOpponent{N: 1}).Apply(ctx); err != nil {
		return err
	}
	if g.ResolvedThisTurn(item.SourceCardID, captainBurnhamLandfallLabel) != 2 {
		return nil
	}
	return CounterOnThis{Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
}
