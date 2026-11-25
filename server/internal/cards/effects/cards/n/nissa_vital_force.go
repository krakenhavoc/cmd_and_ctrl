package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nissa, Vital Force — Legendary Planeswalker — Nissa {3}{G}{G}, starting
// loyalty 5 (Kaladesh, #5539):
//
//	"+1: Untap target land you control. Until your next turn, it becomes a 5/5
//	 Elemental creature with haste. It's still a land.
//	 −3: Return target permanent card from your graveyard to your hand.
//	 −6: You get an emblem with "Whenever a land you control enters, you may
//	 draw a card.""
//
// The +1 is Wrenn and Realmbreaker's animation (one data ScopedEffectFor: type,
// subtype, base 5/5 and haste, pinned to the land until your next turn) after
// an untap of the same target; a target that has gone does neither (CR 608.2b).
// The −3 is a graveyard-to-hand return of a permanent card. The −6's emblem
// carries Aesi's landfall draw as a "you may" (CR 114.3, 603.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "31d4d8ce-36ba-4297-8c7a-34a83ba25255",
		Name:            "Nissa, Vital Force",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Emblem: &effects.EmblemSpec{
			Label: "Nissa, Vital Force emblem",
			Text:  "Whenever a land you control enters, you may draw a card.",
			Triggered: []game.TriggeredAbility{
				effects.OptionalDraw(
					effects.LandfallPaying("Nissa, Vital Force emblem — you may draw a card (landfall)", game.LandfallPayoff{Draws: 1},
						effects.Do(effects.DrawCards{N: 1})),
					"Nissa, Vital Force emblem — draw a card?", 1),
			},
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Untap target land you control. Until your next turn, it becomes a 5/5 Elemental creature with haste. It's still a land.",
				Cost:    effects.LoyaltyCost(1),
				Targets: effects.TargetPermanent("target land you control", effects.And(effects.Land(), effects.YouControl())),
				Effect:  nissaVitalForcePlusOne,
			},
			{
				Label:   "−3: Return target permanent card from your graveyard to your hand.",
				Cost:    effects.LoyaltyCost(-3),
				Targets: effects.TargetCardInGraveyard("target permanent card in your graveyard", effects.YouOwn(), effects.Permanent()),
				Effect:  effects.ReturnTargetGraveyardCardToHand,
			},
			{
				Label:  "−6: You get an emblem with \"Whenever a land you control enters, you may draw a card.\"",
				Cost:   effects.LoyaltyCost(-6),
				Effect: effects.Do(effects.CreateEmblem{}),
			},
		},
	})
}

// nissaVitalForcePlusOne untaps the chosen land, then makes it a 5/5 Elemental
// creature with haste until its controller's next turn. One data
// ScopedEffectFor registration, so the record survives a snapshot restore.
func nissaVitalForcePlusOne(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (effects.UntapTarget{Target: id}).Apply(ctx); err != nil {
		return err
	}
	return effects.ScopedEffectFor{
		Target: id,
		Mods: []game.Mod{
			game.AddTypesMod("Creature"),
			game.AddSubtypesMod(ctx.TextSubtypes("Elemental")...),
			game.SetBasePowerMod(5),
			game.SetBaseToughnessMod(5),
			game.AddKeywordsMod("haste"),
		},
		Duration: effects.DurationUntilYourNextTurn(ctx, ctx.Controller()),
		Label:    "Nissa, Vital Force — becomes a 5/5 Elemental until your next turn",
	}.Apply(ctx)
}
