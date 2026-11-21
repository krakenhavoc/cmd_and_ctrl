package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Captain Ripley Vance — Legendary Creature — Human Pirate {2}{R}, 3/2
// (Lost Caverns of Ixalan Commander):
//
//	"Whenever you cast your third spell each turn, put a +1/+1 counter on
//	 Captain Ripley Vance, then it deals damage equal to its power to any
//	 target."
//
// Cori Mountain Stalwart's cast-count clock at three. The counter lands
// first, then the damage reads the power as it resolves, or as Ripley last
// existed if he has left (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "eefeeab1-b9f2-451a-9250-eb7b54d00fc1",
		Name:         "Captain Ripley Vance",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(
				On(game.EventCast, emeritusOfConflictThirdSpell,
					"Captain Ripley Vance — put a +1/+1 counter on him, then he deals damage equal to his power to any target",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						if onBattlefield(g, item.SourceCardID) && !sourceIsNewObject(g, item) {
							if err := (AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
								return err
							}
						}
						return sourceDealsDamageEqualToItsPowerToTarget(g, item)
					}),
				TargetAny()), game.Purpose{Answers: game.AnswerRemove}),
		},
	})
}
