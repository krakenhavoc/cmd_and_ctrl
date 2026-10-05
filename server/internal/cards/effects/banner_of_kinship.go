package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Banner of Kinship — Artifact {5}:
//
//	"As this artifact enters, choose a creature type. This artifact
//	 enters with a fellowship counter on it for each creature you
//	 control of the chosen type.
//	 Creatures you control of the chosen type get +1/+1 for each
//	 fellowship counter on this artifact."
//
// The one as-enters choice whose answer is read by a SECOND clause: the
// counters depend on the type, so the prompt cannot be the plain stored
// one. It is the resolution-time prompt (#2382) queued from the
// AsEnters hook, and its continuation stores the type and places the
// counters in one step.
//
// "Enters with" is modelled as counters placed the moment the answer
// arrives, through the ordinary counter pipeline so Doubling Season
// applies. The window between entering and answering is closed to
// everyone (an open prompt stops priority), and an unchosen type
// reads as no type, so the lord applies to nothing in it.
func init() {
	Register(Spec{
		OracleID:     "8c220dbd-6572-4715-aae6-dd09a4252d68",
		Name:         "Banner of Kinship",
		Completeness: CompletenessFull,
		AsEnters: func(card *game.Card, ctx *Context) error {
			controller, self := card.Controller, card.InstanceID
			ChooseCreatureTypeThen(ctx.Game, controller, self, "Banner of Kinship — choose a creature type",
				func(g *game.Game, t string) error {
					if t == "" {
						return nil
					}
					return storeTribeAndCountFellowship(g, controller, self, t)
				})
			return nil
		},
		Static: []game.StaticAbility{
			TribalScalingAnthem(
				TribeFilter{Chosen: true, YoursOnly: true},
				func(source *game.Card, _ *game.Game) int {
					return source.Counters["fellowship"]
				},
			),
		},
	})
}

func storeTribeAndCountFellowship(g *game.Game, controller, self uuid.UUID, t string) error {
	if !g.SetNamedTribeForEffect(self, t) {
		return nil
	}
	n := len(creaturesOfTypeYouControl(g, controller, t))
	if n == 0 {
		return nil
	}
	return g.AddCounterForEffect(self, "fellowship", n)
}
