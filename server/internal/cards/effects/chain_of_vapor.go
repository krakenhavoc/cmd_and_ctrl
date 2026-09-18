package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chain of Vapor — Instant {U}:
//
//	"Return target nonland permanent to its owner's hand. Then that
//	 permanent's controller may sacrifice a land of their choice. If
//	 the player does, they may copy this spell and may choose a new
//	 target for that copy."
//
// The card #920 was written for. Its last clause is a spell copying
// ITSELF, which CR 707.10 allows and which this engine could not do:
// resolveTopOfStackLocked removed the item's StackMeta entry before
// calling OnResolve, and CopySpellForEffect looked the source up
// there. The resolving-item slot (`game/resolving_item.go`) is where
// it looks now.
//
// Three things about the chain, all of them printed:
//
//   - The questions go to the BOUNCED permanent's controller, not to
//     the caster — read off the permanent before it moves, since a
//     card in a hand has no controller to ask.
//   - The copy is theirs (CR 707.10b): the effect says THEY may copy
//     it, so the copy is created under their control and they choose
//     its new target. Chaining a Chain of Vapor around the table is
//     the whole card, and it only works because the copy changes
//     hands.
//   - The sacrifice gates the copy. A player who controls no land is
//     asked nothing at all rather than being offered a "yes" that
//     sacrifices nothing and copies anyway.
//
// The copy is made while the answer arrives, which is after this
// spell has already been routed to its owner's graveyard — so the
// copy is built from last-known information. In the rules the spell
// is still on the stack at that moment (CR 608.2m puts it in the
// graveyard as the final step of its own resolution); see
// resolving_item.go on why the difference is spelled as LKI here.
func init() {
	Register(Spec{
		OracleID:     "2167ac25-d042-4b48-b770-8b94acc1a965",
		Name:         "Chain of Vapor",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target nonland permanent", Nonland()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return chainOfVaporBounce(ctx)
		},
	})
}

// chainOfVaporBounce returns the permanent and opens the chain.
func chainOfVaporBounce(ctx *Context) error {
	t, ok := ctx.ClauseTarget(0)
	if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
		return nil
	}
	// Read the controller BEFORE the bounce: once the permanent is in
	// a hand there is no controller left to ask.
	var controller uuid.UUID
	if c, found := ctx.Game.LookupCardForEffect(t.ID); found {
		controller = c.Controller
	}
	if err := (BounceToHand{Target: t.ID}).Apply(ctx); err != nil {
		return err
	}
	if controller == uuid.Nil {
		return nil
	}
	lands := landsControlledByPlayer(ctx.Game, controller)
	if len(lands) == 0 {
		// "May sacrifice a land" with no land is not a question.
		return nil
	}
	return MayChoice{
		Player:   controller,
		Question: "Chain of Vapor — sacrifice a land? (if you do, you may copy this spell)",
		OnYes:    chainOfVaporSacrifice(controller),
	}.Apply(ctx)
}

// chainOfVaporSacrifice is the yes branch: they pick the land, and the
// copy question follows once it has actually gone.
//
// A package-level function closing over one scalar — the
// StackItem.Effect contract, so an undo across either prompt resolves
// it against the restored game. The candidate list is recomputed at
// this point rather than captured, because the board can change
// between the question and the answer.
func chainOfVaporSacrifice(player uuid.UUID) func(ctx *Context) error {
	return func(ctx *Context) error {
		return SacrificeChoice{
			Player:     player,
			Candidates: landsControlledByPlayer(ctx.Game, player),
			Question:   "Chain of Vapor — sacrifice a land",
			Then:       chainOfVaporMayCopy(player),
		}.Apply(ctx)
	}
}

// chainOfVaporMayCopy is "they may copy this spell and may choose a
// new target for that copy" (CR 707.10, 707.10b).
func chainOfVaporMayCopy(player uuid.UUID) func(ctx *Context) error {
	return func(ctx *Context) error {
		return MayChoice{
			Player:   player,
			Question: "Chain of Vapor — copy it? (you may choose a new target for the copy)",
			OnYes:    copyThisSpellFor(player),
		}.Apply(ctx)
	}
}

// copyThisSpellFor copies the spell currently resolving under
// `player`'s control, with the CR 707.10c re-target prompt.
//
// Shared by Chain of Vapor and Chain of Smog: "they may copy this
// spell and may choose a new target for that copy" is the cycle's
// common sentence, and ctx.Item.ID is what "this spell" means.
func copyThisSpellFor(player uuid.UUID) func(ctx *Context) error {
	return func(ctx *Context) error {
		if ctx.Item == nil {
			return nil
		}
		return CopySpell{
			StackID:          ctx.Item.ID,
			Controller:       player,
			ChooseNewTargets: true,
		}.Apply(ctx)
	}
}

// landsControlledByPlayer lists the lands a player controls, in
// battlefield order — the candidate set behind "sacrifice a land of
// their choice".
//
// Caller must hold g.mu — it is an effect-time read.
func landsControlledByPlayer(g *game.Game, playerID uuid.UUID) []uuid.UUID {
	if playerID == uuid.Nil {
		return nil
	}
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != playerID || !c.IsLand() {
			continue
		}
		out = append(out, c.InstanceID)
	}
	return out
}
