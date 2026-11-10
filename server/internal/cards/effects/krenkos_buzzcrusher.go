package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Krenko's Buzzcrusher — Artifact Creature — Insect Thopter {2}{R}{R},
// 4/4 (Murders at Karlov Manor, #4469):
//
//	"Flying, trample
//	 When this creature enters, for each player, destroy up to one
//	 nonbasic land that player controls. For each land destroyed this way,
//	 its controller may search their library for a basic land card, put it
//	 onto the battlefield tapped, then shuffle."
//
// Nothing is targeted (the printed ruling). Buzzcrusher's controller
// chooses, as it resolves, up to one nonbasic land for each player, in
// APNAP order (stdB41BuzzcrusherRun, the keyed run of legs Slaughter the
// Strong asks, here with the controller choosing from each player's
// lands), then the chosen lands are destroyed as one event. Each land that
// was really destroyed (an indestructible one is not, CR 702.12b) lets ITS
// controller, as it was last known, search for a basic land (Path to
// Exile's search): a land destroyed under an opponent's control gives that
// opponent the search, and one destroyed under the controller's gives the
// controller's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "79a63064-f25b-4c34-867e-a9d32da97b14",
		Name:            "Krenko's Buzzcrusher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Krenko's Buzzcrusher — for each player, destroy up to one nonbasic land; its controller may search for a basic land",
				func(g *game.Game, item *game.StackItem) error {
					return stdB41BuzzcrusherAsk(g, item)
				}),
		},
	})
}

// stdB41BuzzcrusherAsk queues one "up to one nonbasic land" leg per
// player, answered by the controller, in APNAP order. A player with no
// nonbasic land is not asked.
func stdB41BuzzcrusherAsk(g *game.Game, item *game.StackItem) error {
	var legs []game.ChooseCardsPrompt
	for _, seat := range apnapPlayers(g) {
		lands := permanentsControlledByMatching(g, seat, NonbasicLand())
		if len(lands) == 0 {
			continue
		}
		legs = append(legs, game.ChooseCardsPrompt{
			Chooser:    item.Controller,
			FromPlayer: seat,
			Source:     item.SourceCardID,
			Question:   "Krenko's Buzzcrusher — destroy up to one nonbasic land this player controls",
			Cards:      lands,
			Min:        0,
			Max:        1,
			Zone:       game.ZoneBattlefield,
			ThenKey:    game.ChooseCardsRunLeg,
		})
	}
	if len(legs) == 0 {
		return nil
	}
	ctx := NewContext(g, item)
	_, err := g.ChooseCardsRunThenKeyedForEffect(legs, stdB41BuzzcrusherRun, item.SourceCardID, ctx.RunData([]uuid.UUID{item.Controller}))
	return err
}

// stdB41BuzzcrusherRun destroys the chosen lands as one event and hands
// the ones destroyed to the search clause. An on-disk key: never renamed,
// never reused.
var stdB41BuzzcrusherRun = RunThen("run/krenkos-buzzcrusher-destroy", func(ctx *Context, r game.RunDone) error {
	ids := r.Picks().Cards()
	if len(ids) == 0 {
		return nil
	}
	return ctx.Game.DestroyPermanentsKeyedForEffect(ids, ctx.ItemTail(stdB41BuzzcrusherSearch))
})

// stdB41BuzzcrusherSearch is "for each land destroyed this way, its
// controller may search their library for a basic land card, put it onto
// the battlefield tapped, then shuffle", keyed. An on-disk key: never
// renamed, never reused.
var stdB41BuzzcrusherSearch = MassTail("mass/krenkos-buzzcrusher-search", func(ctx *Context, swept []game.Card, _ int) error {
	for _, land := range swept {
		if err := (SearchLibrary{
			Player:        land.Controller,
			Source:        ctx.Source(),
			Predicate:     IsBasicLand,
			Dest:          game.ZoneBattlefield,
			Limit:         1,
			Reveal:        false,
			Shuffle:       true,
			TappedOnEntry: true,
			Optional:      true,
			Reason:        "Krenko's Buzzcrusher — you may search for a basic land card",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
