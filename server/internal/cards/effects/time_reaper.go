package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Time Reaper — Creature — Alien Horror {3}{B}{B}, 4/4 (Doctor Who, #5948):
//
//	"Flying, haste
//	 Consume Anomaly — Whenever this creature deals combat damage to a player,
//	 put target face-up card they own in exile on the bottom of their library.
//	 If you do, you gain 3 life."
//
// The target clause is built per trigger (TargetsFrom) from the damaged
// player: a face-up exiled card that player OWNS, in any exile pile. The card
// goes to the bottom of its owner's library and the life is gained from the
// move's keyed tail, so a card that was no longer there (illegal target, CR
// 608.2b) or a move that did not happen gains nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a3f9126f-5797-4436-ac5c-c6911d22ffc2",
		Name:            "Time Reaper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			func() game.TriggeredAbility {
				t := WheneverThisDealsCombatDamageToAPlayer("Time Reaper — Consume Anomaly: put target face-up card they own in exile on the bottom of their library; if you do, you gain 3 life",
					timeReaperConsume)
				t.TargetsFrom = func(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
					return TargetCardInExile("target face-up card they own in exile", FaceUp(), OwnedBy(tc.Event.Target))
				}
				return t
			}(),
		},
	})
}

func timeReaperConsume(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return g.TuckCardsToLibraryKeyedForEffect([]uuid.UUID{id}, game.TuckOptions{ToBottom: true}, ctx.Tail(timeReaperGainLife))
}

// timeReaperGainLife is "if you do, you gain 3 life", told the cards that
// reached the library. An on-disk key: never renamed, never reused.
var timeReaperGainLife = CardTail("tail/time-reaper-gain-life", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return GainLife{Player: ctx.Controller(), Amount: 3}.Apply(ctx)
})
