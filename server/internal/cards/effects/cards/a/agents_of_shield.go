package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Agents of S.H.I.E.L.D. — Creature — Human Spy Hero {2}{W}, 2/4 (Marvel
// Super Heroes, #4455):
//
//	"Whenever a creature you control attacks alone, that creature gets
//	 +1/+1 until end of turn."
//
// Squall's Rough Divide shape with a pump: "alone" is game.AttackedAlone,
// read when the trigger is harvested, after every attacker of the
// declaration is stamped, so a lone declaration is genuinely alone. The
// creature that gets the bonus is the one that attacked, read off the
// triggering event; one removed in response gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "203b72d2-a273-49cc-bd10-2fbc07ed2fc8",
		Name:         "Agents of S.H.I.E.L.D.",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AttackDeclaredByYou(ev, source.Controller) && game.AttackedAlone(g)
			}, "Agents of S.H.I.E.L.D. — the lone attacker gets +1/+1 until end of turn", agentsOfShieldPump),
		},
	})
}

func agentsOfShieldPump(g *game.Game, item *game.StackItem) error {
	attacker := item.Trigger.Event.CardID
	if !effects.OnBattlefield(g, attacker) {
		return nil
	}
	return effects.BoostUntilEOT{
		Target:    attacker,
		Power:     1,
		Toughness: 1,
		Label:     "Agents of S.H.I.E.L.D. — +1/+1",
	}.Apply(effects.NewContext(g, item))
}
