package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spirit-Sister's Call — Enchantment {3}{W}{B} (Duskmourn Commander):
//
//	"At the beginning of your end step, choose target permanent card in your
//	 graveyard. You may sacrifice a permanent that shares a card type with
//	 the chosen card. If you do, return the chosen card from your graveyard
//	 to the battlefield and it gains 'If this permanent would leave the
//	 battlefield, exile it instead of putting it anywhere else.'"
//
// The chosen card is the trigger's target, so it is picked as the trigger
// goes on the stack and re-checked as it resolves (CR 608.2b). As it resolves
// you may sacrifice a permanent you control that shares a card type with the
// chosen card (game.SharedCardTypes, the layered types of the permanent and
// the printed ones of the card), asked over what you control then, with a
// floor of zero for the "may". "If you do" is the sacrifice that really
// happened (maySacrificeOne's keyed tail): the chosen card returns under your
// control, if it is still in your graveyard, and the exile-instead clause is
// a replacement pinned to the returned object (ExileInsteadOfLeavingBattlefield,
// Whip of Erebos' shape), which lasts as long as that object is on the
// battlefield.
//
// No simplification.
func init() {
	const name = "Spirit-Sister's Call"
	Register(Spec{
		OracleID:     "83914a83-4de2-4d64-9770-f68a12014167",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				AtYourEndStep(name+" — choose a permanent card in your graveyard; you may sacrifice a permanent sharing a card type with it to return it",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						chosen, ok := b16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						card, ok := g.LookupCardForEffect(chosen)
						if !ok {
							return nil
						}
						share := func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
							return len(game.SharedCardTypes([]game.Card{card, c})) > 0
						}
						return maySacrificeOne(g, item, name+" — sacrifice a permanent that shares a card type with "+card.Name+"?",
							share, false, spiritSistersCallReturn)
					}),
				TargetCardInGraveyard("target permanent card in your graveyard", YouOwn(), Permanent())),
		},
	})
}

// spiritSistersCallReturn is "if you do, return the chosen card from your
// graveyard to the battlefield and it gains ...". Keyed: an on-disk identity,
// never renamed or reused.
var spiritSistersCallReturn = RegisterSacrificedOne("prompt-run/spirit-sisters-call-return", func(ctx *Context, _ uuid.UUID) error {
	chosen, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (ReturnFromGraveyard{Target: chosen, Dest: game.ZoneBattlefield, Controller: ctx.Controller()}).Apply(ctx); err != nil {
		return err
	}
	if z := ctx.Game.FindCardZoneForEffect(chosen); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	ExileInsteadOfLeavingBattlefield(ctx.Game, chosen, ctx.Controller(),
		"Spirit-Sister's Call: if it would leave the battlefield, exile it instead")
	return nil
})
