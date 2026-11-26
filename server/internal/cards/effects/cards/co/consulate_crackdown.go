package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Consulate Crackdown — Enchantment {3}{W}{W} (Aether Revolt, S82 older sets,
// #5536):
//
//	"When this enchantment enters, exile all artifacts your opponents control
//	 until this enchantment leaves the battlefield."
//
// Pinnacle Starcage's linked exile with another set of permanents: the
// artifacts are exiled together as the trigger resolves, each is recorded
// to return when this enchantment leaves (CR 610.3, 611.2c), and one that
// is already gone, or an enchantment that left before the trigger resolved,
// exiles nothing at all (CR 610.3c). A card an opponent owned comes back
// under its owner's control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "128f7e5a-17e0-4069-a185-c9f874c7d234",
		Name:         "Consulate Crackdown",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepArtifacts, How: game.SweepExile, Partial: true}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Consulate Crackdown — exile all artifacts your opponents control until this enchantment leaves the battlefield", consulateCrackdownExile),
		},
	})
}

func consulateCrackdownExile(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	u := game.UntilReturn{Controller: item.Controller, SourceCardID: item.SourceCardID, Leaves: ref}
	if g.UntilEventHappenedForEffect(u, item) {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsArtifact() && c.Controller != item.Controller {
			ids = append(ids, c.InstanceID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return g.ExileLinkedKeyedForEffect(ref, ids, game.KeyedTail{
		Ref: consulateCrackdownRecorded, Player: item.Controller, Source: item.SourceCardID,
		Carry: []uuid.UUID{ref.ID}, N: ref.Epoch,
	})
}

// consulateCrackdownRecorded schedules each exiled artifact's return for when
// the enchantment leaves. An on-disk key: never renamed.
var consulateCrackdownRecorded = game.RegisterTail("consulate-crackdown/record-returns", func(g *game.Game, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	for _, id := range o.IDs {
		g.ScheduleUntilReturnForEffect(id, game.UntilReturn{
			Controller:   o.Tail.Player,
			SourceCardID: o.Tail.Source,
			Label:        "Consulate Crackdown — the exiled artifact returns when this enchantment leaves the battlefield",
			Leaves:       game.ObjectRef{ID: o.Tail.Carry[0], Epoch: o.Tail.N},
		})
	}
	return nil
})
