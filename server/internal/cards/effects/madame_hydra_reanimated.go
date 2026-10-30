package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Madame Hydra, Reanimated — Legendary Creature — Human Horror Villain
// {3}{B}, 2/2 (Marvel Super Heroes Commander, #4455):
//
//	"Menace
//	 Madame Hydra gets +1/+1 for each creature card in your graveyard.
//	 Whenever Madame Hydra enters or attacks, mill two cards."
//
// The pump is a layer 7c modifier recounted on every pass (Jarad's
// shape), so a creature card that lands from the mill is counted the
// moment it arrives. The mill is one ability watching enters and
// attacks.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7274f6ed-310d-4b65-8083-eedb8e885a33",
		Name:            "Madame Hydra, Reanimated",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Static: []game.StaticAbility{{
			Label:    "Gets +1/+1 for each creature card in your graveyard.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := b11CreatureCardsInGraveyard(g, source.Controller)
				c.Power += n
				c.Toughness += n
			},
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Madame Hydra — mill two cards", Do(MillCards{N: 2})),
		},
	})
}
