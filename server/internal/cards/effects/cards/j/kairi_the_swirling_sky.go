package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kairi, the Swirling Sky — {4}{U}{U} Legendary Creature — Dragon Spirit
// 6/6 (Kamigawa: Neon Dynasty, S82 NEO batch 20):
//
//	"Flying, ward {3}
//	 When Kairi dies, choose one —
//	 • Return any number of target nonland permanents with total mana
//	   value 6 or less to their owners' hands.
//	 • Mill six cards, then return up to two instant and/or sorcery cards
//	   from your graveyard to your hand."
//
// The first mode targets with the total-mana-value set rule (#4616); the
// bounce is the permanents still legal as it resolves. The second mode
// mills first and then offers the graveyard, so a card the mill put there
// can be taken.
//
// No simplification.
func init() {
	dies := effects.WhenThisDies("Kairi, the Swirling Sky — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	dies.Modes = effects.ChooseOne(
		effects.ModeDoing("Return any number of target nonland permanents with total mana value 6 or less to their owners' hands.",
			effects.TargetPermanent("any number of target nonland permanents with total mana value 6 or less", effects.Nonland()).
				WithCount(0, 0).WithTotal(effects.TotalManaValueAtMost(6)),
			func(_ *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.NeoB20BounceTargetedPermanents(ctx)
			}),
		effects.ModeDoing("Mill six cards, then return up to two instant and/or sorcery cards from your graveyard to your hand.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				if err := (effects.MillCards{Player: item.Controller, N: 6}).Apply(ctx); err != nil {
					return err
				}
				// The pick waits for any prompt the mill asks (CR 608.2c, #4155).
				return effects.ReturnChosenFromGraveyard{
					Player:   item.Controller,
					Question: "Kairi, the Swirling Sky — you may return up to two instant and/or sorcery cards from your graveyard to your hand",
					Match:    func(c game.Card) bool { return c.IsInstant() || c.IsSorcery() },
					Max:      2,
					ToHand:   true,
				}.Apply(ctx)
			}),
	)
	effects.Register(effects.Spec{
		OracleID:        "85aa53e8-f062-4d02-bc69-1413f7f0dcf1",
		Name:            "Kairi, the Swirling Sky",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{3}"), "Kairi, the Swirling Sky — ward {3}"),
			dies,
		},
	})
}
