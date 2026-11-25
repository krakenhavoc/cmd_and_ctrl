package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elite Spellbinder — Creature — Human Cleric {2}{W}, 3/1 (Strixhaven: School
// of Mages, #5502):
//
//	"Flying
//	 When this creature enters, look at target opponent's hand. You may exile
//	 a nonland card from it. For as long as that card remains exiled, its owner
//	 may play it. A spell cast this way costs {2} more to cast."
//
// Word of Command's look (only you learn the hand) over a targeted opponent,
// with Thoughtseize's nonland pick made optional. The card stays in the hand
// through the pick (PickKeep) and is then exiled by Zero, Incorporeal Sensate's
// primitive, ExileWithPermission: its owner (not you) may play it for as long as
// it remains exiled, and a spell cast that way costs {2} more (CR 118.8,
// CR 601.2f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8c4d7cbc-9cfc-4f0a-8844-e458ab806056",
		Name:            "Elite Spellbinder",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Elite Spellbinder — look at target opponent's hand; you may exile a nonland card from it",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						opp, ok := effects.FirstLegalPlayerTarget(ctx)
						if !ok {
							return nil
						}
						return effects.ChooseFromRevealedHand{
							Player:   opp,
							Filter:   effects.Nonland(),
							Label:    "nonland card",
							Look:     true,
							Optional: true,
							Keep:     true,
							Then:     eliteSpellbinderExile,
						}.Apply(ctx)
					}),
				effects.TargetPlayer("target opponent", effects.Opponent()),
			),
		},
	})
}

// eliteSpellbinderExile is the rest of the ability: the chosen card, still in
// the hand, is exiled with its owner's permission to play it (and a {2}
// surcharge) for as long as it stays there.
var eliteSpellbinderExile = effects.RevealedPickThen("revealed-pick/elite-spellbinder-exile",
	func(ctx *effects.Context, pick game.RevealedPick) error {
		for _, c := range pick.Chosen {
			if err := (effects.ExileWithPermission{Target: c.InstanceID, WhileExiled: true, CostMore: 2}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
