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
//     card in a hand has no controller to ask. They are asked from the
//     return's CONTINUATION (#993): a hand is a CR 903.9 destination,
//     so a bounced commander's owner is already answering a question,
//     and "then that permanent's controller may sacrifice a land"
//     cannot be put to them on top of it (ADR 0013 §5m item 2). The
//     answer to the return is ignored — the second sentence is about a
//     PLAYER and is not introduced by an "if you do", so it happens
//     whichever way the command-zone question goes.
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
// is still on the stack at that moment (CR 608.2n puts it in the
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
	return BounceToHand{
		Target:   t.ID,
		ThenTail: withCarry(ctx.ItemTail(chainOfVaporBounced), controller),
	}.Apply(ctx)
}

// chainOfVaporBounced opens the chain once the return has settled, keyed
// (#4253): the spell rides the tail as data and the bounced permanent's
// controller its Carry. An on-disk key: never renamed, never reused.
var chainOfVaporBounced = CardTail("then/chain-of-vapor-offer-chain", func(ctx *Context, o game.TailOutcome) error {
	return chainSacrificeALandToCopy(ctx, tailCarry0(o), "Chain of Vapor")
})

// chainSacrificeALandToCopy is the chain's first question, put to the
// bounced permanent's controller once the return has settled: "that
// permanent's controller may sacrifice a land of their choice. If the
// player does, they may copy this spell and may choose a new target for
// that copy." Chain of Silence (ADR 0108 PR 7b, #1904) prints the same
// sentences about the shielded creature's controller; `name` is the card
// asking.
func chainSacrificeALandToCopy(ctx *Context, controller uuid.UUID, name string) error {
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
		Question: name + " — sacrifice a land? (if you do, you may copy this spell)",
		Then:     chainSacrificeLandAnswer,
		Word:     name,
	}.Apply(ctx)
}

// chainSacrificeLand is the yes branch: they pick the land, and the
// copy question follows once it has actually gone. The pick is a keyed
// choose_cards prompt (#4253), as SacrificeChoice's was: the asking spell
// rides it as data and the card's name as its word, so a table waiting on
// the pick is a restore point. The candidate list is read when the prompt
// is queued. The answer is not read: the prompt is mandatory, so the copy
// clause is "then", not "if you do".
func chainSacrificeLand(ctx *Context, player uuid.UUID, name string) error {
	var lands []uuid.UUID
	for _, id := range landsControlledByPlayer(ctx.Game, player) {
		// #3275, CR 101.2: a permanent that can't be sacrificed is not a
		// legal choice.
		if !ctx.Game.CantBeSacrificedForEffect(id) {
			lands = append(lands, id)
		}
	}
	if len(lands) == 0 {
		return chainOfferCopy(ctx.Game, ctx.Item, player, name)
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    player,
		FromPlayer: player,
		Source:     ctx.Source(),
		Question:   name + " — sacrifice a land",
		Cards:      lands,
		Min:        1,
		Max:        1,
		ThenKey:    chainLandPicked,
		LiveZone:   game.ZoneBattlefield,
		CarryWord:  name,
		Carry:      []uuid.UUID{player},
		Asked:      game.NewAskedItem(ctx.Item),
	})
	return nil
}

// chainLandPicked sacrifices the pick and, once it has gone, offers the
// copy. An on-disk key: never renamed, never reused.
var chainLandPicked = game.RegisterChooseCardsThen("choose-cards/chain-sacrifice-a-land", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Carry) == 0 || r.Asked == nil {
		return nil
	}
	// The offer rides the sacrifice as a keyed tail (#4253): the player is
	// its carry, the spell's name its word and the spell its asked item.
	return g.SacrificeAllKeyedForEffect(uuid.Nil, r.Picked, game.KeyedTail{
		Ref: chainSacrificed, Player: r.Chooser, Source: r.Source,
		Carry: []uuid.UUID{r.Carry[0]}, Word: r.Word, Asked: game.NewAskedItem(r.Asked),
	})
})

// chainSacrificed is the land gone: offer the copy. An on-disk key: never
// renamed, never reused.
var chainSacrificed = game.RegisterTail("then/chain-sacrificed-offer-copy", func(g *game.Game, o game.TailOutcome) error {
	item := o.Tail.Asked.StackItem()
	if item == nil || len(o.Tail.Carry) == 0 {
		return nil
	}
	return chainOfferCopy(g, item, o.Tail.Carry[0], o.Tail.Word)
})

// chainOfferCopy is "they may copy this spell and may choose a new target
// for that copy" (CR 707.10, 707.10b).
func chainOfferCopy(g *game.Game, item *game.StackItem, player uuid.UUID, name string) error {
	return MayChoice{
		Player:   player,
		Question: name + " — copy it? (you may choose a new target for the copy)",
		Then:     mayCopyThisSpell,
	}.Apply(NewContext(g, item))
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

// chainSacrificeLandAnswer is the "yes": they pick the land (Word is the
// card asking), then may copy the spell. Keyed (#4291).
var chainSacrificeLandAnswer = MayYes("confirm/chain-sacrifice-a-land", func(ctx *Context, r game.ConfirmAnswer) error {
	return chainSacrificeLand(ctx, r.Chooser, r.Word)
})
