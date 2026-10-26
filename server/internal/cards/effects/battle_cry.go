package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// battle_cry.go — Battle cry (CR 702.91, #4307).
//
//	702.91a Battle cry is a triggered ability. "Battle cry" means
//	        "Whenever this creature attacks, each other attacking
//	        creature gets +1/+0 until end of turn."
//	702.91b If a creature has multiple instances of battle cry, each
//	        triggers separately.
//
// Hero of Bladehold spells the same trigger out; BattleCry is the
// keyword as a row, for an ability given to a creature as a bundle (an
// ability sticker, ADR 0208: Space Fungus Snickerdoodle's) or printed by
// the next card that has it. The set of "each other attacking creature" is
// fixed as the trigger resolves (CR 611.2c), so a creature put onto the
// battlefield attacking afterwards gets nothing.

// BattleCry is the keyword's triggered ability for the card named `name`.
func BattleCry(name string) game.TriggeredAbility {
	label := name + " — battle cry: each other attacking creature gets +1/+0 until end of turn"
	return WheneverThisAttacks(label, func(g *game.Game, item *game.StackItem) error {
		return BoostUntilEOT{
			Match: And(AttackingCreature(), OtherThan(item.SourceCardID)),
			Power: 1,
			Label: name + " — battle cry",
		}.Apply(NewContext(g, item))
	})
}
