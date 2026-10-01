package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Palace Jailer — Creature — Human Soldier {2}{W}{W}, 2/2:
//
//	"When this creature enters, you become the monarch.
//	 When this creature enters, exile target creature an opponent
//	 controls until an opponent becomes the monarch."
//
// Two ETB triggers, ordered by the controller like any pair (CR
// 603.3b). The monarch one is WhenThisEntersYouBecomeTheMonarch.
//
// # "Until an opponent becomes the monarch" (#1722, ADR 0096)
//
// The exile is a one-shot with an "until" (CR 610.3): the card comes
// back when the named event happens, and the event is about the crown,
// not about the Jailer — so the return must not live on the Jailer.
// An Oblivion Ring's "until this leaves the battlefield" is a leave
// trigger on the permanent (Ossification, Hostage Taker); a Jailer
// that dies keeps its creature exiled, and one that is bounced and
// recast exiles a second creature without releasing the first.
//
// So the exile schedules an EVENT-conditioned delayed trigger (#663's
// On/Condition, the Doublecast shape) on EventMonarchChanged, with the
// condition "monarch/an-opponent-became": the new monarch is somebody
// other than the trigger's controller, who is the player that
// controlled the Jailer's ability (CR 603.7d) and so the one "an
// opponent" is relative to. Its Duration is Indefinite — the
// "this turn" default event-conditioned triggers get would release
// the creature at cleanup. The body is flicker's return-to-owners, so
// the creature comes back under its OWNER's control as a new object
// (CR 400.7, CR 610.3c).
//
// An opponent already wearing the crown when the exile resolves does
// not release anything: they did not BECOME the monarch after it. The
// usual line is to order the monarch trigger to resolve first, and
// then the creature stays gone until somebody takes the crown from you.
//
// Two ways this differs from the printed text:
//
//   - The return goes on the stack as a trigger, where CR 610.3c makes
//     it an immediate one-shot. The table gets a window in which the
//     creature is still in exile, and nothing in the catalog can stop
//     it coming back. The Oblivion Ring family (Ossification, Hostage
//     Taker) takes the same posture and is not caveated for it.
//   - THE DECLARED CAVEAT. A delayed trigger belongs to its controller,
//     and eliminatePlayerLocked drops a departed player's delayed
//     triggers (CR 800.4a) BEFORE the CR 725.4 hand-on crowns somebody
//     else. So if the Jailer's controller leaves the game, the creature
//     stays exiled for good — even though the player who takes the
//     crown as they leave is an opponent of theirs, and the printed
//     "until" would be over. Pinned by
//     TestPalaceJailerControllerLeavingKeepsTheCreatureExiled; ADR 0096
//     records it as open. It never helps the Jailer's controller, who
//     has left.
func init() {
	Register(Spec{
		OracleID:     "180eda7c-fca2-403b-85cd-8ffebaf9f408",
		Name:         "Palace Jailer",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"If Palace Jailer's controller leaves the game, the exiled creature stays in exile.",
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouBecomeTheMonarch("Palace Jailer"),
			Targeting(WhenThisEnters(palaceJailerExileLabel, palaceJailerExile),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}

const palaceJailerExileLabel = "Palace Jailer — exile target creature an opponent controls until an opponent becomes the monarch"

// palaceJailerExile exiles the chosen creature and, if it reached
// exile, schedules its release.
func palaceJailerExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		exiled := t.ID
		controller := item.Controller
		return ExileTarget{Target: exiled, Then: func(ctx *Context, ok bool) error {
			if !ok {
				return nil
			}
			duration := game.IndefiniteDuration()
			ctx.Game.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
				Controller:   controller,
				SourceCardID: ctx.Source(),
				Label:        "Palace Jailer — an opponent became the monarch: return the exiled creature",
				On:           []game.EventKind{game.EventMonarchChanged},
				Condition:    anOpponentBecameTheMonarchCondition,
				Cards:        []uuid.UUID{exiled},
				Body:         returnExiledToOwnersBody,
				Duration:     &duration,
			})
			return nil
		}}.Apply(ctx)
	}
	return nil
}
