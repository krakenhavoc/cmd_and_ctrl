package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Remorseless Punishment — Sorcery {3}{B}{B} (Oath of the Gatewatch, S82 OLD
// #5545):
//
//	"Target opponent loses 5 life unless that player discards two cards or
//	 sacrifices a creature or planeswalker of their choice. Repeat this
//	 process once."
//
// A three-way punisher asked of the one targeted opponent twice, each
// question queued by the answer to the one before it (the chain Torment of
// Hailfire's punisher uses, punisher_repeat.go). An opponent with fewer than
// two cards in hand isn't offered the discard, and one with no creature or
// planeswalker isn't offered the sacrifice; "lose 5 life" is always there.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "42e07167-b490-46ad-9b5c-b483a4bb64e3",
		Name:         "Remorseless Punishment",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return remorselessPunishmentOnTarget(ctx)
		},
	})
}

const (
	remorselessPunishmentName         = "Remorseless Punishment"
	remorselessPunishmentLife         = 5
	remorselessPunishmentSacLabel     = "Sacrifice a creature or planeswalker"
	remorselessPunishmentDiscardLabel = "Discard two cards"
)

var (
	remorselessPunishmentAnswered  game.OptionPickThen
	remorselessPunishmentDiscarded game.DiscardThen
	remorselessPunishmentSacrifice game.TailRef
)

// The keyed continuations are registered in init, not at var initialisation:
// the step and the answers refer to each other. On-disk keys: never renamed,
// never reused.
func init() {
	remorselessPunishmentSacrifice = effects.CardTail("then/remorseless-punishment-sacrificed", func(ctx *effects.Context, o game.TailOutcome) error {
		return remorselessPunishmentStep(ctx, o.Tail.Carry)
	})
	remorselessPunishmentDiscarded = game.RegisterDiscardThen("choose-cards/remorseless-punishment-discarded", func(g *game.Game, r game.DiscardsDone) error {
		return remorselessPunishmentStep(effects.KeyedContext(g, r.Carry[0], r.Source), r.Carry[1:])
	})
	remorselessPunishmentAnswered = effects.OptionPickThen("option-pick/remorseless-punishment", remorselessPunishmentAnswer)
}

// remorselessPunishmentCandidates is the creatures and planeswalkers the
// victim could sacrifice.
func remorselessPunishmentCandidates(g *game.Game, victim uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == victim && (c.IsCreature() || c.IsPlaneswalker()) && !g.CantBeSacrificedForEffect(c.InstanceID) {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// remorselessPunishmentOnTarget asks the targeted opponent twice. A player no
// longer a legal target by resolution is asked nothing (CR 608.2b).
//
// Caller holds g.mu.
func remorselessPunishmentOnTarget(ctx *effects.Context) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			return remorselessPunishmentStep(ctx, []uuid.UUID{t.ID, t.ID})
		}
	}
	return nil
}

// remorselessPunishmentStep asks the first question still owed and hands the
// rest to its answer. "Lose 5 life" is first, the option pick's always-legal
// answer; the other two are offered only if the player can do them.
//
// Caller holds g.mu.
func remorselessPunishmentStep(ctx *effects.Context, remaining []uuid.UUID) error {
	for len(remaining) > 0 {
		victim, rest := remaining[0], remaining[1:]
		pl := ctx.Game.PlayerByIDForEffect(victim)
		if pl == nil || pl.Eliminated {
			remaining = rest
			continue
		}
		options := []game.ChoiceOption{{Label: "Lose 5 life", LifeCost: remorselessPunishmentLife}}
		if len(remorselessPunishmentCandidates(ctx.Game, victim)) > 0 {
			options = append(options, game.ChoiceOption{Label: remorselessPunishmentSacLabel})
		}
		if pl.Hand != nil && pl.Hand.Size() >= 2 {
			options = append(options, game.ChoiceOption{Label: remorselessPunishmentDiscardLabel})
		}
		return effects.PickOption{
			Player:   victim,
			Question: remorselessPunishmentName + " — lose 5 life, or...",
			Options:  options,
			ThenKey:  remorselessPunishmentAnswered,
			Carry:    rest,
		}.Apply(ctx)
	}
	return nil
}

func remorselessPunishmentAnswer(ctx *effects.Context, r game.OptionPicked) error {
	victim, rest := r.Chooser, r.Carry
	if r.Option == nil {
		return remorselessPunishmentStep(ctx, rest)
	}
	switch r.Option.Label {
	case remorselessPunishmentSacLabel:
		return effects.SacrificeChoice{
			Player:     victim,
			Candidates: remorselessPunishmentCandidates(ctx.Game, victim),
			Question:   remorselessPunishmentName + " — sacrifice a creature or planeswalker",
			ThenTail:   effects.WithCarry(ctx.Tail(remorselessPunishmentSacrifice), rest...),
		}.Apply(ctx)
	case remorselessPunishmentDiscardLabel:
		return ctx.Game.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
			Player:   victim,
			Source:   ctx.Source(),
			N:        2,
			Question: remorselessPunishmentName + " — discard two cards",
		}, remorselessPunishmentDiscarded, append([]uuid.UUID{ctx.Controller()}, rest...))
	default:
		if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), victim, -remorselessPunishmentLife); err != nil {
			return err
		}
		return remorselessPunishmentStep(ctx, rest)
	}
}
