package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Biolume Egg // Biolume Serpent — {2}{U} Creature — Serpent Egg 0/4 //
// Creature — Serpent 4/4 (Innistrad: Crimson Vow, #5495):
//
//	Front: "Defender
//	        When this creature enters, scry 2.
//	        When you sacrifice this creature, return it to the battlefield
//	        transformed under its owner's control at the beginning of the
//	        next end step."
//	Back:  "Sacrifice two Islands: This creature can't be blocked this
//	        turn."
//
// The sacrifice trigger is Ordeal of Nylea's (EventSacrifice, fired while
// the Egg is still on the battlefield, for a sacrifice by its controller),
// and it creates a delayed trigger pinned to the card in the graveyard as
// the object it is then (CR 400.7): if the card has left the graveyard by
// the next end step nothing returns. The return is the back face up
// (CR 712.14a). The Serpent's cost sacrifices two Islands, itself
// included if it is one (it is not).
//
// No simplification.
func init() {
	const oracle = "7af3876e-5297-46e6-936e-6c8344f4b6ac"
	Register(Spec{
		OracleID:        oracle,
		Name:            "Biolume Egg",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Purpose:         game.Purpose{Scry: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Biolume Egg — scry 2", Do(Scry{N: 2})),
			On(game.EventSacrifice, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID && ev.Actor == source.Controller
			}, "Biolume Egg — return it to the battlefield transformed at the beginning of the next end step",
				biolumeEggSchedule),
		},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Biolume Serpent",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice two Islands: This creature can't be blocked this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    SacrificeN(2, "two Islands", HasSubtype("Island")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return RestrictUntilEOT{Target: ctx.Source(), Restrictions: game.CantBeBlocked, Label: "Biolume Serpent — can't be blocked"}.Apply(ctx)
			},
		}},
	})
}

// biolumeEggReturnBody is the delayed trigger's "return it to the
// battlefield transformed under its owner's control". An on-disk key.
var biolumeEggReturnBody = game.DelayedBody("delayed/biolume-egg-return-transformed",
	func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		if !objectStillInAGraveyard(g, p.Object) {
			return nil
		}
		return ReturnFromGraveyard{Target: p.Object.ID, Dest: game.ZoneBattlefield, Controller: uuid.Nil, Transformed: true}.Apply(NewContext(g, item))
	})

// biolumeEggSchedule follows the sacrificed card into the graveyard as it
// is as the trigger resolves, and does nothing if it is not there.
func biolumeEggSchedule(g *game.Game, item *game.StackItem) error {
	id := item.SourceCardID
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	return ScheduleDelayedTrigger{
		Label:  "Biolume Egg — return it to the battlefield transformed",
		Cards:  []uuid.UUID{id},
		Body:   biolumeEggReturnBody,
		Params: game.EffectParams{Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}},
	}.Apply(NewContext(g, item))
}
