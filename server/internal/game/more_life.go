package game

import "github.com/google/uuid"

// more_life.go — #5626: two player restrictions that compare life totals.
//
//	"At the beginning of each player's upkeep, that player chooses target
//	 player who has more life than they do and is their opponent." (Oath of
//	 Mages)
//	"{R}, {T}: Choose target opponent who has more life than you do as you
//	 activate this ability. This creature deals 2 damage to that player."
//	 (Keeper of the Flame)
//
// MoreLifeThanChooser is Oath of Druids' creature count (more_creatures_than_chooser.go)
// read off life: a player other than the clause's chooser, with more life than
// the chooser has, in a free-for-all game every other player still in it being
// an opponent. It is judged on every walk, the CR 608.2b re-check included, as
// the legal set, the announce check and the re-check are one answer.
//
// MoreLifeAsActivated is "as you activate this ability": the totals are read
// only as the ability is announced (CR 602.2b), so it is judged only on the walk
// TargetSource.announcing marks and never by the CR 608.2b re-check. The Exodus
// ruling says the ability still resolves if the life totals change in response.

// moreLifeAdmits gates one player candidate by MoreLifeThanChooser and
// MoreLifeAsActivated: true when neither is set.
//
// Caller must hold g.mu.
func (s *TargetSpec) moreLifeAdmits(g *Game, src TargetSource, p *Player) bool {
	if p == nil {
		return false
	}
	if s.MoreLifeThanChooser {
		who := triggerChooserOf(src)
		if who == uuid.Nil || p.ID == who || p.Eliminated {
			return false
		}
		chooser := g.playerByIDLocked(who)
		if chooser == nil || p.Life <= chooser.Life {
			return false
		}
	}
	if s.MoreLifeAsActivated && src.announcing {
		if p.ID == src.Controller || p.Eliminated {
			return false
		}
		you := g.playerByIDLocked(src.Controller)
		if you == nil || p.Life <= you.Life {
			return false
		}
	}
	return true
}
