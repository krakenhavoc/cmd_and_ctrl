package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Suki, Courageous Rescuer — Legendary Creature — Human Warrior Ally
// {1}{W}{W}, 2/4 (Avatar: The Last Airbender, #4459):
//
//	"Other creatures you control get +1/+0.
//	 Whenever another permanent you control leaves the battlefield during
//	 your turn, create a 1/1 white Ally creature token. This ability
//	 triggers only once each turn."
//
// The anthem is Sarcophagus's TribalAnthem over every other creature you
// control. The trigger watches every departure — dies, exiled, bounced,
// sacrificed — of another permanent that was yours as it left (the
// last-known controller, CR 603.10a), and only while it is your turn.
// "Only once each turn" is TriggersOnlyOnceEachTurn, the per-object tally
// (a new Suki is a new tally, CR 400.7), with OncePerBatch so two
// permanents leaving at once do not both get past the tally before the
// first trigger is put on the stack (Battle-Scarred Survivalist's pair).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "70ea1a9d-ec52-4c11-8c4d-0d957e5491ad",
		Name:         "Suki, Courageous Rescuer",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{TribalAnthem(TribeFilter{Others: true, YoursOnly: true}, 1, 0)},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(TriggersOnlyOnceEachTurn(On(game.EventLTB, sukiAnotherPermanentLeftOnYourTurn,
				"Suki, Courageous Rescuer — create a 1/1 white Ally creature token",
				Do(CreateToken{Template: TokenCard("1/1 white Ally"), N: 1})))),
		},
	})
}

// sukiAnotherPermanentLeftOnYourTurn is "another permanent you control
// leaves the battlefield during your turn".
func sukiAnotherPermanentLeftOnYourTurn(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventLTB || ev.CardID == uuid.Nil || ev.CardID == source.InstanceID {
		return false
	}
	if !IsYourTurn(g, source.Controller) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && leftUnderControlOf(ev, c) == source.Controller
}
