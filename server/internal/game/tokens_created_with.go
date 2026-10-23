package game

import "github.com/google/uuid"

// tokens_created_with.go — #1997: "tokens created with this", the link
// CR 607.1d draws between a token and the object whose ability created
// it. Tombstone Stairwell is the first card:
//
//	"At the beginning of each upkeep, if this enchantment is on the
//	 battlefield, each player creates a 2/2 black Zombie creature token
//	 with haste named Tombspawn for each creature card in their
//	 graveyard.
//	 At the beginning of each end step and when this enchantment leaves
//	 the battlefield, destroy all tokens created with this enchantment.
//	 They can't be regenerated."
//
// The link is on the TOKEN (Card.CreatedWith), stamped as it is minted
// from TokenGroup.CreatedWith, for the reason Card.ExiledWith is on the
// exiled card: the set has to survive a restore, a change of control
// and the source leaving, and a list on the source would have to be
// pruned on every token's exit. It names the source OBJECT, so a
// Stairwell that leaves and comes back is a new object whose end-step
// trigger leaves the old one's Tombspawn alone (CR 400.7), while the
// leaves-the-battlefield trigger of the one that left still finds them
// through its own item's SourceObject.

// TokensCreatedWithForEffect lists the tokens on the battlefield linked
// to the source OBJECT `source` (Card.CreatedWith == source), in
// battlefield order, whoever controls them now. A zero ref lists
// nothing.
//
// Caller must hold g.mu (read or write).
func (g *Game) TokensCreatedWithForEffect(source PermissionCardRef) []uuid.UUID {
	if g.Battlefield == nil || source.ID == uuid.Nil {
		return nil
	}
	var out []uuid.UUID
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if IsTokenPtr(c) && c.CreatedWith == source {
			out = append(out, c.InstanceID)
		}
	}
	return out
}
