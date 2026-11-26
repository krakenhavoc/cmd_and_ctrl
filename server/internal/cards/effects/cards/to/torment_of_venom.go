package cardsto

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Torment of Venom — Instant {2}{B}{B} (Hour of Devastation, #5534):
//
//	"Put three -1/-1 counters on target creature. Its controller loses 3 life
//	 unless they sacrifice another nonland permanent of their choice or
//	 discard a card."
//
// The counters go on first. Then the creature's controller, as it was (CR
// 608.2h), picks one of three: lose 3 life, sacrifice a nonland permanent
// other than that creature, or discard a card. The two payments are offered
// only to a player who can make them (CR 118.3), and losing the life is
// always offered. Choosing a payment pays it; the life is lost only if they
// choose to. A target that has gone makes the spell do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9e9a7f0b-ffba-43d6-892a-95fd0b8b39f8",
		Name:         "Torment of Venom",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, 0, 0)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			victim := c.Controller
			if err := (effects.AddCounter{Target: id, Kind: game.CounterMinusOne, N: 3}).Apply(ctx); err != nil {
				return err
			}
			if p := ctx.Game.PlayerByIDForEffect(victim); p == nil || p.Eliminated {
				return nil
			}
			options := []game.ChoiceOption{{Label: tormentLoseLife}}
			if len(tormentSacrificeCandidates(ctx.Game, victim, id)) > 0 {
				options = append(options, game.ChoiceOption{Label: tormentSacrifice})
			}
			if p := ctx.Game.PlayerByIDForEffect(victim); p != nil && p.Hand.Size() > 0 {
				options = append(options, game.ChoiceOption{Label: tormentDiscard})
			}
			return effects.PickOption{
				Player:   victim,
				Question: "Torment of Venom — lose 3 life, sacrifice another nonland permanent, or discard a card?",
				Options:  options,
				ThenKey:  tormentChosen,
				Carry:    []uuid.UUID{victim, id},
				WithItem: true,
			}.Apply(ctx)
		},
	})
}

const (
	tormentLoseLife  = "Lose 3 life"
	tormentSacrifice = "Sacrifice another nonland permanent"
	tormentDiscard   = "Discard a card"
)

// tormentSacrificeCandidates is the nonland permanents `victim` controls other
// than the creature that took the counters.
func tormentSacrificeCandidates(g *game.Game, victim, except uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, id := range effects.NonlandPermanentsControlledBy(g, victim) {
		if id != except {
			out = append(out, id)
		}
	}
	return out
}

var tormentChosen = effects.OptionPickThen("option-pick/torment-of-venom-pay-or-lose", func(ctx *effects.Context, r game.OptionPicked) error {
	if len(r.Carry) < 2 {
		return nil
	}
	victim, creature := r.Carry[0], r.Carry[1]
	label := tormentLoseLife
	if r.Option != nil {
		label = r.Option.Label
	}
	switch label {
	case tormentSacrifice:
		return effects.SacrificeChoice{
			Player:     victim,
			Candidates: tormentSacrificeCandidates(ctx.Game, victim, creature),
			Question:   "Torment of Venom — sacrifice another nonland permanent",
		}.Apply(ctx)
	case tormentDiscard:
		ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
			Player:   victim,
			Source:   ctx.Source(),
			N:        1,
			Question: "Torment of Venom — discard a card",
		})
		return nil
	}
	return effects.GainLife{Player: victim, Amount: -3}.Apply(ctx)
})
