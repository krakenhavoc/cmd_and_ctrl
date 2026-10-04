package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Memory Theft — Sorcery {2}{B}:
//
//	"Target opponent reveals their hand. You choose a nonland card from
//	 it. That player discards that card. You may put a card that has an
//	 Adventure that player owns from exile into that player's
//	 graveyard."
//
// Thoughtseize's pick (ADR 0116), then an optional second choice. "A
// card that has an Adventure" is any face-up adventurer card in exile
// that the targeted opponent owns, whether or not it was ever cast as
// an Adventure (the 2019-10-04 rulings); a face-down card shows no
// characteristics and is never offered. You may choose one even when
// the hand had nothing to discard, and you may choose none.
//
// The exile choice is printed after the discard and is queued on the
// next line, after the pick (ADR 0116 §6). It neither reads the chosen
// card nor changes which cards may be chosen: the candidates are in
// exile, and a discarded card goes to the graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1d60e2a7-2059-48cb-a6ab-36ba25b80b3a",
		Name:         "Memory Theft",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			victim := TargetedPlayer(ctx)
			if victim == uuid.Nil {
				return nil
			}
			if err := (ChooseFromRevealedHand{Player: victim, Filter: Nonland(), Label: "nonland card"}).Apply(ctx); err != nil {
				return err
			}
			return memoryTheftAdventure(ctx.Game, item, victim)
		},
	})
}

// memoryTheftAdventure offers the victim's exiled adventurer cards and
// moves the one chosen, if any, to that player's graveyard.
func memoryTheftAdventure(g *game.Game, item *game.StackItem, victim uuid.UUID) error {
	var offer []uuid.UUID
	if g.Exile != nil {
		for _, c := range g.Exile.Cards {
			if c.Owner == victim && c.Layout == game.LayoutAdventure && !c.FaceDown {
				offer = append(offer, c.InstanceID)
			}
		}
	}
	if len(offer) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    item.Controller,
		FromPlayer: victim,
		Source:     item.SourceCardID,
		Question:   "Memory Theft — you may put a card that has an Adventure from exile into its owner's graveyard",
		Cards:      offer,
		Min:        0,
		Max:        1,
		Zone:       game.ZoneExile,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			for _, id := range picked {
				if err := g.PutIntoGraveyardForEffect(id); err != nil {
					return err
				}
			}
			return nil
		},
	})
	return nil
}
