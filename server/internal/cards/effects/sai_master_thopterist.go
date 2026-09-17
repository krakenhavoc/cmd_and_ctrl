package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sai, Master Thopterist — Legendary Creature — Human Artificer
// {2}{U}, 1/4 (EDHREC rank 820):
//
//	"Whenever you cast an artifact spell, create a 1/1 colorless
//	 Thopter artifact creature token with flying.
//	 {1}{U}, Sacrifice two artifacts: Draw a card."
//
// The artifact deck's token engine: every rock, every Treasure spell,
// every equipment is a flier. The cast trigger is Beast Whisperer's
// shape; the Thopter is itself an artifact, so it feeds the
// artifact-ETB payoffs (Reckless Fireweaver) as printed.
//
// "{1}{U}, Sacrifice two artifacts: Draw a card." is a mana component
// plus a sacrifice cost with a count of two (#747, SacrificeN). Sai is
// not an artifact, so she is never one of the two.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "52241b9c-7a69-4176-9234-8bdab09d8e64",
		Name:         "Sai, Master Thopterist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller {
					return false
				}
				spell, ok := g.LookupCardForEffect(ev.CardID)
				return ok && spell.IsArtifact()
			}, "Sai, Master Thopterist — create a Thopter", Do(CreateToken{
				Template: TokenCard("1/1 colorless Thopter artifact with flying"),
				N:        1,
			})),
		},
		Activated: []ActivatedAbility{{
			Label:  "{1}{U}, Sacrifice two artifacts: Draw a card.",
			Cost:   Plus(ManaCost("{1}{U}"), SacrificeN(2, "two artifacts", Artifact())),
			Effect: Do(DrawCards{N: 1}),
		}},
	})
}
