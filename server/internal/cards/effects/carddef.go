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

// activatedShapes projects declared activated abilities into the
// engine's shapes. Shared by buildDef and by buildGrantDef, because
// a CR 707.9a granted activated ability is declared exactly as a
// card's own is (ability_grant.go). Nil in, nil out.
func activatedShapes(in []ActivatedAbility) []game.ActivatedAbilityShape {
	if len(in) == 0 {
		return nil
	}
	out := make([]game.ActivatedAbilityShape, len(in))
	for i, a := range in {
		out[i] = game.ActivatedAbilityShape{
			Label:        a.Label,
			Cost:         a.Cost,
			Targets:      a.Targets,
			Modes:        a.Modes,
			SorcerySpeed: a.SorcerySpeed,
			Zones:        a.Zones,
			Cycling:      a.Cycling,
			Condition:    a.Condition,
			ActiveWhen:   a.ActiveWhen,
			Exhaust:      a.Exhaust,
			Effect:       a.Effect,
		}
	}
	return out
}

// buildDef projects one Spec into the shape the engine reads.
func buildDef(spec Spec) *game.CardDef {
	d := &game.CardDef{
		StartingLoyalty:            spec.StartingLoyalty,
		TargetMode:                 spec.TargetMode,
		Targets:                    spec.Targets,
		Modes:                      spec.Modes,
		Replacements:               spec.Replacements,
		EntersWithCountersFromCast: spec.EntersWithCountersFromCast,
		PrintedKeywords:            spec.PrintedKeywords,
		Triggered:                  spec.Triggered,
		ManaTriggers:               spec.ManaTriggers,
		TriggerDoublers:            spec.TriggerDoublers,
		AdditionalCost:             spec.AdditionalCost,
		OptionalCosts:              spec.OptionalCosts,
		AlternativeCosts:           spec.AlternativeCosts,
		TapCost:                    spec.TapCost,
		CostModifiers:              spec.CostModifiers,
		SelfCostModifiers:          spec.SelfCostModifiers,
		ExhaustPermissions:         spec.ExhaustPermissions,
		AttackTaxes:                spec.AttackTaxes,
		CastableZones:              spec.CastableZones,
		SpecialActions:             spec.SpecialActions,
		UntapStep:                  spec.UntapStep,
		UntapStepRestrictions:      spec.UntapStepRestrictions,
		UntapCaps:                  spec.UntapCaps,
		UntapOptOuts:               spec.UntapOptOuts,
		CantBeCountered:            spec.CantBeCountered,
		NoMaxHandSize:              spec.NoMaxHandSize,
		PlayerKeywords:             spec.PlayerKeywords,
		WantsDistinctColors:        spec.WantsDistinctColors,
		AdditionalLandPlays:        spec.AdditionalLandPlays,
		XMatters:                   spec.XMatters,
		CastPermissions:            standingCastPermissions(spec.CastPermissions),
		LibraryTopVisible:          spec.LibraryTopVisible,
		CastCondition:              spec.CastCondition,
		CastConditionLabel:         spec.CastConditionLabel,
		CastRestrictions:           spec.CastRestrictions,
		ActivationRestrictions:     spec.ActivationRestrictions,
	}
	if spec.Battle != nil {
		d.BattleDefense = spec.Battle.Defense
	}
	// #659 / CR 702.62b: suspend's two triggered abilities are the
	// KEYWORD's, not the card's. Grown here from the declaration so
	// three card files cannot spell them three ways and the fourth
	// cannot forget one, exactly as the Cycling constructor stamps
	// its own zone and discard cost.
	//
	// #990 made it two: the upkeep countdown and "when the last time
	// counter is removed". They are separate abilities in print, they
	// go on the stack at different moments and either can be
	// countered without the other — see game/suspend.go's header.
	for _, sa := range spec.SpecialActions {
		if sa.Kind == game.SpecialActionSuspend {
			d.Triggered = append(append([]game.TriggeredAbility(nil), d.Triggered...),
				game.SuspendUpkeepTrigger(), game.SuspendLastCounterTrigger())
			break
		}
	}
	// #657 / CR 702.35a: madness is a replacement and a trigger, and
	// both belong to the KEYWORD. Grown here from the one-string
	// declaration for the reason suspend's pair is, and appended to
	// whatever the card declares itself — Big Game Hunter has an ETB
	// trigger of its own and keeps it.
	if spec.Madness != "" {
		d.Replacements = append(append([]game.ReplacementEffect(nil), d.Replacements...),
			game.MadnessReplacement())
		d.Triggered = append(append([]game.TriggeredAbility(nil), d.Triggered...),
			game.MadnessTrigger(spec.Madness))
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
	d.Activated = activatedShapes(spec.Activated)
	if len(spec.ManaAbilities) > 0 {
		d.ManaAbilities = make([]game.ManaAbilityShape, len(spec.ManaAbilities))
		for i, a := range spec.ManaAbilities {
			d.ManaAbilities[i] = game.ManaAbilityShape{
				TapCost:                   a.Cost.Tap,
				SacrificeCost:             a.Cost.Sacrifice,
				SacrificeOther:            a.Cost.SacrificeOther,
				LifeCost:                  a.Cost.Life,
				ManaCost:                  a.Cost.Mana,
				RemoveCounters:            a.Cost.RemoveCounters,
				AddCounter:                a.Cost.AddCounter,
				Produced:                  a.Produced,
				Label:                     a.Label,
				Exhaust:                   a.Exhaust,
				Rider:                     a.Rider,
				NarrowToCommanderIdentity: a.NarrowToCommanderIdentity,
				Condition:                 a.Condition,
				ProducedFunc:              a.ProducedFunc,
				ProducedForPaid:           a.ProducedForPaid,
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
				appendKeywordsTo(c, kws)
			},
		}
		d.Static = append(append([]game.StaticAbility(nil), spec.Static...), synth)
	}
	if len(d.Static) == 0 {
		d.Static = nil
	}
	return d
}

// standingCastPermissions normalises a Spec's declared permissions
// into the shape the engine derives them in (ADR 0066): a permanent's
// static ability is always a STANDING permission and always lasts for
// as long as the permanent remains.
//
// Forced here rather than checked in Register, because there is no
// other thing a card file could have meant: a Spec slot is read off
// the battlefield, and the two fields it would be setting are the two
// the derivation owns. A per-instance permission does not come from a
// Spec at all.
func standingCastPermissions(in []game.CastPermission) []game.CastPermission {
	if len(in) == 0 {
		return nil
	}
	out := make([]game.CastPermission, len(in))
	copy(out, in)
	for i := range out {
		out[i].Scope = game.ScopeStanding
		out[i].Duration = game.WhileInZoneDuration()
	}
	return out
}
