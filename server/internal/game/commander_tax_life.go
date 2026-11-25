package game

import "github.com/google/uuid"

// commander_tax_life.go — "Rather than pay {2} for each previous time
// you've cast this spell from the command zone this game, pay 2 life that
// many times." (Liesa, Shroud of Dusk, #6151).
//
// CR 903.8: "A commander cast from the command zone costs an additional
// {2} for each previous time the player casting it has cast it from the
// command zone that game. This additional cost is informally known as the
// 'commander tax.'" The card replaces that additional cost: the same count
// of previous casts (Player.CommanderCasts) charges 2 life each instead of
// {2}. It stays an additional cost of the cast, so it is paid under an
// alternative cost too (CR 118.9d), checked against the caster's life with
// every other life cost of the cast (CR 119.4), and paid with them while
// the spell is on the stack.
//
// Catalog data on the card's definition (CardDef.CommanderTaxInLife), read
// through CommanderTaxInLifeFor. Nothing is added to game state: the count
// is the existing CommanderCasts tally.

// CatalogCommanderTaxInLife is the catalog hook for a card whose commander
// tax is paid in life; carddef.go's init wires it. Nil (no catalog) means
// no card does.
var CatalogCommanderTaxInLife func(oracleID string) bool

// CommanderTaxInLifeFor reports whether the card keyed `key` pays its
// commander tax in life.
func CommanderTaxInLifeFor(key string) bool {
	return CatalogCommanderTaxInLife != nil && key != "" && CatalogCommanderTaxInLife(key)
}

// CommanderTaxLifePerCast is the life each previous cast charges in place
// of {2}.
const CommanderTaxLifePerCast = 2

// commanderTaxLocked is the commander tax a cast of `card` by `p` from
// `fromZone` (the cast_spell wire name) owes: `generic` mana for an
// ordinary commander, or `life` for one whose tax is paid in life. Both are
// zero off the command zone and on a first cast. Caller must hold g.mu.
func commanderTaxLocked(p *Player, card Card, fromZone string) (generic, life int) {
	if p == nil || fromZone != "command" {
		return 0, 0
	}
	casts := p.CommanderCasts[card.InstanceID]
	if casts <= 0 {
		return 0, 0
	}
	if CommanderTaxInLifeFor(CatalogKey(card)) {
		return 0, casts * CommanderTaxLifePerCast
	}
	return casts * 2, 0
}

// CommanderTaxLifeForEffect is the life `playerID` pays as commander tax to
// cast `card` from `fromZone` (0 for every card but one whose tax is paid
// in life). The view's price and the bot enumerator read it, so neither
// offers a cast the life total refuses. Caller must hold g.mu.
func (g *Game) CommanderTaxLifeForEffect(playerID uuid.UUID, card Card, fromZone string) int {
	_, life := commanderTaxLocked(g.playerByIDLocked(playerID), card, fromZone)
	return life
}
