package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Call for Aid — Sorcery {4}{R} (Tales of Middle-earth Commander):
//
//	"Gain control of all creatures target opponent controls until end
//	 of turn. Untap those creatures. They gain haste until end of turn.
//	 You can't attack that player this turn. You can't sacrifice those
//	 creatures this turn."
//
// Insurrection's three primitives (gain control, untap, haste) run over
// the creatures that player controls as the spell resolves, locked in
// at that moment (CR 611.2c). "You can't attack that player this turn"
// is the CR 508.1c grant Web of Inertia uses, from the caster to the
// opponent, covering the player and the permanents they control. "You
// can't sacrifice those creatures" is Slicer, Hired Muscle's
// pinned "can't be sacrificed this turn", read by every cost, choice
// and instruction (CR 101.2); the victim cannot sacrifice them either
// way, since they no longer control them.
//
// If the target is illegal on resolution the spell does nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	const name = "Call for Aid"
	effects.Register(effects.Spec{
		OracleID:     "6d650eef-1156-4e3b-8fa8-90a6ee04d9a8",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			t, ok := effects.FirstLegalTarget(ctx)
			if !ok || t.Kind != game.TargetPlayer {
				return nil
			}
			victim := t.ID
			taken := effects.MatchingBattlefield(ctx, effects.And(effects.Creature(), func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
				return c.Controller == victim
			}))
			for _, c := range taken {
				if err := (effects.GainControl{
					Target:   c.InstanceID,
					Duration: effects.DurationUntilEndOfTurn(ctx),
					Label:    name + " — gain control until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			for _, c := range taken {
				if err := (effects.UntapTarget{Target: c.InstanceID}).Apply(ctx); err != nil {
					return err
				}
			}
			for _, c := range taken {
				if err := (effects.GrantKeywordUntilEOT{
					Target:   c.InstanceID,
					Keywords: []string{"haste"},
					Label:    name + " — haste until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			ctx.Game.GrantCantAttackPlayerThisTurnForEffect(item.Controller, victim, game.CantAttackScope{},
				name+" — you can't attack that player this turn", item.SourceCardID)
			for _, c := range taken {
				if err := (effects.CantBeSacrificedThisTurn{Target: c.InstanceID, Label: name + " — can't be sacrificed this turn"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
