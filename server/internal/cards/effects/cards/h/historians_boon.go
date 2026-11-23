package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Historian's Boon — Enchantment {3}{W} (Dominaria United Commander):
//
//	"Whenever this enchantment or another nontoken enchantment you
//	 control enters, create a 1/1 white Soldier creature token.
//	 Whenever the final chapter ability of a Saga you control triggers,
//	 create a 4/4 white Angel creature token with flying and vigilance."
//
// The first ability is the Archon of Redemption shape: one trigger, two
// ways in (its own entry, or another nontoken enchantment under its
// controller's control, judged as it entered). The second watches the
// chapter event itself, which is the moment the chapter ability triggers
// (CR 714.2b): it is a final chapter ability when the chapter number is
// the Saga's final chapter number (CR 714.2d, 714.2e). "A Saga you
// control" is the controller the Saga had as the chapter triggered, which
// the event carries. A read-ahead Saga that skips to its last chapter
// triggers it too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ce1e6ae2-78c1-4bc6-bb73-bb025dd67a90",
		Name:         "Historian's Boon",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, historiansBoonEnchantmentEntered, "Historian's Boon — create a 1/1 white Soldier",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Soldier"), N: 1})),
			effects.On(game.EventSagaChapter, historiansBoonFinalChapter, "Historian's Boon — create a 4/4 white Angel with flying and vigilance",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("4/4 white Angel with flying and vigilance"), N: 1})),
		},
	})
}

// historiansBoonEnchantmentEntered: this enchantment, or another nontoken
// enchantment its controller controls, entering.
func historiansBoonEnchantmentEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind == game.EventETB && ev.CardID == source.InstanceID {
		return true
	}
	c, ok := effects.EnteredUnderYourControl(ev, source, g, true)
	return ok && c.IsEnchantment() && !c.IsToken()
}

// historiansBoonFinalChapter: the chapter that just triggered is the final
// chapter of a Saga the source's controller controlled as it triggered.
func historiansBoonFinalChapter(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventSagaChapter || ev.CardID == uuid.Nil || ev.Actor != source.Controller {
		return false
	}
	saga, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !game.IsSaga(saga) {
		return false
	}
	final := game.SagaFinalChapter(saga)
	return final > 0 && ev.Amount == final
}
