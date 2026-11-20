package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Moira Brown, Guide Author — Legendary Creature — Human Citizen {1}{R}{W},
// 2/3 (Fallout, #5948):
//
//	"When Moira Brown enters, create a colorless Book Equipment artifact token
//	 named Wasteland Survival Guide with "Equipped creature gets +1/+1 for each
//	 quest counter among permanents you control" and equip {1}.
//	 Whenever you attack, put a quest counter on target nonland permanent you
//	 control."
//
// The token is the printed one (GapB64SurvivalGuideToken): a layer 7c pump
// counted over the quest counters on every permanent its controller controls,
// and the shared equip ability. "Whenever you attack" is once per declaration
// (OncePerBatch over attackDeclaredByYou); the target is a nonland permanent of
// its controller's, chosen as the trigger goes on the stack, and the counter
// goes on only while it is still legal (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "db02644b-820f-4102-b1f6-ed8fac40baf8",
		Name:         "Moira Brown, Guide Author",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Moira Brown — create a Wasteland Survival Guide Equipment token",
				Do(CreateToken{Template: GapB64SurvivalGuideToken(), N: 1})),
			OncePerBatch(Targeting(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller)
			}, "Moira Brown — put a quest counter on target nonland permanent you control",
				putACounterOnTheTarget("quest")),
				TargetPermanent("target nonland permanent you control", Nonland(), YouControl()))),
		},
	})
}
