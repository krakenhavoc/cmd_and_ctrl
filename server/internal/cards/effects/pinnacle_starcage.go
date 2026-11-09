package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pinnacle Starcage — Artifact {1}{W}{W} (Edge of Eternities):
//
//	"When this artifact enters, exile all artifacts and creatures with
//	 mana value 2 or less until this artifact leaves the battlefield.
//	 {6}{W}{W}: Put each card exiled with this artifact into its owner's
//	 graveyard, then create a 2/2 colorless Robot artifact creature token
//	 for each card put into a graveyard this way. Sacrifice this
//	 artifact."
//
// The seam is the pairing of two engine halves no card had used together:
// ONE simultaneous exile of a whole group that is both "linked" (CR
// 607.2a: Card.ExiledWith names this Starcage, so the second ability can
// read "each card exiled with this artifact") and "until" (CR 610.3: a
// record per card, returning it when this object leaves the battlefield).
// Game.ExileLinkedKeyedForEffect does the exile and hands the cards that
// reached exile to a tail, which writes the return records. Tokens exiled
// this way cease to exist and are never recorded or returned; a token's
// mana value is 0 unless it copies something, and an X in a cost is 0.
// An entry trigger that resolves after the Starcage has left exiles
// nothing (CR 610.3a).
//
// The activated ability puts the linked cards into their owners'
// graveyards as one move (a card that left exile since is no longer
// linked), makes a Robot for each card that really arrived, and only then
// sacrifices the Starcage, so the sacrifice does not return the cards it
// just buried. Both halves read the Starcage as the object that made
// them: a Starcage that left and came back claims nothing (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "10e0b563-4dfc-4530-a010-0ba40bdf02ac",
		Name:         "Pinnacle Starcage",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Pinnacle Starcage — exile all artifacts and creatures with mana value 2 or less until this artifact leaves the battlefield", starcageExile),
		},
		Activated: []ActivatedAbility{{
			Label:  "{6}{W}{W}: Put each card exiled with this artifact into its owner's graveyard, then create a 2/2 colorless Robot artifact creature token for each card put into a graveyard this way. Sacrifice this artifact.",
			Cost:   ManaCost("{6}{W}{W}"),
			Effect: starcageBury,
		}},
	})
}

func starcageExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
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
		if (c.IsArtifact() || c.IsCreature()) && c.ManaValue() <= 2 {
			ids = append(ids, c.InstanceID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return g.ExileLinkedKeyedForEffect(ref, ids, game.KeyedTail{
		Ref: starcageRecorded, Player: item.Controller, Source: item.SourceCardID,
		Carry: []uuid.UUID{ref.ID}, N: ref.Epoch,
	})
}

// starcageRecorded writes the "until this leaves the battlefield" record
// of every card that reached exile. An on-disk key: never renamed, never
// reused.
var starcageRecorded = game.RegisterTail("pinnacle-starcage/record-returns", func(g *game.Game, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	for _, id := range o.IDs {
		g.ScheduleUntilReturnForEffect(id, game.UntilReturn{
			Controller:   o.Tail.Player,
			SourceCardID: o.Tail.Source,
			Label:        "Pinnacle Starcage — the exiled card returns when it leaves the battlefield",
			Leaves:       game.ObjectRef{ID: o.Tail.Carry[0], Epoch: o.Tail.N},
		})
	}
	return nil
})

func starcageBury(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		ids = append(ids, c.InstanceID)
	}
	tail := ctx.ItemTail(starcageRobots)
	if len(ids) == 0 {
		return tail.Run(g, nil)
	}
	return g.PutCardsIntoGraveyardKeyedForEffect(ids, tail)
}

// starcageRobots is "then create a Robot for each card put into a
// graveyard this way. Sacrifice this artifact." An on-disk key: never
// renamed, never reused.
var starcageRobots = CardTail("pinnacle-starcage/robots-then-sacrifice", func(ctx *Context, o game.TailOutcome) error {
	if n := len(o.IDs); n > 0 {
		if err := (CreateToken{Template: TokenCard("2/2 colorless Robot artifact"), N: n}).Apply(ctx); err != nil {
			return err
		}
	}
	return SacrificePermanent{Target: ctx.Source()}.Apply(ctx)
})
