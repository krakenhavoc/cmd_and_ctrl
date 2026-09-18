package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// carddef.go — builds the engine's view of a Spec once, at Register
// (#622). Everything the twenty-two per-slot hooks used to compute on
// every call is computed here, one time per card: the mana and
// activated ability shapes, the printed-keyword static (closure and
// all), the resolve and as-enters wrappers, the target-mode hint.

// defs is the engine-facing catalog, keyed exactly as registry is.
var defs = map[string]*game.CardDef{}

// lookupDef is the game.CatalogLookup implementation.
func lookupDef(key string) *game.CardDef { return defs[key] }

// buildDef projects one Spec into the shape the engine reads.
func buildDef(spec Spec) *game.CardDef {
	d := &game.CardDef{
		StartingLoyalty:       spec.StartingLoyalty,
		TargetMode:            spec.TargetMode,
		Targets:               spec.Targets,
		Modes:                 spec.Modes,
		Replacements:          spec.Replacements,
		PrintedKeywords:       spec.PrintedKeywords,
		Triggered:             spec.Triggered,
		TriggerDoublers:       spec.TriggerDoublers,
		AdditionalCost:        spec.AdditionalCost,
		AlternativeCosts:      spec.AlternativeCosts,
		TapCost:               spec.TapCost,
		CostModifiers:         spec.CostModifiers,
		SelfCostModifiers:     spec.SelfCostModifiers,
		CastableZones:         spec.CastableZones,
		UntapStep:             spec.UntapStep,
		UntapStepRestrictions: spec.UntapStepRestrictions,
		CantBeCountered:       spec.CantBeCountered,
		NoMaxHandSize:         spec.NoMaxHandSize,
		AdditionalLandPlays:   spec.AdditionalLandPlays,
		XMatters:              spec.XMatters,
	}
	if spec.Battle != nil {
		d.BattleDefense = spec.Battle.Defense
	}
	// S20: a structured TargetSpec is the source of truth for the
	// client hint too.
	if spec.Targets != nil {
		d.TargetMode = spec.Targets.Mode
	}
	if onResolve := spec.OnResolve; onResolve != nil {
		d.Resolve = func(g *game.Game, item *game.StackItem) error {
			return onResolve(item, NewContext(g, item))
		}
	}
	if asEnters := spec.AsEnters; asEnters != nil {
		d.AsEnters = func(g *game.Game, cardID uuid.UUID) error {
			// Find the live card so the hook reads the current state
			// (post-move) rather than a copy captured before it.
			zone := g.FindCardZoneForEffect(cardID)
			if zone == nil {
				return nil
			}
			for i := range zone.Cards {
				if zone.Cards[i].InstanceID == cardID {
					return asEnters(&zone.Cards[i], NewContext(g, nil))
				}
			}
			return nil
		}
	}
	if len(spec.Activated) > 0 {
		d.Activated = make([]game.ActivatedAbilityShape, len(spec.Activated))
		for i, a := range spec.Activated {
			d.Activated[i] = game.ActivatedAbilityShape{
				Label:        a.Label,
				Cost:         a.Cost,
				Targets:      a.Targets,
				Modes:        a.Modes,
				SorcerySpeed: a.SorcerySpeed,
				Condition:    a.Condition,
				Effect:       a.Effect,
			}
		}
	}
	if len(spec.ManaAbilities) > 0 {
		d.ManaAbilities = make([]game.ManaAbilityShape, len(spec.ManaAbilities))
		for i, a := range spec.ManaAbilities {
			d.ManaAbilities[i] = game.ManaAbilityShape{
				TapCost:                   a.Cost.Tap,
				SacrificeCost:             a.Cost.Sacrifice,
				SacrificeOther:            a.Cost.SacrificeOther,
				LifeCost:                  a.Cost.Life,
				ManaCost:                  a.Cost.Mana,
				Produced:                  a.Produced,
				Label:                     a.Label,
				Rider:                     a.Rider,
				NarrowToCommanderIdentity: a.NarrowToCommanderIdentity,
				Condition:                 a.Condition,
				ProducedFunc:              a.ProducedFunc,
				DerivesFromOtherSources:   a.DerivesFromOtherSources,
				Restrictions:              a.Restrictions,
				RestrictionsFunc:          a.RestrictionsFunc,
			}
		}
	}
	// S18 sub-PR 2: the printed keywords become one self-only Layer 6
	// static, appended after the hand-written ones. Built here, once,
	// rather than on every layer recompute.
	d.Static = spec.Static
	if len(spec.PrintedKeywords) > 0 {
		kws := append([]string(nil), spec.PrintedKeywords...)
		synth := game.StaticAbility{
			Layer:     game.Layer6Ability,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				for _, kw := range kws {
					if !keywordSliceContains(c.Abilities, kw) {
						c.Abilities = append(c.Abilities, kw)
					}
				}
			},
		}
		d.Static = append(append([]game.StaticAbility(nil), spec.Static...), synth)
	}
	if len(d.Static) == 0 {
		d.Static = nil
	}
	return d
}
