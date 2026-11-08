package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hunter's Blowgun — Artifact — Equipment {1} (The Lost Caverns of
// Ixalan, #4470):
//
//	"Equipped creature gets +1/+1.
//	 Equipped creature has deathtouch during your turn. Otherwise, it has
//	 reach.
//	 Equip {2}"
//
// Two layer-6 keyword grants on the equipped creature (Long-Lost Lances'
// "during your turn" read), split on whose turn it is: "your" is the
// Blowgun's controller, so the creature has exactly one of the two at any
// time, deathtouch on its controller's turn and reach on everyone else's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a0292e0b-7bb3-46d3-a29f-b54bce5b20f0",
		Name:         "Hunter's Blowgun",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			Labelled(KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return AttachedToSource(target, g, source) && IsYourTurn(g, source.Controller)
			}, "deathtouch"), "Equipped creature has deathtouch during your turn."),
			Labelled(KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return AttachedToSource(target, g, source) && !IsYourTurn(g, source.Controller)
			}, "reach"), "Otherwise, equipped creature has reach."),
		},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}
