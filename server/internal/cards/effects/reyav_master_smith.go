package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reyav, Master Smith — {R}{W} Legendary Creature — Dwarf Artificer, 2/2:
//
//	"Whenever a creature you control that's enchanted or equipped attacks,
//	 that creature gains double strike until end of turn."
//
// One trigger per attacker: a creature you control (Reyav herself
// included) that has an Aura or an Equipment attached as the attack is
// declared. The condition is read when the attack is declared, so an
// attachment made in response does not retroactively trigger it, and
// the ability resolves on the attacker if it is still on the battlefield
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d1fc2549-40d5-41c8-b3ab-07ea13e9381d",
		Name:         "Reyav, Master Smith",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller) && rprB03IsEnchantedOrEquipped(g, ev.CardID)
			}, "Reyav, Master Smith — the enchanted or equipped attacker gains double strike until end of turn", reyavDoubleStrike),
		},
	})
}

func reyavDoubleStrike(g *game.Game, item *game.StackItem) error {
	attacker := item.Trigger.Event.CardID
	if !onBattlefield(g, attacker) {
		return nil
	}
	return GrantKeywordUntilEOT{
		Target:   attacker,
		Keywords: []string{"double strike"},
		Label:    "Reyav, Master Smith — double strike",
	}.Apply(NewContext(g, item))
}
