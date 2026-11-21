package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Markov Enforcer — Creature — Vampire Soldier {4}{R}{R}, 6/6 (Innistrad:
// Crimson Vow Commander, #5495):
//
//	"Whenever this creature or another Vampire you control enters, this
//	 creature fights up to one target creature an opponent controls.
//	 Whenever a creature dealt damage by this creature this turn dies,
//	 create a Blood token."
//
// Two triggers. The fight is Surly Badgersaur's (this creature, if it is
// still on the battlefield, fights the one announced target while both are
// legal, CR 701.14b); it fires for the Enforcer's own arrival and for every
// other Vampire that enters under your control, the Vampire word read
// through the source's text. The Blood trigger is Zurgo Helmsmasher's
// record of who this creature damaged this turn (each object as it is now
// or last was, CR 400.7; prevented damage was never dealt), so a creature the
// Enforcer's fight or combat damage hurt makes a Blood token when it dies.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e262ea6f-c8c7-45c7-a9a6-29dcbe6300a4",
		Name:         "Markov Enforcer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventETB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
					return ok && (c.InstanceID == source.InstanceID || c.HasSubtype(source.TextSubtype("Vampire")))
				},
				Targets: effects.TargetCreature("up to one target creature an opponent controls", effects.OpponentControls()).WithCount(0, 1),
				Key:     "Markov Enforcer — fights up to one target creature an opponent controls",
				Effect:  effects.B30SourceFightsFirstLegalTarget,
			},
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, died := effects.DiedCreature(ev, g)
				return died && ev.CardID != source.InstanceID && effects.DealtDamageByThisTurn(g, source.InstanceID, ev.CardID)
			}, "Markov Enforcer — create a Blood token", effects.Do(effects.CreateToken{Template: effects.BloodToken(), N: 1})),
		},
	})
}
