package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Idris, Soul of the TARDIS (#6092): "Vanishing 3. Imprint — When Idris
// enters, exile another artifact you control until Idris leaves the
// battlefield. Idris has all activated and triggered abilities of the
// exiled card and gets +X/+X, where X is the exiled card's mana value."
//
// The imprint exile is linked (CR 607.2a: the triggered ability exiles,
// the static refers to "the exiled card"), so the borrow (ADR 0155 and its
// 2026-11-05 amendment: activated AND triggered abilities) and the +X/+X
// read exactly the card this Idris object exiled. The card returns when
// Idris leaves (CR 610.3), at which point it is a new object with no link
// (CR 400.7) and Idris lends and gets nothing from it.
func init() {
	const clause = "Idris has all activated and triggered abilities of the exiled card."
	borrow := effects.HasActivatedAndTriggeredAbilitiesOf(game.AbilityBorrow{
		Label:   clause,
		Lenders: effects.ExiledWithLenders(func(*game.Game, *game.Card, game.Card) bool { return true }),
	})
	borrow.DependsOnExile = true
	effects.Register(effects.Spec{
		OracleID:     "abb5210f-597c-450d-9239-6253a9136680",
		Name:         "Idris, Soul of the TARDIS",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.Cmdr2B17VanishingEntry(3)},
		Triggered: append(effects.Cmdr2B17VanishingTriggers(),
			effects.WhenThisEnters(idrisImprintLabel, idrisChoose)),
		Static: []game.StaticAbility{
			borrow,
			{
				Label:          "Idris gets +X/+X, where X is the exiled card's mana value.",
				Layer:          game.Layer7PT,
				SubLayer:       game.SubLayer7C_Modify,
				AppliesTo:      effects.SelfOnly,
				DependsOnExile: true,
				Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
					x := 0
					for _, e := range effects.Cmdr2S36CardsExiledWithThis(g, source) {
						x += e.ManaValue()
					}
					c.Power += x
					c.Toughness += x
				},
			},
		},
	})
}

const idrisImprintLabel = "Idris, Soul of the TARDIS — exile another artifact you control until Idris leaves the battlefield"

func idrisChoose(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	src := item.SourceCardID
	return effects.ChoosePermanents{
		Question: idrisImprintLabel,
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			ids := effects.PermanentsControlledByMatching(g, of, effects.And(effects.Artifact(), effects.NotSelf(src)))
			if len(ids) == 0 {
				return nil, 0, 0
			}
			return ids, 1, 1
		},
		Keyed: idrisChosen,
	}.Apply(ctx)
}

var idrisChosen = effects.RegisterPermanentsPick("prompt-run/idris-soul-of-the-tardis-imprint", effects.PermanentsDef{
	ThenRun: func(ctx *effects.Context, picked game.PromptedPicks, _ game.RunDone) error {
		cards := picked.Cards()
		if len(cards) == 0 || !effects.OnBattlefield(ctx.Game, cards[0]) {
			return nil
		}
		return effects.ExileUntil{
			Target:     cards[0],
			ThisLeaves: true,
			Linked:     true,
			Label:      "Idris, Soul of the TARDIS — the exiled card returns when Idris leaves the battlefield",
		}.Apply(ctx)
	},
})
