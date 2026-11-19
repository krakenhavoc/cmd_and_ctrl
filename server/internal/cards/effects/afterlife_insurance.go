package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Afterlife Insurance — Instant {1}{W/B} (Ravnica: Clue Edition):
//
//	"Creatures you control gain afterlife 1 until end of turn. Draw a card."
//
// Whisperwood Elemental's mass grant: every creature its controller controls
// as the spell resolves (CR 611.2c) gains the afterlife trigger (CR 702.135a,
// Tithe Taker's dies-trigger shape) until end of turn; a creature that enters
// later gets nothing.
//
// No simplification.
const afterlifeInsuranceAfterlife = "afterlife-insurance/afterlife-1"

func init() {
	Register(Spec{
		OracleID:     "05501e88-d4c3-4474-92a9-c02ab15b107b",
		Name:         "Afterlife Insurance",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: afterlifeInsuranceAfterlife,
			Triggered: []game.TriggeredAbility{
				WhenThisDies("Afterlife Insurance — afterlife 1: create a 1/1 white and black Spirit creature token with flying",
					Do(CreateToken{Template: kayasGuileSpirit(), N: 1})),
			},
			Text: "Afterlife 1 (When this creature dies, create a 1/1 white and black Spirit creature token with flying.)",
		}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Controller != ctx.Controller() || !c.IsCreature() {
					continue
				}
				if err := (GrantAbilitiesFor{
					Target: c.InstanceID,
					Keys:   []string{afterlifeInsuranceAfterlife},
					Label:  "Afterlife Insurance — afterlife 1 until end of turn",
				}).Apply(ctx.asGroupMember()); err != nil {
					return err
				}
			}
			return DrawCards{N: 1}.Apply(ctx)
		},
	})
}
