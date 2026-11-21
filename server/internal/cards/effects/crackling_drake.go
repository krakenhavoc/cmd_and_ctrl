package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Crackling Drake — Creature — Drake {U}{U}{R}{R}, */4 (Ravnica Remastered,
// #5948):
//
//	"Flying
//	 Crackling Drake's power is equal to the total number of instant and
//	 sorcery cards you own in exile and in your graveyard.
//	 When this creature enters, draw a card."
//
// A characteristic-defining ability (CR 604.3), applied in layer 7a as Enigma
// Drake's is, setting the power only. The count is read live: the instant and
// sorcery cards in the controller's graveyard plus those in the exile zone that
// the controller owns (Cosmogoyf's count), so a card the controller owns but
// an opponent put into exile counts, and a face-down exiled card, which has no
// types (CR 708.2), does not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c8b4eba6-88ce-42d7-a235-5a4765a4a981",
		Name:            "Crackling Drake",
		Completeness:    CompletenessFull,
		Purpose:         game.Purpose{Draws: 1},
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{{
			Label:     "Crackling Drake's power is equal to the total number of instant and sorcery cards you own in exile and in your graveyard.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power = fandanielInstantsAndSorceries(g, source.Controller) + gapB74InstantsAndSorceriesOwnedInExile(g, source.Controller)
			},
			// A card arriving in or leaving exile from a zone other than the
			// battlefield or a graveyard changes the count (the graveyard half
			// and the battlefield half already invalidate the layer pass).
			DependsOnExile: true,
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Crackling Drake — draw a card", Do(DrawCards{N: 1})),
		},
	})
}

// gapB74InstantsAndSorceriesOwnedInExile counts the face-up instant and
// sorcery cards in exile that `owner` owns.
func gapB74InstantsAndSorceriesOwnedInExile(g *game.Game, owner uuid.UUID) int {
	n := 0
	if g == nil || g.Exile == nil {
		return 0
	}
	for i := range g.Exile.Cards {
		c := &g.Exile.Cards[i]
		if c.Owner == owner && !c.FaceDown && (c.IsInstant() || c.IsSorcery()) {
			n++
		}
	}
	return n
}
