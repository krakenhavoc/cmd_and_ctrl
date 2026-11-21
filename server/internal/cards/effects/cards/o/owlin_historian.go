package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Owlin Historian — Creature — Bird Cleric {2}{W}, 2/3 (Secrets of
// Strixhaven, #4456):
//
//	"Flying
//	 When this creature enters, surveil 1.
//	 Whenever one or more cards leave your graveyard, this creature gets
//	 +1/+1 until end of turn."
//
// The second trigger is Ark of Hunger's batch trigger (OncePerBatch over
// leftGraveyardEvents / whenCardsLeaveYourGraveyard): cards that leave as one
// event trigger it once, and a card cast from the graveyard counts. The
// bonus is for the Historian that triggered, so a copy that left or came
// back as a new object gets nothing (CR 400.7). The surveil's own cards go
// into the graveyard, not out of it, so it never triggers the second ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e1ebb60e-fbf3-4f8f-a76f-d637ec59ffb0",
		Name:            "Owlin Historian",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Owlin Historian — surveil 1", effects.Do(effects.Surveil{N: 1})),
			effects.OncePerBatch(effects.OnAny(effects.LeftGraveyardEvents, effects.WhenCardsLeaveYourGraveyard,
				"Owlin Historian — gets +1/+1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.SourceIsStillThisPermanent(g, item) {
						return nil
					}
					return effects.BoostUntilEOT{
						Target:    item.SourceCardID,
						Power:     1,
						Toughness: 1,
						Label:     "Owlin Historian — +1/+1 until end of turn",
					}.Apply(effects.NewContext(g, item))
				})),
		},
	})
}
