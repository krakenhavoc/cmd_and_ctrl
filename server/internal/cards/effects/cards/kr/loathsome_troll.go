package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Loathsome Troll — Creature — Troll {3}{G}{G}, 6/2 (Adventures in the
// Forgotten Realms, S82):
//
//	"{3}{G}: Roll a d20. Activate only if this card is in your graveyard.
//	 1—9 | Put this card on top of your library.
//	 10—19 | Return this card to your hand.
//	 20 | Return this card to the battlefield tapped."
//
// An ability that works from the graveyard (Zones, CR 113.6); the zone is
// the activation restriction, so a Troll anywhere else cannot activate it.
// The d20 goes through the one dice door and the tail moves "this card" only
// if it is still the card that was in the graveyard when the ability was
// activated (CR 400.7): the Troll gone from the graveyard, or back as a new
// object, is left alone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d360ce89-d80d-4be6-be8c-7e7758cd5840",
		Name:         "Loathsome Troll",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{G}: Roll a d20. Activate only if this card is in your graveyard.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{3}{G}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  effects.Do(effects.RollDice{Sides: 20, N: 1, Then: loathsomeTrollRolled}),
		}},
	})
}

// loathsomeTrollRolled is the results table, after the d20. An on-disk key.
var loathsomeTrollRolled = effects.DiceTail("dice/loathsome-troll-rolled", func(ctx *effects.Context, rolls []int) error {
	g, item := ctx.Game, ctx.Item
	id := ctx.Source()
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard || effects.SourceIsNewObject(g, item) {
		return nil
	}
	switch r := rolls[0]; {
	case r <= 9:
		return g.TuckCardsToLibraryKeyedForEffect([]uuid.UUID{id}, game.TuckOptions{}, game.KeyedTail{})
	case r <= 19:
		return effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}.Apply(ctx)
	}
	return effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield, Controller: ctx.Controller(), Tapped: true}.Apply(ctx)
})
