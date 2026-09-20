package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wight of the Reliquary — Creature — Zombie Knight {B}{G}, 2/2
// (EDHREC rank 1259):
//
//	"Vigilance
//	 This creature gets +1/+1 for each creature card in your graveyard.
//	 {T}, Sacrifice another creature: Search your library for a land
//	 card, put it onto the battlefield tapped, then shuffle."
//
// A Knight of the Reliquary whose currency is creatures rather than
// lands. The pump is a Layer 7c self-static over the controller's
// graveyard; the ramp is a CR 602 ability whose sacrifice is a cost,
// paid at announce, so the dies-triggers land above the ability and
// the fetched land comes after them.
//
// Two declared gaps, both weaker than printed:
//
//   - "Another creature" is by name (b03NotNamed): a cost's clause
//     never receives its source, and in a singleton format the name
//     is the creature. A token copy of the Wight could not be fed to
//     it either.
//
// It also carried the graveyard-staleness simplification until
// #1117 — the layer engine recomputed on battlefield and counter
// events, not on graveyard traffic, so a creature card milled or
// discarded grew the Wight at the next battlefield event rather than
// at once. A graveyard crossing now bumps the layer version itself.
func init() {
	Register(Spec{
		OracleID:        "4507df69-6bf7-43d6-a609-c032b61835d5",
		Name:            "Wight of the Reliquary",
		Completeness:    CompletenessCaveats,
		PrintedKeywords: []string{"vigilance"},
		Caveats: []string{
			"A second copy or token copy of Wight of the Reliquary can't be sacrificed to its own ability.",
		},
		Static: []game.StaticAbility{{
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
		Activated: []ActivatedAbility{{
			Label: "{T}, Sacrifice another creature: Search your library for a land card, put it onto the battlefield tapped, then shuffle.",
			Cost: Plus(TapCost(), game.AbilityCost{
				SacrificeOther: sacrificeSpec("another creature", Creature(), b03NotNamed("Wight of the Reliquary")),
			}),
			Effect: b11FetchLandTapped("Wight of the Reliquary — a land card, onto the battlefield tapped"),
		}},
	})
}
