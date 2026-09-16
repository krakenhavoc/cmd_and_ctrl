package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Liliana of the Veil — Legendary Planeswalker — Liliana for
// {1}{B}{B}, starting loyalty 3 (EDHREC rank 2971):
//
//	"+1: Each player discards a card.
//	 −2: Target player sacrifices a creature.
//	 −6: Separate all permanents target player controls into two
//	     piles. That player sacrifices all permanents in the pile of
//	     their choice."
//
// The two abilities the card is played for are complete, and both
// are choices made by the player who loses the card rather than by
// Liliana's controller:
//
//   - The +1 is the pending-discard modal, once per seat, the
//     controller included. It is symmetric on purpose — that is the
//     whole reason Liliana is a hard card to build around — and a
//     player with an empty hand discards nothing rather than erroring.
//   - The −2 is an edict, so it is not targeted at a creature and
//     hexproof does not save one. The victim picks from their own
//     creatures through the sacrifice prompt (CR 701.21); a player
//     with no creature sacrifices nothing, and the loyalty is still
//     paid, exactly as in paper.
//
// THE −6 IS NOT REGISTERED. "Separate all permanents into two piles"
// is a two-stage choice by two different players — Liliana's
// controller divides, then the victim picks a pile — and the engine
// has no pile-division prompt of any kind. The alternatives are all
// worse than the omission: dividing the piles automatically makes
// the most important decision on the card for the player who should
// be making it, and a −6 that sacrificed some fixed fraction would
// be a different card. The loyalty still accrues past 6; the day a
// division prompt exists, this file gains one entry.
func init() {
	Register(Spec{
		OracleID:     "0ba134d8-ee7d-48ec-8dc6-57942b8e9261",
		Name:         "Liliana of the Veil",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The -6 isn't offered — splitting a player's permanents into two piles has no prompt yet."},
		// Printed loyalty reaches the card through deck import
		// (ADR 0032 §1); this is the fallback for tokens, fixtures
		// and the dev spawner.
		StartingLoyalty: 3,
		Activated: []ActivatedAbility{
			{
				Label: "+1: Each player discards a card.",
				Cost:  LoyaltyCost(1),
				Effect: func(g *game.Game, _ *game.StackItem) error {
					for _, p := range g.Seats {
						if p == nil || p.Eliminated {
							continue
						}
						g.DiscardChoiceForEffect(p.ID, 1)
					}
					return nil
				},
			},
			{
				Label:   "−2: Target player sacrifices a creature.",
				Cost:    LoyaltyCost(-2),
				Targets: TargetPlayer("target player"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetPlayer {
							continue
						}
						g.PlayerSacrificesForEffect(item.SourceCardID, t.ID,
							sacrificeSpec("a creature", Creature()),
							"Liliana of the Veil — sacrifice a creature")
					}
					return nil
				},
			},
		},
	})
}
