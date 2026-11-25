package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blood Baron of Vizkopa — Creature — Vampire {3}{W}{B}, 4/4 (Commander 2017,
// #5533):
//
//	"Lifelink, protection from white and from black
//	 As long as you have 30 or more life and an opponent has 10 or less
//	 life, this creature gets +6/+6 and has flying."
//
// Serra Ascendant's life-gated pump and flying with Lurking Jackals's
// "an opponent has 10 or less life": both halves are read live through the
// layers, so the creature shrinks the moment either life total moves out of
// range. An opponent who has left the game does not count.
//
// No simplification.
func init() {
	whileBoth := func(target *game.Card, g *game.Game, source *game.Card) bool {
		if target.InstanceID != source.InstanceID || !effects.ControllerLifeAtLeast(g, source, 30) {
			return false
		}
		for _, p := range effects.LivePlayers(g) {
			if p.ID != source.Controller && p.Life <= 10 {
				return true
			}
		}
		return false
	}
	effects.Register(effects.Spec{
		OracleID:        "326adf88-3f6d-4bc2-aafa-dd55c6c95784",
		Name:            "Blood Baron of Vizkopa",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink", "protection from white", "protection from black"},
		Static: []game.StaticAbility{
			effects.LifeGatedPump(whileBoth, 6, 6),
			effects.LifeGatedKeyword(whileBoth, "flying"),
		},
	})
}
