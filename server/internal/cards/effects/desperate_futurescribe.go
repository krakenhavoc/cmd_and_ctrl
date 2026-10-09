package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Desperate Futurescribe — Creature — Kor Scout {2}{W}{U}, 3/4:
//
//	"Flying
//	 At the beginning of combat on your turn, another target creature
//	 you control gets +1/+1 until end of turn. If you've scried or
//	 surveilled this turn, put a +1/+1 counter on that creature
//	 instead."
//
// A targeted beginning-of-combat trigger (the target is chosen as it
// goes on the stack, and removed if you control no other creature,
// CR 603.3d). The "instead" is a clause of the effect, read as the
// trigger RESOLVES: a scry or surveil done in response upgrades the
// pump to a counter. "Scried or surveilled" reads this turn's
// EventScry / EventSurveil by the controller; neither is emitted for a
// scry that looked at an empty library, because none happened.
//
// No simplification.
func init() {
	combat := AtBeginningOfYourCombat("Desperate Futurescribe — another target creature you control gets +1/+1",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if rfCreatureBScriedOrSurveilledThisTurn(g, item.Controller) {
					return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
				}
				return BoostUntilEOT{
					Target: t.ID, Power: 1, Toughness: 1,
					Label: "Desperate Futurescribe — +1/+1",
				}.Apply(ctx)
			}
			return nil
		})
	combat.TargetsFrom = AnotherTarget(func(other CardPredicate) *game.TargetSpec {
		return TargetCreature("another target creature you control", YouControl(), other)
	})
	Register(Spec{
		OracleID:        "83c12f50-e5a1-479d-afab-edb0b40412f2",
		Name:            "Desperate Futurescribe",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{combat},
	})
}
