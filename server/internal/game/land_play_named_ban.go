package game

import "github.com/google/uuid"

// land_play_named_ban.go — S82 OLD batch 573 (Conjurer's Ban): "lands with
// the chosen name can't be played" until a stated duration ends (CR 101.2,
// CR 611.2b). The record is a ModCantPlayLandsNamed ScopedEffect naming the
// card name in Text, one mod per name, banning every player; the land-play
// gate reads it beside the stored "this turn" records.

// CantPlayLandsNamedForEffect registers the ban. `label` is the source's
// name for the refusal. Reports whether a record was written; a blank name
// bans nothing.
//
// Caller must hold g.mu (write).
func (g *Game) CantPlayLandsNamedForEffect(sourceID uuid.UUID, name string, d Duration, label string) bool {
	if name == "" {
		return false
	}
	return g.RegisterScopedRuleEffectForEffect(sourceID, ScopeGame, uuid.Nil,
		[]Mod{{Kind: ModCantPlayLandsNamed, Text: name}}, d, label)
}

// landPlayNamedBanLocked returns the refusal for `card` under a live named
// land-play ban, if any. The duration is asked here rather than left to the
// sweep, for the reason graveyardPlayBanLocked gives.
//
// Caller must hold g.mu (read or write).
func (g *Game) landPlayNamedBanLocked(card Card) (reason string, source uuid.UUID, banned bool) {
	for i := range g.ScopedEffects {
		e := &g.ScopedEffects[i]
		for _, m := range e.Mods {
			if m.Kind != ModCantPlayLandsNamed || !SpellOrLandHasName(card, m.Text) {
				continue
			}
			if g.durationExpiredLocked(e.Duration, false) {
				continue
			}
			reason = "Lands named " + m.Text + " can't be played"
			if e.SourceName != "" {
				reason += " — " + e.SourceName
			}
			return reason, e.Source.ID, true
		}
	}
	return "", uuid.Nil, false
}
