package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Profane Memento — Artifact {1} (Magic 2015, #5556):
//
//	"Whenever a creature card is put into an opponent's graveyard from
//	 anywhere, you gain 1 life."
//
// Vulturous Zombie's arrival reading (S82OldB256CardPutIntoAnOpponentsGraveyard:
// a discard, a mill, a death, a countered spell and a card put there from
// exile each count, a token never does, a card that only changes graveyard
// does not) narrowed to creature cards, one trigger per card (CR 603.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c81d1890-e741-49ec-a6ce-86c27fd4d502",
		Name:         "Profane Memento",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny(append([]game.EventKind{game.EventCounterSpell}, effects.GraveyardArrivalKinds...),
				effects.S82OldB358CreatureCardPutIntoAnOpponentsGraveyard,
				"Profane Memento — you gain 1 life",
				effects.Do(effects.GainLife{Amount: 1})),
		},
	})
}
