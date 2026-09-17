package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Karn Liberated — Legendary Planeswalker — Karn for {7}, starting
// loyalty 6 (EDHREC rank 7173):
//
//	"+4: Target player exiles a card from their hand.
//	 −3: Exile target permanent.
//	 −14: Restart the game, leaving in exile all non-Aura permanent
//	      cards exiled with Karn. Then put those cards onto the
//	      battlefield under your control."
//
// The −3 is complete and is what the card does at a four-player
// table: colorless unconditional removal that answers anything,
// including an indestructible or regenerating permanent, because
// exile is not destruction. Routed through the engine's shared exit,
// so a COMMANDER exiled by it gets the CR 903.9 prompt and its owner
// may choose the command zone instead (#529 / #539).
//
// THE +4 IS WEAKER THAN PRINTED, AND DELIBERATELY SO: the target
// player DISCARDS the card instead of exiling it, so it lands in
// their graveyard where they can still get it back. The choice is
// still theirs — it is the ordinary pending-discard modal, the same
// one Mind Rot uses — and the only hand picker in the engine is that
// modal, with no exile lane on it. Every way this differs from the
// printed card favours the player being attacked: a discarded card
// can be reanimated, escaped or flashed back, and a card in exile
// cannot. That is the right direction for a simplification, and a
// Karn with no plus at all — able only to tick down toward an
// ultimate he does not have — would be the worse card.
//
// THE −14 IS NOT REGISTERED, and this is the one omission in the
// catalog that needs no argument. "Restart the game" is not an
// effect; it is a second game, seeded from the exile pile of the
// first, with new hands, new mulligans and a new turn order. There
// is no approximation of it that is recognisably the same ability,
// and a −14 that did something smaller would be a different card
// wearing Karn's text. The loyalty still accrues past 14.
func init() {
	Register(Spec{
		OracleID:     "0ca233f4-1b7f-4807-ab6e-2b1f5439b3db",
		Name:         "Karn Liberated",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The +4 makes the target player discard a card instead of exiling it, so it goes to their graveyard.",
			"The -14 isn't offered — restarting the game isn't something this engine can do.",
		},
		// Printed loyalty reaches the card through deck import
		// (ADR 0032 §1); this is the fallback for tokens, fixtures
		// and the dev spawner.
		StartingLoyalty: 6,
		Activated: []ActivatedAbility{
			{
				Label:   "+4: Target player exiles a card from their hand.",
				Cost:    LoyaltyCost(4),
				Targets: TargetPlayer("target player"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetPlayer {
							continue
						}
						g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
							Player: t.ID,
							Source: item.SourceCardID,
							N:      1,
						})
					}
					return nil
				},
			},
			{
				Label:   "−3: Exile target permanent.",
				Cost:    LoyaltyCost(-3),
				Targets: TargetPermanent("target permanent"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := (ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				},
			},
		},
	})
}
