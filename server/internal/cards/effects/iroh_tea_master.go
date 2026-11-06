package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Iroh, Tea Master — Legendary Creature — Human Citizen Ally {1}{R}{W},
// 2/2:
//
//	"When Iroh enters, create a Food token.
//	 At the beginning of combat on your turn, you may have target
//	 opponent gain control of target permanent you control. When you do,
//	 create a 1/1 white Ally creature token. Put a +1/+1 counter on that
//	 token for each permanent you own that your opponents control."
//
// The combat trigger targets twice (Dragonspark Reactor's Clauses): the
// opponent and a permanent you control, chosen every combat whether or
// not you mean to give anything (the ruling). The "you may" is asked as
// it resolves (Ratchet, Field Medic's MayChoice), and only while both
// targets are still legal: if either is not, the ability does nothing
// (the ruling, CR 608.2b). On a yes the opponent gains control with no
// end (CR 611.2a), so the permanent comes back only if that opponent
// leaves the game (the ruling, CR 800.4a), and a CR 603.12 reflexive
// trigger makes the Ally and then counts, as it resolves, every
// permanent you own that an opponent controls, the one just given
// included.
//
// No simplifications. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:     "87adda41-113b-4126-8978-25a258b31bdd",
		Name:         "Iroh, Tea Master",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Iroh, Tea Master — create a Food token", Do(CreateToken{Template: FoodToken(), N: 1})),
			Targeting(AtBeginningOfYourCombat("Iroh, Tea Master — you may have target opponent gain control of target permanent you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if _, ok := ctx.ClauseTarget(0); !ok {
						return nil
					}
					if _, ok := ctx.ClauseTarget(1); !ok {
						return nil
					}
					return MayChoice{
						Question: "Iroh, Tea Master — have target opponent gain control of target permanent you control? (When you do, create a 1/1 white Ally creature token.)",
						Then:     irohTeaMasterGiveAnswer,
					}.Apply(ctx)
				}),
				Clauses(
					TargetPlayer("target opponent", Opponent()),
					TargetPermanent("target permanent you control", YouControl()),
				)),
		},
	})
}

// irohTeaMasterGiveAnswer is the "yes": the opponent gains control, and
// when you do, the Ally. Keyed (#4291).
var irohTeaMasterGiveAnswer = MayYes("confirm/iroh-tea-master-give", func(ctx *Context, _ game.ConfirmAnswer) error {
	opp, ok := ctx.ClauseTarget(0)
	if !ok {
		return nil
	}
	perm, ok := ctx.ClauseTarget(1)
	if !ok {
		return nil
	}
	if err := (GainControl{
		Target:     perm.ID,
		Controller: opp.ID,
		Duration:   game.IndefiniteDuration(),
		Label:      "Iroh, Tea Master — target opponent gains control",
	}).Apply(ctx); err != nil {
		return err
	}
	return WhenYouDo("Iroh, Tea Master — create a 1/1 white Ally creature token with a +1/+1 counter for each permanent you own that your opponents control",
		irohTeaMasterAllyBody).Apply(ctx)
})

var irohTeaMasterAllyBody = game.SimpleDelayedBody("iroh-tea-master/ally", func(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	made, err := g.CreateTokensForEffect(ctx.Controller(), ctx.TextToken(TokenCard("1/1 white Ally")), 1, game.TokenEntryOptions{})
	if err != nil || len(made) == 0 {
		return err
	}
	n := irohTeaMasterOwnedButOpponentsControl(g, ctx.Controller())
	if n == 0 {
		return nil
	}
	return AddCounter{Target: made[0], Kind: game.CounterPlusOne, N: n}.Apply(ctx)
})

// irohTeaMasterOwnedButOpponentsControl is "each permanent you own that
// your opponents control".
func irohTeaMasterOwnedButOpponentsControl(g *game.Game, you uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Owner == you && c.Controller != you {
			n++
		}
	}
	return n
}
