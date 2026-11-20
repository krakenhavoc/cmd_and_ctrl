package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grip of Phyresis — Instant {2}{U} (Commander 2016, #5537):
//
//	"Gain control of target Equipment, then create a 0/0 black Phyrexian
//	 Germ creature token and attach that Equipment to it."
//
// The theft has no duration (Coveted Falcon's GainControl). The Germ is
// living weapon's token, created keyed so the attach runs once it is on
// the battlefield, and the Equipment it attaches is the spell's target
// rather than its source. Nothing checks state-based actions until the
// spell has finished resolving, so a Germ the Equipment gives no
// toughness then dies and you keep the Equipment, as the ruling says.
// A target that became illegal in response makes the spell fizzle.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4c95b2fe-d107-44f9-b877-8bd4799f5a65",
		Name:         "Grip of Phyresis",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target Equipment", HasSubtype("Equipment")),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (GainControl{Target: id, Duration: game.IndefiniteDuration(), Label: "Grip of Phyresis — gain control"}).Apply(ctx); err != nil {
				return err
			}
			return ctx.Game.CreateTokensKeyedForEffect(game.TokenCreation{
				Controller: ctx.Controller(),
				Source:     item.SourceCardID,
				Groups:     []game.TokenGroup{{Template: ctx.TextToken(TokenCard(livingWeaponGermToken)), Count: 1}},
			}, ctx.ItemTail(gripOfPhyresisAttach))
		},
	})
}

// gripOfPhyresisAttach is "and attach that Equipment to it", once the
// Germ is on the battlefield. An on-disk key.
var gripOfPhyresisAttach = CardTail("grip-of-phyresis/attach", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 || len(ctx.Item.Targets) == 0 {
		return nil
	}
	equipment := ctx.Item.Targets[0].ID
	if !onBattlefield(ctx.Game, equipment) {
		return nil
	}
	return ctx.Game.AttachForEffect(equipment, game.TargetRef{Kind: game.TargetCard, ID: o.IDs[0]})
})
