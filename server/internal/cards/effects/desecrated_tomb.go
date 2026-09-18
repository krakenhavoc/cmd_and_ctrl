package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Desecrated Tomb — Artifact for {3} (EDHREC rank 4368):
//
//	"Whenever one or more creature cards leave your graveyard,
//	 create a 1/1 black Bat creature token with flying."
//
// A colourless engine for any deck that recurs creatures: every
// Reanimate, every Eternal Witness loop, every graveyard-exiling
// escape cost hands you a flying body. Roadmap batch 42 (#449), "no
// new machinery".
//
// "ONE OR MORE … LEAVE" is a batch trigger, not a per-card one: a
// mass reanimation that returns four creatures makes ONE Bat, and
// Living Death makes one Bat rather than a swarm. The engine emits
// one zone-move event per card, so OncePerBatch is what turns that
// stream back into the single trigger the card prints — it fires on
// the first event of a batch and declines every later event of the
// same batch (#829, AGENTS.md §7).
//
// LEAVE, in any direction: to the battlefield, to hand, to exile, to
// the library. The Tomb does not care where they went, only that they
// are no longer in your graveyard — which is why it is as good with
// Scavenging Ooze as with Animate Dead.
//
// YOUR graveyard. An opponent exiling their own graveyard gives you
// nothing, and an opponent exiling YOURS gives you a Bat for their
// trouble.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "07ff619a-21ee-44b6-b666-ceab4a78096a",
		Name:         "Desecrated Tomb",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventZoneMove,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b02CreatureCardLeftYourGraveyard(ev, source, g)
				},
				"Desecrated Tomb — create a 1/1 black Bat with flying",
				func(g *game.Game, item *game.StackItem) error {
					return CreateToken{
						Controller: item.Controller,
						Template:   b42BlackBatFlyingToken(),
						N:          1,
					}.Apply(NewContext(g, item))
				})),
		},
	})
}
