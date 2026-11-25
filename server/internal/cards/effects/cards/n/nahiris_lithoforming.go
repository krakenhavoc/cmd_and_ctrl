package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nahiri's Lithoforming — Sorcery {X}{R}{R} (Zendikar Rising):
//
//	"Sacrifice X lands. For each land sacrificed this way, draw a card.
//	 You may play X additional lands this turn. Lands you control enter
//	 tapped this turn."
//
// The sacrifice is on resolution and the caster chooses the lands
// (ChoosePermanents, not a target): with fewer than X lands they all go,
// and the draw counts the ones that really left. The X additional land
// plays and the enters-tapped replacement (#6147,
// PermanentsEnterTappedThisTurnForEffect over "land", the caster's) are
// made whatever was sacrificed, X = 0 included.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "888e711a-1251-4704-8ef1-b4b98ec54b5a",
		Name:         "Nahiri's Lithoforming",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			x := ctx.X()
			me := ctx.Controller()
			var lands []uuid.UUID
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Controller == me && c.IsLand() {
					lands = append(lands, c.InstanceID)
				}
			}
			if n := min(x, len(lands)); n > 0 {
				return effects.ChoosePermanents{
					Question: "Nahiri's Lithoforming — sacrifice X lands",
					Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
						var out []uuid.UUID
						for _, c := range g.BattlefieldCardsForEffect() {
							if c.Controller == of && c.IsLand() {
								out = append(out, c.InstanceID)
							}
						}
						k := min(x, len(out))
						return out, k, k
					},
					Keyed: nahirisLithoformingChosen,
				}.Apply(ctx)
			}
			return nahirisLithoformingRest(ctx, 0)
		},
	})
}

var nahirisLithoformingChosen = effects.RegisterPermanentsPick("prompt-run/nahiris-lithoforming-sacrifice", effects.PermanentsDef{
	Then: func(ctx *effects.Context, picked game.PromptedPicks) error {
		if picked.Count() == 0 {
			return nahirisLithoformingRest(ctx, 0)
		}
		return ctx.Game.SacrificeAllKeyedForEffect(ctx.Source(), picked.Cards(), ctx.ItemTail(nahirisLithoformingDraw))
	},
})

// nahirisLithoformingDraw is the rest of the spell, run once the lands
// have gone; its draw counts how many really left. An on-disk key.
var nahirisLithoformingDraw = effects.CardTail("then/nahiris-lithoforming-draw", func(ctx *effects.Context, o game.TailOutcome) error {
	return nahirisLithoformingRest(ctx, len(o.IDs))
})

// nahirisLithoformingRest draws one card per land sacrificed, grants X
// additional land plays this turn, and makes the caster's lands enter
// tapped this turn.
func nahirisLithoformingRest(ctx *effects.Context, sacrificed int) error {
	me := ctx.Controller()
	if sacrificed > 0 {
		if err := (effects.DrawCards{Player: me, N: sacrificed}).Apply(ctx); err != nil {
			return err
		}
	}
	ctx.Game.GrantAdditionalLandPlayForEffect(me, ctx.X())
	ctx.Game.PermanentsEnterTappedThisTurnForEffect(ctx.Source(), me, []string{"land"},
		"Nahiri's Lithoforming — lands you control enter tapped this turn")
	return nil
}
