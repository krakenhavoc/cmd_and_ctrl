package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sokenzan Smelter — Creature — Goblin Artificer {1}{R}, 2/2 (Kamigawa: Neon
// Dynasty, S82):
//
//	"At the beginning of combat on your turn, you may pay {1} and sacrifice an
//	 artifact. If you do, create a 3/1 red Construct artifact creature token
//	 with haste."
//
// Killer Service's pay-and-sacrifice: both halves are costs, so it is offered
// only when there is an artifact to sacrifice; the sacrifice is forced once
// the {1} is paid. The Smelter itself is not an artifact, so it is never its
// own fodder.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dd645241-2992-4c4e-a6a8-c2215d2340ae",
		Name:         "Sokenzan Smelter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat("Sokenzan Smelter — you may pay {1} and sacrifice an artifact to create a 3/1 Construct", sokenzanSmelterOffer),
		},
	})
}

func sokenzanSmelterOffer(g *game.Game, item *game.StackItem) error {
	has := false
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsArtifact() {
			has = true
			break
		}
	}
	if !has {
		return nil
	}
	return MayPay{
		Chooser:  item.Controller,
		Cost:     "{1}",
		Question: "Sokenzan Smelter — pay {1} and sacrifice an artifact to create a 3/1 red Construct artifact creature token with haste?",
		Then:     sokenzanSmelterPaid,
	}.Apply(NewContext(g, item))
}

var sokenzanSmelterPaid = PayUnlessBranches("pay-unless/s82-neo-sokenzan-smelter",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return mustSacrificeOne(ctx.Game, ctx.Item, "Sokenzan Smelter — sacrifice an artifact", Artifact(), sokenzanSmelterConstruct)
	}, nil)

var sokenzanSmelterConstruct = RegisterSacrificedOne("prompt-run/s82-neo-sokenzan-smelter-construct",
	func(ctx *Context, _ uuid.UUID) error {
		return CreateToken{Controller: ctx.Controller(), Template: TokenCard("3/1 red Construct artifact with haste"), N: 1}.Apply(ctx)
	})
