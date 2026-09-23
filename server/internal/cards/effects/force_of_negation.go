package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Force of Negation — Instant {1}{U}{U} (EDHREC rank 266):
//
//	"If it's not your turn, you may exile a blue card from your hand
//	 rather than pay this spell's mana cost.
//	 Counter target noncreature spell. If that spell is countered
//	 this way, exile it instead of putting it into its owner's
//	 graveyard."
//
// The Pitch shape (alternative_cost.go) with a Condition instead of a
// life cost — Force of Will's exile-a-blue-card cost, gated on "not
// your turn" through the same isActivePlayer read
// OpponentsCantCastDuringYourTurn uses.
//
// The exile clause is CounterTarget.Dest (#1230): counterSpellLocked
// already accepts a destination zone through
// `Game.CounterTargetToZoneForEffect`, the same shape Devious
// Cover-Up and Remand use — see the sweep note on
// docs/engine-seams.md's now-closed "Counter-to-hand / counter-to-zone
// effect surface" row.
func init() {
	Register(Spec{
		OracleID:     "ac2173f9-f223-440a-9231-fd98762bdc6f",
		Name:         "Force of Negation",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target noncreature spell", Noncreature()),
		AlternativeCosts: []game.AlternativeCost{
			{
				Key:           "pitch",
				Label:         "If it's not your turn, exile a blue card from your hand",
				ExileFromHand: CardInYourHand("a blue card from your hand", OfColor("U")),
				PayLabel:      "a blue card from your hand",
				Condition: func(g *game.Game, controller uuid.UUID) bool {
					return !isActivePlayer(g, controller)
				},
			},
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return CounterTarget{
				StackID: item.Targets[0].ID,
				Dest:    game.ZoneRef{Kind: game.ZoneExile},
			}.Apply(ctx)
		},
	})
}
