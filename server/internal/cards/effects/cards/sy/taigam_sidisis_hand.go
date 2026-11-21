package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Taigam, Sidisi's Hand — Legendary Creature — Human Wizard {3}{U}{B}, 3/4
// (Commander 2017, #5533):
//
//	"Skip your draw step.
//	 At the beginning of your upkeep, look at the top three cards of your
//	 library. Put one of them into your hand and the rest into your
//	 graveyard.
//	 {B}, {T}, Exile X cards from your graveyard: Target creature gets
//	 -X/-X until end of turn."
//
// "Skip your draw step" is Necropotence's SkipYourDrawStep (CR 614.10),
// lasting as long as Taigam is on the battlefield. The upkeep look is
// Professor Onyx's +1: exactly one card into the hand and the rest into the
// graveyard (takeRestGraveyard), not milled or discarded. The activated
// ability is Winter, Cursed Rider's "Exile X cards" cost with no {X} in the
// mana: the activator announces X by naming the cards, and the effect reads
// it back as the number exiled; a target gone by resolution makes it do
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "677aa518-072b-46d2-b0da-da57c7bc2bba",
		Name:         "Taigam, Sidisi's Hand",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SkipYourDrawStep()},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Taigam, Sidisi's Hand — look at the top three cards; one into your hand, the rest into your graveyard",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					player := ctx.Controller()
					return effects.TakeFromLibraryToHand{
						Player: player,
						Cards:  g.LookAtTopOfLibraryForEffect(player, 3),
						Max:    1,
						Label:  "Taigam, Sidisi's Hand — put one of them into your hand and the rest into your graveyard",
						Key:    effects.TakeRestGraveyard,
					}.Apply(ctx)
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}, {T}, Exile X cards from your graveyard: Target creature gets -X/-X until end of turn.",
			Purpose: effects.ForTargets(effects.HarmUntilEOT(0, -1, -1)),
			Cost: effects.Plus(effects.ManaCost("{B}"), effects.TapCost(),
				effects.ExileXFromGraveyardAsCost("X cards", func(game.Card) bool { return true })),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				x := len(ctx.Exiled())
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok || x == 0 {
					return nil
				}
				return effects.BoostUntilEOT{Target: id, Power: -x, Toughness: -x, Label: "Taigam, Sidisi's Hand — -X/-X"}.Apply(ctx)
			},
		}},
	})
}
