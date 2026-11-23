package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Frantic Scapegoat — Creature — Goat {R}, 1/1:
//
//	"Haste
//	 When this creature enters, suspect it. (It has menace and can't
//	 block.)
//	 Whenever one or more other creatures you control enter, if this
//	 creature is suspected, you may suspect one of the other creatures.
//	 If you do, this creature is no longer suspected."
//
// The second ability is a "one or more" trigger (OncePerBatch): one
// trigger for a whole simultaneous entry. "If this creature is
// suspected" is an intervening if (CR 603.4), read as the trigger would
// fire and again as it resolves. Entering together with other creatures
// does not trigger it, because the Goat's own "suspect it" is still on
// the stack then.
//
// "One of the other creatures" is chosen as the trigger resolves, from
// the creatures that entered under your control in the batch that
// triggered it (enteredInTriggeringBatch, #2733) and are still on the
// battlefield. Only a creature that can become suspected is offered
// (Game.CanBecomeSuspected): one already suspected, or under Airtight
// Alibi, would not be suspected, so choosing it could never be the "if
// you do". The choice does not target, so hexproof does not stop it.
// Declining, or having nothing to choose, leaves the Goat suspected.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d181310f-404b-4f26-8024-b7b537d1fd90",
		Name:            "Frantic Scapegoat",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Frantic Scapegoat — suspect it", func(g *game.Game, item *game.StackItem) error {
				return effects.Suspect{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			}),
			effects.OncePerBatch(effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, true)
				return ok && c.IsCreature() && source.Suspected
			}, "Frantic Scapegoat — you may suspect one of the other creatures", franticScapegoatPassTheBlame)),
		},
	})
}

// franticScapegoatPassTheBlame is the batch trigger's resolution: if
// the Goat is still suspected, offer the other creatures that entered,
// and move the suspicion onto the one chosen.
func franticScapegoatPassTheBlame(g *game.Game, item *game.StackItem) error {
	goat := item.SourceCardID
	// CR 603.4: re-checked on resolution. A Goat that left and came
	// back is a new object (CR 400.7), not "this creature" (#1432).
	if effects.SourceIsNewObject(g, item) || !g.IsSuspected(goat) {
		return nil
	}
	ctx := effects.NewContext(g, item)
	you := ctx.Controller()
	var offer []uuid.UUID
	for _, e := range effects.EnteredInTriggeringBatch(ctx) {
		if e.ID == goat || e.Controller != you || !g.CanBecomeSuspected(e.ID) {
			continue
		}
		offer = append(offer, e.ID)
	}
	if len(offer) == 0 {
		return nil
	}
	return effects.ChoosePermanents{
		Player:   you,
		Question: "Frantic Scapegoat: you may suspect one of the creatures that entered. If you do, Frantic Scapegoat is no longer suspected.",
		Candidates: func(*game.Game, uuid.UUID) ([]uuid.UUID, int, int) {
			return offer, 0, 1
		},
		Keyed: franticScapegoatChosen,
	}.Apply(ctx)
}

// franticScapegoatChosen is suspecting the chosen creature and, if that
// worked, no longer the Goat, keyed (#4290); the Goat is the run's source.
var franticScapegoatChosen = effects.RegisterPermanentsPick("prompt-run/frantic-scapegoat-suspect", effects.PermanentsDef{
	Then: func(ctx *effects.Context, picked game.PromptedPicks) error {
		for _, id := range picked.Cards() {
			if ctx.Game.SuspectForEffect(id) {
				ctx.Game.UnsuspectForEffect(ctx.Source())
			}
		}
		return nil
	},
})
