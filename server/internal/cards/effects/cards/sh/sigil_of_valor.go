package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sigil of Valor — Artifact — Equipment {2}:
//
//	"Whenever equipped creature attacks alone, it gets +1/+1 until end
//	 of turn for each other creature you control.
//	 Equip {1}"
//
// "Alone" is CR 506.5 (game.AttackedAlone, read when the attack trigger
// is harvested, after the whole declaration). The count is read as the
// ability resolves, over the Sigil's controller's creatures other than
// the attacker.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b9447f25-f285-4b62-8976-4a49363dfdd7",
		Name:         "Sigil of Valor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return source.IsAttachedTo(ev.CardID) && game.AttackedAlone(g)
			}, "Sigil of Valor — equipped creature gets +1/+1 until end of turn for each other creature you control", sigilOfValorPump),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{1}")},
	})
}

func sigilOfValorPump(g *game.Game, item *game.StackItem) error {
	attacker := item.Trigger.Event.CardID
	if !effects.OnBattlefield(g, attacker) {
		return nil
	}
	n := 0
	for _, id := range effects.PermanentsControlledByMatching(g, item.Controller, effects.Creature()) {
		if id != attacker {
			n++
		}
	}
	return effects.BoostUntilEOT{
		Target:    attacker,
		Power:     n,
		Toughness: n,
		Label:     "Sigil of Valor — +1/+1 for each other creature you control",
	}.Apply(effects.NewContext(g, item))
}
