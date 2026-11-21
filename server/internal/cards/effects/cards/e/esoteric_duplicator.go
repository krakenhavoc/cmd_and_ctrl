package cardse

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Esoteric Duplicator — Artifact — Clue {2}{U} (The Big Score, #4468):
//
//	"Whenever you sacrifice this artifact or another artifact, you may pay
//	 {2}. If you do, at the beginning of the next end step, create a token
//	 that's a copy of that artifact.
//	 {2}, Sacrifice this artifact: Draw a card."
//
// The sacrificed artifact is read as it last existed (CR 603.10a, 608.2h):
// a Clue or Treasure token that has already ceased to exist still counts and
// is still copied. Paying {2} is the trigger's own "you may pay" (MayPay);
// the paid half schedules a delayed trigger (CR 603.7) for the next end
// step, carrying the object (with its epoch) as data, and the copy takes the
// artifact's copiable values (CR 707.2): a token's original characteristics,
// no counters, and an X of 0.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0752fe67-9b8a-42c0-90e3-087d426b2f3b",
		Name:         "Esoteric Duplicator",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventSacrifice, b35YouSacrificedAnArtifact,
				"Esoteric Duplicator — you may pay {2}; if you do, at the beginning of the next end step, create a token that's a copy of that artifact",
				func(g *game.Game, item *game.StackItem) error {
					id := effects.NewContext(g, item).Trigger().Event.CardID
					ref, ok := g.PermanentRefForEffect(id)
					if !ok {
						return nil
					}
					return effects.MayPay{
						Chooser:      item.Controller,
						Cost:         "{2}",
						Question:     "Esoteric Duplicator — pay {2} to copy that artifact at the beginning of the next end step?",
						Then:         esotericDuplicatorPaid,
						CarryObjects: []game.ObjectRef{ref},
					}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, Sacrifice this artifact: Draw a card.",
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.SacrificeThis()),
			Effect:  effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}

// b35YouSacrificedAnArtifact is "you sacrifice an artifact": the sacrifice
// event of the source's controller, the permanent read as it last existed on
// the battlefield (a token that has ceased to exist included), else as it
// stands.
func b35YouSacrificedAnArtifact(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventSacrifice || ev.CardID == uuid.Nil || source == nil || ev.Actor != source.Controller {
		return false
	}
	if info, ok := g.LastKnownPermanentForEffect(ev.CardID); ok {
		for _, t := range info.Characteristic.Types {
			if strings.EqualFold(t, "artifact") {
				return true
			}
		}
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsArtifact()
}

// esotericDuplicatorPaid is the paid half: the delayed copy at the next end
// step. An on-disk key.
var esotericDuplicatorPaid = effects.PayUnlessBranches("pay-unless/esoteric-duplicator-copy-at-end-step",
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		if len(r.Objects) == 0 {
			return nil
		}
		return effects.ScheduleDelayedTrigger{
			Label:  "Esoteric Duplicator — create a token that's a copy of that artifact",
			Body:   esotericDuplicatorCopyBody,
			Params: game.EffectParams{Object: r.Objects[0]},
		}.Apply(ctx)
	}, nil)

// esotericDuplicatorCopyBody creates the token copy of the artifact the
// delayed trigger names, as it last existed. An on-disk key.
var esotericDuplicatorCopyBody = game.DelayedBody("esoteric-duplicator/copy-the-artifact",
	func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		tmpl, ok := g.TokenCopyTemplateForObjectForEffect(p.Object)
		if !ok {
			return nil
		}
		return effects.CreateTokenCopies(effects.NewContext(g, item), item.Controller, tmpl, 1)
	})
