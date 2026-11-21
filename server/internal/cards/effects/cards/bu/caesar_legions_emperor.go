package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Caesar, Legion's Emperor — Legendary Creature — Human Soldier
// {1}{R}{W}{B}, 4/4 (Fallout, #5948):
//
//	"Whenever you attack, you may sacrifice another creature. When you
//	 do, choose two —
//	 • Create two 1/1 red and white Soldier creature tokens with haste
//	   that are tapped and attacking.
//	 • You draw a card and you lose 1 life.
//	 • Caesar deals damage equal to the number of creature tokens you
//	   control to target opponent."
//
// "Whenever you attack" is once per declaration (OncePerBatch over
// attackDeclaredByYou). The sacrifice is asked as the trigger resolves
// (maySacrificeOne: another creature, none means no question), and only
// a creature that really went raises the CR 603.12 reflexive trigger,
// whose two bullets are chosen as it goes on the stack (CR 603.3c), so
// the opponent for the third is picked then and the table can respond.
// Each bullet runs in printed order (CR 608.2c): the Soldiers (made
// tapped and attacking by the controller's choice of what each attacks,
// CR 508.4, so they never "attacked"), the draw then the life loss, and
// the damage counted as it resolves, the new Soldiers and the Soldiers
// of earlier bullets included. Caesar deals it as he last existed.
//
// No simplification.
func init() {
	const label = "Caesar, Legion's Emperor — you may sacrifice another creature; when you do, choose two"
	effects.Register(effects.Spec{
		OracleID:     "8e62d05a-6efd-4764-bca8-97895e0cb613",
		Name:         "Caesar, Legion's Emperor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclaredByYou(ev, source.Controller)
			}, label, func(g *game.Game, item *game.StackItem) error {
				return effects.MaySacrificeOne(g, item, "Caesar, Legion's Emperor — sacrifice another creature? When you do, choose two",
					effects.Creature(), true, caesarSacrificed)
			})),
		},
	})
}

// caesarSacrificed raises the reflexive modal trigger once a creature has
// really gone, keyed.
var caesarSacrificed = effects.RegisterSacrificedOne("prompt-run/caesar-legions-emperor-sacrifice", func(ctx *effects.Context, _ uuid.UUID) error {
	return effects.ReflexiveTrigger{
		Label: "Caesar, Legion's Emperor — choose two",
		Body:  caesarModesBody,
	}.Apply(ctx)
})

var caesarModesBody = game.ReflexiveModalBody("caesar-legions-emperor/modes", func(_ uuid.UUID, _ game.EffectParams) *game.ModeSpec {
	return effects.ChooseN("Choose two", 2, 2,
		effects.ModeDoing("Create two 1/1 red and white Soldier creature tokens with haste that are tapped and attacking.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.CreateTokensAttackingYourChoice{
					Controller: item.Controller,
					Template:   effects.TokenCard("1/1 red and white Soldier with haste"),
					N:          2,
					Tapped:     true,
					Label:      "Caesar, Legion's Emperor",
				}.Apply(ctx)
			}),
		effects.ModeDoing("You draw a card and you lose 1 life.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				if err := (effects.DrawCards{N: 1}).Apply(ctx); err != nil {
					return err
				}
				return ctx.Game.ChangePlayerLifeForEffect(item.SourceCardID, ctx.Controller(), -1)
			}),
		effects.ModeDoing("Caesar deals damage equal to the number of creature tokens you control to target opponent.",
			effects.TargetPlayer("target opponent", effects.Opponent()),
			func(item *game.StackItem, ctx *effects.Context, occ int) error {
				t, ok := effects.ModeTarget(ctx, occ)
				if !ok || t.Kind != game.TargetPlayer {
					return nil
				}
				n := len(effects.MatchingBattlefield(ctx, effects.And(effects.Creature(), effects.YouControl(), effects.IsTokenPredicate())))
				return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: n}.Apply(ctx)
			}),
	)
})
