package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cid, Timeless Artificer — Legendary Creature — Human Artificer {2}{W}{U},
// 4/4 (Final Fantasy, #4462):
//
//	"Artifact creatures and Heroes you control get +1/+1 for each Artificer
//	 you control and each Artificer card in your graveyard.
//	 A deck can have any number of cards named Cid, Timeless Artificer.
//	 Cycling {W}{U}"
//
// The anthem is a layer 7c modification of every artifact creature and every
// Hero creature its controller controls, read every pass: the count is the
// Artificers (creature type, a changeling included) on the battlefield under
// the controller plus the Artificer cards in their graveyard, so Cid counts
// himself, and each point of it is +1/+1. The deckbuilding permission is
// read from the card's oracle text by deck.Validate (CR 113.6n). Cycling is
// the engine's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "87050537-99c9-4993-a770-4329b2e749e4",
		Name:         "Cid, Timeless Artificer",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "Artifact creatures and Heroes you control get +1/+1 for each Artificer you control and each Artificer card in your graveyard.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller && target.IsCreature() &&
					(target.IsArtifact() || target.HasSubtype(source.TextSubtype("Hero")))
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := artificersYouControlAndInYourGraveyard(g, source.Controller, source.TextSubtype("Artificer"))
				c.Power += n
				c.Toughness += n
			},
		}},
		Activated: []ActivatedAbility{Cycling("{W}{U}")},
	})
}

// artificersYouControlAndInYourGraveyard is "each Artificer you control and
// each Artificer card in your graveyard".
func artificersYouControlAndInYourGraveyard(g *game.Game, player uuid.UUID, artificer string) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsCreature() && c.HasSubtype(artificer) {
			n++
		}
	}
	if p := g.PlayerByIDForEffect(player); p != nil && p.Graveyard != nil {
		for _, c := range p.Graveyard.Cards {
			if c.HasSubtype(artificer) {
				n++
			}
		}
	}
	return n
}
