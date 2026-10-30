package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Strategic Intervention — Enchantment {1}{W} (Marvel Super Heroes
// Commander, #4455):
//
//	"Whenever a creature you control attacks alone, it gets +1/+1 until end
//	 of turn. Tap up to one target creature defending player controls."
//
// Exalted's trigger shape (mshB1YouAttackAlone: game.AttackedAlone is read
// when the attack trigger is harvested, after every attacker is stamped)
// with an "up to one" target. The defending player is the one the attacking
// creature is attacking (TargetOfDefendingPlayer, read off the ATTACKER's
// attack, not the enchantment's), and no target is legal. The pump goes to
// the attacker if it is still on the battlefield, and the tap is a separate
// instruction, so a target that became illegal leaves the pump alone and a
// departed attacker leaves the tap alone (CR 608.2b).
//
// No simplification.
func init() {
	const label = "Strategic Intervention — the attacker gets +1/+1 until end of turn; tap up to one target creature defending player controls"
	trigger := On(game.EventAttack, mshB1YouAttackAlone, label,
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			if onBattlefield(g, item.Trigger.Event.CardID) {
				if err := (BoostUntilEOT{Target: item.Trigger.Event.CardID, Power: 1, Toughness: 1, Label: label}).Apply(ctx); err != nil {
					return err
				}
			}
			return b36TapChosenCreature(g, item)
		})
	trigger.TargetsFrom = upToOne(func(tc game.TriggerContext, _ *game.Card, g *game.Game) *game.TargetSpec {
		attacker, ok := g.LookupCardForEffect(tc.Event.CardID)
		if !ok {
			return nil
		}
		return TargetOfDefendingPlayer("creature", "target creature defending player controls", Creature())(tc, &attacker, g)
	})
	Register(Spec{
		OracleID:     "b8d019fa-3893-453e-a46d-aa53fc3f2f0e",
		Name:         "Strategic Intervention",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{trigger},
	})
}
