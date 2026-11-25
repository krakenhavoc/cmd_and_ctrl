package game

import "github.com/google/uuid"

// draw_reveal.go — #6238: "The chosen player reveals each card they
// draw" (Booby Trap). A static ability, not a trigger: the card is
// revealed as it is drawn (CR 701.20a: "show that card to all players
// for a brief time"), with nothing on the stack. Each permanent whose
// CardDef.RevealsChosenPlayersDraws is set and whose ChosenPlayer is the
// drawer reveals the card once, in its own name. A permanent that has
// lost its abilities reveals nothing (catalogAbilityKeyOf).

// revealDrawnCardLocked reveals `cardID`, just drawn by `player`, for
// every permanent that says so. Caller must hold g.mu in write mode.
func (g *Game) revealDrawnCardLocked(player, cardID uuid.UUID) {
	if g.Battlefield == nil || CatalogLookup == nil {
		return
	}
	for _, src := range g.drawRevealSourcesLocked(player) {
		g.RevealForEffect(RevealSpec{
			Player: player,
			Source: src.id,
			Cards:  []uuid.UUID{cardID},
			Reason: src.name + " — the chosen player reveals each card they draw",
		})
	}
}

type drawRevealSource struct {
	id   uuid.UUID
	name string
}

// drawRevealSourcesLocked is every permanent on the battlefield that
// reveals `player`'s draws, in battlefield order. Collected before any
// reveal so an event's listeners cannot reshape the walk. Caller must
// hold g.mu.
func (g *Game) drawRevealSourcesLocked(player uuid.UUID) []drawRevealSource {
	var out []drawRevealSource
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.ChosenPlayer != player || player == uuid.Nil {
			continue
		}
		d := catalogDef(catalogAbilityKeyOf(c))
		if d == nil || !d.RevealsChosenPlayersDraws {
			continue
		}
		out = append(out, drawRevealSource{id: c.InstanceID, name: EffectiveOf(c).Name})
	}
	return out
}

// RevealsDrawsOfForEffect reports whether some permanent on the
// battlefield reveals `player`'s draws (the view's and a bot's reader).
// Caller must hold g.mu.
func (g *Game) RevealsDrawsOfForEffect(player uuid.UUID) bool {
	return g.Battlefield != nil && CatalogLookup != nil && len(g.drawRevealSourcesLocked(player)) > 0
}
