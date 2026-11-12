package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Iron-Craw Crusher — Artifact Creature — Wurm {7}, 4/6 (The Brothers'
// War, #5656): "Prototype {2}{G}{G} — 2/5", "Whenever this creature
// attacks, target attacking creature gets +X/+0 until end of turn, where
// X is this creature's power."
//
// Prototype is the card's second face (CR 718, game/prototype.go), built
// by the deck importer from the inset; nothing here spells it out.
// X is the Crusher's power as the trigger resolves (CR 608.2h), so a
// prototyped Crusher gives +2/+0 and one cast normally +4/+0; it may
// target itself.
//
// No simplification.
const ironCrawCrusherLabel = "Iron-Craw Crusher — target attacking creature gets +X/+0"

func init() {
	Register(Spec{
		OracleID:     "b04f45de-cb35-45a4-b039-0bf008c3ca56",
		Name:         "Iron-Craw Crusher",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisAttacks(ironCrawCrusherLabel, ironCrawCrusherPump),
				TargetCreature("target attacking creature", AttackingCreature())),
		},
	})
}

// ironCrawCrusherPump reads the Crusher's power as the trigger resolves:
// the permanent as it is now, or as it last existed if it has left. A
// power of zero or less, or an illegal target, does nothing.
func ironCrawCrusherPump(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	info, ok := ctx.SourcePermanent()
	if !ok || info.Power <= 0 {
		return nil
	}
	return vsPumpTheTarget(ctx, info.Power, 0, ironCrawCrusherLabel)
}
