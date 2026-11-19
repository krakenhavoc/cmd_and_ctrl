package game

import "github.com/google/uuid"

// legend_rule_exemption.go — "the legend rule doesn't apply" (CR 704.5j,
// #2177): Mirror Box and Sakashima of a Thousand Faces for permanents
// their controller controls, Mirror Gallery for every permanent.
//
// Read live off the battlefield each time the legend-rule state-based
// action runs, keyed by CatalogAbilityKey, and never stored (the
// AnyColorSpend pattern): two exempting permanents compose, one leaving
// cannot revoke the other's exemption, and one that lost its abilities
// exempts nothing. When the last exempting permanent leaves, the very
// next SBA check applies the rule to the duplicates that remain, and
// their controller chooses which to keep (CR 704.5j).

// LegendRuleExemptionWhose is whose permanents a printed exemption covers.
type LegendRuleExemptionWhose uint8

const (
	// LegendRuleExemptYours is "…doesn't apply to permanents you control":
	// the exempting permanent's controller (Mirror Box, Sakashima).
	LegendRuleExemptYours LegendRuleExemptionWhose = iota
	// LegendRuleExemptEveryone is "The legend rule doesn't apply"
	// (Mirror Gallery).
	LegendRuleExemptEveryone
	// LegendRuleExemptYourCreatures is "The legend rule doesn't apply to
	// creatures you control" (Council of Reeds, #4707): only the
	// controller's permanents that are creatures right now. A legendary
	// artifact or planeswalker they control, and anything an opponent
	// controls, still answers to CR 704.5j. Appended last so no existing
	// value moves.
	LegendRuleExemptYourCreatures
	// LegendRuleExemptYourSubtype is "The legend rule doesn't apply to
	// <Spiders> you control" (Spider-Verse, #4460): only the controller's
	// permanents that have LegendRuleExemption.Subtype right now, read
	// through the exempting permanent's text (CR 612.2). Appended last so
	// no existing value moves.
	LegendRuleExemptYourSubtype
	// LegendRuleExemptYourTokens is "The legend rule doesn't apply to
	// tokens you control" (Cadric, Soul Kindler): only the controller's
	// permanents that are tokens right now. A legendary card they control,
	// and anything an opponent controls, still answers to CR 704.5j.
	// Appended last so no existing value moves.
	LegendRuleExemptYourTokens
	// LegendRuleExemptNamedPair is "If there are exactly two permanents
	// named <this card> on the battlefield, the legend rule doesn't apply
	// to them" (Brothers Yamazaki): read live, the exempting permanent's
	// own effective name; covers both permanents, whoever controls them,
	// while exactly two bear that name. A third (or a lone one) switches
	// it off, and CR 704.5j applies. Appended last so no existing value
	// moves.
	LegendRuleExemptNamedPair
)

// LegendRuleExemption is one printed legend-rule exemption. Catalog data.
type LegendRuleExemption struct {
	// Label is the clause as printed.
	Label string
	Whose LegendRuleExemptionWhose
	// Subtype is LegendRuleExemptYourSubtype's subtype, as printed.
	Subtype string
}

// CatalogLegendRuleExemptions returns the exemptions a permanent with the
// given catalog key has. carddef.go sets it from CardDef.
var CatalogLegendRuleExemptions func(key string) []LegendRuleExemption

// legendRuleExempt is the set of controllers the exemptions on the
// battlefield cover right now.
type legendRuleExempt struct {
	all         bool
	controllers map[uuid.UUID]bool
	// creatureControllers are the controllers whose CREATURES are covered
	// (LegendRuleExemptYourCreatures).
	creatureControllers map[uuid.UUID]bool
	// subtypeControllers are the subtypes each controller's permanents
	// are covered for (LegendRuleExemptYourSubtype).
	subtypeControllers map[uuid.UUID][]string
	// tokenControllers are the controllers whose TOKENS are covered
	// (LegendRuleExemptYourTokens).
	tokenControllers map[uuid.UUID]bool
	// pairNames are the names of exactly-two pairs the exemption covers
	// (LegendRuleExemptNamedPair).
	pairNames map[string]bool
}

func (e legendRuleExempt) covers(controller uuid.UUID) bool {
	return e.all || e.controllers[controller]
}

// coversPermanent is covers for one permanent: its controller is exempt
// outright, or it is a creature and the controller's creatures are.
func (e legendRuleExempt) coversPermanent(c *Card) bool {
	if e.covers(c.Controller) || (e.creatureControllers[c.Controller] && c.IsCreature()) ||
		(e.tokenControllers[c.Controller] && c.IsToken()) {
		return true
	}
	if len(e.pairNames) > 0 {
		name := c.Effective().Name
		if name == "" {
			name = c.Name
		}
		if e.pairNames[name] {
			return true
		}
	}
	for _, sub := range e.subtypeControllers[c.Controller] {
		if c.HasSubtype(sub) {
			return true
		}
	}
	return false
}

// legendRuleExemptControllersLocked reads every exemption on the
// battlefield. Caller must hold g.mu (read or write). Reads only.
func (g *Game) legendRuleExemptControllersLocked() legendRuleExempt {
	var out legendRuleExempt
	// The stored "this turn" records (Hall of Echoes, #3515): a
	// player-scoped exemption an ability registered, swept at cleanup.
	// Read before the battlefield walk so it still applies when no
	// catalog is loaded.
	for i := range g.ScopedEffects {
		for _, m := range g.ScopedEffects[i].Mods {
			if m.Kind != ModLegendRuleDoesntApply {
				continue
			}
			if m.Player == uuid.Nil {
				out.all = true
				continue
			}
			if out.controllers == nil {
				out.controllers = map[uuid.UUID]bool{}
			}
			out.controllers[m.Player] = true
		}
	}
	if CatalogLegendRuleExemptions == nil || g.Battlefield == nil {
		return out
	}
	for i := range g.Battlefield.Cards {
		src := &g.Battlefield.Cards[i]
		key := catalogAbilityKeyOf(src)
		if key == "" {
			continue
		}
		for _, s := range CatalogLegendRuleExemptions(key) {
			switch s.Whose {
			case LegendRuleExemptEveryone:
				out.all = true
			case LegendRuleExemptYours:
				if out.controllers == nil {
					out.controllers = map[uuid.UUID]bool{}
				}
				out.controllers[src.Controller] = true
			case LegendRuleExemptYourCreatures:
				if out.creatureControllers == nil {
					out.creatureControllers = map[uuid.UUID]bool{}
				}
				out.creatureControllers[src.Controller] = true
			case LegendRuleExemptYourTokens:
				if out.tokenControllers == nil {
					out.tokenControllers = map[uuid.UUID]bool{}
				}
				out.tokenControllers[src.Controller] = true
			case LegendRuleExemptNamedPair:
				name := src.Effective().Name
				if name == "" {
					name = src.Name
				}
				if name == "" || g.battlefieldNameCountLocked(name) != 2 {
					continue
				}
				if out.pairNames == nil {
					out.pairNames = map[string]bool{}
				}
				out.pairNames[name] = true
			case LegendRuleExemptYourSubtype:
				if s.Subtype == "" {
					continue
				}
				if out.subtypeControllers == nil {
					out.subtypeControllers = map[uuid.UUID][]string{}
				}
				out.subtypeControllers[src.Controller] = append(out.subtypeControllers[src.Controller], src.TextSubtype(s.Subtype))
			}
		}
	}
	return out
}

// LegendRuleDoesntApplyThisTurnForEffect registers "the legend rule
// doesn't apply to permanents you control this turn" (Hall of Echoes,
// #3515; ADR 0032's amendment of 2026-10-14): a ModLegendRuleDoesntApply
// record naming `player`, swept at cleanup (CR 514.2). The printed
// Spec.LegendRuleExemptions are a static read live off the battlefield
// with no duration, which would switch the rule off for the whole game;
// this is the version an ability starts and the turn ends.
//
// `label` is the source's clause for the log. Registers nothing for a
// player who is not seated. Reports whether a record was written. The
// rule is not re-run here: the next state-based action check sees the
// record, and the first one after cleanup sees it gone and applies
// CR 704.5j with the controller choosing which to keep.
//
// Caller must hold g.mu (write).
func (g *Game) LegendRuleDoesntApplyThisTurnForEffect(sourceID, player uuid.UUID, label string) bool {
	if player == uuid.Nil || g.playerByIDLocked(player) == nil {
		return false
	}
	return g.RegisterScopedRuleEffectForEffect(sourceID, ScopeGame, uuid.Nil,
		[]Mod{{Kind: ModLegendRuleDoesntApply, Player: player}}, g.UntilEndOfTurnDuration(), label)
}

// battlefieldNameCountLocked counts the permanents whose effective name
// is `name` (CR 201.2). Caller must hold g.mu.
func (g *Game) battlefieldNameCountLocked(name string) int {
	n := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		got := c.Effective().Name
		if got == "" {
			got = c.Name
		}
		if got == name {
			n++
		}
	}
	return n
}
