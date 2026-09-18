package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Manglehorn — Creature — Beast {2}{G}, 2/2 (EDHREC rank 2253):
//
//	"When this creature enters, you may destroy target artifact.
//	 Artifacts your opponents control enter tapped."
//
// The green artifact hoser. Two halves:
//
//   - The ETB is Acidic Slime's targeted trigger with a "you may":
//     the controller is asked, then picks the artifact on the board,
//     and the destroy resolves off the stack (CR 608.2b re-checks
//     the target).
//   - The tapped entry is Authority of the Consuls' CR 614
//     replacement narrowed to artifacts — the entering permanent's
//     controller must not be Manglehorn's. A cast, reanimated,
//     fetched or flickered artifact runs the entry pipeline and
//     arrives tapped, as printed; an artifact creature is an
//     artifact and counts.
//
// An opponent's artifact TOKEN — a Treasure, a Clue — enters tapped
// too, since #762: a created token now runs the same
// battlefield-entry pipeline every other permanent runs, so this
// replacement sees it exactly as it sees a cast artifact.
func init() {
	Register(Spec{
		OracleID:     "b67db32b-30a9-49d2-b4c2-90a9f80c36eb",
		Name:         "Manglehorn",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:        []game.EventKind{game.EventETB},
			AppliesTo:      b06SelfETB,
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Manglehorn — destroy target artifact?"},
			Targets:        TargetPermanent("target artifact", Artifact()),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return destroyChosenTargetTrigger(source, "Manglehorn — destroy target artifact")
			},
		}},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield {
					return false
				}
				entering, ok := g.LookupCardForEffect(ev.CardID)
				if !ok || entering.Controller == src.Controller {
					return false
				}
				return entering.IsArtifact()
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.EntersTapped = true
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Manglehorn: artifacts your opponents control enter tapped",
		}},
	})
}
