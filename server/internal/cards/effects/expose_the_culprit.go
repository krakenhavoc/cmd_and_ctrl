package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Expose the Culprit — Instant {1}{R} (Murders at Karlov Manor, #4469):
//
//	"Choose one or both —
//	 • Turn target face-down creature face up.
//	 • Exile any number of face-up creatures you control with disguise in a
//	   face-down pile, shuffle that pile, then cloak them. (To cloak a
//	   card, put it onto the battlefield face down as a 2/2 creature with
//	   ward {2}. Turn it face up any time for its mana cost if it's a
//	   creature card.)"
//
// The first mode is Hauntwoods Shrieker's free turn (TurnFaceUp): a
// creature the engine refuses to turn (a cloaked noncreature card, a
// permanent held down by Karlov Watchdog) stays as it is. The second asks
// for any number of the controller's face-up creatures with disguise (the
// card prints the disguise keyword; a face-down creature is not face up),
// exiles them together, shuffles the pile (Game.CloakShuffledFromExileForEffect)
// and cloaks them one at a time (CR 701.58a, e). Both modes run in printed
// order when both are chosen, and a creature that has been turned face up by
// the first can be picked by the second.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4f7e4aff-ef68-44fe-a0dc-b9ab0ecd7146",
		Name:         "Expose the Culprit",
		Completeness: CompletenessFull,
		Modes: ChooseOneOrMore(
			ModeDoing("Turn target face-down creature face up.",
				TargetCreature("target face-down creature", FaceDown()),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					for _, t := range ctx.ModeTargets(occ) {
						if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
							continue
						}
						if err := (TurnFaceUp{Targets: []uuid.UUID{t.ID}}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
			ModeDoing("Exile any number of face-up creatures you control with disguise in a face-down pile, shuffle that pile, then cloak them.",
				nil, func(item *game.StackItem, ctx *Context, _ int) error {
					return exposeTheCulpritAsk(ctx, item)
				}),
		),
	})
}

// faceUpCreaturesWithDisguise lists the face-up creatures `player`
// controls that print disguise.
func faceUpCreaturesWithDisguise(g *game.Game, player uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != player || !c.IsCreature() || c.FaceDownIsPermanent() {
			continue
		}
		if ac := game.AlternativeCostByKey(c.OracleID, "disguise"); ac == nil || ac.FaceDown == nil {
			continue
		}
		out = append(out, c.InstanceID)
	}
	return out
}

// exposeTheCulpritAsk offers the controller their face-up disguise
// creatures, any number.
func exposeTheCulpritAsk(ctx *Context, item *game.StackItem) error {
	candidates := faceUpCreaturesWithDisguise(ctx.Game, item.Controller)
	if len(candidates) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    item.Controller,
		FromPlayer: item.Controller,
		Source:     item.SourceCardID,
		Question:   "Expose the Culprit — exile any number of face-up creatures you control with disguise, then cloak them",
		Cards:      candidates,
		Min:        0,
		Max:        len(candidates),
		LiveZone:   game.ZoneBattlefield,
		ThenKey:    exposeTheCulpritExile,
		Asked:      game.NewAskedItem(item),
	})
	return nil
}

// exposeTheCulpritExile exiles the chosen creatures together. Keyed: an
// on-disk identity, never renamed or reused.
var exposeTheCulpritExile = askedPick("choose-cards/expose-the-culprit-exile", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return ctx.Game.ExileCardsKeyedForEffect(r.Picked, ctx.ItemTail(exposeTheCulpritCloak))
})

// exposeTheCulpritCloak shuffles the exiled pile and cloaks it. Keyed: an
// on-disk identity, never renamed or reused.
var exposeTheCulpritCloak = CardTail("then/expose-the-culprit-cloak", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	_, err := ctx.Game.CloakShuffledFromExileForEffect(ctx.Controller(), o.IDs)
	return err
})
