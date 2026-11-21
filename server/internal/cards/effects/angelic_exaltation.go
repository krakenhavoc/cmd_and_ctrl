package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Angelic Exaltation — Enchantment {3}{W} (Ravnica Remastered):
//
//	"Whenever a creature you control attacks alone, it gets +X/+X until end
//	 of turn, where X is the number of creatures you control."
//
// Exalted's trigger shape (Agents of S.H.I.E.L.D.): "alone" is
// game.AttackedAlone (CR 506.5), read when the trigger is harvested, and the
// creature that gets the bonus is the one that attacked. X is counted as the
// ability resolves (CR 608.2h). An attacker that has left gets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "507d2f27-08a9-4938-b191-5734cda42ef2",
		Name:         "Angelic Exaltation",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller) && game.AttackedAlone(g)
			}, "Angelic Exaltation — the lone attacker gets +X/+X until end of turn, X the number of creatures you control",
				angelicExaltationPump),
		},
	})
}

func angelicExaltationPump(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	attacker := item.Trigger.Event.CardID
	if !onBattlefield(g, attacker) {
		return nil
	}
	x := countControlled(g, item.Controller, func(c game.Card) bool { return c.IsCreature() })
	return BoostUntilEOT{
		Target:    attacker,
		Power:     x,
		Toughness: x,
		Label:     "Angelic Exaltation — +X/+X",
	}.Apply(NewContext(g, item))
}
