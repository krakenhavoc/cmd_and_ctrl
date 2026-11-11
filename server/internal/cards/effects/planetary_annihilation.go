package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Planetary Annihilation — Sorcery {3}{R}{R} (Edge of Eternities
// Commander):
//
//	"Each player chooses six lands they control, then sacrifices the rest.
//	 Planetary Annihilation deals 6 damage to each creature."
//
// Archfiend of Depravity's keep-set shape over lands: each seat with more
// than six (APNAP, CR 101.4) is asked once for the six it keeps, a seat
// with six or fewer is not asked, and only when every seat has answered
// does everything else they control go in one simultaneous sacrifice.
// Nothing targets. The 6 damage is dealt to each creature once the
// sacrifices have settled (the sentence's second half), as one damage
// instance; a table with no seat over six deals it at once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1d45f248-5b83-4377-a3ba-4eae48e63ca1",
		Name:         "Planetary Annihilation",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 6}},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			g := ctx.Game
			var legs []game.ChooseCardsPrompt
			var asked []uuid.UUID
			for _, seat := range apnapPlayers(g) {
				lands := permanentsControlledByMatching(g, seat, Land())
				if len(lands) <= 6 {
					continue
				}
				asked = append(asked, seat)
				legs = append(legs, game.ChooseCardsPrompt{
					Chooser:    seat,
					FromPlayer: seat,
					Source:     item.SourceCardID,
					Question:   "Planetary Annihilation — choose six lands you control; the rest are sacrificed",
					Cards:      lands,
					Min:        6,
					Max:        6,
					Zone:       game.ZoneBattlefield,
					ThenKey:    game.ChooseCardsRunLeg,
				})
			}
			if len(legs) == 0 {
				return damageEachMatching(ctx, Creature(), 6)
			}
			_, err := g.ChooseCardsRunThenKeyedForEffect(legs, planetaryAnnihilationRun, item.SourceCardID, ctx.RunData(asked))
			return err
		},
	})
}

// planetaryAnnihilationRun sacrifices the lands the asked seats did not
// keep, then deals the damage. An on-disk key.
var planetaryAnnihilationRun = RunThen("prompt-run/planetary-annihilation-keep-lands", func(ctx *Context, r game.RunDone) error {
	kept := map[uuid.UUID]bool{}
	for _, id := range r.Picks().Cards() {
		kept[id] = true
	}
	var rest []uuid.UUID
	for _, seat := range r.Carry {
		for _, id := range permanentsControlledByMatching(ctx.Game, seat, Land()) {
			if !kept[id] {
				rest = append(rest, id)
			}
		}
	}
	if len(rest) == 0 {
		return damageEachMatching(ctx, Creature(), 6)
	}
	return ctx.Game.SacrificeAllKeyedForEffect(r.Source, rest, ctx.ItemTail(planetaryAnnihilationDamage))
})

// planetaryAnnihilationDamage is the 6 damage to each creature, after the
// sacrifices. An on-disk key.
var planetaryAnnihilationDamage = CardTail("then/planetary-annihilation-damage", func(ctx *Context, _ game.TailOutcome) error {
	return damageEachMatching(ctx, Creature(), 6)
})
