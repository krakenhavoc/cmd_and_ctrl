package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Megaflora Jungle — Plane — Gargantikar (Commander Masters, #4870):
//
//	"Each creature with mana value 2 or less gets +2/+2.
//	 Whenever chaos ensues, create a 1/1 green Insect creature token with
//	 flying named Butterfly."
//
// A plane (ADR 0216 §6). The static is a layer-7c anthem over every
// creature, whoever controls it, whose mana value is 2 or less, read live
// each pass (a token has mana value 0, CR 202.3). The chaos ability (CR
// 311.7) makes the planar controller one Butterfly, a green Insect with
// flying whose name is Butterfly.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c907333b-eae8-444f-9da3-b193289f19e2",
		Name:         "Megaflora Jungle",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{b16Anthem(func(target *game.Card, g *game.Game, source *game.Card) bool {
			return target.IsCreature() && target.ManaValue() <= 2
		}, 2, 2)},
		Triggered: []game.TriggeredAbility{
			WheneverChaosEnsues("Megaflora Jungle — create a 1/1 green Insect creature token with flying named Butterfly",
				Do(CreateToken{Template: TokenCard("1/1 green Insect with flying named Butterfly"), N: 1})),
		},
	})
}
