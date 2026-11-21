package cardsex

import (
	"encoding/binary"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Expert-Level Safe — Artifact {2} (Fallout, #5948):
//
//	"When this artifact enters, exile the top two cards of your library face
//	 down.
//	 {1}, {T}: You and target opponent each secretly choose 1, 2, or 3. Then
//	 those choices are revealed. If they match, sacrifice this artifact and
//	 put all cards exiled with it into their owners' hands. Otherwise,
//	 exile the top card of your library face down."
//
// The exiles are Asmodeus the Archfiend's face-down linked exile (CR 406.3a,
// CR 607.2a): nobody may look at the cards, and they are linked to this
// Safe as the object it is now. The choice is Prisoner's Dilemma's secret
// choice, asked of you and then the target opponent only (a choice, not a
// vote). The Safe's object epoch rides the choice (s82GapB60EncodeEpoch) so
// the cards exiled with it are still found when the Safe has left in the
// meantime (CR 608.2h); a Safe that is no longer on the battlefield as that
// object is not sacrificed, and the cards go to their owners' hands all the
// same.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cf9129a3-25fe-4787-b41c-0b6968ba5aee",
		Name:         "Expert-Level Safe",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Expert-Level Safe — exile the top two cards of your library face down",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					return s82GapB60ExileTopFaceDown(ctx, 2)
				}))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: You and target opponent each secretly choose 1, 2, or 3. Then those choices are revealed. If they match, sacrifice this artifact and put all cards exiled with it into their owners' hands. Otherwise, exile the top card of your library face down.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetPlayer {
					return nil
				}
				epoch := 0
				if c, ok := g.LookupCardForEffect(item.SourceCardID); ok {
					epoch = c.ObjectEpoch
				}
				if ref, ok := ctx.SourceRef(); ok {
					epoch = ref.Epoch
				}
				return effects.Vote{
					Question:      "Expert-Level Safe — secretly choose 1, 2, or 3",
					Words:         []string{"1", "2", "3"},
					Voters:        []uuid.UUID{ctx.Controller(), t.ID},
					Choice:        true,
					Secret:        true,
					ForController: []int{1, 1, 1},
					ForOpponents:  []int{1, 1, 1},
					Then:          expertLevelSafeChosen,
					Carry:         []uuid.UUID{s82GapB60EncodeEpoch(epoch)},
				}.Apply(ctx)
			},
		}},
	})
}

// expertLevelSafeChosen is the rest of the ability, told both choices. An
// on-disk key: never renamed, never reused.
var expertLevelSafeChosen = effects.VoteResultThen("vote/expert-level-safe", func(ctx *effects.Context, r game.VoteResult) error {
	if len(r.Ballots) < 2 {
		return nil
	}
	g := ctx.Game
	if r.Ballots[0].Option != r.Ballots[1].Option {
		return s82GapB60ExileTopFaceDownFrom(g, r.Source, r.Controller, 1)
	}
	epoch := 0
	if len(r.Carry) > 0 {
		epoch = s82GapB60DecodeEpoch(r.Carry[0])
	}
	// The sacrifice first, as printed. The cards are found by the object they
	// are linked to, which the Safe leaving does not change (CR 607.2a).
	if c, ok := g.LookupCardForEffect(r.Source); ok && effects.OnBattlefield(g, r.Source) && c.ObjectEpoch == epoch {
		if err := g.SacrificePermanentForEffect(r.Source); err != nil {
			return err
		}
	}
	var ids []uuid.UUID
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef{ID: r.Source, Epoch: epoch}) {
		ids = append(ids, c.InstanceID)
	}
	if len(ids) == 0 {
		return nil
	}
	g.BounceCardsToHandForEffect(ids)
	return nil
})

// s82GapB60ExileTopFaceDown is "exile the top N cards of your library face
// down", linked to the resolving item's source.
func s82GapB60ExileTopFaceDown(ctx *effects.Context, n int) error {
	return s82GapB60ExileTopFaceDownFrom(ctx.Game, ctx.Source(), ctx.Controller(), n)
}

// s82GapB60ExileTopFaceDownFrom exiles the top n cards of `player`'s
// library one at a time, face down and linked to `source` (Asmodeus the
// Archfiend's exile). A library with fewer cards gives up what it has.
func s82GapB60ExileTopFaceDownFrom(g *game.Game, source, player uuid.UUID, n int) error {
	for i := 0; i < n; i++ {
		p := g.PlayerByIDForEffect(player)
		if p == nil || p.Library == nil || len(p.Library.Cards) == 0 {
			return nil
		}
		top := p.Library.Cards[len(p.Library.Cards)-1].InstanceID
		if _, err := g.ExileFaceDownLinkedForEffect(source, player, top); err != nil {
			return err
		}
	}
	return nil
}

// s82GapB60EncodeEpoch carries a permanent's object epoch through a vote's
// Carry, which holds IDs: the epoch in the first eight bytes.
func s82GapB60EncodeEpoch(epoch int) uuid.UUID {
	var id uuid.UUID
	binary.BigEndian.PutUint64(id[:8], uint64(epoch))
	return id
}

func s82GapB60DecodeEpoch(id uuid.UUID) int {
	return int(binary.BigEndian.Uint64(id[:8]))
}
