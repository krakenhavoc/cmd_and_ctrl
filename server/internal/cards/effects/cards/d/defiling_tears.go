package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Defiling Tears — Instant {2}{B} (Invasion, S82 OLD #5616):
//
//	"Until end of turn, target creature becomes black, gets +1/-1, and gains
//	 '{B}: Regenerate this creature.'"
//
// Three until-end-of-turn changes to one object (CR 611.2c). It becomes
// black and no other colour (layer 5, the colour-setting mod Spiritmonger
// uses), gets +1/-1 (layer 7c), and gains a regenerate ability as a duration
// grant, as Retraction Helix grants its bounce: the creature's controller
// activates it for {B}. All three end at the cleanup step. The +1/-1 can
// kill a creature with toughness 1 as state-based actions are checked.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0decd8ea-a3d6-436b-81e9-bdd836537a1c",
		Name:         "Defiling Tears",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Grants: []effects.AbilityGrant{{
			Key:       defilingTearsRegenerate,
			Activated: []effects.ActivatedAbility{effects.RegenerateThisRow("{B}")},
			Text:      "{B}: Regenerate this creature.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.ScopedEffectFor{
				Target:   id,
				Mods:     []game.Mod{game.SetColorsMod(ctx.TextColor("B")), game.ModifyPTMod(1, -1)},
				Duration: effects.DurationUntilEndOfTurn(ctx),
				Label:    "Defiling Tears — black, +1/-1",
			}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantAbilitiesFor{
				Target: id,
				Keys:   []string{defilingTearsRegenerate},
				Label:  "Defiling Tears — {B}: Regenerate this creature",
			}.Apply(ctx)
		},
	})
}

const defilingTearsRegenerate = "defiling-tears/regenerate"
