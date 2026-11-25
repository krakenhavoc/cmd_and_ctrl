package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Clara Oswald — Legendary Creature — Human Advisor {6}, 2/6 (Doctor Who,
// #5948):
//
//	"Impossible Girl — If Clara Oswald is your commander, choose a color
//	 before the game begins. Clara Oswald is the chosen color.
//	 If a triggered ability of a Doctor you control triggers, that ability
//	 triggers an additional time.
//	 Doctor's companion (You can have two commanders if the other is the
//	 Doctor.)"
//
// Impossible Girl is CR 903.4b (ChoosesCommanderColor,
// game/commander_color.go): the deck checker reads the chosen colour off
// the deck's cards, and as the game begins the engine records it, or asks
// its owner when the deck shows none. Clara is that colour in every zone,
// and the colour joins her colour identity. "Impossible Girl" is an
// ability word (CR 207.2c).
//
// The second paragraph is Harmonic Prodigy's doubler shape
// (DoublesAbilitiesOf with SourceMatch, CR 603.2d): the ability's source
// must be a permanent (a Doctor you control), a Doctor by its last-known
// characteristics, and controlled by Clara's controller (the default
// check). "Doctor" is a word of Clara's text (CR 612.2).
//
// Doctor's companion is a deck-construction rule (CR 702.124m, partner.go
// in internal/deck); a search for "a card with doctor's companion" reads
// gapDoctorsCompanionCards, which lists her.
//
// No simplification.
func init() {
	d := effects.DoublesAbilitiesOf(nil, effects.DoublesAbilitiesOfOptions{
		SourceMatch: func(_ *game.Game, q game.TriggerDoublingQuery) bool {
			if !q.SourceIsPermanent {
				return false
			}
			src := effects.CharacteristicCard(q.Source.InstanceID, q.Source, q.SourceLKI)
			return src.HasSubtype(q.Doubler.TextSubtype("Doctor"))
		},
	})
	d.Label = "Clara Oswald"
	d.RowLabel = "If a triggered ability of a Doctor you control triggers, that ability triggers an additional time."
	effects.Register(effects.Spec{
		OracleID:              "542eae01-058d-4959-92a8-49d6b1e33ecd",
		Name:                  "Clara Oswald",
		Completeness:          effects.CompletenessFull,
		ChoosesCommanderColor: true,
		TriggerDoublers:       []game.TriggerDoubler{d},
	})
}
