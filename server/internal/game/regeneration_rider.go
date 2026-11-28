package game

import "github.com/google/uuid"

// regeneration_rider.go — "Regenerate target creature. You gain control of
// that creature if it regenerates this way" (Debt of Loyalty, CR 701.19a).
//
// The shield is the ordinary one (Card.RegenerationShields); what the spell
// adds is a rider that fires when THAT shield is the one that replaces a
// destruction. The rider is a ScopedEffect record (ModRegenerationGainsControl)
// pinned to the permanent and ending with the turn, as the shield does, so it
// is data a snapshot carries and a cleanup sweep removes with no new field on
// the card.
//
// WHICH shield is used. A permanent can carry plain shields and riders at once;
// the permanent's controller chooses the order its replacements apply (CR
// 616.1), and would spend a plain shield before one that costs them the
// creature. So a rider is spent only when there are more riders than shields
// left after this one is used: with one plain and one rider-bearing shield the
// first destruction is replaced by the plain one and the creature stays put.

// RegenerateGainingControlForEffect is RegenerateForEffect with the rider:
// one shield for the permanent and, if that shield replaces a destruction,
// `player` gains control of it with no duration. Returns ErrCardNotFound for a
// card that is not on the battlefield (CR 701.19b).
//
// Caller must hold g.mu (write).
func (g *Game) RegenerateGainingControlForEffect(cardID, player, source uuid.UUID, label string) error {
	if err := g.RegenerateForEffect(cardID); err != nil {
		return err
	}
	affected := g.PinnedObjectsLocked(cardID)
	if len(affected) == 0 || player == uuid.Nil {
		return nil
	}
	g.appendScopedEffectLocked(source, affected, ScopeNone, uuid.Nil,
		[]Mod{{Kind: ModRegenerationGainsControl, Player: player}},
		g.PinnedTo(g.UntilEndOfTurnDuration(), cardID), label, timeNowUnixNano())
	return nil
}

// regenerationRiderIndexesLocked lists, oldest first, the registry indexes of
// the rider records that cover `c`.
//
// Caller must hold g.mu.
func (g *Game) regenerationRiderIndexesLocked(c *Card) []int {
	var out []int
	for i := range g.ScopedEffects {
		e := &g.ScopedEffects[i]
		if e.Scope == ScopeNone && scopedEffectHasMod(e, ModRegenerationGainsControl) && affectedPredicate(e.Affected)(c, g, nil) {
			out = append(out, i)
		}
	}
	return out
}

// spendRegenerationRiderLocked runs as a shield on `c` has just been spent
// (its count already lowered): if riders outnumber the shields left, the
// shield that was used carried one, so the oldest rider is removed and its
// player gains control of the permanent (CR 701.19a, "if it regenerates this
// way"), with no duration (CR 611.2).
//
// Caller must hold g.mu (write).
func (g *Game) spendRegenerationRiderLocked(c *Card) {
	idx := g.regenerationRiderIndexesLocked(c)
	if len(idx) == 0 || len(idx) <= c.RegenerationShields {
		return
	}
	rec := g.ScopedEffects[idx[0]]
	var player uuid.UUID
	for _, m := range rec.Mods {
		if m.Kind == ModRegenerationGainsControl {
			player = m.Player
		}
	}
	g.dropScopedEffectsLocked(func(i int, _ ScopedEffect) bool { return i == idx[0] })
	if p := g.playerByIDLocked(player); p == nil || p.Eliminated {
		return
	}
	g.GainControlForEffect(rec.Source.ID, c.InstanceID, player, IndefiniteDuration(), "regenerated this way — gain control")
}
