package cardsto

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tolarianContemptCounter is the rejection counter's kind: a plain string,
// like every counter the engine has no rule for.
const tolarianContemptCounter = "rejection"

// Tolarian Contempt — Enchantment {3}{U}{U} (March of the Machine: The
// Aftermath, #5476):
//
//	"When this enchantment enters, put a rejection counter on each creature
//	 your opponents control.
//	 At the beginning of your end step, for each opponent, choose up to one
//	 target creature they control with a rejection counter on it. That
//	 creature's owner puts it on their choice of the top or bottom of their
//	 library."
//
// The enters trigger counts the creatures opponents control as it resolves.
// The end-step trigger is Molten Primordial's shape: one "up to one target
// creature that player controls with a rejection counter on it" clause per
// opponent seat, each bound to its opponent (TargetsFrom, CR 603.3d), so a
// creature that changed hands in response is no longer a legal target
// (CR 608.2b). Each creature still legal is tucked by ITS OWNER, who
// chooses the top or the bottom of their library (Endless Detour's route,
// so a commander's owner is offered the command zone, CR 903.9).
//
// No simplification.
func init() {
	endStep := effects.AtYourEndStep("Tolarian Contempt — for each opponent, choose up to one target creature they control with a rejection counter; its owner puts it on the top or bottom of their library",
		tolarianContemptTuck)
	endStep.TargetsFrom = tolarianContemptClauses
	effects.Register(effects.Spec{
		OracleID:     "831623c9-8279-439e-a42b-ac47cacdb058",
		Name:         "Tolarian Contempt",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Tolarian Contempt — put a rejection counter on each creature your opponents control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, c := range effects.MatchingBattlefield(ctx, effects.And(effects.OpponentControls(), effects.Creature())) {
						if err := (effects.AddCounter{Target: c.InstanceID, Kind: tolarianContemptCounter, N: 1}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
			endStep,
		},
	})
}

// tolarianContemptClauses is one "up to one target creature they control
// with a rejection counter on it" clause per opponent seat, in seat order.
// It reads only the seat list, which a started game never changes (see
// moltenPrimordialClauses), so a restore rebuilds the same list.
func tolarianContemptClauses(tc game.TriggerContext, source *game.Card, g *game.Game) *game.TargetSpec {
	controller := source.Controller
	if obj := tc.Object; obj != nil && obj.ID == source.InstanceID && obj.Controller != uuid.Nil {
		controller = obj.Controller
	}
	withCounter := func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.Counters[tolarianContemptCounter] > 0 }
	var clauses []*game.TargetSpec
	for _, p := range g.Seats {
		if p == nil || p.ID == controller {
			continue
		}
		clauses = append(clauses, effects.TargetCreature("up to one target creature "+p.Name+" controls with a rejection counter on it",
			effects.ControlledBy(p.ID), withCounter).WithCount(0, 1))
	}
	if len(clauses) == 0 {
		return nil
	}
	return effects.Clauses(clauses[0], clauses[1:]...)
}

// tolarianContemptTuck has each chosen creature's owner put it on the top
// or the bottom of their library.
func tolarianContemptTuck(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		c, ok := g.LookupCardForEffect(t.ID)
		if !ok || c.Owner == uuid.Nil {
			continue
		}
		if err := (effects.PutInLibraryInAnyOrder{
			Chooser:   c.Owner,
			Cards:     []uuid.UUID{t.ID},
			Placement: game.LibraryPlaceTopOrBottom,
			Label:     "Tolarian Contempt — put it on the top or bottom of your library",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
