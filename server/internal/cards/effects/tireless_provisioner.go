package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tireless Provisioner — Creature — Elf Scout {2}{G}, 3/2 (EDHREC
// rank 184):
//
//	"Landfall — Whenever a land you control enters, create a Food
//	 token or a Treasure token."
//
// A ramp creature that turns every land drop into an extra mana —
// or, when the mana isn't needed, three life. Landfall is an ETB
// trigger filtered to lands, and it fires for any way a land enters
// under your control: played, fetched, put onto the battlefield by
// a spell. The fetchland case fires twice, once for the fetch land
// and once for the basic it finds, as in paper.
//
// The choice between the two tokens is made as the trigger resolves
// (PickOption, CR 608.2c): Treasure first, which is the pick a bot makes
// and the one the card is played for, then Food.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ab8d5f5c-1976-4f77-8ed2-8d28ee666741",
		Name:         "Tireless Provisioner",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			LandfallPaying("Tireless Provisioner — create a Treasure (landfall)", game.LandfallPayoff{Tokens: 1},
				func(g *game.Game, item *game.StackItem) error {
					return PickOption{
						Question: "Tireless Provisioner — create a Food token or a Treasure token",
						Options: []game.ChoiceOption{
							{Label: "Create a Treasure token"},
							{Label: "Create a Food token"},
						},
						ThenKey: tirelessProvisionerToken,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// tirelessProvisionerToken is the answer to "Food or Treasure?", keyed
// (#4253): the controller is the stand-in context's, as the old closure's
// item.Controller was.
var tirelessProvisionerToken = OptionPickThen("option-pick/tireless-provisioner-token",
	func(ctx *Context, r game.OptionPicked) error {
		template := TreasureToken()
		if r.Index == 1 {
			template = FoodToken()
		}
		return CreateToken{Controller: ctx.Controller(), Template: template, N: 1}.Apply(ctx)
	})
