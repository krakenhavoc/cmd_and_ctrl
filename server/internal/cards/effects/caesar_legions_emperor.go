package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "8e62d05a-6efd-4764-bca8-97895e0cb613",
		Name:         "Caesar, Legion's Emperor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller)
			}, label, func(g *game.Game, item *game.StackItem) error {
				return maySacrificeOne(g, item, "Caesar, Legion's Emperor — sacrifice another creature? When you do, choose two",
					Creature(), true, caesarSacrificed)
			})),
		},
	})
}

// caesarSacrificed raises the reflexive modal trigger once a creature has
// really gone, keyed.
var caesarSacrificed = RegisterSacrificedOne("prompt-run/caesar-legions-emperor-sacrifice", func(ctx *Context, _ uuid.UUID) error {
	return ReflexiveTrigger{
		Label: "Caesar, Legion's Emperor — choose two",
		Body:  caesarModesBody,
	}.Apply(ctx)
})

var caesarModesBody = game.ReflexiveModalBody("caesar-legions-emperor/modes", func(_ uuid.UUID, _ game.EffectParams) *game.ModeSpec {
	return ChooseN("Choose two", 2, 2,
		ModeDoing("Create two 1/1 red and white Soldier creature tokens with haste that are tapped and attacking.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return CreateTokensAttackingYourChoice{
					Controller: item.Controller,
					Template:   TokenCard("1/1 red and white Soldier with haste"),
					N:          2,
					Tapped:     true,
					Label:      "Caesar, Legion's Emperor",
				}.Apply(ctx)
			}),
		ModeDoing("You draw a card and you lose 1 life.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				if err := (DrawCards{N: 1}).Apply(ctx); err != nil {
					return err
				}
				return ctx.Game.ChangePlayerLifeForEffect(item.SourceCardID, ctx.Controller(), -1)
			}),
		ModeDoing("Caesar deals damage equal to the number of creature tokens you control to target opponent.",
			TargetPlayer("target opponent", Opponent()),
			func(item *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
				if !ok || t.Kind != game.TargetPlayer {
					return nil
				}
				n := len(MatchingBattlefield(ctx, And(Creature(), YouControl(), IsTokenPredicate())))
				return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: n}.Apply(ctx)
			}),
	)
})
