package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brightspear Zealot — Creature — Human Soldier {2}{W}, 2/4 (Edge of
// Eternities, #4461):
//
//	"Vigilance
//	 This creature gets +2/+0 as long as you've cast two or more spells
//	 this turn."
//
// Vigilance is a printed keyword. The pump is a layer 7c modification
// on the Zealot alone, gated on its controller's cast tally for the
// turn (any spell, creature or not), Lyse Hext's shape: the static
// declares DependsOnSpellsCast, so the layers are recomputed the moment
// the second spell is cast rather than on some later event, and the
// turn change that clears the tally ends it.
//
// No simplification.
func init() {
	pump := game.StaticAbility{
		Label:    "Gets +2/+0 as long as you've cast two or more spells this turn.",
		Layer:    game.Layer7PT,
		SubLayer: game.SubLayer7C_Modify,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			return selfCreature(target, g, source) && g.CastTallyFor(source.Controller).Total >= 2
		},
		Apply:               pumpBy(2, 0),
		DependsOnSpellsCast: true,
	}
	Register(Spec{
		OracleID:        "31a0c9a2-c9bd-47f7-872b-a43fb64d873c",
		Name:            "Brightspear Zealot",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Static:          []game.StaticAbility{pump},
	})
}
