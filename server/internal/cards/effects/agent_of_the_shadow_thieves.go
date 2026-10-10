package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Agent of the Shadow Thieves — Legendary Enchantment — Background
// {1}{B} (EDHREC rank 5427):
//
//	"Commander creatures you own have "Whenever this creature attacks
//	 a player, if no opponent has more life than that player, put a
//	 +1/+1 counter on this creature. It gains deathtouch and
//	 indestructible until end of turn.""
//
// An ADR 0093 grant to each commander creature you own, under anyone's
// control (choose_a_background.go). The intervening if is read as the
// attack is declared and as the trigger resolves (CR 603.4). The
// counter and both keywords go on the creature that attacked.
//
// No simplification.
const agentOfTheShadowThievesGrant = "agent-of-the-shadow-thieves/counter-deathtouch-indestructible"

func init() {
	Register(Spec{
		OracleID:     "b2142d8b-ea53-443e-bd60-9e70d9da9bd7",
		Name:         "Agent of the Shadow Thieves",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: agentOfTheShadowThievesGrant,
			Triggered: []game.TriggeredAbility{
				wheneverThisAttacksAPlayerNoOpponentRicher("Agent of the Shadow Thieves — +1/+1 counter, deathtouch and indestructible", func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{
						Target:   ctx.Source(),
						Keywords: []string{"deathtouch", "indestructible"},
						Label:    "Agent of the Shadow Thieves — deathtouch and indestructible",
					}.Apply(ctx)
				}),
			},
			Text: "Whenever this creature attacks a player, if no opponent has more life than that player, put a +1/+1 counter on this creature. It gains deathtouch and indestructible until end of turn.",
		}},
		Static: []game.StaticAbility{grantToCommanderCreaturesYouOwn(agentOfTheShadowThievesGrant)},
	})
}
