package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Breeches, the Blastmaker — Legendary Creature — Goblin Pirate
// {1}{U}{R}, 3/3:
//
//	"Menace
//	 Whenever you cast your second spell each turn, you may sacrifice
//	 an artifact. If you do, flip a coin. When you win the flip, copy
//	 that spell. You may choose new targets for the copy. When you
//	 lose the flip, Breeches deals damage equal to that spell's mana
//	 value to any target."
//
// Four printed sentences and four separate stack objects, which is
// what makes the card a card rather than a coin-flip slot machine:
//
//   - THE TRIGGER is mandatory and fires on the second spell only.
//     "Your second spell each turn" is the per-turn cast tally, which
//     the cast path bumps BEFORE it emits EventCast — so a count of
//     exactly two means "this is the second", the same clock Maelstrom
//     Nexus reads for "first".
//   - THE SACRIFICE is the resolution's own "you may", not the
//     trigger's, so the question is put when the trigger resolves and
//     the controller picks WHICH artifact. Asking it as a trigger-level
//     prompt would put it before the response window and offer no
//     choice of artifact. A controller with no artifact is asked
//     nothing at all rather than offered a "yes" that sacrifices
//     nothing and flips anyway — the coin is gated on the sacrifice.
//   - THE FLIP is the keyed, rewindable one (ADR 0054): the result is
//     drawn when the call is answered, and a changed call after an
//     undo changes the face without turning a loss into a win.
//   - THE TWO PAYOFFS are CR 603.12 reflexive triggers, not the tail
//     of the flip's continuation. Each goes on the stack above the
//     spell that triggered this, so the table gets a window to answer
//     the copy or the damage on its own, and the damage's "any
//     target" is chosen as that trigger goes on the stack (CR 603.3d)
//     rather than before anyone knew whether the flip was lost.
//
// "That spell" is still on the stack the whole way through: the
// trigger went on the stack above it, and every question in the chain
// is a pending choice, which stops priority from passing — so nothing
// under it can resolve while the controller is deciding. The copy
// therefore has something to copy, and the mana value has something
// to read.
//
// Mana value is read at the FLIP rather than at the damage trigger's
// resolution, which is the one place the spell is guaranteed to still
// be on the stack (CR 202.3e: {X} counts only there). A spell
// countered in response to the reflexive trigger is answered with
// last-known information, which is what CR 608.2h asks for, instead
// of silently costing the value of its X.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8514f20a-50c7-4319-84cb-2bf263548234",
		Name:            "Breeches, the Blastmaker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventCast},
			AppliesTo: YouCastYourSecondSpellEachTurn,
			Key:       "Breeches, the Blastmaker — sacrifice an artifact and flip a coin",
			Effect: func(g *game.Game, item *game.StackItem) error {
				if item.Trigger == nil {
					return nil
				}
				return breechesOffer(item.Trigger.Event.CardID)(g, item)
			},
		}},
	})
}

// breechesOffer is the trigger body: "you may sacrifice an artifact".
//
// A package-level function closing over the spell's ID — the
// StackItem.Effect contract, so an undo anywhere in the chain
// resolves it against the restored game.
func breechesOffer(spell uuid.UUID) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		if len(permanentsControlledByMatching(g, item.Controller, Artifact())) == 0 {
			// "You may sacrifice an artifact" with no artifact is not
			// a question, and without the sacrifice there is no flip.
			return nil
		}
		return MayChoice{
			Question: "Breeches, the Blastmaker — sacrifice an artifact? (if you do, flip a coin)",
			Then:     breechesAnswer,
			Carry:    []uuid.UUID{spell},
		}.Apply(ctx)
	}
}

// breechesSacrifice is the yes branch: they pick which artifact, and
// the flip follows once it has actually gone.
//
// The candidate list is recomputed here rather than captured, because
// the board can change between the question and the answer — and an
// artifact that left in that window means there is nothing to
// sacrifice and therefore nothing to flip for.
func breechesSacrifice(spell uuid.UUID) func(ctx *Context) error {
	return func(ctx *Context) error {
		artifacts := permanentsControlledByMatching(ctx.Game, ctx.Controller(), Artifact())
		if len(artifacts) == 0 {
			return nil
		}
		// #3275, CR 101.2: a permanent that can't be sacrificed is not a
		// legal choice, and with none left the flip follows at once.
		var candidates []uuid.UUID
		for _, id := range artifacts {
			if !ctx.Game.CantBeSacrificedForEffect(id) {
				candidates = append(candidates, id)
			}
		}
		if len(candidates) == 0 {
			return breechesFlip(spell)(ctx)
		}
		// The pick is a keyed choose_cards pick (#4253): the spell that
		// triggered the ability is the carry and the asking item rides the
		// prompt as data.
		ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:    ctx.Controller(),
			FromPlayer: ctx.Controller(),
			Source:     ctx.Source(),
			Question:   "Breeches, the Blastmaker — sacrifice an artifact",
			Cards:      candidates,
			Min:        1,
			Max:        1,
			LiveZone:   game.ZoneBattlefield,
			ThenKey:    breechesSacrificed,
			Carry:      []uuid.UUID{spell},
			Asked:      game.NewAskedItem(ctx.Item),
		})
		return nil
	}
}

// breechesSacrificed is the artifact chosen, keyed (#4253): it goes through
// the sacrifice's own continuation, so the coin is flipped once the
// permanent has gone. The spell that triggered the ability is the carry,
// and rides the sacrifice's keyed tail with the trigger as its asked item.
var breechesSacrificed = askedPick("choose-cards/breeches-the-blastmaker-sacrifice", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Carry) == 0 {
		return nil
	}
	t := ctx.ItemTail(breechesSacrificeDone)
	t.Carry = []uuid.UUID{r.Carry[0]}
	return ctx.Game.SacrificeAllKeyedForEffect(uuid.Nil, r.Picked, t)
})

// breechesSacrificeDone is the sacrifice settled: flip the coin. An on-disk
// key: never renamed, never reused.
var breechesSacrificeDone = CardTail("then/breeches-the-blastmaker-flip", func(ctx *Context, o game.TailOutcome) error {
	return breechesFlip(carriedID(o.Tail.Carry, 0))(ctx)
})

// breechesFlip is "if you do, flip a coin". The call is keyed (#4253 slice
// 10b): the spell rides it as its carry and the trigger as its asked item.
func breechesFlip(spell uuid.UUID) func(ctx *Context) error {
	return func(ctx *Context) error {
		ctx.Game.FlipCoinForEffect(game.CoinFlipSpec{
			Flipper:  ctx.Controller(),
			Source:   ctx.Source(),
			Question: "Breeches, the Blastmaker — call the coin flip",
			ThenKey:  breechesFlipped,
			Carry:    []uuid.UUID{spell},
			Asked:    game.NewAskedItem(ctx.Item),
		})
		return nil
	}
}

// breechesFlipped is the settled call. An on-disk key: never renamed, never
// reused.
var breechesFlipped = CoinFlipThen("coin-flip/breeches-the-blastmaker", func(ctx *Context, r game.CoinFlipped) error {
	return breechesFlipResult(ctx.Game, ctx.Item, carriedID(r.Carry, 0), r.Result)
})

// breechesFlipResult creates whichever of the two reflexive triggers
// the flip called for.
//
// The Context is rebuilt from the game the answer arrived in, the way
// every asynchronous clause in the catalog ends; the stack item it is
// bound to is the trigger that created the flip, which is what CR
// 603.12 makes the reflexive trigger's source and controller.
func breechesFlipResult(g *game.Game, item *game.StackItem, spell uuid.UUID, result game.CoinFlipResult) error {
	if len(result.Won) == 0 {
		return nil
	}
	ctx := NewContext(g, item)
	if result.Won[0] {
		return ReflexiveTrigger{
			Label: "Breeches, the Blastmaker — copy that spell",
			Cards: []uuid.UUID{spell},
			Body:  breechesCopyThatSpellBody,
		}.Apply(ctx)
	}
	return ReflexiveTrigger{
		Label:  "Breeches, the Blastmaker — damage equal to that spell's mana value",
		Body:   breechesBlastBody,
		Params: game.EffectParams{Amount: spellManaValueForEffect(g, spell)},
	}.Apply(ctx)
}

// breechesCopyThatSpell is "when you win the flip, copy that spell.
// You may choose new targets for the copy." The spell rides the
// trigger's payload rather than a captured variable, which is the
// slot reflexive triggers carry for exactly this.
//
// "That spell" NAMES the spell; the reflexive trigger does not target
// it. So a spell countered before this trigger resolves is copied from
// last-known information (CR 608.2h, #1288) — storm's and Doublecast's
// reading (ADR 0043 decision 17), not Reverberate's CR 608.2b one.
func breechesCopyThatSpell(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	cards := ctx.PayloadCards()
	if len(cards) == 0 {
		return nil
	}
	return CopySpell{
		StackID:          cards[0],
		Controller:       item.Controller,
		ChooseNewTargets: true,
		FromLastKnown:    true,
	}.Apply(ctx)
}

// breechesBlastEffect (reflexive_bodies.go) is "when you lose the
// flip, Breeches deals damage equal to that spell's mana value to any
// target": the amount rides Params.Amount, fixed when the flip was
// lost; the target is the one chosen as this trigger went on the
// stack.

// breechesAnswer is the "yes": sacrifice an artifact, then flip a coin for
// the spell that triggered it (the first carried ID). Keyed (#4291).
var breechesAnswer = MayYes("confirm/breeches-sacrifice", func(ctx *Context, r game.ConfirmAnswer) error {
	return breechesSacrifice(carryAt(r, 0))(ctx)
})
