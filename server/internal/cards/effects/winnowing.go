package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Winnowing — Sorcery {4}{W}{W} (Lorwyn Eclipsed, #4458):
//
//	"Convoke (Your creatures can help cast this spell. Each creature you
//	 tap while casting this spell pays for {1} or one mana of that
//	 creature's color.)
//	 For each player, you choose a creature that player controls. Then
//	 each player sacrifices all other creatures they control that don't
//	 share a creature type with the chosen creature they control."
//
// Convoke is the shared TapCost (CR 702.51a). The choice is Tragic
// Arrogance's seam (ChoosePermanents over every player, APNAP from the
// active player, every leg answered by the caster) with Humiliate's offer:
// exactly one creature of each player that controls any, and a player
// with none is skipped. Nothing is targeted, so hexproof and shroud don't
// stop the choice.
//
// Then, once every leg is answered, each player sacrifices at the same
// time every creature they control other than their chosen one that
// shares no creature type with it (game.SharesCreatureType on the live,
// layered characteristics): a changeling has every creature type
// (CR 702.73a), so it shares a type with any creature that has one, and
// a creature with no creature type shares none, so it goes unless it was
// chosen. Lands and other noncreature permanents are untouched.
//
// ADR 0126: it reads as a board wipe, declared as a partial sacrifice
// sweep of creatures (what shares a type with a chosen creature stays).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8b834f83-951b-4b5c-b9f1-3d860baa3c0e",
		Name:         "Winnowing",
		Completeness: CompletenessFull,
		TapCost:      Convoke(),
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepSacrifice, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ChoosePermanents{
				Question:   "Winnowing — choose a creature that player controls; their creatures that share no creature type with it are sacrificed",
				Of:         seatsFromActive(ctx.Game),
				Candidates: humiliateCreatures,
				Keyed:      winnowingChosen,
			}.Apply(ctx)
		},
	})
}

var winnowingChosen = RegisterPermanentsPick("prompt-run/winnowing-sacrifice-unshared",
	PermanentsDef{Then: winnowingSacrificeUnshared})

// winnowingSacrificeUnshared is the second sentence, as one simultaneous
// sacrifice across every player.
//
// Caller holds g.mu.
func winnowingSacrificeUnshared(ctx *Context, picked game.PromptedPicks) error {
	g := ctx.Game
	chosen := map[uuid.UUID]*game.Card{}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if ids := picked.By(c.Controller); len(ids) == 1 && ids[0] == c.InstanceID {
			chosen[c.Controller] = c
		}
	}
	var doomed []uuid.UUID
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if !c.IsCreature() || picked.Contains(c.InstanceID) {
			continue
		}
		keep, ok := chosen[c.Controller]
		if !ok && !picked.Picked(c.Controller) {
			// Nothing was chosen for this player: they controlled no
			// creature when the choice was made.
			continue
		}
		if ok && game.SharesCreatureType(c, keep) {
			continue
		}
		doomed = append(doomed, c.InstanceID)
	}
	if len(doomed) == 0 {
		return nil
	}
	return g.SacrificeAllKeyedForEffect(ctx.Source(), doomed, game.KeyedTail{})
}
