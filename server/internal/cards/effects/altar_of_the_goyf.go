package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Altar of the Goyf — Kindred Artifact — Lhurgoyf {5}:
//
//	"Whenever a creature you control attacks alone, it gets +X/+X until
//	 end of turn, where X is the number of card types among cards in all
//	 graveyards.
//	 Lhurgoyf creatures you control have trample."
//
// The attack trigger fires once the whole declaration is known and only
// when exactly one creature was declared as an attacker (Sigil of Valor's
// alone test), a creature its controller controls. X is the Tarmogoyf
// count (CR 205.2a types, a card of two types counting for both) read as
// the trigger resolves (CR 608.2h), and the pump is on the attacker for
// the turn. The trample grant is a Lhurgoyf lord's.
func init() {
	const name = "Altar of the Goyf"
	Register(Spec{
		OracleID:     "51d7bd70-0874-41d4-94bc-c3e14c968668",
		Name:         name,
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Lhurgoyf"}, YoursOnly: true}, "trample"),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !attackDeclaredByYou(ev, source.Controller) || !game.AttackedAlone(g) {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.Controller == source.Controller
			}, name+" — the lone attacker gets +X/+X until end of turn, where X is the number of card types among cards in all graveyards",
				func(g *game.Game, item *game.StackItem) error {
					attacker, ok := neoB09LoneAttacker(g, item)
					if !ok {
						return nil
					}
					x := game.DistinctCardTypesInAllGraveyards(g)
					return BoostUntilEOT{Target: attacker.InstanceID, Power: x, Toughness: x, Label: name + " — +X/+X"}.Apply(NewContext(g, item))
				}),
		},
	})
}
