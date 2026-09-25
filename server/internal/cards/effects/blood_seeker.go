package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blood Seeker — Creature — Vampire Shaman {1}{B}, 1/1 (EDHREC rank
// 2592):
//
//	"Whenever a creature an opponent controls enters, you may have
//	 that player lose 1 life."
//
// Suture Priest's second ability on a Vampire. An optional trigger
// that is the Seeker's controller's choice — "you may HAVE that
// player lose", so the yes/no goes to the controller — one prompt
// per creature, as printed. The losing player is captured in Build
// off the entering creature's controller, so a creature that changes
// hands before the trigger resolves still costs the player who
// brought it in. Life loss, not damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "41087db6-34c4-4e2b-9f54-5e4488ca9c0b",
		Name:         "Blood Seeker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature() && c.Controller != source.Controller
			},
			Key:            "Blood Seeker — that player loses 1 life",
			Effect:         thatPlayerLosesOneLife,
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Blood Seeker — that player loses 1 life?"},
		}},
	})
}
