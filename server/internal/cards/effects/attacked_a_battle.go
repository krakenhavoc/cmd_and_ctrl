package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attacked_a_battle.go — the three ways a card asks about an attack on
// a battle (#5649, March of the Machine):
//
//   - "Whenever this creature attacks a battle" (WheneverThisAttacksABattle):
//     the creature was declared as an attacker attacking a battle (CR
//     508.3a). A creature put onto the battlefield attacking one never
//     "attacked" (CR 508.4), and one reselected onto a battle did not
//     attack a second time (CR 508.7a); neither announces an EventAttack.
//   - "If it's attacking a battle" (isAttackingABattle): what the
//     creature is attacking right now, so a reselect counts and a battle
//     removed from combat does not (CR 506.4c).
//   - "As long as it attacked a battle this turn"
//     (SelfKeywordWhileItAttackedABattleThisTurn): the turn's attack
//     records (game.AttackedBattleThisTurn), true from the declaration
//     until the turn ends whatever became of the battle.

// ThisAttackedABattle — "this creature attacks a battle": the source was
// declared as an attacker and what it was declared at is a battle.
func ThisAttackedABattle(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return attackDeclared(ev, source) && g.ClassifyAttackTargetForEffect(ev.Target) == game.AttackTargetBattle
}

// WheneverThisAttacksABattle — "Whenever this creature attacks a battle".
func WheneverThisAttacksABattle(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventAttack, ThisAttackedABattle, label, effect)
}

// isAttackingABattle reports whether `c` is attacking a battle now.
func isAttackingABattle(g *game.Game, c game.Card) bool {
	return g.ClassifyAttackTargetForEffect(c.AttackingTarget) == game.AttackTargetBattle
}

// SelfKeywordWhileItAttackedABattleThisTurn is "This creature has
// <keyword> as long as it attacked a battle this turn" (War Historian): a
// layer 6 grant read from the turn's attack records. DependsOnEvents
// wakes the layer cache as the attack is announced, and the turn's own
// advance clears it.
func SelfKeywordWhileItAttackedABattleThisTurn(keyword string) game.StaticAbility {
	ab := KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return selfCreature(target, g, source) && g.AttackedBattleThisTurn(source.InstanceID)
	}, keyword)
	ab.Label = "Has " + keyword + " as long as it attacked a battle this turn."
	ab.DependsOnEvents = []game.EventKind{game.EventAttack}
	return ab
}
