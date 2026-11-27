package game

import "github.com/google/uuid"

// more_cards_in_hand_than_chooser.go — S82 OLD batch 742: Oath of Scholars'
// "that player chooses target player who has more cards in hand than they do
// and is their opponent", the sibling of Oath of Druids' creature count
// (more_creatures_than_chooser.go). Both halves are relative to the player
// who chooses (the upkeep player), not to the Oath's controller.

// moreCardsInHandThanChooserAdmits is MoreCardsInHandThanChooser's gate on
// one player candidate: true without the flag; with it, a player other than
// the clause's chooser who has more cards in hand than the chooser has. A
// walk with no chooser admits nobody.
//
// Caller must hold g.mu.
func (s *TargetSpec) moreCardsInHandThanChooserAdmits(g *Game, src TargetSource, p *Player) bool {
	if !s.MoreCardsInHandThanChooser {
		return true
	}
	who := triggerChooserOf(src)
	if who == uuid.Nil || p == nil || p.ID == who || p.Eliminated {
		return false
	}
	chooser := g.playerByIDLocked(who)
	if chooser == nil || chooser.Hand == nil || p.Hand == nil {
		return false
	}
	return p.Hand.Size() > chooser.Hand.Size()
}
