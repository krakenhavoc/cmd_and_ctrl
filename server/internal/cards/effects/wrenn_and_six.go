package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wrenn and Six — Legendary Planeswalker — Wrenn for {R}{G},
// starting loyalty 3:
//
//	"+1: Return up to one target land card from your graveyard to
//	     your hand.
//	 −1: Wrenn and Six deals 1 damage to any target.
//	 −7: You get an emblem with 'Instant and sorcery cards in your
//	     graveyard have retrace.'"
//
// The two abilities that make the card are wired in full, and the −1
// is the reason S27's #406 fix had to land first: before it, a
// Wrenn's −1 aimed at an opposing planeswalker incremented a number
// nothing read.
//
// "Up to one target" is a Min 0 clause, so the +1 is activatable with
// an empty graveyard and still gains the loyalty. That is the printed
// card and it is most of why Wrenn is played — the plus is a
// loyalty engine first and a land recursion second.
//
// The −7 is the emblem "Instant and sorcery cards in your graveyard have
// retrace." (#2528). It is a STANDING cast permission declared on the
// emblem (EmblemSpec.CastPermissions) and derived from its owner's
// Player.Emblems on every query, so it covers an instant that reaches the
// graveyard after the ultimate, never ends (CR 114.2), and composes with
// a Six on the same table. Retrace is the card's printed mana cost plus a
// discarded land card, claimed as a priced offer (AltCostKey "retrace",
// DiscardLandCard); there is no turn restriction, so an instant is
// retraced at instant speed on any turn. See six.go and ADR 0066's
// 2026-10-07 amendment.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "108ae90a-50fa-4cfd-b751-d630e41425fe",
		Name:         "Wrenn and Six",
		Completeness: CompletenessFull,
		// Printed loyalty reaches the card through deck import
		// (ADR 0032 §1); this is the fallback for tokens, fixtures
		// and the dev spawner.
		StartingLoyalty: 3,
		Emblem: &EmblemSpec{
			Label: "Wrenn and Six emblem",
			Text:  "Instant and sorcery cards in your graveyard have retrace.",
			CastPermissions: []game.CastPermission{{
				Zone:            game.ZoneGraveyard,
				Filter:          game.PermissionFilter{InstantOrSorceryOnly: true},
				AltCostKey:      game.AltCostKeyRetrace,
				DiscardLandCard: true,
				Label:           "Retrace — discard a land card (Wrenn and Six emblem)",
			}},
		},
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Return up to one target land card from your graveyard to your hand.",
				Cost:    LoyaltyCost(1),
				Targets: upToOneLandInYourGraveyard(),
				Effect:  returnTargetedGraveyardCardsToHand,
			},
			{
				Label:   "−1: Wrenn and Six deals 1 damage to any target.",
				Cost:    LoyaltyCost(-1),
				Targets: TargetAny(),
				Purpose: ForTargets(DamageToTarget(0, 1)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if len(item.Targets) == 0 {
						return nil
					}
					return DealDamage{
						Source: item.SourceCardID,
						Target: item.Targets[0].ID,
						Amount: 1,
					}.Apply(ctx)
				},
			},
			{
				Label: "−7: You get an emblem with \"Instant and sorcery cards in your graveyard have retrace.\"",
				Cost:  LoyaltyCost(-7),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateEmblem{}.Apply(NewContext(g, item))
				},
			},
		},
	})
}

// upToOneLandInYourGraveyard is Wrenn's +1 clause. Min 0 is the "up
// to one": the picker's Done button is live with nothing selected, so
// the loyalty gain is available on an empty graveyard.
func upToOneLandInYourGraveyard() *game.TargetSpec {
	spec := TargetCardInGraveyard("up to one target land card in your graveyard", YouOwn(), Land())
	spec.Min = 0
	return spec
}

// returnTargetedGraveyardCardsToHand returns each still-legal targeted
// card from its owner's graveyard to their hand (Wrenn and Six's +1,
// Tamiyo, Collector of Tales' −3).
func returnTargetedGraveyardCardsToHand(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.Targets() {
		if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
			continue
		}
		if err := (ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
