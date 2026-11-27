package game

import "github.com/google/uuid"

// more_lands_than_chooser.go — S82 OLD batch 743: Oath of Lieges' "that
// player chooses target player who controls more lands than they do and is
// their opponent", the sibling of Oath of Druids' creature count
// (more_creatures_than_chooser.go). Both halves are relative to the player who
// chooses (the upkeep player), not to the Oath's controller.

// moreLandsThanChooserAdmits is MoreLandsThanChooser's gate on one player
// candidate: true without the flag; with it, a player other than the clause's
// chooser who controls more lands than the chooser does. A walk with no
// chooser admits nobody.
//
// Caller must hold g.mu.
func (s *TargetSpec) moreLandsThanChooserAdmits(g *Game, src TargetSource, p *Player) bool {
	if !s.MoreLandsThanChooser {
		return true
	}
	who := triggerChooserOf(src)
	if who == uuid.Nil || p == nil || p.ID == who || p.Eliminated {
		return false
	}
	return g.landsControlledByLocked(p.ID) > g.landsControlledByLocked(who)
}

// landsControlledByLocked is how many lands `player` controls, by the types
// after the layers.
//
// Caller must hold g.mu.
func (g *Game) landsControlledByLocked(player uuid.UUID) int {
	if g.Battlefield == nil {
		return 0
	}
	n := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == player && c.IsLand() {
			n++
		}
	}
	return n
}
