package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eye of Doom — Artifact {4} (Commander 2013, S82 older sets):
//
//	"When this artifact enters, each player chooses a nonland permanent and
//	 puts a doom counter on it.
//	 {2}, {T}, Sacrifice this artifact: Destroy each permanent with a doom
//	 counter on it."
//
// Nothing targets. Each player in turn order (APNAP, CR 101.4) is asked for
// one nonland permanent, any player's, and every choice is made before the
// counters go on, so the players choose from the same board. Two players
// may choose the same permanent; it gets a counter for each. The destroy
// is a simultaneous sweep of every permanent holding a doom counter.
//
// No simplification.
var eyeOfDoomCounters = effects.RunThen("prompt-run/eye-of-doom-counters", func(ctx *effects.Context, r game.RunDone) error {
	for _, id := range r.Picks().Cards() {
		if err := (effects.AddCounter{Target: id, Kind: "doom", N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})

func eyeOfDoomEnters(g *game.Game, item *game.StackItem) error {
	var cands []uuid.UUID
	for _, seat := range effects.ApnapPlayers(g) {
		cands = append(cands, effects.PermanentsControlledByMatching(g, seat, effects.Nonland())...)
	}
	if len(cands) == 0 {
		return nil
	}
	var legs []game.ChooseCardsPrompt
	for _, seat := range effects.ApnapPlayers(g) {
		legs = append(legs, game.ChooseCardsPrompt{
			Chooser:    seat,
			FromPlayer: seat,
			Source:     item.SourceCardID,
			Question:   "Eye of Doom — choose a nonland permanent; it gets a doom counter",
			Cards:      cands,
			Min:        1,
			Max:        1,
			Zone:       game.ZoneBattlefield,
			ThenKey:    game.ChooseCardsRunLeg,
		})
	}
	_, err := g.ChooseCardsRunThenKeyedForEffect(legs, eyeOfDoomCounters, item.SourceCardID, game.RunData{})
	return err
}

func init() {
	effects.Register(effects.Spec{
		OracleID:     "e808a11e-29bd-4e99-a24e-67fa8f6fe502",
		Name:         "Eye of Doom",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Eye of Doom — each player chooses a nonland permanent and puts a doom counter on it", eyeOfDoomEnters),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice this artifact: Destroy each permanent with a doom counter on it.",
			Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepAllPermanents, How: game.SweepDestroy, Partial: true}},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeThis()),
			Effect:  effects.Do(effects.DestroyAllMatching{Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.Counters["doom"] > 0 }}),
		}},
	})
}
