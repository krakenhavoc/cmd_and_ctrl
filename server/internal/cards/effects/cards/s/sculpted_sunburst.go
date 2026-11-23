package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sculpted Sunburst — Sorcery {3}{W}{W} (Commander Legends: Battle for
// Baldur's Gate, #5948):
//
//	"Choose a creature you control, then each opponent chooses a creature
//	 they control with equal or lesser power. If you chose a creature this
//	 way, exile each creature not chosen by any player this way."
//
// Two prompts in turn. I am asked for one of my creatures; then, reading that
// creature's power as the opponents choose (through the layers), each
// opponent (APNAP) is asked for one of theirs with equal or lesser power, an
// opponent with none being asked nothing. When the last answer is in, every
// creature not chosen is exiled at once. If I controlled no creature to
// choose, nothing was chosen and nothing is exiled. Nothing targets.
//
// ADR 0126: it reads as a board wipe, declared as a partial exile sweep of
// creatures.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aee0388f-544e-4c28-a34f-29cc0a97d424",
		Name:         "Sculpted Sunburst",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepExile, Partial: true}},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			g := ctx.Game
			mine := effects.PermanentsControlledByMatching(g, item.Controller, effects.Creature())
			if len(mine) == 0 {
				return nil
			}
			leg := game.ChooseCardsPrompt{
				Chooser:    item.Controller,
				FromPlayer: item.Controller,
				Source:     item.SourceCardID,
				Question:   "Sculpted Sunburst — choose a creature you control; each opponent then chooses a creature with equal or lesser power, and every creature not chosen is exiled",
				Cards:      mine,
				Min:        1,
				Max:        1,
				Zone:       game.ZoneBattlefield,
				ThenKey:    game.ChooseCardsRunLeg,
			}
			_, err := g.ChooseCardsRunThenKeyedForEffect([]game.ChooseCardsPrompt{leg}, sculptedBurstMine, item.SourceCardID, game.RunData{Carry: []uuid.UUID{item.Controller}})
			return err
		},
	})
}

// sculptedBurstMine is my choice made: each opponent is asked for a
// creature of equal or lesser power. An on-disk key: never renamed, never
// reused.
var sculptedBurstMine = effects.RunThen("prompt-run/sculpted-sunburst-mine", func(ctx *effects.Context, r game.RunDone) error {
	g := ctx.Game
	chosen := r.Picks().Cards()
	if len(chosen) == 0 || len(r.Carry) == 0 {
		return nil
	}
	mine, ok := g.LookupCardForEffect(chosen[0])
	if !ok {
		// It left before the opponents chose; its last known power is not
		// kept, so nothing is on offer to them and the rest are exiled.
		return sculptedBurstExile(g, chosen)
	}
	power := mine.CurrentPower()
	var legs []game.ChooseCardsPrompt
	for _, seat := range effects.ApnapLivingSeats(g) {
		if seat == r.Carry[0] {
			continue
		}
		var ids []uuid.UUID
		for _, id := range effects.PermanentsControlledByMatching(g, seat, effects.Creature()) {
			if c, ok := g.LookupCardForEffect(id); ok && c.CurrentPower() <= power {
				ids = append(ids, id)
			}
		}
		if len(ids) == 0 {
			continue
		}
		legs = append(legs, game.ChooseCardsPrompt{
			Chooser:    seat,
			FromPlayer: seat,
			Source:     r.Source,
			Question:   "Sculpted Sunburst — choose a creature you control with power equal to or less than the one chosen; the rest are exiled",
			Cards:      ids,
			Min:        1,
			Max:        1,
			Zone:       game.ZoneBattlefield,
			ThenKey:    game.ChooseCardsRunLeg,
		})
	}
	if len(legs) == 0 {
		return sculptedBurstExile(g, chosen)
	}
	_, err := g.ChooseCardsRunThenKeyedForEffect(legs, sculptedBurstExileRest, r.Source, game.RunData{Carry: chosen})
	return err
})

// sculptedBurstExileRest is every answer in: exile each creature that no
// player chose. Carry holds my choice. An on-disk key: never renamed, never
// reused.
var sculptedBurstExileRest = effects.RunThen("prompt-run/sculpted-sunburst-exile", func(ctx *effects.Context, r game.RunDone) error {
	return sculptedBurstExile(ctx.Game, append(append([]uuid.UUID(nil), r.Carry...), r.Picks().Cards()...))
})

// sculptedBurstExile exiles every creature on the battlefield not in kept.
func sculptedBurstExile(g *game.Game, kept []uuid.UUID) error {
	keep := map[uuid.UUID]bool{}
	for _, id := range kept {
		keep[id] = true
	}
	var doomed []uuid.UUID
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.IsCreature() && !keep[c.InstanceID] {
			doomed = append(doomed, c.InstanceID)
		}
	}
	if len(doomed) > 0 {
		g.ExileCardsForEffect(doomed)
	}
	return nil
}
