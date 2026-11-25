package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Liliana, Waker of the Dead — Legendary Planeswalker — Liliana {2}{B}{B},
// loyalty 4 (Core Set 2021, #5511):
//
//	"+1: Each player discards a card. Each opponent who can't loses 3 life.
//	 −3: Target creature gets -X/-X until end of turn, where X is the number
//	 of cards in your graveyard.
//	 −7: You get an emblem with "At the beginning of combat on your turn, put
//	 target creature card from a graveyard onto the battlefield under your
//	 control. It gains haste.""
//
// The +1 sorts the opponents first: one with no card in hand can't discard and
// loses 3; every player with a card discards one (a prompt each, the
// controller's too). The −3 counts your graveyard as it resolves. The emblem
// trigger chooses a creature card in any graveyard as it goes on the stack,
// returns it under the emblem owner's control and gives it haste with no
// duration (a CR 611.2a effect on that object).
//
// No simplification.
func init() {
	const emblemText = "At the beginning of combat on your turn, put target creature card from a graveyard onto the battlefield under your control. It gains haste."
	effects.Register(effects.Spec{
		OracleID:        "e0df8e2a-96d2-4e4d-b065-aa2e162abbfe",
		Name:            "Liliana, Waker of the Dead",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Emblem: &effects.EmblemSpec{
			Label: "Liliana, Waker of the Dead emblem",
			Text:  emblemText,
			Triggered: []game.TriggeredAbility{
				effects.Targeting(
					effects.AtBeginningOfYourCombat("Liliana, Waker of the Dead emblem — put target creature card from a graveyard onto the battlefield under your control; it gains haste",
						func(g *game.Game, item *game.StackItem) error {
							ctx := effects.NewContext(g, item)
							return effects.ReturnFromGraveyardTogether{
								Targets:    effects.LegalTargetCardIDs(ctx),
								Controller: ctx.Controller(),
								ThenTail:   lilianaWakerHaste,
							}.Apply(ctx)
						}),
					effects.TargetCardInGraveyard("target creature card from a graveyard", effects.Creature())),
			},
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:  "+1: Each player discards a card. Each opponent who can't loses 3 life.",
				Cost:   effects.LoyaltyCost(1),
				Effect: effects.Do(effects.Instruction(lilianaWakerDiscard)),
			},
			{
				Label:   "−3: Target creature gets -X/-X until end of turn, where X is the number of cards in your graveyard.",
				Cost:    effects.LoyaltyCost(-3),
				Targets: effects.TargetCreature("target creature"),
				Purpose: effects.ForTargets(effects.HarmUntilEOT(0, 0, 0)),
				Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					t, ok := ctx.ClauseTarget(0)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
					if p == nil {
						return nil
					}
					x := p.Graveyard.Size()
					return effects.BoostUntilEOT{Target: t.ID, Power: -x, Toughness: -x, Label: "Liliana, Waker of the Dead — -X/-X"}.Apply(ctx)
				})),
			},
			{
				Label:  "−7: You get an emblem with \"" + emblemText + "\"",
				Cost:   effects.LoyaltyCost(-7),
				Effect: effects.Do(effects.CreateEmblem{}),
			},
		},
	})
}

// lilianaWakerDiscard is the +1: opponents who can't discard lose 3 life,
// then every player who can discards a card.
func lilianaWakerDiscard(ctx *effects.Context) error {
	var discarders []uuid.UUID
	for _, p := range ctx.Game.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		if p.Hand.Size() == 0 {
			if p.ID != ctx.Controller() {
				if err := (effects.GainLife{Player: p.ID, Amount: -3}).Apply(ctx); err != nil {
					return err
				}
			}
			continue
		}
		discarders = append(discarders, p.ID)
	}
	for _, id := range discarders {
		ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
			Player:   id,
			Source:   ctx.Source(),
			N:        1,
			Question: "Liliana, Waker of the Dead — discard a card",
		})
	}
	return nil
}

// lilianaWakerHaste is "It gains haste." over the card the emblem returned,
// for as long as it is that object. An on-disk key.
var lilianaWakerHaste = effects.CardTail("then/s82-old-b107-liliana-waker-haste", func(ctx *effects.Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		if err := (effects.ScopedEffectFor{
			Target:   id,
			Mods:     []game.Mod{game.AddKeywordsMod("haste")},
			Duration: ctx.Game.PinnedTo(game.IndefiniteDuration(), id),
			Label:    "Liliana, Waker of the Dead emblem — haste",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
