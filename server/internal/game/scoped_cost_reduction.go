package game

import (
	"fmt"

	"github.com/google/uuid"
)

// scoped_cost_reduction.go — "until <duration>, spells <a player> casts
// cost {N} less to cast", a timed cost reduction held as a scoped record
// (#6093). The twin of scoped_cost_tax.go's increase: an effect of a
// resolving spell or ability that lasts for a while and outlives its
// source, so it is data in a ScopedEffect record (ADR 0041) read by the one
// place every cast is priced (activeCostModifiersLocked).
//
//   - Commander Liara Portyr: "spells you cast from exile this turn cost
//     {X} less to cast, where X is the number of players being attacked".
//     X is counted as the trigger resolves and written into the record.
//   - Will Kenrith's −2: "Until your next turn, instant, sorcery, and
//     planeswalker spells that player casts cost {2} less to cast".
//
// The record names the casting player in Mod.Player; Mod.Types narrows
// the spells to those with at least one of the card types listed (empty
// is every spell). "Spells you cast from exile" is a kind of its own,
// ModExiledSpellsCostLessForPlayer, rather than a zone field: Mod's one
// zone field, SourceZone, is the damage-source shield's and is refused on
// every other kind. Amount is generic mana removed (CR 601.2f), the reduction
// a printed "cost {N} less" makes: it joins the reduction pass and is
// floored at the cost's generic component like every other. It applies to
// a spell's casting only, never to an ability's activation. Mod kinds are
// on-disk identities: never renamed, never reused.

// ModSpellsCostLessForPlayer is "spells <Player> casts [of these types]
// cost {Amount} less to cast" for as long as the record lasts. Not a layer
// operation: the cast pricer reads it.
const ModSpellsCostLessForPlayer ModKind = "spellsCostLessForPlayer"

// ModExiledSpellsCostLessForPlayer is "spells <Player> casts from exile
// [of these types] cost {Amount} less to cast": ModSpellsCostLessForPlayer
// narrowed to spells cast from exile (CostQuery.FromZone).
const ModExiledSpellsCostLessForPlayer ModKind = "exiledSpellsCostLessForPlayer"

// ModNamedSpellsCostLess is "spells with the chosen name cost {Amount} less to
// cast" for as long as the record lasts, whoever casts them (Cheering
// Fanatic, S82 batch 206). The name rides Mod.Text; Player stays empty.
const ModNamedSpellsCostLess ModKind = "namedSpellsCostLess"

// NamedSpellsCostLessMod is the record's mod: spells named `name` (any face,
// CR 201.2b), cast by any player, cost `n` less.
func NamedSpellsCostLessMod(name string, n int) Mod {
	return Mod{Kind: ModNamedSpellsCostLess, Text: name, Amount: n}
}

// SpellsCostLessForPlayerMod is the record's mod: spells `player` casts
// cost `n` less, narrowed to `types` (any of them; none is every spell).
func SpellsCostLessForPlayerMod(player uuid.UUID, n int, types ...string) Mod {
	return Mod{Kind: ModSpellsCostLessForPlayer, Player: player, Amount: n, Types: append([]string(nil), types...)}
}

// ExiledSpellsCostLessForPlayerMod is the from-exile form: spells `player`
// casts from exile cost `n` less, narrowed to `types` as above.
func ExiledSpellsCostLessForPlayerMod(player uuid.UUID, n int, types ...string) Mod {
	return Mod{Kind: ModExiledSpellsCostLessForPlayer, Player: player, Amount: n, Types: append([]string(nil), types...)}
}

// costLessForPlayer reports whether m is one of the two reduction kinds.
func costLessForPlayer(m Mod) bool {
	return m.Kind == ModSpellsCostLessForPlayer || m.Kind == ModExiledSpellsCostLessForPlayer || m.Kind == ModNamedSpellsCostLess
}

// scopedCostReductionModProblem is the registration check: "" when the mod
// is sound. A reduction of zero is not refused, because Liara's X is the
// number of players being attacked and the caller decides whether to
// register one; a negative amount would be an increase and is refused.
func scopedCostReductionModProblem(m Mod) string {
	if !costLessForPlayer(m) {
		return ""
	}
	if m.Kind == ModNamedSpellsCostLess {
		if m.Text == "" {
			return fmt.Sprintf("a %s mod names no card", m.Kind)
		}
	} else if m.Player == uuid.Nil {
		return fmt.Sprintf("a %s mod names no player", m.Kind)
	}
	if m.Amount < 0 {
		return fmt.Sprintf("a %s mod needs an Amount of at least 0", m.Kind)
	}
	return ""
}

// appendScopedCostReductionLocked adds one cost modifier for each live
// record that reduces a player's casts, bound to a stand-in for the
// effect's source whose controller is the record's controller. Caller must
// hold g.mu.
func (g *Game) appendScopedCostReductionLocked(out []boundCostModifier, q CostQuery) []boundCostModifier {
	for i := range g.ScopedEffects {
		e := g.ScopedEffects[i]
		for _, m := range e.Mods {
			if !costLessForPlayer(m) || m.Amount < 1 {
				continue
			}
			player, n, types := m.Player, m.Amount, m.Types
			fromExile := m.Kind == ModExiledSpellsCostLessForPlayer
			if m.Kind == ModNamedSpellsCostLess {
				name := m.Text
				cm := CostModifier{
					Kind:      CostReduction,
					Label:     e.Label,
					AppliesTo: func(q CostQuery) bool { return CardNameMatches(q.Card, name) },
					Amount:    func(CostQuery) int { return n },
				}
				if cm.pricesAnnouncement(q) {
					out = append(out, boundCostModifier{modifier: cm, source: Card{Controller: e.Controller, Name: e.Label}})
				}
				continue
			}
			cm := CostModifier{
				Kind:  CostReduction,
				Label: e.Label,
				AppliesTo: func(q CostQuery) bool {
					if player == uuid.Nil || q.Controller != player {
						return false
					}
					if fromExile && q.FromZone != ZoneExile {
						return false
					}
					if len(types) == 0 {
						return true
					}
					for _, t := range types {
						if q.Card.HasCardType(t) {
							return true
						}
					}
					return false
				},
				Amount: func(CostQuery) int { return n },
			}
			if !cm.pricesAnnouncement(q) {
				continue
			}
			out = append(out, boundCostModifier{modifier: cm, source: Card{Controller: e.Controller, Name: e.Label}})
		}
	}
	return out
}
