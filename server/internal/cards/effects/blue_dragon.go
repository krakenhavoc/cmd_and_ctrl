package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blue Dragon — Creature — Dragon {5}{U}{U}, 5/5 (Adventures in the
// Forgotten Realms):
//
//	"Flying
//	 Lightning Breath — When this creature enters, until your next turn,
//	 target creature an opponent controls gets -3/-0, up to one other
//	 target creature gets -2/-0, and up to one other target creature gets
//	 -1/-0."
//
// One trigger with three target clauses in printed order (CR 115.3): the
// first is required, the other two are "up to one" and each must differ from
// the clauses before it ("other"). Each pick shrinks until its controller's
// next turn begins (CR 611.2b); a pick that is no longer legal is skipped,
// the others still apply (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1d00e54a-5897-43dc-9c38-a0e2ce6a7433",
		Name:            "Blue Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Blue Dragon — until your next turn, target creature an opponent controls gets -3/-0, up to one other target creature gets -2/-0, and up to one other target creature gets -1/-0",
					blueDragonBreath),
				Clauses(
					TargetCreature("target creature an opponent controls", OpponentControls()),
					Distinct(UpToOneTargetCreature("up to one other target creature")),
					Distinct(UpToOneTargetCreature("up to one other target creature")),
				)),
		},
	})
}

func blueDragonBreath(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for slot, shrink := range []int{-3, -2, -1} {
		t, ok := ctx.ClauseTarget(slot)
		if !ok {
			continue
		}
		err := ScopedEffectFor{
			Target:   t.ID,
			Mods:     []game.Mod{game.ModifyPTMod(shrink, 0)},
			Duration: DurationUntilYourNextTurn(ctx, item.Controller),
			Label:    "Blue Dragon — Lightning Breath, until your next turn",
		}.Apply(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}
