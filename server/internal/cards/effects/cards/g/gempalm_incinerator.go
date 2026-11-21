package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gempalm Incinerator — Creature — Goblin {2}{R}, 2/1 (Dominaria
// Remastered, #5948):
//
//	"Cycling {1}{R} ({1}{R}, Discard this card: Draw a card.)
//	 When you cycle this card, you may have it deal X damage to target
//	 creature, where X is the number of Goblins on the battlefield."
//
// The cycle trigger fires from the graveyard (CR 702.29c) with a real
// target, like Agonasaur Rex's. The "may" is asked on resolution; X is
// counted then (the Incinerator is in the graveyard, so it is not one of
// the Goblins), and the damage is dealt by the card, whose last-known
// information is used (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "467ea3e3-7767-405a-9d1a-8fa5de696376",
		Name:         "Gempalm Incinerator",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Cycling("{1}{R}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.InGraveyard(effects.On(game.EventCycle, effects.Self,
					"Gempalm Incinerator — you may have it deal X damage to target creature, where X is the number of Goblins on the battlefield",
					func(g *game.Game, item *game.StackItem) error {
						return effects.MayChoice{
							Question: "Gempalm Incinerator — deal damage to the target creature equal to the number of Goblins on the battlefield?",
							Then:     gempalmIncineratorDamage,
						}.Apply(effects.NewContext(g, item))
					})),
				effects.TargetCreature("target creature"),
			),
		},
	})
}

var gempalmIncineratorDamage = effects.MayYes("may/gempalm-incinerator-damage", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	ts := ctx.LegalTargets()
	if len(ts) == 0 {
		return nil
	}
	goblin := ctx.TextSubtype("Goblin")
	n := 0
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.HasSubtype(goblin) {
			n++
		}
	}
	return effects.DealDamage{Source: ctx.Item.SourceCardID, Target: ts[0].ID, Amount: n}.Apply(ctx)
})
