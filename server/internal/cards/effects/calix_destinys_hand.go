package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Calix, Destiny's Hand — Legendary Planeswalker — Calix {2}{G}{W},
// loyalty 4 (Commander Masters, #5948):
//
//	"+1: Look at the top four cards of your library. You may reveal an
//	 enchantment card from among them and put that card into your hand. Put
//	 the rest on the bottom of your library in a random order.
//	 −3: Exile target creature or enchantment you don't control until target
//	 enchantment you control leaves the battlefield.
//	 −7: Return all enchantment cards from your graveyard to the battlefield."
//
// The −3 has two targets (CR 115.1): the permanent to exile and the
// enchantment whose leaving brings it back. The return is tied to that
// second object (CR 610.3): if it has already left when the ability
// resolves, nothing is exiled (CR 610.3c). If the first target is illegal
// the enchantment's leaving still has nothing to return.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d9f01177-2237-475e-ad91-114bfae077ce",
		Name:            "Calix, Destiny's Hand",
		Completeness:    CompletenessFull,
		StartingLoyalty: 4,
		Activated: []ActivatedAbility{
			{
				Label: "+1: Look at the top four cards of your library. You may reveal an enchantment card from among them and put that card into your hand. Put the rest on the bottom of your library in a random order.",
				Cost:  LoyaltyCost(1),
				Effect: Do(Instruction(func(ctx *Context) error {
					player := ctx.Controller()
					return TakeFromLibraryToHand{
						Player:   player,
						Cards:    ctx.Game.LookAtTopOfLibraryForEffect(player, 4),
						Match:    Enchantment(),
						Max:      1,
						Optional: true,
						Label:    "Calix, Destiny's Hand — reveal an enchantment card and put it into your hand",
						Key:      takeRevealRestBottomRandom,
					}.Apply(ctx)
				})),
			},
			{
				Label: "−3: Exile target creature or enchantment you don't control until target enchantment you control leaves the battlefield.",
				Cost:  LoyaltyCost(-3),
				Targets: Clauses(
					TargetPermanent("target creature or enchantment you don't control", Or(Creature(), Enchantment()), Not(YouControl())),
					TargetPermanent("target enchantment you control", Enchantment(), YouControl()),
				),
				Effect: calixExileUntilEnchantmentLeaves,
			},
			{
				Label: "−7: Return all enchantment cards from your graveyard to the battlefield.",
				Cost:  LoyaltyCost(-7),
				Effect: Do(Instruction(func(ctx *Context) error {
					return b36ReturnAllEnchantmentCards(ctx)
				})),
			},
		},
	})
}

// calixExileUntilEnchantmentLeaves exiles the first target until the second
// leaves the battlefield. Either target gone (CR 608.2b) means no exile.
func calixExileUntilEnchantmentLeaves(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, t := range item.Targets {
		if t.Kind == game.TargetCard && ctx.IsTargetLegal(t) {
			ids = append(ids, t.ID)
		} else {
			ids = append(ids, uuid.Nil)
		}
	}
	if len(ids) != 2 || ids[0] == uuid.Nil || ids[1] == uuid.Nil {
		return nil
	}
	ref, ok := g.PermanentRefForEffect(ids[1])
	if !ok {
		return nil
	}
	return ExileUntil{
		Target:      ids[0],
		UntilLeaves: ref,
		Label:       "Calix, Destiny's Hand — returns when the enchantment leaves the battlefield",
	}.Apply(ctx)
}
