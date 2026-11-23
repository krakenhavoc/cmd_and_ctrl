package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Old Man Willow — Legendary Creature — Treefolk {2}{B}{G}, */* (The Lord of
// the Rings: Tales of Middle-earth):
//
//	"Old Man Willow's power and toughness are each equal to the number of
//	 lands you control.
//	 Whenever Old Man Willow attacks, you may sacrifice another creature or
//	 a token. When you do, target creature an opponent controls gets -2/-2
//	 until end of turn."
//
// The characteristic-defining P/T is Beanstalk Giant's (LandsYouControlPT).
// The optional sacrifice is a card-set pick of its own with a minimum of
// zero (Sothera's chain): another creature, or any token, the controller
// controls, Willow itself excluded; declining is choosing none. Only a
// permanent that really was sacrificed makes "when you do" (a CR 603.12
// reflexive trigger, SacrificePermanent's tail) go on the stack, targeting a
// creature an opponent controls as it does (CR 603.3d), and the -2/-2 is
// applied if it is still a legal target as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "96c96842-42c2-4aa3-95e2-6186d4ab5693",
		Name:         "Old Man Willow",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.LandsYouControlPT("Old Man Willow's power and toughness are each equal to the number of lands you control."),
		},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Old Man Willow — you may sacrifice another creature or a token; when you do, target creature an opponent controls gets -2/-2 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					me := ctx.Controller()
					var cands []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == me && oldManWillowMaySacrifice(c, item.SourceCardID) {
							cands = append(cands, c.InstanceID)
						}
					}
					if len(cands) == 0 {
						return nil
					}
					g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
						Chooser:  me,
						Source:   ctx.Source(),
						Question: "Old Man Willow — you may sacrifice another creature or a token",
						Cards:    cands,
						Min:      0,
						Max:      1,
						Zone:     game.ZoneBattlefield,
						LiveZone: game.ZoneBattlefield,
						ThenKey:  oldManWillowPick,
						Asked:    game.NewAskedItem(ctx.Item),
					})
					return nil
				}),
		},
	})
}

// oldManWillowMaySacrifice is "another creature or a token": a creature
// other than Willow, or a token of any type.
func oldManWillowMaySacrifice(c game.Card, willow uuid.UUID) bool {
	return (c.IsCreature() && c.InstanceID != willow) || c.IsToken()
}

// oldManWillowPick is the controller's choice answered: the sacrifice, then
// the reflexive trigger. An on-disk key: never renamed, never reused.
var oldManWillowPick = effects.AskedPick("choose-cards/old-man-willow-sacrifice", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) != 1 {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(r.Picked[0])
	if !ok || c.Controller != r.Chooser || !effects.OnBattlefield(ctx.Game, r.Picked[0]) || !oldManWillowMaySacrifice(c, ctx.Source()) {
		return nil
	}
	return effects.SacrificePermanent{Target: r.Picked[0], ThenTail: ctx.Tail(oldManWillowSacrificed)}.Apply(ctx)
})

var oldManWillowSacrificed = effects.CardTail("then/old-man-willow-sacrificed", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return effects.ReflexiveTrigger{Label: "Old Man Willow — target creature an opponent controls gets -2/-2 until end of turn", Body: oldManWillowShrink}.Apply(ctx)
})

var oldManWillowShrink = game.ReflexiveBody("old-man-willow/shrink",
	func(g *game.Game, item *game.StackItem, _ game.EffectParams) error {
		ctx := effects.NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			return effects.BoostUntilEOT{Target: t.ID, Power: -2, Toughness: -2, Label: "Old Man Willow — -2/-2 until end of turn"}.Apply(ctx)
		}
		return nil
	},
	effects.ConstTargets(func() *game.TargetSpec {
		return effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())
	}))
