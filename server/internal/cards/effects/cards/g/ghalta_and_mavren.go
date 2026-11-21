package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ghalta and Mavren — Legendary Creature — Dinosaur Vampire {3}{G}{G}{W}{W},
// 12/12 (March of the Machine Commander):
//
//	"Trample
//	 Whenever you attack, choose one —
//	 • Create a tapped and attacking X/X green Dinosaur creature token with
//	   trample, where X is the greatest power among other attacking creatures.
//	 • Create X 1/1 white Vampire creature tokens with lifelink, where X is the
//	   number of other attacking creatures."
//
// "Whenever you attack" is one trigger per declaration (CR 603.2c), however
// many creatures attack, and fires for an attack by other creatures alone. X is
// read as the trigger resolves from the creatures attacking then (CR 608.2h
// does not reach here: the source's own power never enters it). The Dinosaur
// enters tapped and attacking the player (or planeswalker or battle) the
// attack was declared on, or the one its controller names when that is gone
// (CR 508.4); it is "other" than Ghalta and Mavren, which is never counted.
//
// The printed Dinosaur token is */*, the size being the card's X, so it is
// the 3/3 green Dinosaur with trample row resized to X.
//
// No simplification.
func init() {
	t := effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return effects.AttackDeclaredByYou(ev, source.Controller)
	}, "Ghalta and Mavren — choose one", func(*game.Game, *game.StackItem) error { return nil }))
	t.Modes = effects.ChooseOne(
		effects.ModeDoing("Create a tapped and attacking X/X green Dinosaur creature token with trample, where X is the greatest power among other attacking creatures.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				x := 0
				for _, c := range effects.S82OtherAttackers(ctx.Game, item.SourceCardID) {
					x = max(x, c.CurrentPower())
				}
				tmpl := effects.TokenCard("3/3 green Dinosaur with trample")
				tmpl.Power, tmpl.Toughness = x, x
				return effects.CreateTokensAttackingYourChoice{
					Template: tmpl,
					N:        1,
					Tapped:   true,
					Prefer:   effects.TriggerAttackTarget(item),
					Label:    "Ghalta and Mavren",
				}.Apply(ctx)
			}),
		effects.ModeDoing("Create X 1/1 white Vampire creature tokens with lifelink, where X is the number of other attacking creatures.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.CreateToken{Template: effects.TokenCard("1/1 white Vampire with lifelink"),
					N: len(effects.S82OtherAttackers(ctx.Game, item.SourceCardID))}.Apply(ctx)
			}),
	)
	effects.Register(effects.Spec{
		OracleID:        "756042c5-e444-45d7-8dc6-ae7c83c98438",
		Name:            "Ghalta and Mavren",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered:       []game.TriggeredAbility{t},
	})
}
