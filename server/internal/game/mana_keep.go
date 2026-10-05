package game

import "github.com/google/uuid"

// mana_keep.go — mana a player does not lose as steps and phases end
// (CR 106.4, CR 500.4; #2166, ADR 0040 amendment 2026-10-05).
//
// emptyAllManaPoolsLocked used to drop every token at every step entry.
// Two shapes of printed text change that, and they are two readers over
// the same sweep:
//
//   - A STATIC over a player's pool, derived from the battlefield and
//     never stored (CatalogManaPool, the CatalogHandSize pattern):
//     Upwelling "players don't lose unspent mana", Omnath "unspent green
//     mana", Leyline Tyrant "unspent red mana", and Kruphix, God of
//     Horizons' "if you would lose unspent mana, that mana becomes
//     colorless instead". The static leaving ends it at once, so the next
//     boundary empties normally.
//
//   - MANA MARKED when it is produced: "until end of turn, you don't lose
//     this mana as steps and phases end" (Karn, Legacy Reforged, Savage
//     Ventmaw). The mark is a ManaRiderKeepUntilEndOfTurn rider on the
//     token, so it rides every mint route and the snapshot. It expires
//     when the cleanup step begins (CR 514.2 ends "until end of turn"
//     effects there) — or at the next untap, for a turn that ended
//     without one — and the mana is then an ordinary token, lost (or
//     converted) like any other.
//
// A granted statement with a duration ("until end of turn, you don't
// lose unspent red mana", The Last Agni Kai) is a third source:
// PlayerStatic.KeepManaColors, read through the same duration test every
// player-level grant uses.
//
// Kruphix's conversion keeps the token's restrictions and riders: only
// its colour changes, so a Karn-style restricted mana stays restricted.

// ManaPoolPlayers is who a mana-pool static reaches, relative to the
// controller of the permanent that has it.
type ManaPoolPlayers int

const (
	// ManaPoolYou — "you don't lose …". The zero value.
	ManaPoolYou ManaPoolPlayers = iota
	// ManaPoolEachPlayer — "players don't lose …" (Upwelling).
	ManaPoolEachPlayer
)

// ManaPoolKind is what a mana-pool static does to mana that would be lost.
type ManaPoolKind int

const (
	// ManaPoolKeep — the mana stays. Colors narrows it; nil is all mana,
	// colourless included.
	ManaPoolKeep ManaPoolKind = iota
	// ManaPoolBecomesColorless — the mana that would be lost becomes
	// colourless and stays (Kruphix). Colors is unused.
	ManaPoolBecomesColorless
)

// ManaPoolStatic is ONE printed static over a player's mana pool.
// Catalog data, declared on effects.Spec.ManaPool; never serialised.
type ManaPoolStatic struct {
	Players ManaPoolPlayers
	Kind    ManaPoolKind
	// Colors are the colours a ManaPoolKeep covers ("G", "R", …); empty
	// is every mana.
	Colors []string
	// When is the designation gate (ADR 0071); the zero value is none.
	When Designation
}

// reaches reports whether the static, on a permanent `controller`
// controls, applies to player `p`.
func (s ManaPoolStatic) reaches(controller, p uuid.UUID) bool {
	if s.Players == ManaPoolEachPlayer {
		return true
	}
	return p == controller
}

// CatalogManaPool returns the mana-pool statics a catalog entry
// declares. A nil hook means no catalog is wired and nothing is kept.
var CatalogManaPool func(oracleID string) []ManaPoolStatic

// manaPoolRules is what the battlefield and the player's grants say
// about a pool right now.
type manaPoolRules struct {
	keepAll    bool
	keepColors map[string]bool
	convert    bool
}

func (r manaPoolRules) keeps(color string) bool {
	return r.keepAll || r.keepColors[color]
}

// manaPoolRulesLocked derives the rules for `p`. Derived on every call
// and stored nowhere, so a static that leaves stops applying at the very
// next boundary. endOfTurn is true when the boundary is the one that ends
// "until end of turn" grants.
func (g *Game) manaPoolRulesLocked(p *Player, endOfTurn bool) manaPoolRules {
	var r manaPoolRules
	add := func(s ManaPoolStatic) {
		switch s.Kind {
		case ManaPoolBecomesColorless:
			r.convert = true
		case ManaPoolKeep:
			if len(s.Colors) == 0 {
				r.keepAll = true
				return
			}
			if r.keepColors == nil {
				r.keepColors = map[string]bool{}
			}
			for _, c := range s.Colors {
				r.keepColors[c] = true
			}
		}
	}
	if CatalogManaPool != nil && g.Battlefield != nil {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			// CatalogAbilityKey: these are static abilities, and a
			// permanent that has lost its abilities gives nothing.
			key := catalogAbilityKeyOf(c)
			if key == "" {
				continue
			}
			for _, s := range CatalogManaPool(key) {
				if s.reaches(c.Controller, p.ID) && s.When.Active(*c) {
					add(s)
				}
			}
		}
	}
	for _, s := range p.Statics {
		if len(s.KeepManaColors) == 0 || g.durationExpiredLocked(s.Duration, endOfTurn) {
			continue
		}
		add(ManaPoolStatic{Kind: ManaPoolKeep, Colors: s.KeepManaColors})
	}
	return r
}

// GrantKeepManaForEffect is "until <duration>, you don't lose unspent
// <colours> mana as steps and phases end" (The Last Agni Kai). Caller
// must hold g.mu (write). Empty colours is a no-op.
func (g *Game) GrantKeepManaForEffect(player uuid.UUID, colors []string, label string, source uuid.UUID, d Duration) {
	p := g.playerByIDLocked(player)
	if p == nil || len(colors) == 0 {
		return
	}
	p.Statics = append(p.Statics, PlayerStatic{
		KeepManaColors: append([]string(nil), colors...),
		Source:         source,
		Label:          label,
		Duration:       d,
	})
}

// hasKeepMark reports whether the token carries the keep-until-end-of-turn
// mark.
func (t ManaToken) hasKeepMark() bool {
	for _, r := range t.Riders {
		if r.Kind == ManaRiderKeepUntilEndOfTurn {
			return true
		}
	}
	return false
}

// withoutKeepMark returns the token with the mark stripped (fresh rider
// slice), for the moment it expires.
func (t ManaToken) withoutKeepMark() ManaToken {
	var riders []ManaSpendRider
	for _, r := range t.Riders {
		if r.Kind != ManaRiderKeepUntilEndOfTurn {
			riders = append(riders, r)
		}
	}
	t.Riders = copyManaRiders(riders)
	return t
}

// sweepManaPoolLocked is the CR 106.4 step-boundary sweep for one
// player: it keeps what the player's rules and marks allow, converts
// what Kruphix converts, and returns how many tokens were LOST (a
// converted token is not lost). Caller must hold g.mu (write).
func (g *Game) sweepManaPoolLocked(p *Player) int {
	// The boundary that ends "until end of turn": cleanup begins (CR
	// 514.2), or a new turn's untap when no cleanup ran in between.
	endOfTurn := g.Turn.Step == StepCleanup || g.Turn.Step == StepUntap
	rules := g.manaPoolRulesLocked(p, endOfTurn)
	var kept ManaPool
	lost := 0
	for _, t := range p.ManaPool {
		if t.hasKeepMark() {
			if !endOfTurn {
				kept = append(kept, t)
				continue
			}
			t = t.withoutKeepMark()
		}
		switch {
		case rules.keeps(t.Color):
			kept = append(kept, t)
		case rules.convert:
			t.Color = "C"
			kept = append(kept, t)
		default:
			lost++
		}
	}
	p.ManaPool = kept
	// A sweep that only kept or converted emits no emptied event, but
	// a conversion changes the colour counts a pool-reading static sees.
	invalidateLayersForManaPoolLocked(g)
	return lost
}
