package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chromium, the Mutable — Legendary Creature — Elder Dragon {4}{W}{U}{B}, 7/7
// (Core Set 2019, S82 OLD batch 199, #5527):
//
//	"Flash
//	 This spell can't be countered.
//	 Flying
//	 Discard a card: Until end of turn, Chromium becomes a Human with base
//	 power and toughness 1/1, loses all abilities, and gains hexproof. It
//	 can't be blocked this turn."
//
// Flash and flying are printed keywords and "can't be countered" is the
// spec's CantBeCountered (Balustrade Wurm's), working on the stack. The
// activated ability is Snakeform's one continuous effect on itself until end
// of turn: layer 4 the creature type becomes Human (it stays legendary), layer
// 6 all abilities are lost and then hexproof is granted by a later effect, so
// it keeps it, and layer 7b the base power and toughness are 1/1. "It can't be
// blocked this turn" is Bessie's CantBeBlocked restriction, a rule of the turn
// rather than an ability, so losing all abilities does not lift it. The
// ability is the cost of one card, discarded as it is activated.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "58d8ff36-4a4d-40e0-b812-3d23b2c7606a",
		Name:            "Chromium, the Mutable",
		Completeness:    effects.CompletenessFull,
		CantBeCountered: true,
		PrintedKeywords: []string{"flash", "flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Discard a card: Until end of turn, Chromium becomes a Human with base power and toughness 1/1, loses all abilities, and gains hexproof. It can't be blocked this turn.",
			Cost:    effects.DiscardACard(),
			Purpose: game.Purpose{Answers: game.AnswerProtect | game.AnswerCombatGrant},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id := item.SourceCardID
				if !effects.OnBattlefield(g, id) || ctx.IsNewSourceObject(id) {
					return nil
				}
				if err := effects.UntilEndOfTurn(ctx, id, nil,
					"Chromium, the Mutable — becomes a Human with base power and toughness 1/1 and loses all abilities until end of turn",
					game.SetCreatureTypesMod(ctx.TextSubtypes("Human")...),
					game.LoseAllAbilitiesMod(),
					game.SetBasePowerMod(1), game.SetBaseToughnessMod(1)); err != nil {
					return err
				}
				if err := (effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"hexproof"}, Label: "Chromium, the Mutable — hexproof until end of turn"}).Apply(ctx); err != nil {
					return err
				}
				return effects.RestrictUntilEOT{Target: id, Restrictions: game.CantBeBlocked, Label: "Chromium, the Mutable — can't be blocked this turn"}.Apply(ctx)
			},
		}},
	})
}
