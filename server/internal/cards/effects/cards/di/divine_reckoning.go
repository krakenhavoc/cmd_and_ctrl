package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Divine Reckoning — Sorcery {2}{W}{W} (Commander 2019):
//
//	"Each player chooses a creature they control. Destroy the rest.
//	 Flashback {5}{W}{W} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// Single Combat's keep-set shape with a destroy: each seat that controls two or
// more creatures is asked once, in turn order from the active player
// (CR 101.4), for the one it keeps; a seat with one or none is not asked and
// keeps what it has. Only when every asked seat has answered is everything
// else those seats control destroyed in one sweep (regeneration may save a
// creature, as printed). Nothing targets. Flashback is the shared constructor,
// with the graveyard opened as a cast zone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "15d8b339-7520-43c9-b2dc-4a936357042b",
		Name:         "Divine Reckoning",
		Completeness: effects.CompletenessFull,
		Purpose: game.Purpose{
			Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true},
		},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{5}{W}{W}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			g := ctx.Game
			var legs []game.ChooseCardsPrompt
			var asked []uuid.UUID
			for _, seat := range effects.ApnapPlayers(g) {
				ids := effects.PermanentsControlledByMatching(g, seat, effects.Creature())
				if len(ids) <= 1 {
					continue
				}
				asked = append(asked, seat)
				legs = append(legs, game.ChooseCardsPrompt{
					Chooser:    seat,
					FromPlayer: seat,
					Source:     item.SourceCardID,
					Question:   "Divine Reckoning — choose a creature you control; the rest are destroyed",
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
			_, err := g.ChooseCardsRunThenKeyedForEffect(legs, divineReckoningDestroy, item.SourceCardID, game.RunData{Carry: asked})
			return err
		},
	})
}

// divineReckoningDestroy destroys every creature the asked seats control that
// they did not keep. An on-disk key: never renamed, never reused.
var divineReckoningDestroy = effects.RunThen("prompt-run/divine-reckoning-destroy", func(ctx *effects.Context, r game.RunDone) error {
	kept := map[uuid.UUID]bool{}
	for _, id := range r.Picks().Cards() {
		kept[id] = true
	}
	rest := map[uuid.UUID]bool{}
	for _, seat := range r.Carry {
		for _, id := range effects.PermanentsControlledByMatching(ctx.Game, seat, effects.Creature()) {
			if !kept[id] {
				rest[id] = true
			}
		}
	}
	if len(rest) == 0 {
		return nil
	}
	return effects.DestroyAllMatching{
		Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return rest[c.InstanceID] },
	}.Apply(ctx)
})
