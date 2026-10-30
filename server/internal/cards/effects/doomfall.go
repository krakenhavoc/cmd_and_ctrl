package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Doomfall — Sorcery {2}{B} (Marvel Super Heroes Commander, #4455):
//
//	"Choose one —
//	 • Target opponent exiles a creature they control.
//	 • Target opponent reveals their hand. You choose a nonland card from
//	   it. Exile that card."
//
// The first bullet is an exile edict: the TARGETED opponent is the chooser
// (not the caster) of one of the creatures they control, queued as a
// choose_cards prompt with a registered continuation that exiles the pick
// (the sacrifice edicts' shape, with exile as the verb). It is a choice,
// so a creature that left in the meantime is withdrawn from the pick
// (LiveZone) and an opponent with no creature is simply not asked. The
// second bullet is Null Summoner's revealed-hand exile, unlinked.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ba2c2a65-256a-46db-bf2b-b02b7d114a42",
		Name:         "Doomfall",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Target opponent exiles a creature they control.",
				TargetPlayer("target opponent", Opponent()),
				func(item *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					return msc17ExileEdict(ctx, t.ID, "Doomfall — exile a creature you control", Creature(), doomfallExile)
				}),
			ModeDoing("Target opponent reveals their hand. You choose a nonland card from it. Exile that card.",
				TargetPlayer("target opponent", Opponent()),
				func(item *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					return ChooseFromRevealedHand{
						Player: t.ID,
						Filter: Nonland(),
						Label:  "nonland card",
						Exile:  true,
					}.Apply(ctx)
				}),
		),
	})
}

// doomfallExile exiles the creature the targeted opponent chose.
var doomfallExile = game.RegisterChooseCardsThen("choose-cards/doomfall-exile",
	func(g *game.Game, r game.CardsChosen) error {
		return g.ExileCardsKeyedForEffect(r.Picked, game.KeyedTail{})
	})

// msc17ExileEdict asks `victim` to choose one permanent they control that
// matches (an exile edict: they choose, they exile), then runs `then` on
// the pick. An empty candidate set asks nothing.
func msc17ExileEdict(ctx *Context, victim uuid.UUID, question string, match CardPredicate, then game.ChooseCardsThen) error {
	var cards []uuid.UUID
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == victim && match(ctx.Game, victim, c) {
			cards = append(cards, c.InstanceID)
		}
	}
	if len(cards) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    victim,
		FromPlayer: victim,
		Source:     ctx.Source(),
		Question:   question,
		Cards:      cards,
		Min:        1,
		Max:        1,
		ThenKey:    then,
		LiveZone:   game.ZoneBattlefield,
	})
	return nil
}
