package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Canal Courier — Creature — Human Rogue {5}{U}, 3/5 (Conspiracy: Take the
// Crown, S82 older sets batch 272, #5540):
//
//	"When this creature enters, you become the monarch.
//	 Whenever this creature and another creature attack different players,
//	 this creature can't be blocked this combat."
//
// The monarch ETB is the shared one. The second trigger fires on the
// Courier's own attack declaration when, in that declaration, another
// attacking creature is attacking a player other than the one the Courier
// attacks (an attack on a planeswalker or a battle is not an attack on a
// player, CR 506.2); the declaration is one action, so every attacker's
// target is already set when the trigger is checked. The Courier then can't
// be blocked until the end of that combat phase (CR 511.3), as Noctis's
// does, so an additional combat phase is a fresh combat.
//
// No simplification.
func init() {
	const label = "Canal Courier — this creature can't be blocked this combat"
	effects.Register(effects.Spec{
		OracleID:     "efaa1262-b7ac-49ab-831a-d10b59680691",
		Name:         "Canal Courier",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouBecomeTheMonarch("Canal Courier"),
			effects.On(game.EventAttack, canalCourierAttacksADifferentPlayerThanAnother, label,
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					d, ok := effects.DurationUntilEndOfCombat(ctx)
					if !ok || !effects.OnBattlefield(g, item.SourceCardID) {
						return nil
					}
					return effects.ScopedEffectFor{
						Target:   item.SourceCardID,
						Mods:     []game.Mod{game.AddRestrictionsMod(game.CantBeBlocked)},
						Duration: d,
						Label:    label,
					}.Apply(ctx)
				}),
		},
	})
}

// canalCourierAttacksADifferentPlayerThanAnother is the trigger condition:
// the Courier was declared attacking a player, and another attacking
// creature is attacking a different player.
func canalCourierAttacksADifferentPlayerThanAnother(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !effects.AttackDeclared(ev, source) {
		return false
	}
	mine := effects.AttackedPlayer(g, ev)
	if mine == uuid.Nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.InstanceID == source.InstanceID || !c.IsCreature() || c.AttackingTarget == uuid.Nil || c.AttackingTarget == mine {
			continue
		}
		if g.PlayerByIDForEffect(c.AttackingTarget) != nil {
			return true
		}
	}
	return false
}
