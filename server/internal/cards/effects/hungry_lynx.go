package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hungry Lynx — Creature — Cat {1}{G}, 2/2 (Commander 2017, #5533):
//
//	"Cats you control have protection from Rats. (They can't be blocked,
//	 targeted, or dealt damage by Rats.)
//	 At the beginning of your end step, target opponent creates a 1/1 black
//	 Rat creature token with deathtouch.
//	 Whenever a Rat dies, put a +1/+1 counter on each Cat you control."
//
// The protection is Yawgmoth, Thran Physician's subtype quality (CR 702.16,
// ADR 0072) granted in layer 6 to every Cat its controller controls, the
// Lynx included. The end-step token goes to the TARGET opponent, who
// controls it (Dowsing Dagger's shape). "Whenever a Rat dies" is any
// player's Rat, read from last-known information (CR 603.10a), so the Rats
// it hands out feed it; the counters go on each Cat you control as the
// ability resolves.
//
// No simplification.
func init() {
	protection := TribalKeywordGrant(TribeFilter{Tribes: []string{"Cat"}, YoursOnly: true}, "protection from Rats")
	protection.Label = "Cats you control have protection from Rats."
	Register(Spec{
		OracleID:     "a5fdc358-25de-4a5b-b965-53a01df6019c",
		Name:         "Hungry Lynx",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{protection},
		Triggered: []game.TriggeredAbility{
			Targeting(AtYourEndStep("Hungry Lynx — target opponent creates a 1/1 black Rat with deathtouch",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetPlayer {
							return CreateToken{Controller: t.ID, Template: TokenCard("1/1 black Rat with deathtouch"), N: 1}.Apply(ctx)
						}
					}
					return nil
				}),
				TargetPlayer("target opponent", Opponent())),
			On(game.EventLTB, c17ARatDied, "Hungry Lynx — put a +1/+1 counter on each Cat you control",
				avB30aPutCounterOnEachOfSubtypeYouControl("Cat")),
		},
	})
}
