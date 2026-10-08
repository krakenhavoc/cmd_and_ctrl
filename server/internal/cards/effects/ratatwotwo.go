package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ratatwotwo — Legendary Creature — Rat {3}{W}, 2/2 (an Unknown Event
// card, not legal in Commander):
//
//	"When Ratatwotwo enters the battlefield, create a Chef Role token
//	 attached to up to one other target creature you control. (Enchanted
//	 creature gets +1/+1 and has "Whenever this creature attacks, create
//	 a Food token.")
//	 Whenever a creature you control enchanted by a Chef Role attacks or
//	 blocks, it gets +X/+X until end of turn, where X is Ratatwotwo's
//	 power."
//
// The second ability is one trigger with two conditions, once per
// creature that attacks, and once per creature that blocks however many
// attackers it blocks (CR 509.3a). "Enchanted by a Chef Role" means a
// Chef Role is attached to it, whoever controls the Role.
//
// Last-known information (CR 608.2h) for X is not modelled: if Ratatwotwo
// has left the battlefield when the bonus resolves, the creature gets
// nothing rather than the power Ratatwotwo last had.
func init() {
	Register(Spec{
		OracleID:     "0edd7d9a-af37-4d3b-8b4b-913c36dadf82",
		Name:         "Ratatwotwo",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"If Ratatwotwo has left the battlefield by the time the bonus resolves, the creature gets no bonus instead of using Ratatwotwo's last power."},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Ratatwotwo — create a Chef Role token attached to up to one other target creature you control",
					createRoleOnFirstTarget(RoleChef)),
				Another(TargetCreature("up to one other target creature you control", YouControl()).WithCount(0, 1))),
			OnAny([]game.EventKind{game.EventAttack, game.EventBlock}, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller {
					return false
				}
				if ev.Kind == game.EventBlock && ev.Amount > 1 {
					return false
				}
				return enchantedByRoleNamed(g, ev.CardID, "Chef Role")
			}, "Ratatwotwo — the creature gets +X/+X until end of turn, X being Ratatwotwo's power",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					self, ok := g.LookupCardForEffect(item.SourceCardID)
					if !ok {
						return nil
					}
					x := self.CurrentPower()
					return BoostUntilEOT{
						Target: item.Trigger.Event.CardID, Power: x, Toughness: x,
						Label: "Ratatwotwo — +X/+X",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// enchantedByRoleNamed reports whether a Role with the given printed
// name is attached to the permanent.
func enchantedByRoleNamed(g *game.Game, id uuid.UUID, name string) bool {
	if g == nil || g.Battlefield == nil || id == uuid.Nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		a := &g.Battlefield.Cards[i]
		if a.Name == name && a.IsRole() && a.IsAttachedTo(id) {
			return true
		}
	}
	return false
}
