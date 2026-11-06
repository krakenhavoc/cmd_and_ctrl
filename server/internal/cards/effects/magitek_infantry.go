package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Magitek Infantry — Artifact Creature — Robot Soldier {W}, 1/1 (Final
// Fantasy, #4462):
//
//	"This creature gets +1/+0 as long as you control another artifact.
//	 {2}{W}: Search your library for a card named Magitek Infantry, put it
//	 onto the battlefield tapped, then shuffle."
//
// The pump is a layer 7c modification on the Infantry itself, read on
// every recompute, so it comes and goes with the other artifact (a
// second Infantry counts). The search is Bogbrew Witch's: by name, onto
// the battlefield tapped, then shuffle; finding nothing still shuffles.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "563992ff-6fdf-45a0-afc2-7c73c2ec6f74",
		Name:         "Magitek Infantry",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "This creature gets +1/+0 as long as you control another artifact.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: magitekInfantryWhileAnotherArtifact,
			Apply:     pumpBy(1, 0),
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}{W}: Search your library for a card named Magitek Infantry, put it onto the battlefield tapped, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ManaCost("{2}{W}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:        item.Controller,
					Predicate:     func(c game.Card) bool { return game.HasName(c, "Magitek Infantry") },
					Dest:          game.ZoneBattlefield,
					Limit:         1,
					Shuffle:       true,
					TappedOnEntry: true,
					Source:        item.SourceCardID,
					Reason:        "Magitek Infantry — search for a card named Magitek Infantry",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}

// magitekInfantryWhileAnotherArtifact is "this creature … as long as you
// control another artifact": the source alone, while its controller
// controls an artifact other than it. It runs inside a layer recompute,
// so it walks the live battlefield (as artifactsControlledBy does), whose
// cards carry what the earlier layers wrote this pass.
func magitekInfantryWhileAnotherArtifact(target *game.Card, g *game.Game, source *game.Card) bool {
	if !selfCreature(target, g, source) {
		return false
	}
	for i := range g.Battlefield.Cards {
		other := &g.Battlefield.Cards[i]
		if other.InstanceID != source.InstanceID && other.Controller == source.Controller && other.IsArtifact() {
			return true
		}
	}
	return false
}
