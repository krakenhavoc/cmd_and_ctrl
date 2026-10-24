package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// carddef.go — builds the engine's view of a Spec once, at Register
// (#622). Everything the twenty-two per-slot hooks used to compute on
// every call is computed here, one time per card: the mana and
// activated ability shapes, the resolve and as-enters wrappers, the
// target-mode hint.

// defs is the engine-facing catalog, keyed exactly as registry is.
var defs = map[string]*game.CardDef{}

// lookupDef is the game.CatalogLookup implementation.
func lookupDef(key string) *game.CardDef { return defs[key] }

// fileDef is the one writer of `defs`: it stamps every triggered row of
// the definition with its catalog identity (the key and its index —
// game.IdentifyCatalogRows, ADR 0041 P9) and files it. A row the engine
// can name is a row whose waiting stack item can be restored. It also
// marks which rows are source-blind (#1968, source_blind.go), on the
// copy IdentifyCatalogRows made, so the card file's slice is untouched.
func fileDef(key string, d *game.CardDef) {
	game.IdentifyCatalogRows(key, d)
	classifySourceBlind(d)
	classifyFootprint(d)
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
			Reconfigure:   a.Reconfigure,
			UntapSelf:     a.UntapSelf,
			Condition:     a.Condition,
			ActiveWhen:    a.ActiveWhen,
			Exhaust:       a.Exhaust,
			Boast:         a.Boast,
			CostModifiers: a.CostModifiers,
			Uncopyable:    a.Uncopyable,
			AnyPlayer:     a.AnyPlayer,
			OpponentsOnly: a.OpponentsOnly,
			OwnerOnly:     a.OwnerOnly,
			GrantorOnly:   a.GrantorOnly,
			Purpose:       a.Purpose,
			MayQuestion:   a.May.question(),
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
			Answers:                   a.Answers,
			TapOthers:                 a.Cost.TapOthers,
			LifeCost:                  a.Cost.Life,
			EnergyCost:                a.Cost.Energy,
			ExertCost:                 a.Cost.Exert,
			ManaCost:                  a.Cost.Mana,
			RemoveCounters:            a.Cost.RemoveCounters,
			AddCounter:                a.Cost.AddCounter,
			DiscardCards:              a.Cost.DiscardCards,
			ExileCards:                a.Cost.ExileCards,
			ExilePermanents:           a.Cost.ExilePermanents,
			Produced:                  a.Produced,
			Label:                     a.Label,
			Exhaust:                   a.Exhaust,
			OncePerTurn:               a.OncePerTurn,
			AnyPlayer:                 a.AnyPlayer,
			Rider:                     manaRider(a),
			RiderSelfDamage:           a.PainToYou,
			RiderSelfLifeGain:         a.GainLife,
			RiderOpponentLifeGain:     a.OpponentsGainLife,
			RiderOpponentDamage:       a.PingOpponents,
			PreRider:                  a.PreRider,
			NarrowToCommanderIdentity: a.NarrowToCommanderIdentity,
			Condition:                 manaAbilityCondition(a),
			ProducedFunc:              foldProduced(a),
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

// manaRider is the built ability's rider: the declared Rider, the
// PainRider a PainToYou declaration stands for (#2392), the
// GainLifeRider a GainLife declaration stands for (#4185), or the
// opponent-facing riders OpponentsGainLife and PingOpponents stand for
// (#4189). Register refuses a spec that sets two, so at most one is
// non-zero here.
func manaRider(a ManaAbility) func(g *game.Game, controller, source uuid.UUID) error {
	if a.PainToYou > 0 {
		return PainRider(a.PainToYou)
	}
	if a.GainLife > 0 {
		return GainLifeRider(a.GainLife)
	}
	if a.OpponentsGainLife > 0 {
		return EachOpponentGainsLifeRider(a.OpponentsGainLife)
	}
	if a.PingOpponents > 0 {
		return PingEachOpponentRider(a.PingOpponents)
	}
	return a.Rider
}

// GainLifeRider is the mana-ability rider "You gain N life": the
// ability's controller gains n life once the mana is in the pool
// (CR 605.3b). Declare it with ManaAbility.GainLife, which also tells
// the auto-tapper what it does.
func GainLifeRider(n int) func(g *game.Game, controller, source uuid.UUID) error {
	return func(g *game.Game, controller, source uuid.UUID) error {
		return g.ChangePlayerLifeForEffect(source, controller, n)
	}
}

// manaAbilityCondition is the built ability's whole activation gate:
// the declared Condition, and — for a mana ability that sets
// OncePerTurn — "this object has not activated the ability with this
// label this turn" (#1621).
//
// The declaration and its enforcement are one statement on purpose.
// game.ManaAbilityShape.OncePerTurn is a bit the AUTO-TAPPER reads to
// know a costless ability is bounded; every other reader honours the
// rule through Condition, which it already asks. Setting the bit here
// and the gate here is what keeps "the planner thinks it is bounded"
// and "the engine bounds it" from being two facts that can drift.
//
// The count is the per-turn half of the activation record
// (game.ActivatedThisTurn), keyed by the OBJECT and the LABEL: both
// ActivateManaAbility and the auto-tapper's executor write it for
// every mana activation, before anything is paid.
func manaAbilityCondition(a ManaAbility) ActivationCondition {
	if !a.OncePerTurn {
		return a.Condition
	}
	label := a.Label
	once := func(g *game.Game, _, source uuid.UUID) bool {
		return g.ActivatedThisTurn(source, label) == 0
	}
	if a.Condition == nil {
		return once
	}
	return AllConditions(once, a.Condition)
}

// buildDef projects one Spec into the shape the engine reads.
func buildDef(spec Spec) *game.CardDef {
	d := &game.CardDef{
		StartingLoyalty:                     spec.StartingLoyalty,
		TargetMode:                          spec.TargetMode,
		Targets:                             spec.Targets,
		ReturnedAura:                        spec.ReturnedAura,
		Modes:                               spec.Modes,
		Purpose:                             spec.Purpose,
		Replacements:                        spec.Replacements,
		EntersWithCountersFromCast:          spec.EntersWithCountersFromCast,
		PrintedKeywords:                     spec.PrintedKeywords,
		Triggered:                           spec.Triggered,
		ManaTriggers:                        spec.ManaTriggers,
		TriggerDoublers:                     spec.TriggerDoublers,
		TriggerSuppressors:                  spec.TriggerSuppressors,
		AdditionalCost:                      spec.AdditionalCost,
		OptionalCosts:                       spec.OptionalCosts,
		AlternativeCosts:                    spec.AlternativeCosts,
		TapCost:                             spec.TapCost,
		Delve:                               spec.Delve,
		Splice:                              spec.Splice,
		SpendOnly:                           spec.SpendOnly,
		SpendOnlySources:                    spec.SpendOnlySources,
		SpellsYouCastHaveDelve:              spec.SpellsYouCastHaveDelve,
		SplitSecondGrants:                   spec.SplitSecondGrants,
		GrantedOptionalCosts:                spec.GrantedOptionalCosts,
		GrantedTapCosts:                     spec.GrantedTapCosts,
		CostModifiers:                       spec.CostModifiers,
		SelfCostModifiers:                   spec.SelfCostModifiers,
		ExhaustPermissions:                  spec.ExhaustPermissions,
		BoastLimits:                         spec.BoastLimits,
		AttackTaxes:                         spec.AttackTaxes,
		TargetTaxes:                         spec.TargetTaxes,
		BlockRules:                          spec.BlockRules,
		AttackLimits:                        spec.AttackLimits,
		ExertOnAttack:                       spec.ExertOnAttack,
		HexproofBypasses:                    spec.HexproofBypasses,
		WardSuppressions:                    spec.WardSuppressions,
		CastableZones:                       spec.CastableZones,
		CastableZonesIf:                     spec.CastableZonesIf,
		SpecialActions:                      spec.SpecialActions,
		SpecialActionGrants:                 spec.SpecialActionGrants,
		UntapStep:                           spec.UntapStep,
		UntapStepRestrictions:               spec.UntapStepRestrictions,
		CounterRemovalLocks:                 spec.CounterRemovalLocks,
		UntapCaps:                           spec.UntapCaps,
		UntapOptOuts:                        spec.UntapOptOuts,
		DrawStep:                            spec.DrawStep,
		CantBeCountered:                     spec.CantBeCountered,
		CantBeCopied:                        spec.CantBeCopied,
		FlashWithCleanupSacrifice:           spec.FlashWithCleanupSacrifice,
		CastTimeValues:                      spec.CastTimeValues,
		CantBeCounteredIf:                   spec.CantBeCounteredIf,
		SpellDamageCantBePrevented:          spec.SpellDamageCantBePrevented,
		SpellsCantBeCountered:               spec.SpellsCantBeCountered,
		HandSize:                            handSizeStatics(spec),
		ManaPool:                            spec.ManaPool,
		PlayerKeywords:                      spec.PlayerKeywords,
		PlayerKeywordsIf:                    spec.PlayerKeywordsIf,
		PlayerLifeTotalLocked:               spec.PlayerLifeTotalLocked,
		DamageStaysThroughCleanup:           spec.DamageStaysThroughCleanup,
		LoyaltyTwiceEachTurn:                spec.LoyaltyTwiceEachTurn,
		ActivateAsThoughHaste:               spec.ActivateAsThoughHaste,
		ExtraVote:                           spec.ExtraVote,
		OpponentsFaceVillainousChoicesAgain: spec.OpponentsFaceVillainousChoicesAgain,
		ControlsOpponentsSearches:           spec.ControlsOpponentsSearches,
		DamageCantBePrevented:               spec.DamageCantBePrevented,
		CantGainLife:                        spec.CantGainLife,
		DrawLimits:                          spec.DrawLimits,
		VentureLimits:                       spec.VentureLimits,
		DamageAsThough:                      spec.DamageAsThough,
		AnyColorSpend:                       spec.AnyColorSpend,
		LifeForMana:                         spec.LifeForMana,
		LegendRuleExemptions:                spec.LegendRuleExemptions,
		ZeroLoyaltyExemptions:               spec.ZeroLoyaltyExemptions,
		OpponentEffectProtections:           spec.OpponentEffectProtections,
		GameEndGates:                        spec.GameEndGates,
		WantsDistinctColors:                 spec.WantsDistinctColors,
		ReadsManaSpent:                      spec.ReadsManaSpent,
		WantsManaFrom:                       spec.WantsManaFrom,
		WantsColors:                         spec.WantsColors,
		AdditionalLandPlays:                 spec.AdditionalLandPlays,
		XMatters:                            spec.XMatters,
		MadnessLife:                         spec.MadnessLife,
		XCeiling:                            spec.XCeiling,
		CastPermissions:                     standingCastPermissions(spec.CastPermissions),
		GatedCastPermissions:                gatedStandingCastPermissions(spec.GatedCastPermissions),
		CastTimings:                         spec.CastTimings,
		GrantedAlternativeCosts:             spec.GrantedAlternativeCosts,
		OpeningHand:                         spec.OpeningHand,
		LibraryTopVisible:                   spec.LibraryTopVisible,
		HandsRevealed:                       spec.HandsRevealed,
		AttackOnlyNearestOpponentInChosenDirection: spec.AttackOnlyNearestOpponentInChosenDirection,
		DiscardByOpponent:                          spec.DiscardByOpponent,
		DiscardByOpponentForHand:                   spec.DiscardByOpponentForHand,
		CastCondition:                              spec.CastCondition,
		CastConditionLabel:                         spec.CastConditionLabel,
		CastRestrictions:                           spec.CastRestrictions,
		LandPlayRestrictions:                       spec.LandPlayRestrictions,
		IgnorableEffect:                            spec.IgnorableEffect,
		EntryRestrictions:                          spec.EntryRestrictions,
		TargetingRestrictions:                      spec.TargetingRestrictions,
		ActivationRestrictions:                     spec.ActivationRestrictions,
		ActivationTimings:                          spec.ActivationTimings,

		// Oath of Teferi (#4184), set apart so the block above keeps its
		// alignment.
		YourPlaneswalkersLoyaltyTwiceEachTurn: spec.YourPlaneswalkersLoyaltyTwiceEachTurn,
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
	// #1665 / CR 702.94a: miracle's trigger belongs to the KEYWORD,
	// grown here from the card's Miracle alternative cost for the
	// reason madness's pair is — the cost is the declaration, and the
	// "reveal it as you draw it" half cannot be forgotten.
	for _, ac := range spec.AlternativeCosts {
		if ac.Key == game.AltCostKeyMiracle {
			d.Triggered = append(append([]game.TriggeredAbility(nil), d.Triggered...),
				game.MiracleTrigger(spec.Name, ac.ManaCost))
			break
		}
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
		// #4019: an OnResolve written as steps can be found again from
		// the spell's key, so its later steps park behind an earlier
		// step's prompt (do_walk.go).
		if steps, ok := resolveStepsOf(onResolve); ok && len(steps) > 0 {
			d.ResumeResolve = resumeResolveSteps(steps)
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
	if asEntersOthers := spec.AsEntersOthers; asEntersOthers != nil {
		d.AsEntersOthers = liveCardHook(asEntersOthers)
	}
	if asTransforms := spec.AsTransformsInto; asTransforms != nil {
		d.AsTransformsInto = liveCardHook(asTransforms)
	}
	if asAttached := spec.AsAttached; asAttached != nil {
		d.AsAttached = liveCardHook(asAttached)
	}
	d.Activated = activatedShapes(spec.Activated)
	d.ManaAbilities = manaShapes(spec.ManaAbilities)
	// #4236: the printed keywords are NOT a static. They are the object's
	// own characteristics, so they are in the baseline the layer pass
	// starts from (CR 613.1; printedInputs.characteristicFrom reads
	// CatalogPrintedKeywords), where every removal, whenever it began,
	// takes them away (CR 613.1f, CR 613.7). They used to be re-applied
	// here as a self-only layer-6 row at the creature's own timestamp,
	// which put a keyword back after any older "loses flying".
	d.Static = spec.Static
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

// foldProduced is the shape's single produce hook (#3042): the spec's
// string ProducedFunc, or its count-driven ProducedAmount, as one
// func returning a game.ManaOutput. Nil when the spec declares neither.
func foldProduced(a ManaAbility) func(g *game.Game, controller, source uuid.UUID) game.ManaOutput {
	switch {
	case a.ProducedAmount != nil:
		f := a.ProducedAmount
		return func(g *game.Game, c, s uuid.UUID) game.ManaOutput { return game.ManaOutput{Amount: f(g, c, s)} }
	case a.ProducedFunc != nil:
		f := a.ProducedFunc
		return func(g *game.Game, c, s uuid.UUID) game.ManaOutput { return game.ManaOutput{Text: f(g, c, s)} }
	}
	return nil
}
