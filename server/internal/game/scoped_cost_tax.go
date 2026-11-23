package game

import "github.com/google/uuid"

// scoped_cost_tax.go — "until <duration>, spells your opponents cast cost {N}
// more to cast", a timed cost increase held as a scoped record (Tax Collector's
// Tax bullet, #5948). A battlefield permanent's cost modifier is a static
// ability and a promise like "the next spell you cast" is held on the caster;
// this is the third kind, an effect of a resolving ability that taxes every
// opponent's casts for a while and outlives its source, so it is data in a
// ScopedEffect record (ADR 0041) read by the one place every cast is priced
// (activeCostModifiersLocked), like the rule gates around it.
//
// The record's Controller is "you": the taxed casters are every player but them.
// The Amount is generic mana added (CR 601.2f), the same increase a printed
// "cost {1} more" makes, and it applies to a spell's casting only, never to an
// ability's activation. Mod kinds are on-disk identities: never renamed, never
// reused.

// ModSpellsCostMoreForOpponents is "spells your opponents cast cost {Amount}
// more to cast" for as long as the record lasts. Reads Amount (at least 1);
// "you" is the record's Controller. Not a layer operation: the cast pricer reads
// it.
const ModSpellsCostMoreForOpponents ModKind = "spellsCostMoreForOpponents"

// SpellsCostMoreForOpponentsMod is the record's mod for a generic increase of n.
func SpellsCostMoreForOpponentsMod(n int) Mod {
	return Mod{Kind: ModSpellsCostMoreForOpponents, Amount: n}
}

// scopedCostTaxModProblem is the registration check: "" when the mod is sound.
func scopedCostTaxModProblem(m Mod) string {
	if m.Kind == ModSpellsCostMoreForOpponents && m.Amount < 1 {
		return "a spellsCostMoreForOpponents mod needs an Amount of at least 1"
	}
	return ""
}

// appendScopedCostTaxLocked adds one cost modifier for each live record that
// taxes opponents' casts, bound to a stand-in for the effect's source whose
// controller is "you". Caller must hold g.mu.
func (g *Game) appendScopedCostTaxLocked(out []boundCostModifier, q CostQuery) []boundCostModifier {
	for i := range g.ScopedEffects {
		e := g.ScopedEffects[i]
		for _, m := range e.Mods {
			if m.Kind != ModSpellsCostMoreForOpponents {
				continue
			}
			you, n := e.Controller, m.Amount
			cm := CostModifier{
				Kind:  CostIncrease,
				Label: e.Label,
				AppliesTo: func(q CostQuery) bool {
					return you != uuid.Nil && q.Controller != uuid.Nil && q.Controller != you
				},
				Amount: func(CostQuery) int { return n },
			}
			if !cm.pricesAnnouncement(q) {
				continue
			}
			out = append(out, boundCostModifier{modifier: cm, source: Card{Controller: you, Name: e.Label}})
		}
	}
	return out
}
