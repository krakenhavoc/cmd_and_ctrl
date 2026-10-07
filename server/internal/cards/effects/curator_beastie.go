package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Curator Beastie — Creature — Beast {4}{G}{G}:
//
//	"Reach
//	 Colorless creatures you control enter with two additional +1/+1
//	 counters on them.
//	 Whenever this creature enters or attacks, manifest dread."
//
// The counters are a real CR 614 entry replacement, like Grumgully's,
// so they arrive with the creature and Hardened Scales applies. The
// replacement reads the permanent AS IT WILL ENTER (CR 614.12): a
// card put onto the battlefield face down is a colorless creature
// (CR 708.2) whatever it is in the library, which is what makes the
// manifest this card makes arrive as a 4/4.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6cf8a502-6620-460f-8fab-b4ede28beb11",
		Name:            "Curator Beastie",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Replacements:    []game.ReplacementEffect{curatorBeastieCounters()},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Curator Beastie — manifest dread", Do(ManifestDread{})),
		},
	})
}

func curatorBeastieCounters() game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield || src == nil {
				return false
			}
			entering, ok := g.LookupCardForEffect(ev.CardID)
			if !ok || entering.Controller != src.Controller {
				return false
			}
			if ev.FaceDown != game.FaceDownNone {
				// A face-down permanent is colorless; it is a creature
				// unless a listing says it is something else.
				if ev.FaceDownListed == nil {
					return true
				}
				for _, t := range ev.FaceDownListed.Types {
					if t == "Creature" {
						return true
					}
				}
				return false
			}
			return entering.IsCreature() && entering.IsColorless()
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.AddCounterAtETB(game.CounterPlusOne, 2)
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: "Curator Beastie: colorless creatures enter with two additional +1/+1 counters",
	}
}
