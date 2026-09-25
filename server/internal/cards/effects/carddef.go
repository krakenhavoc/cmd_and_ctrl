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

// fileDef is the one writer of `defs`: it stamps every triggered row of
// the definition with its catalog identity (the key and its index —
// game.IdentifyCatalogRows, ADR 0041 P9) and files it. A row the engine
// can name is a row whose waiting stack item can be restored.
func fileDef(key string, d *game.CardDef) {
	game.IdentifyCatalogRows(key, d)
	defs[key] = d
}

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
			Label:         a.Label,
			Cost:          a.Cost,
			Targets:       a.Targets,
			Modes:         a.Modes,
			SorcerySpeed:  a.SorcerySpeed,
			Zones:         a.Zones,
			Cycling:       a.Cycling,
			Equip:         a.Equip,
			Condition:     a.Condition,
			ActiveWhen:    a.ActiveWhen,
			Exhaust:       a.Exhaust,
			CostModifiers: a.CostModifiers,
			Uncopyable:    a.Uncopyable,
			Effect:        a.Effect,
		}
	}
	return out
}

// manaShapes projects declared mana abilities into the engine's
// shapes. Shared by buildDef and by buildGrantDef, because a granted
// mana ability (ADR 0093 — Cryptolith Rite's "{T}: Add one mana of any
// color") is declared exactly as a card's own is. Nil in, nil out.
func manaShapes(in []ManaAbility) []game.ManaAbilityShape {
	if len(in) == 0 {
		return nil
	}
	out := make([]game.ManaAbilityShape, len(in))
	for i, a := range in {
		out[i] = game.ManaAbilityShape{
			Zones:                     a.Zones,
			ExileSelf:                 a.Cost.ExileSelf,
			TapCost:                   a.Cost.Tap,
			SacrificeCost:             a.Cost.Sacrifice,
			SacrificeOther:            a.Cost.SacrificeOther,
			TapOthers:                 a.Cost.TapOthers,
			LifeCost:                  a.Cost.Life,
			ManaCost:                  a.Cost.Mana,
			RemoveCounters:            a.Cost.RemoveCounters,
			AddCounter:                a.Cost.AddCounter,
			DiscardCards:              a.Cost.DiscardCards,
			ExileCards:                a.Cost.ExileCards,
			Produced:                  a.Produced,
			Label:                     a.Label,
			Exhaust:                   a.Exhaust,
			Rider:                     a.Rider,
			PreRider:                  a.PreRider,
			NarrowToCommanderIdentity: a.NarrowToCommanderIdentity,
			Condition:                 a.Condition,
			ProducedFunc:              a.ProducedFunc,
			ProducedForPaid:           a.ProducedForPaid,
			DerivesFromOtherSources:   a.DerivesFromOtherSources,
			DerivedMatch:              a.DerivedMatch,
			DerivedColorsOnly:         a.DerivedColorsOnly,
			Restrictions:              a.Restrictions,
			RestrictionsFunc:          a.RestrictionsFunc,
			SpendRiders:               a.SpendRiders,
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
		BlockRules:                 spec.BlockRules,
		AttackLimits:               spec.AttackLimits,
		CastableZones:              spec.CastableZones,
		SpecialActions:             spec.SpecialActions,
		SpecialActionGrants:        spec.SpecialActionGrants,
		UntapStep:                  spec.UntapStep,
		UntapStepRestrictions:      spec.UntapStepRestrictions,
		UntapCaps:                  spec.UntapCaps,
		UntapOptOuts:               spec.UntapOptOuts,
		DrawStep:                   spec.DrawStep,
		CantBeCountered:            spec.CantBeCountered,
		NoMaxHandSize:              spec.NoMaxHandSize,
		PlayerKeywords:             spec.PlayerKeywords,
		PlayerLifeTotalLocked:      spec.PlayerLifeTotalLocked,
		GameEndGates:               spec.GameEndGates,
		WantsDistinctColors:        spec.WantsDistinctColors,
		WantsManaFrom:              spec.WantsManaFrom,
		AdditionalLandPlays:        spec.AdditionalLandPlays,
		XMatters:                   spec.XMatters,
		CastPermissions:            standingCastPermissions(spec.CastPermissions),
		GatedCastPermissions:       gatedStandingCastPermissions(spec.GatedCastPermissions),
		CastTimings:                spec.CastTimings,
		LibraryTopVisible:          spec.LibraryTopVisible,
		CastCondition:              spec.CastCondition,
		CastConditionLabel:         spec.CastConditionLabel,
		CastRestrictions:           spec.CastRestrictions,
		ActivationRestrictions:     spec.ActivationRestrictions,
		ActivationTimings:          spec.ActivationTimings,
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
	// ADR 0089 / CR 702.174a–b: gift is a cost and a gift, and both
	// belong to the KEYWORD. Grown here from the one declaration for
	// the reason madness's pair is. The cost goes LAST so any printed
	// optional cost keeps its index; the entry trigger is installed on
	// every gift card and fires only for a permanent whose record says
	// the gift was promised, and the resolve wrapper stands aside for a
	// permanent spell — so neither half needs to know the card's type.
	if gift := spec.Gift; gift != nil {
		d.OptionalCosts = append(append([]game.AdditionalCost(nil), d.OptionalCosts...), gift.cost())
		d.Triggered = append(append([]game.TriggeredAbility(nil), d.Triggered...), gift.entryTrigger(spec.Name))
	}
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
	// CR 702.174j: an instant or sorcery's gift happens before any of
	// its other spell abilities.
	if spec.Gift != nil {
		d.Resolve = spec.Gift.wrapResolve(spec.OnResolve)
	}
	if asEnters := spec.AsEnters; asEnters != nil {
		d.AsEnters = liveCardHook(asEnters)
	}
	if asTransforms := spec.AsTransformsInto; asTransforms != nil {
		d.AsTransformsInto = liveCardHook(asTransforms)
	}
	d.Activated = activatedShapes(spec.Activated)
	d.ManaAbilities = manaShapes(spec.ManaAbilities)
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

// liveCardHook adapts a card-side "as this …" clause (AsEnters,
// AsTransformsInto) to the engine's (game, card ID) hook. It finds the
// live card so the clause reads the current state (post-move, or
// post-transform) rather than a copy captured before it.
func liveCardHook(hook func(card *game.Card, ctx *Context) error) func(g *game.Game, cardID uuid.UUID) error {
	return func(g *game.Game, cardID uuid.UUID) error {
		zone := g.FindCardZoneForEffect(cardID)
		if zone == nil {
			return nil
		}
		for i := range zone.Cards {
			if zone.Cards[i].InstanceID == cardID {
				return hook(&zone.Cards[i], NewContext(g, nil))
			}
		}
		return nil
	}
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

// gatedStandingCastPermissions is standingCastPermissions for a
// GATED entry (#1314): the same Scope/Duration normalisation, applied
// to the embedded game.CastPermission of each game.CastPermissionGate
// rather than to the gate wrapper itself, which carries no Scope or
// Duration of its own.
func gatedStandingCastPermissions(in []game.CastPermissionGate) []game.CastPermissionGate {
	if len(in) == 0 {
		return nil
	}
	out := make([]game.CastPermissionGate, len(in))
	copy(out, in)
	for i := range out {
		out[i].Permission.Scope = game.ScopeStanding
		out[i].Permission.Duration = game.WhileInZoneDuration()
	}
	return out
}
