package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Razia's Purification — Sorcery {4}{R}{W} (Ravnica: City of Guilds, S82 OLD
// batch 581, #5600):
//
//	"Each player chooses three permanents they control, then sacrifices the
//	 rest."
//
// Planetary Annihilation's keep-set shape over every permanent: each seat with
// more than three permanents (APNAP, CR 101.4) is asked once for the three it
// keeps, a seat with three or fewer is not asked (it keeps them all), and only
// when every seat has answered are all the others sacrificed at once
// (CR 701.21a: indestructible and regeneration do not help).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "60e95faf-1496-4298-9588-52b50b05fc6b",
		Name:         "Razia's Purification",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepAllPermanents, How: game.SweepSacrifice, Partial: true}},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			g := ctx.Game
			var legs []game.ChooseCardsPrompt
			var asked []uuid.UUID
			for _, seat := range effects.ApnapPlayers(g) {
				perms := effects.PermanentsControlledByMatching(g, seat, func(*game.Game, uuid.UUID, game.Card) bool { return true })
				if len(perms) <= 3 {
					continue
				}
				asked = append(asked, seat)
				legs = append(legs, game.ChooseCardsPrompt{
					Chooser:    seat,
					FromPlayer: seat,
					Source:     item.SourceCardID,
					Question:   "Razia's Purification — choose three permanents you control; you sacrifice the rest",
					Cards:      perms,
					Min:        3,
					Max:        3,
					Zone:       game.ZoneBattlefield,
					ThenKey:    game.ChooseCardsRunLeg,
				})
			}
			if len(legs) == 0 {
				return nil
			}
			_, err := g.ChooseCardsRunThenKeyedForEffect(legs, raziasPurificationRun, item.SourceCardID, ctx.RunData(asked))
			return err
		},
	})
}

// raziasPurificationRun is the sacrifice once every seat has answered: every
// permanent of an asked seat that was not kept. An on-disk key: never renamed,
// never reused.
var raziasPurificationRun = effects.RunThen("prompt-run/razias-purification-sacrifice-the-rest", func(ctx *effects.Context, r game.RunDone) error {
	kept := map[uuid.UUID]bool{}
	for _, id := range r.Picks().Cards() {
		kept[id] = true
	}
	var rest []uuid.UUID
	for _, seat := range r.Carry {
		for _, id := range effects.PermanentsControlledByMatching(ctx.Game, seat, func(*game.Game, uuid.UUID, game.Card) bool { return true }) {
			if !kept[id] {
				rest = append(rest, id)
			}
		}
	}
	if len(rest) == 0 {
		return nil
	}
	return ctx.Game.SacrificeAllKeyedForEffect(r.Source, rest, game.KeyedTail{})
})
