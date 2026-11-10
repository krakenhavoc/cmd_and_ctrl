package game

import "github.com/google/uuid"

// standing_permission_live.go — the parts of a STANDING cast permission
// that are facts about the game right now rather than about the granting
// card (#4458, Maralen, Fae Ascendant):
//
//	"Once each turn, you may cast a spell with mana value less than or
//	 equal to the number of Elves and Faeries you control from among
//	 cards exiled with Maralen this turn without paying its mana cost."
//
// A standing permission is derived off the battlefield on every query
// (standingCastPermissionsLocked), so a clause that reads the clock or
// the board is filled in here, as ExiledWith is filled in from the
// granting object: "this turn" is the current Turn.Seq
// (PermissionFilter.ExiledOnTurn), and "the number of Elves and Faeries
// you control" is counted now (MaxSpellManaValueCountOf into
// MaxSpellManaValue). Every surface that asks — CastSpell, the view's
// offers, the bot enumerator — reads the same derived value, so all of
// them judge the cap and the turn the same way.

// fillDerivedLiveLocked fills the live parts of a derived standing
// permission for `holder`. A permission that prints neither clause is
// left as it is.
//
// Caller must hold g.mu.
func (g *Game) fillDerivedLiveLocked(perm *CastPermission, holder uuid.UUID) {
	if perm.Filter.ExiledWithSourceThisTurn {
		perm.Filter.ExiledOnTurn = g.Turn.Seq
	}
	if len(perm.MaxSpellManaValueCountOf) > 0 {
		n := g.countPermanentsWithAnySubtypeLocked(holder, perm.MaxSpellManaValueCountOf)
		perm.MaxSpellManaValue = &n
	}
}

// countPermanentsWithAnySubtypeLocked counts the permanents `controller`
// controls that have at least one of `subtypes` (through the layered
// view, so a changeling counts), each permanent once: "the number of
// Elves and Faeries you control".
//
// Caller must hold g.mu.
func (g *Game) countPermanentsWithAnySubtypeLocked(controller uuid.UUID, subtypes []string) int {
	if g.Battlefield == nil {
		return 0
	}
	n := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller != controller {
			continue
		}
		for _, s := range subtypes {
			if c.HasSubtype(s) {
				n++
				break
			}
		}
	}
	return n
}
