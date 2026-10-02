package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Palace Jailer — Creature — Human Soldier {2}{W}{W}, 2/2:
//
//	"When this creature enters, you become the monarch.
//	 When this creature enters, exile target creature an opponent
//	 controls until an opponent becomes the monarch."
//
// Two ETB triggers, ordered by the controller like any pair (CR
// 603.3b). The monarch one is WhenThisEntersYouBecomeTheMonarch.
//
// # "Until an opponent becomes the monarch" (#1722, ADR 0096, #1729)
//
// The exile is a one-shot with an "until" (CR 610.3): the card comes
// back when the named event happens, and the event is about the crown,
// not about the Jailer — so the return must not live on the Jailer.
// The rulings of 2021-03-19: "Palace Jailer leaving the battlefield
// won't cause the exiled creature to return. The game will continue to
// watch for the next time an opponent becomes the monarch", and "any
// opponent becoming the monarch will cause the card to return".
//
// ExileUntil records the return on EventMonarchChanged with the
// condition "monarch/an-opponent-became": the new monarch is somebody
// other than the player who controlled the Jailer's ability. It is a
// one-shot effect, performed immediately after the crown moves, with
// no stack in between (#1729) — and since it is not a triggered
// ability, it does not leave the game with the Jailer's controller.
// When they leave wearing the crown, CR 725.4 hands it to an opponent
// of theirs at the same moment, and the creature comes back.
//
// "If you're not the monarch as Palace Jailer's second ability
// resolves, the creature will be exiled until there's a new monarch and
// that player is one of your opponents. The creature won't immediately
// return just because an opponent is the monarch." The usual line is to
// order the monarch trigger to resolve first. CR 610.3b covers the
// other order: if an opponent BECOMES the monarch after the exile
// ability triggered and before it resolves, the creature is not exiled
// at all.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "180eda7c-fca2-403b-85cd-8ffebaf9f408",
		Name:         "Palace Jailer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouBecomeTheMonarch("Palace Jailer"),
			Targeting(WhenThisEnters(palaceJailerExileLabel, palaceJailerExile),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}

const palaceJailerExileLabel = "Palace Jailer — exile target creature an opponent controls until an opponent becomes the monarch"

// palaceJailerExile exiles the chosen creature until an opponent of
// the ability's controller becomes the monarch.
func palaceJailerExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return ExileUntil{
			Target:    t.ID,
			On:        []game.EventKind{game.EventMonarchChanged},
			Condition: anOpponentBecameTheMonarchCondition,
			Label:     "Palace Jailer — the exiled creature returns when an opponent becomes the monarch",
		}.Apply(ctx)
	}
	return nil
}
