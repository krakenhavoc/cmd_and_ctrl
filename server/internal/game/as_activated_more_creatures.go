package game

// as_activated_more_creatures.go — #5626: a player restriction that applies
// only as an activated ability is announced.
//
//	"{G}, {T}: Choose target opponent who controls more creatures than you
//	 do as you activate this ability. Create a 2/2 green Beast creature
//	 token." (Keeper of the Beasts)
//
// The words "as you activate this ability" put the count in CR 602.2b's
// announcement (CR 601.2c: targets are chosen then), not in the ability's
// target legality afterwards: the Exodus ruling says the ability still
// resolves if the opponent no longer controls more creatures. So it is read
// only on the walk TargetSource.announcing marks (the activation gate, the
// legal set the picker and the bot see) and never by the CR 608.2b re-check.

// moreCreaturesAsActivatedAdmits is MoreCreaturesAsActivated's gate on one
// player candidate: true without the flag and on any walk but an
// announcement's; on one, a player other than the activating controller
// who controls more creatures than that controller does. In a free-for-all
// game every other player still in the game is an opponent.
//
// Caller must hold g.mu.
func (s *TargetSpec) moreCreaturesAsActivatedAdmits(g *Game, src TargetSource, p *Player) bool {
	if !s.MoreCreaturesAsActivated || !src.announcing {
		return true
	}
	if p == nil || p.ID == src.Controller || p.Eliminated {
		return false
	}
	return g.creaturesControlledByLocked(p.ID) > g.creaturesControlledByLocked(src.Controller)
}
