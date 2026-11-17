package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gimbal, Gremlin Prodigy — Legendary Creature — Gremlin Artificer
// {2}{G}{U}{R}, 4/4 (March of the Machine Commander, #5477):
//
//	"Artifact creatures you control have trample.
//	 At the beginning of your end step, create a 0/0 red Gremlin artifact
//	 creature token. Put X +1/+1 counters on it, where X is the number of
//	 differently named artifact tokens you control."
//
// The grant is a layer-6 keyword read through the layered types, so a
// Vehicle has trample while it is a creature, and Gimbal (not an artifact)
// has none. The end-step trigger is Deekah's two steps
// (s82MomB14CreateGremlin, s82MomB14PutGremlinCounters): the token is
// created, then X is counted, so the new Gremlin counts itself and X is at
// least 1.
//
// No simplification.
func init() {
	grant := KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.IsCreature() && target.IsArtifact() && target.Controller == source.Controller
	}, "trample")
	grant.Label = "Artifact creatures you control have trample."
	Register(Spec{
		OracleID:     "e7441404-71ff-449e-812b-18cd52eabb98",
		Name:         "Gimbal, Gremlin Prodigy",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{grant},
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Gimbal, Gremlin Prodigy — create a 0/0 red Gremlin artifact creature token and put X +1/+1 counters on it, where X is the number of differently named artifact tokens you control",
				Do(
					s82MomB14CreateGremlin{CreateToken{Template: TokenCard("0/0 red Gremlin artifact"), N: 1}},
					s82MomB14PutGremlinCounters{CounterOnThis{Kind: game.CounterPlusOne, N: 1}},
				)),
		},
	})
}
