package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Single Combat — Sorcery {3}{W}{W} (Fallout):
//
//	"Each player chooses a creature or planeswalker they control, then
//	 sacrifices the rest. Players can't cast creature or planeswalker spells
//	 until the end of your next turn."
//
// Archfiend of Depravity's keep-set shape over creatures and planeswalkers
// together: each seat that controls two or more (APNAP, CR 101.4) is asked
// once for the one it keeps, a seat with one or none is not asked, and only
// when every seat has answered does everything else of those two types they
// control go in one simultaneous sacrifice. Nothing targets. The second
// sentence is a cast ban on every player still in the game, creature or
// planeswalker spells, lasting until the end of MY next turn (CR 611.2c: it
// binds the players in the game as it resolves); it is applied first, as the
// sentence order has no bearing on it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "822ef47e-3f1b-40a6-8c5c-fea3d7deba5b",
		Name:         "Single Combat",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			g := ctx.Game
			for _, p := range g.Seats {
				if p == nil || p.Eliminated {
					continue
				}
				if err := (RestrictCasting{
					Player:   p.ID,
					Rule:     game.CastBanRule{Kind: game.CastBanOutright, Filter: game.PermissionFilter{CardTypesAny: "creature planeswalker"}},
					Label:    "Single Combat — can't cast creature or planeswalker spells until the end of your next turn",
					Duration: g.UntilEndOfYourNextTurnDuration(ctx.Controller()),
				}).Apply(ctx); err != nil {
					return err
				}
			}
			var legs []game.ChooseCardsPrompt
			var asked []uuid.UUID
			for _, seat := range apnapPlayers(g) {
				ids := permanentsControlledByMatching(g, seat, Or(Creature(), Planeswalker()))
				if len(ids) <= 1 {
					continue
				}
				asked = append(asked, seat)
				legs = append(legs, game.ChooseCardsPrompt{
					Chooser:    seat,
					FromPlayer: seat,
					Source:     item.SourceCardID,
					Question:   "Single Combat — choose a creature or planeswalker you control; you sacrifice the rest",
					Cards:      ids,
					Min:        1,
					Max:        1,
					Zone:       game.ZoneBattlefield,
					ThenKey:    game.ChooseCardsRunLeg,
				})
			}
			if len(legs) == 0 {
				return nil
			}
			_, err := g.ChooseCardsRunThenKeyedForEffect(legs, singleCombatSacrifice, item.SourceCardID, game.RunData{Carry: asked})
			return err
		},
	})
}

// singleCombatSacrifice sacrifices every creature and planeswalker the asked
// seats control that they did not keep, as one simultaneous exit. An on-disk
// key: never renamed, never reused.
var singleCombatSacrifice = RunThen("prompt-run/single-combat-sacrifice", func(ctx *Context, r game.RunDone) error {
	kept := map[uuid.UUID]bool{}
	for _, id := range r.Picks().Cards() {
		kept[id] = true
	}
	var rest []uuid.UUID
	for _, seat := range r.Carry {
		for _, id := range permanentsControlledByMatching(ctx.Game, seat, Or(Creature(), Planeswalker())) {
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
