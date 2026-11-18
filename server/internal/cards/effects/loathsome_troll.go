package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "d360ce89-d80d-4be6-be8c-7e7758cd5840",
		Name:         "Loathsome Troll",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}{G}: Roll a d20. Activate only if this card is in your graveyard.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ManaCost("{3}{G}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  Do(RollDice{Sides: 20, N: 1, Then: loathsomeTrollRolled}),
		}},
	})
}

// loathsomeTrollRolled is the results table, after the d20. An on-disk key.
var loathsomeTrollRolled = DiceTail("dice/loathsome-troll-rolled", func(ctx *Context, rolls []int) error {
	g, item := ctx.Game, ctx.Item
	id := ctx.Source()
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard || sourceIsNewObject(g, item) {
		return nil
	}
	switch r := rolls[0]; {
	case r <= 9:
		return g.TuckCardsToLibraryKeyedForEffect([]uuid.UUID{id}, game.TuckOptions{}, game.KeyedTail{})
	case r <= 19:
		return ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}.Apply(ctx)
	}
	return ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield, Controller: ctx.Controller(), Tapped: true}.Apply(ctx)
})
