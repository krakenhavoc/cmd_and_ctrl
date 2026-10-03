package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Take the Bait — Instant {2}{R}{W}:
//
//	"Cast this spell only during combat on an opponent's turn.
//	 Prevent all combat damage that would be dealt to you and
//	 planeswalkers you control this turn. Untap all attacking creatures
//	 and goad them. After this phase, there is an additional combat
//	 phase."
//
// ADR 0108 §7, Delivery PR 7 (#1904): the combat-only not-one-use shield
// protecting you and the planeswalkers you control, read as the damage
// would be dealt; then Karlach's untap of every attacking creature, your
// goad on each of them (until your next turn, CR 701.15a), and the extra
// combat phase after this one (ADR 0059's turn plan).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "5d39cce9-afa1-422a-8e6c-8f50c710a681",
		Name:         "Take the Bait",
		Completeness: CompletenessFull,
		CastCondition: func(g *game.Game, controller uuid.UUID, _ game.Card) bool {
			active := activePlayerIDOf(g)
			return game.PhaseOf(g.Turn.Step) == game.PhaseCombat && active != uuid.Nil && active != controller
		},
		CastConditionLabel: "Cast this spell only during combat on an opponent's turn.",
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			shield := PreventDamageFromSource{Protect: ShieldYouAndYourPermanents("planeswalker"), CombatOnly: true}
			if err := shield.Apply(ctx); err != nil {
				return err
			}
			attackers := AttackingCreatures(ctx.Game)
			if err := untapEach(ctx, attackers); err != nil {
				return err
			}
			if err := GoadAllMatching(ctx, func(_ *game.Game, c game.Card) bool {
				return slices.Contains(attackers, c.InstanceID)
			}); err != nil {
				return err
			}
			return ExtraCombatAfterThisPhase().Apply(ctx)
		},
	})
}
