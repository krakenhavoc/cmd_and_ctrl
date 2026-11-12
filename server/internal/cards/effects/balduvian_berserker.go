package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Balduvian Berserker — Creature — Kor Berserker {2}{R}, 1/3 (Dominaria
// United, #5487, seam #5658):
//
//	"Enlist (As this creature attacks, you may tap a nonattacking creature
//	 you control without summoning sickness. When you do, add its power to
//	 this creature's until end of turn.)
//	 When this creature dies, it deals damage equal to its power to any
//	 target."
//
// Enlist is the engine's keyword (CR 702.154, game/enlist.go): the creature
// to tap is chosen with the attack declaration, tapped as the declaration
// locks in, and the linked trigger adds its power as it resolves.
// The dies trigger is Infernal Phantom's (eerieDiesWithPower): the target
// is chosen as the trigger goes on the stack, and the damage is the power
// the Berserker had as it died (CR 603.10), so an enlist's +X/+0 from this
// turn is in it. The damage is dealt by that last-known object.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3a7b1d1c-1724-4208-943b-650a5d773653",
		Name:            "Balduvian Berserker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordEnlist},
		Triggered: []game.TriggeredAbility{
			eerieDiesWithPower("Balduvian Berserker — it deals damage equal to its power to any target", TargetAny(), balduvianBerserkerDamage),
		},
	})
}

// balduvianBerserkerDamage deals the Berserker's last-known power (the
// item's Amount) to the trigger's target.
func balduvianBerserkerDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
	ts := ctx.LegalTargets()
	if !ok || len(ts) == 0 {
		return nil
	}
	return DealDamage{SourceObject: &ref, Target: ts[0].ID, Amount: item.Params.Amount}.Apply(ctx)
}
