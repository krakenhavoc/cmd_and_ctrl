package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "b9447f25-f285-4b62-8976-4a49363dfdd7",
		Name:         "Sigil of Valor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return source.IsAttachedTo(ev.CardID) && game.AttackedAlone(g)
			}, "Sigil of Valor — equipped creature gets +1/+1 until end of turn for each other creature you control", sigilOfValorPump),
		},
		Activated: []ActivatedAbility{EquipAbility("{1}")},
	})
}

func sigilOfValorPump(g *game.Game, item *game.StackItem) error {
	attacker := item.Trigger.Event.CardID
	if !onBattlefield(g, attacker) {
		return nil
	}
	n := 0
	for _, id := range permanentsControlledByMatching(g, item.Controller, Creature()) {
		if id != attacker {
			n++
		}
	}
	return BoostUntilEOT{
		Target:    attacker,
		Power:     n,
		Toughness: n,
		Label:     "Sigil of Valor — +1/+1 for each other creature you control",
	}.Apply(NewContext(g, item))
}
