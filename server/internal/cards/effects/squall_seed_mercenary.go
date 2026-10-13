package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Squall, SeeD Mercenary — Legendary Creature — Human Knight Mercenary
// {2}{W}{B}, 3/4:
//
//	"Rough Divide — Whenever a creature you control attacks alone, it
//	 gains double strike until end of turn.
//	 Whenever Squall deals combat damage to a player, return target
//	 permanent card with mana value 3 or less from your graveyard to the
//	 battlefield."
//
// Rough Divide is exalted's shape without the counters: "alone" is
// game.AttackedAlone, read when the attack trigger is harvested, which is
// after every attacker in the declaration is stamped, so a lone
// declaration is genuinely alone (the Bilbo's Ring note). The creature
// that gets double strike is the one that attacked, read off the
// triggering event, and it keeps the keyword until the turn ends. A
// creature removed in response gets nothing.
//
// The second ability is mandatory and targeted: the target is chosen
// when the trigger goes on the stack, and a graveyard with no permanent
// card of mana value 3 or less puts no trigger on the stack at all
// (CR 603.3d). It fires once per combat damage event, so a double-striking
// Squall returns a second card from the second damage step, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e2348e0e-eeca-47b5-9417-bf4470ec29bb",
		Name:         "Squall, SeeD Mercenary",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller) && game.AttackedAlone(g)
			}, "Squall — Rough Divide: the lone attacker gains double strike", squallRoughDivide),
			Targeting(
				On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return ev.Source == source.InstanceID && combatDamageToPlayerBy(ev, source.Controller, g)
				}, "Squall — return a permanent card with mana value 3 or less from your graveyard", b26ReturnFirstLegalGraveyardTargetToBattlefield),
				TargetCardInGraveyard("target permanent card with mana value 3 or less in your graveyard",
					YouOwn(), Permanent(), ManaValueLE(3))),
		},
	})
}

func squallRoughDivide(g *game.Game, item *game.StackItem) error {
	attacker := item.Trigger.Event.CardID
	if !onBattlefield(g, attacker) {
		return nil
	}
	return GrantKeywordUntilEOT{
		Target:   attacker,
		Keywords: []string{"double strike"},
		Label:    "Squall — Rough Divide",
	}.Apply(NewContext(g, item))
}
