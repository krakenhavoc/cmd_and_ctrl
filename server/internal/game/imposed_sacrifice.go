package game

import (
	"errors"
	"slices"

	"github.com/google/uuid"
)

// imposed_sacrifice.go — #6205: a NON-MANA additional cost that a
// permanent imposes on OTHER objects' spells and activated abilities.
//
//	Brutal Suppression  "Activated abilities of nontoken Rebels cost an
//	                    additional "Sacrifice a land" to activate."
//	Drought             "Spells cost an additional "Sacrifice a Swamp" to
//	                    cast for each black mana symbol in their mana
//	                    costs. / Activated abilities cost an additional
//	                    "Sacrifice a Swamp" to activate for each black mana
//	                    symbol in their activation costs."
//
// CR 601.2f: "Some effects may increase or reduce the cost to pay … Costs
// may include paying mana, tapping permanents, sacrificing permanents,
// discarding cards, and so on." CR 602.2b carries it to activations: "An
// activated ability's analog to a spell's mana cost (as referenced in rule
// 601.2f) is its activation cost."
//
// CostModifier (cost_modifier.go) prices generic mana and TargetTax
// (target_tax.go) prices life; neither can add a permanent to sacrifice.
//
// # Nothing new about paying
//
// The imposed sacrifice is folded into the object's OWN sacrifice clause
// once, as the cost is determined, and from there it is an ordinary
// sacrifice cost: validated by validateSacrificeCostLocked (or the cast's
// additional-cost validator), paid by payCostSacrificesLocked as one
// simultaneous exit, offered by the legal enumerator, shipped by the view
// as `sacrifice_options`, picked in the client's sacrifice picker. The
// three readers of an activation's cost already share one fold for a
// board-modified loyalty cost (LoyaltyAdjustedCostForEffect, #6148); the
// imposed sacrifice is folded at the same three sites, and a cast's at the
// three sites that read the card's AdditionalCost.
//
// The fold is a set rule (TargetSpec.EachOf, #2526): one entry per imposed
// sacrifice ("a Swamp", "a Swamp", "a land"), plus — when the object
// prints a sacrifice of its own — one entry per permanent THAT clause
// sacrifices, marked SacrificeKind.Head. The picks must fill every entry
// one-to-one, so a Swamp sacrificed to Drought is never also the creature
// Bone Splinters asks for (CR 118.3: one permanent pays one cost). The
// client's picker matches by each group's candidate list, so it needs no
// new idea either.
//
// # What the fold refuses
//
// A clause whose COUNT the player announces or the engine fixes ("any
// number", "X", "all", "half") or whose set carries another rule ("with
// different names", "mana value X") has no one-to-one form to join, so
// an object printing one cannot be cast or activated while an imposed
// sacrifice applies to it (ErrImposedSacrificeUnfoldable). No printed
// card under Drought or Brutal Suppression's reach in the catalog needs
// that; when one does, this is where it gets a second clause.

// ErrImposedSacrificeUnfoldable refuses a cast or activation whose own
// sacrifice clause cannot be joined with an imposed one (see above).
var ErrImposedSacrificeUnfoldable = errors.New("an imposed sacrifice cost can't be combined with this cost")

// ImposedSacrificeQuery is what an imposed sacrifice may look at.
type ImposedSacrificeQuery struct {
	// Game is the game the cost is being determined in. Read-only.
	Game *Game
	// Source is the permanent whose static imposes the cost.
	Source Card
	// Payer is the player casting the spell or activating the ability.
	Payer uuid.UUID
	// Object is the spell being cast, or the source of the ability being
	// activated, as it stands now.
	Object Card
	// Spell is true for a cast, false for an activation.
	Spell bool
}

// ImposedSacrifice is one "<objects> cost an additional "Sacrifice a
// <kind>" to <cast / activate>" static of a permanent on the battlefield.
type ImposedSacrifice struct {
	// Label is the clause as printed, for the move log and the picker.
	Label string
	// Kind is what each sacrifice takes ("a Swamp", "a land"). Data, so
	// the folded clause stays a set rule (TargetSpec.EachOf).
	Kind SacrificeKind
	// Spells and Activations say which costs it adds to. At least one
	// must be set; Register refuses neither.
	Spells      bool
	Activations bool
	// PerBlackSymbol is Drought's "for each black mana symbol in their
	// mana costs" / "in their activation costs": the count is the black
	// symbols (hybrid and Phyrexian symbols that can be paid with black
	// count) in the spell's mana cost or the ability's mana cost. False
	// is one sacrifice.
	PerBlackSymbol bool
	// AppliesTo narrows which objects it applies to ("nontoken Rebels").
	// Nil is every spell / every activated ability.
	AppliesTo func(q ImposedSacrificeQuery) bool
}

// CatalogImposedSacrifices returns the imposed sacrifices a catalog entry
// declares. Wired in carddef.go; nil in a test that builds cards by hand.
var CatalogImposedSacrifices func(key string) []ImposedSacrifice

// imposedSacrificesOf is the imposed sacrifices a permanent contributes
// now: its catalog entry's, minus those an ability-removing effect took
// (CR 613.1f, catalogAbilityKeyOf).
func imposedSacrificesOf(c *Card) []ImposedSacrifice {
	if CatalogImposedSacrifices == nil {
		return nil
	}
	key := catalogAbilityKeyOf(c)
	if key == "" {
		return nil
	}
	return CatalogImposedSacrifices(key)
}

// BlackManaSymbols counts the black mana symbols in a mana cost string:
// {B}, a hybrid symbol that offers black ({B/G}, {2/B}), and a Phyrexian
// black symbol ({B/P}). An unparseable or empty cost has none.
func BlackManaSymbols(cost string) int {
	if cost == "" {
		return 0
	}
	pc, err := ParseCost(cost)
	if err != nil {
		return 0
	}
	n := 0
	for _, req := range pc.Required {
		if slices.Contains(req.Options, "B") {
			n++
		}
	}
	return n
}

// imposedSacrificeKindsLocked lists the sacrifices every permanent on the
// battlefield imposes on `object` (a spell when `spell`, else an ability's
// source) paid by `payer`, whose mana cost — the spell's mana cost, or the
// ability's activation mana — is `mana`. One entry per permanent to
// sacrifice. Nil for nearly every cost. Caller must hold g.mu.
func (g *Game) imposedSacrificeKindsLocked(payer uuid.UUID, object Card, spell bool, mana string) []SacrificeKind {
	if g == nil || g.Battlefield == nil || CatalogImposedSacrifices == nil {
		return nil
	}
	var out []SacrificeKind
	for i := range g.Battlefield.Cards {
		src := &g.Battlefield.Cards[i]
		for _, is := range imposedSacrificesOf(src) {
			if (spell && !is.Spells) || (!spell && !is.Activations) {
				continue
			}
			if is.AppliesTo != nil && !is.AppliesTo(ImposedSacrificeQuery{Game: g, Source: *src, Payer: payer, Object: object, Spell: spell}) {
				continue
			}
			n := 1
			if is.PerBlackSymbol {
				n = BlackManaSymbols(mana)
			}
			for range n {
				out = append(out, is.Kind)
			}
		}
	}
	return out
}

// ImposedSacrificesExistForEffect reports whether any permanent on the
// battlefield imposes a sacrifice right now, so the hot paths can skip the
// fold on the boards that have none. Caller must hold g.mu.
func (g *Game) ImposedSacrificesExistForEffect() bool {
	if g == nil || g.Battlefield == nil || CatalogImposedSacrifices == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		if len(imposedSacrificesOf(&g.Battlefield.Cards[i])) > 0 {
			return true
		}
	}
	return false
}

// zoneCostClauseHead is the zone a folded clause's CardOK is asked in when
// the question is "does this permanent pay the object's OWN clause?" — a
// Head entry of the set rule. Only foldImposedSacrifice's closure reads
// it; no zone walk ever produces it.
const zoneCostClauseHead ZoneKind = "cost_clause_head"

// foldImposedSacrifice joins `kinds` (the imposed sacrifices) to `own`
// (the object's own sacrifice clause, nil for none), paid by `payer` for
// an ability of `sourceID` (uuid.Nil for a spell). The result is a fresh
// fixed-count clause whose EachOf names every permanent to sacrifice;
// `own` is never mutated. ok is false when `own` cannot be folded (see
// the file comment).
func (g *Game) foldImposedSacrifice(payer, sourceID uuid.UUID, own *TargetSpec, kinds []SacrificeKind, label string) (*TargetSpec, bool) {
	if len(kinds) == 0 {
		return own, true
	}
	imposed := slices.Clone(kinds)
	matchKinds := func(c Card) bool {
		for _, k := range imposed {
			if k.Matches(c) {
				return true
			}
		}
		return false
	}
	if own == nil {
		spec := &TargetSpec{
			Mode:  "permanent",
			Label: label,
			Zones: []ZoneKind{ZoneBattlefield},
			CardOK: func(_ *Game, _ uuid.UUID, c Card, zone ZoneKind) bool {
				return zone != zoneCostClauseHead && matchKinds(c)
			},
			EachOf: imposed,
		}
		return spec.WithCount(len(imposed), len(imposed)), true
	}
	if SacrificeCostVariable(own) || SacrificeAnyNumber(own) || own.CountFromX || SacrificeManaValueX(own) ||
		own.Different != nil || own.Same != nil || own.Total != nil || len(own.Rest) > 0 {
		return nil, false
	}
	n := SacrificeCostCount(own)
	var entries []SacrificeKind
	if len(own.EachOf) > 0 {
		// Already a set rule of kinds ("a Swamp and a Forest"): its
		// entries join the imposed ones as they are.
		entries = append(slices.Clone(own.EachOf), imposed...)
	} else {
		for range n {
			entries = append(entries, SacrificeKind{Label: own.Label, Head: true})
		}
		entries = append(entries, imposed...)
	}
	// The own clause's whole per-permanent test (its CardOK and its data
	// filters), asked of the clause as printed, so "another creature" or
	// "an untapped Cleric" still means what it said.
	ownMatch := func(gg *Game, c Card) bool {
		if own.ExcludeSource && c.InstanceID == sourceID {
			return false
		}
		if !gg.AttachedToSourceOKForEffect(own, sourceID, c.InstanceID) {
			return false
		}
		return gg.specMatchLocked(gg.costSourceLocked(payer, sourceID), own, TargetRef{Kind: TargetCard, ID: c.InstanceID}, false)
	}
	spec := &TargetSpec{
		Mode:  own.Mode,
		Label: own.Label,
		Zones: []ZoneKind{ZoneBattlefield},
		CardOK: func(gg *Game, _ uuid.UUID, c Card, zone ZoneKind) bool {
			if zone == zoneCostClauseHead {
				return ownMatch(gg, c)
			}
			return ownMatch(gg, c) || matchKinds(c)
		},
		EachOf: entries,
	}
	if spec.Mode == "" {
		spec.Mode = "permanent"
	}
	return spec.WithCount(len(entries), len(entries)), true
}

// costClauseHeadMatchLocked answers a Head entry of `spec`'s set rule:
// whether `c` pays the clause the imposed sacrifices were folded into.
// Caller must hold g.mu.
func (g *Game) costClauseHeadMatchLocked(spec *TargetSpec, c Card) bool {
	return spec != nil && spec.CardOK != nil && spec.CardOK(g, uuid.Nil, c, zoneCostClauseHead)
}

// ImposedSacrificeAbilityCostForEffect is `cost` — an activated ability
// of `source` activated by `activator` — with every imposed sacrifice
// folded into its sacrifice clause (CR 602.2b, 601.2f). ok is false when
// the fold is impossible and the ability cannot be activated now.
// Unchanged, and ok, on every board with no imposed sacrifice. Caller
// must hold g.mu.
func (g *Game) ImposedSacrificeAbilityCostForEffect(activator uuid.UUID, source Card, cost AbilityCost) (AbilityCost, bool) {
	if !g.ImposedSacrificesExistForEffect() {
		return cost, true
	}
	kinds := g.imposedSacrificeKindsLocked(activator, source, false, cost.Mana)
	if len(kinds) == 0 {
		return cost, true
	}
	if cost.SacrificeAll {
		return cost, false
	}
	folded, ok := g.foldImposedSacrifice(activator, source.InstanceID, cost.SacrificeOther, kinds, kinds[0].Label)
	if !ok {
		return cost, false
	}
	cost.SacrificeOther = folded
	return cost, true
}

// ImposedSacrificeManaAbilityForEffect is ImposedSacrificeAbilityCostForEffect
// for a mana ability (CR 605: a mana ability is an activated ability, so
// Drought's "activated abilities" reaches a filter land's {B/R}). ok is
// false when the fold is impossible. Caller must hold g.mu.
func (g *Game) ImposedSacrificeManaAbilityForEffect(activator uuid.UUID, source Card, ab ManaAbilityShape) (ManaAbilityShape, bool) {
	if !g.ImposedSacrificesExistForEffect() {
		return ab, true
	}
	kinds := g.imposedSacrificeKindsLocked(activator, source, false, ab.ManaCost)
	if len(kinds) == 0 {
		return ab, true
	}
	folded, ok := g.foldImposedSacrifice(activator, source.InstanceID, ab.SacrificeOther, kinds, kinds[0].Label)
	if !ok {
		return ab, false
	}
	ab.SacrificeOther = folded
	return ab, true
}

// ImposedSacrificeCastCostForEffect is `ac` — the additional cost of
// `card` cast by `caster`, its either/or branch already chosen — with
// every imposed sacrifice folded into its sacrifice clause (CR 601.2f). A
// card with no additional cost of its own gets one that only sacrifices.
// The catalog's AdditionalCost is never mutated. ok is false when the
// fold is impossible and the card cannot be cast now. Caller must hold
// g.mu.
func (g *Game) ImposedSacrificeCastCostForEffect(caster uuid.UUID, card Card, ac *AdditionalCost) (*AdditionalCost, bool) {
	if !g.ImposedSacrificesExistForEffect() {
		return ac, true
	}
	kinds := g.imposedSacrificeKindsLocked(caster, card, true, card.ManaCost)
	if len(kinds) == 0 {
		return ac, true
	}
	var own *TargetSpec
	if ac != nil {
		if ac.SacrificeAll || ac.SacrificeHalfRoundedUp || ac.CasualtyX || ac.ReducedBySacrificedManaCost {
			return ac, false
		}
		own = ac.Sacrifice
	}
	folded, ok := g.foldImposedSacrifice(caster, uuid.Nil, own, kinds, kinds[0].Label)
	if !ok {
		return ac, false
	}
	var out AdditionalCost
	if ac != nil {
		out = *ac
	}
	if out.Label == "" {
		out.Label = "Sacrifice " + kinds[0].Label
	}
	out.Sacrifice = folded
	return &out, true
}
