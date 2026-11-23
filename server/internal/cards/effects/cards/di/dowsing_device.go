package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const DowsingDeviceOracleID = "2f4374f6-c695-4a5d-a6d6-0e41eaa587ca"

// Dowsing Device — Artifact {1}{R}, the front of a transforming card whose
// back is Geode Grotto (The Lost Caverns of Ixalan, #4470):
//
//	"Whenever this artifact or another artifact you control enters, up to
//	 one target creature you control gets +1/+0 and gains haste until end
//	 of turn. Then transform this artifact if you control four or more
//	 artifacts."
//
// A trigger on any artifact entering under your control, this one
// included, with an "up to one" target (CR 601.2c: none is a legal
// choice, and the second sentence still happens). The pump goes first
// when the target is still legal (CR 608.2b), then the artifacts are
// counted as the trigger resolves, by their layered types, so an artifact
// creature counts once and the Device itself counts. A Device that has
// left does not transform.
//
// Geode Grotto:
//
//	"{T}: Add {R}.
//	 {2}{R}, {T}: Until end of turn, target creature gains haste and gets
//	 +X/+0, where X is the number of artifacts you control. Activate only
//	 as a sorcery."
//
// X is counted as the ability resolves (CR 608.2b), and the target is
// re-checked then.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     DowsingDeviceOracleID,
		Name:         "Dowsing Device",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
				return ok && c.IsArtifact()
			}, "Dowsing Device — up to one target creature you control gets +1/+0 and gains haste until end of turn; transform it if you control four or more artifacts",
				dowsingDeviceResolve),
				effects.TargetPermanent("up to one target creature you control", effects.Creature(), effects.YouControl()).WithCount(0, 1)),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     DowsingDeviceOracleID + "#1",
		Name:         "Geode Grotto",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "{T}: Add {R}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:        "{2}{R}, {T}: Until end of turn, target creature gains haste and gets +X/+0, where X is the number of artifacts you control. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{2}{R}"), effects.TapCost()),
			SorcerySpeed: true,
			Targets:      effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				x := len(effects.MatchingBattlefield(ctx, effects.And(effects.Artifact(), effects.YouControl())))
				if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"haste"}, Label: "Geode Grotto — haste"}).Apply(ctx); err != nil {
					return err
				}
				return effects.BoostUntilEOT{Target: t.ID, Power: x, Label: "Geode Grotto — +X/+0"}.Apply(ctx)
			},
		}},
	})
}

func dowsingDeviceResolve(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
		if err := (effects.BoostUntilEOT{Target: t.ID, Power: 1, Label: "Dowsing Device — +1/+0"}).Apply(ctx); err != nil {
			return err
		}
		if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"haste"}, Label: "Dowsing Device — haste"}).Apply(ctx); err != nil {
			return err
		}
	}
	if effects.SourceIsNewObject(g, item) || !effects.OnBattlefield(g, ctx.Source()) {
		return nil
	}
	if len(effects.MatchingBattlefield(ctx, effects.And(effects.Artifact(), effects.YouControl()))) < 4 {
		return nil
	}
	return effects.TransformThis{}.Apply(ctx)
}
