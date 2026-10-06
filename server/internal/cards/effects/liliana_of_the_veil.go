package effects

import (
	"github.com/google/uuid"

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
//
//   - The −2 is an edict, so it is not targeted at a creature and
//     hexproof does not save one. The victim picks from their own
//     creatures through the sacrifice prompt (CR 701.21); a player
//     with no creature sacrifices nothing, and the loyalty is still
//     paid, exactly as in paper.
//
//   - The −6 is the PileSplit machinery Do or Die uses, over every
//     permanent the target player controls: Liliana's controller
//     separates them into two piles (either may be empty), then the
//     target player chooses which pile they sacrifice. That is the
//     printed order of decisions — you divide, they choose. The
//     sacrifice is one simultaneous event (CR 701.21), it does not
//     target a permanent, and a permanent that left before the choice
//     was answered is not sacrificed. A target who controls nothing
//     resolves nothing and asks nothing. Until #2154 this ability was
//     omitted because no pile prompt existed; Do or Die (#2084) shipped
//     the first one over permanents.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0ba134d8-ee7d-48ec-8dc6-57942b8e9261",
		Name:         "Liliana of the Veil",
		Completeness: CompletenessFull,
		// Printed loyalty reaches the card through deck import
		// (ADR 0032 §1); this is the fallback for tokens, fixtures
		// and the dev spawner.
		StartingLoyalty: 3,
		Activated: []ActivatedAbility{
			{
				Label: "+1: Each player discards a card.",
				Cost:  LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					for _, p := range g.Seats {
						if p == nil || p.Eliminated {
							continue
						}
						g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
							Player: p.ID,
							Source: item.SourceCardID,
							N:      1,
						})
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
			{
				Label: "−6: Separate all permanents target player controls into two piles. That player sacrifices all permanents in the pile of their choice.",
				// ADR 0126 §6: one of two piles of the target player's permanents.
				Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepAllPermanents, How: game.SweepSacrifice, OpponentsOnly: true, Partial: true}},
				Cost:    LoyaltyCost(-6),
				Targets: TargetPlayer("target player"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetPlayer {
							continue
						}
						return PileSplit{
							Splitter:      ctx.Controller(),
							Chooser:       t.ID,
							Owner:         t.ID,
							SplitQuestion: "Liliana of the Veil — separate these permanents into two piles",
							PickQuestion:  "Liliana of the Veil — choose a pile; you sacrifice every permanent in it",
							Cards:         permanentsControlledByPlayer(g, t.ID),
							Then: func(ctx *Context, sacrificed, _ []uuid.UUID) error {
								ctx.Game.SacrificeAllForEffect(ctx.Source(), sacrificed)
								return nil
							},
						}.Apply(ctx)
					}
					return nil
				},
			},
		},
	})
}

// permanentsControlledByPlayer is every permanent on the battlefield
// the player controls, in battlefield order.
func permanentsControlledByPlayer(g *game.Game, playerID uuid.UUID) []uuid.UUID {
	if playerID == uuid.Nil {
		return nil
	}
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == playerID {
			out = append(out, c.InstanceID)
		}
	}
	return out
}
