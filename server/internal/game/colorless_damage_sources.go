package game

import "slices"

// colorless_damage_sources.go — "Black and/or red permanents and spells
// are colorless sources of damage" (Ghostly Flame, #6230).
//
// A rule-modifying static (CR 613.11), not a colour change: the object
// stays black or red for everything else (a "destroy target black
// creature", a colour count, a devotion), and only as a SOURCE OF DAMAGE
// is it colourless (CR 105.2c: a colorless object has no color). So it is
// read at the one place the engine looks at a damage source's
// characteristics, damageSourceLKILocked, which every damage reader goes
// through: protection (CR 702.16e), prevention filters ("damage a black
// source would deal"), the tallies ("dealt damage by a red source this
// turn"). The snapshot is taken as the damage event opens, so it is the
// statics on the battlefield then that count.
//
// It reaches a permanent (on the battlefield, or remembered from there)
// and a spell (on the stack); a card in any other zone is neither and
// keeps its colours. Read live through CatalogAbilityKey, so a Ghostly
// Flame that has lost its abilities stops.

// CatalogColorlessDamageSources returns the colours a battlefield
// permanent with this catalog key makes colourless as sources of damage.
// Populated from effects.Spec.ColorlessDamageSources. A nil hook finds
// none.
var CatalogColorlessDamageSources func(key string) []string

// colorlessDamageSourceColorsLocked is every colour some permanent on the
// battlefield makes colourless as a source of damage.
//
// Caller must hold g.mu.
func (g *Game) colorlessDamageSourceColorsLocked() []string {
	if g.Battlefield == nil || CatalogColorlessDamageSources == nil {
		return nil
	}
	var out []string
	for i := range g.Battlefield.Cards {
		key := catalogAbilityKeyOf(&g.Battlefield.Cards[i])
		if key == "" {
			continue
		}
		for _, c := range CatalogColorlessDamageSources(key) {
			if !slices.Contains(out, c) {
				out = append(out, c)
			}
		}
	}
	return out
}

// asDamageSourceLocked is ch as a source of damage: a copy with no
// colours when the source is a permanent or a spell with one of the
// colours colorlessDamageSourceColorsLocked names, else ch itself.
// permanentOrSpell says whether the object is one.
//
// Caller must hold g.mu.
func (g *Game) asDamageSourceLocked(ch *Characteristic, permanentOrSpell bool) *Characteristic {
	if ch == nil || !permanentOrSpell || len(ch.Colors) == 0 {
		return ch
	}
	colors := g.colorlessDamageSourceColorsLocked()
	if len(colors) == 0 || !slices.ContainsFunc(ch.Colors, func(c string) bool { return slices.Contains(colors, c) }) {
		return ch
	}
	out := *ch
	out.Colors = nil
	return &out
}
