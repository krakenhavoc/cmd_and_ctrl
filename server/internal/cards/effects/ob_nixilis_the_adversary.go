package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ob Nixilis, the Adversary — Legendary Planeswalker — Nixilis {1}{B}{R},
// loyalty 3:
//
//	"Casualty X. The copy isn't legendary and has starting loyalty X.
//	 +1: Each opponent loses 2 life unless they discard a card. If you
//	 control a Demon or Devil, you gain 2 life.
//	 −2: Create a 1/1 red Devil creature token with "When this token
//	 dies, it deals 1 damage to any target."
//	 −7: Target player draws seven cards and loses 7 life."
//
// Casualty X is CasualtyX (casualty.go, #5668): the caster may sacrifice
// a creature as the spell is cast, X is that creature's power, and the
// engine's casualty trigger copies the spell with both printed
// exceptions. The copy resolves as a token (CR 608.3f) that is not
// legendary and enters with X loyalty counters; with X 0 it enters with
// none and CR 704.5i puts it into the graveyard.
//
// The +1's "unless" is each opponent's own choice, asked in APNAP order
// (CR 101.4), one question at a time so a later opponent sees what the
// earlier ones did (Court of Ambition's chain). An opponent with no card
// in hand cannot discard, so they are not asked and lose the life (CR
// 118.3's reasoning). "If you control a Demon or Devil" is read after
// the last opponent, as the last instruction (CR 608.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a6abb6af-7ea2-43f3-8718-2a077ad0c947",
		Name:            "Ob Nixilis, the Adversary",
		Completeness:    CompletenessFull,
		StartingLoyalty: 3,
		OptionalCosts:   []game.AdditionalCost{CasualtyX()},
		Activated: []ActivatedAbility{
			{
				Label: "+1: Each opponent loses 2 life unless they discard a card. If you control a Demon or Devil, you gain 2 life.",
				Cost:  LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					var victims []uuid.UUID
					for _, id := range apnapPlayers(g) {
						if id != item.Controller {
							victims = append(victims, id)
						}
					}
					return obNixilisAdversaryStep(NewContext(g, item), victims)
				},
			},
			{
				Label:   "−2: Create a 1/1 red Devil creature token with \"When this token dies, it deals 1 damage to any target.\"",
				Cost:    LoyaltyCost(-2),
				Purpose: game.Purpose{Tokens: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateToken{Template: devilToken(), N: 1}.Apply(NewContext(g, item))
				},
			},
			{
				Label:   "−7: Target player draws seven cards and loses 7 life.",
				Cost:    LoyaltyCost(-7),
				Targets: TargetPlayer("target player"),
				Purpose: ForTargets(game.TargetPurpose{Slot: 0, Draws: 7, LifeLoss: 7}),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return b32TargetPlayerDrawsAndLosesLife(item, NewContext(g, item), 7, 7)
				},
			},
		},
	})
	obNixilisAdversaryLoss = RegisterThenSteps("then-steps/ob-nixilis-the-adversary-loss", obNixilisAdversaryLossSteps()...)
	obNixilisAdversaryDiscarded = game.RegisterDiscardThen("choose-cards/ob-nixilis-the-adversary-discarded",
		func(g *game.Game, r game.DiscardsDone) error {
			if len(r.Carry) == 0 {
				return nil
			}
			return obNixilisAdversaryStep(keyedContext(g, r.Carry[0], r.Source), r.Carry[1:])
		})
	obNixilisAdversaryAnswered = OptionPickThen("option-pick/ob-nixilis-the-adversary-answer",
		func(ctx *Context, r game.OptionPicked) error {
			if len(r.Carry) == 0 {
				return nil
			}
			victim, rest := r.Carry[0], r.Carry[1:]
			switch r.Index {
			case 1:
				return ctx.Game.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
					Player:   victim,
					Source:   ctx.Source(),
					N:        1,
					Question: "Ob Nixilis, the Adversary — discard a card",
				}, obNixilisAdversaryDiscarded, append([]uuid.UUID{ctx.Controller()}, rest...))
			case 0:
				return obNixilisAdversaryLoseThenNext(ctx, victim, rest)
			default:
				// index -1: the player left before the question could be
				// asked (CR 800.4a); the rest of the run still happens.
				return obNixilisAdversaryStep(ctx, rest)
			}
		})
}

// The +1's keyed continuations, registered in init because each answer
// asks the next opponent and so names itself. The keys are on-disk
// identities: never renamed, never reused.
var (
	// obNixilisAdversaryAnswered is an opponent's choice between losing 2
	// life and discarding; its Carry is the opponent, then the opponents
	// still owed a question.
	obNixilisAdversaryAnswered game.OptionPickThen
	// obNixilisAdversaryDiscarded is the discard behind the choice; its
	// Carry is the effect's controller, then the opponents still owed a
	// question.
	obNixilisAdversaryDiscarded game.DiscardThen
	// obNixilisAdversaryLoss is one opponent's life loss, then the next
	// opponent's question.
	obNixilisAdversaryLoss ThenSteps
)

const obNixilisAdversaryRecord = "ob-nixilis-the-adversary/loss"

// obNixilisAdversaryStep asks the next opponent still owed a question,
// and after the last one reads "If you control a Demon or Devil, you
// gain 2 life." Package-level over frozen scalars, so an undo that
// replays an answer resolves it against the restored game.
//
// Caller holds g.mu.
func obNixilisAdversaryStep(ctx *Context, remaining []uuid.UUID) error {
	for len(remaining) > 0 {
		victim, rest := remaining[0], remaining[1:]
		p := ctx.Game.PlayerByIDForEffect(victim)
		if p == nil || p.Eliminated {
			remaining = rest
			continue
		}
		if p.Hand == nil || p.Hand.Size() < 1 {
			return obNixilisAdversaryLoseThenNext(ctx, victim, rest)
		}
		return PickOption{
			Player:   victim,
			Question: "Ob Nixilis, the Adversary — lose 2 life, or discard a card?",
			Options: []game.ChoiceOption{
				{Label: "Lose 2 life", LifeCost: 2},
				{Label: "Discard a card"},
			},
			ThenKey: obNixilisAdversaryAnswered,
			Carry:   append([]uuid.UUID{victim}, rest...),
		}.Apply(ctx)
	}
	you := ctx.Controller()
	if ControlsA(ctx.TextSubtype("Demon"))(ctx.Game, you) || ControlsA(ctx.TextSubtype("Devil"))(ctx.Game, you) {
		return GainLife{Amount: 2}.Apply(ctx)
	}
	return nil
}

// obNixilisAdversaryLoseThenNext is one opponent's life loss, then the
// next opponent's question, as continuation steps (CR 608.2c): the
// question waits for a prompt the loss asks (a CR 616 order of two
// life-loss replacements). The record carries the opponent and the
// opponents still owed a question.
func obNixilisAdversaryLoseThenNext(ctx *Context, victim uuid.UUID, rest []uuid.UUID) error {
	ctx.RecordStep(obNixilisAdversaryRecord, append([]uuid.UUID{victim}, rest...))
	return obNixilisAdversaryLoss.Run(ctx)
}

func obNixilisAdversaryLossSteps() []Applier {
	return []Applier{
		Instruction(func(ctx *Context) error {
			r := ctx.StepRecord(obNixilisAdversaryRecord)
			if len(r.IDs) == 0 {
				return nil
			}
			return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), r.IDs[0], -2)
		}),
		Instruction(func(ctx *Context) error {
			r := ctx.StepRecord(obNixilisAdversaryRecord)
			if len(r.IDs) == 0 {
				return nil
			}
			return obNixilisAdversaryStep(ctx, r.IDs[1:])
		}),
	}
}
