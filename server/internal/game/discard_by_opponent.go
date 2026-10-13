package game

// discard_by_opponent.go — what a card declares about "if a spell or
// ability an opponent controls causes you to discard" (ADR 0153 §5).
//
// The bot prices a discard an opponent forced differently from a loot of
// its own: a Baloth pitched to Mind Rot comes back as a free 4/4. The
// fact has to reach the wire as data the catalog constructor declared,
// never as something the heuristic infers from a name or from oracle
// text (ADR 0126), so the card declares which of three shapes it is.

// DiscardByOpponentKind names what an opponent-caused discard does for
// the card, in the three shapes the 16 cards fall into.
type DiscardByOpponentKind string

const (
	// DiscardByOpponentNone is every card that prints no such clause.
	DiscardByOpponentNone DiscardByOpponentKind = ""
	// DiscardByOpponentBattlefield: the discard puts the card onto the
	// battlefield instead of the graveyard (Obstinate Baloth, Dodecapod).
	DiscardByOpponentBattlefield DiscardByOpponentKind = "battlefield"
	// DiscardByOpponentLibrary: the discard puts the card on top of its
	// owner's library instead (Nephalia Academy, which declares it for
	// every card in its controller's hand).
	DiscardByOpponentLibrary DiscardByOpponentKind = "library"
	// DiscardByOpponentTrigger: the discard triggers an ability of the
	// card (Guerrilla Tactics, Sand Golem).
	DiscardByOpponentTrigger DiscardByOpponentKind = "trigger"
)

// CatalogDiscardByOpponent returns what the card with this catalog key
// declares: `own` for the card itself in a hand, and `forHand` for every
// card in its controller's hand while it is on the battlefield (Nephalia
// Academy). Populated at init from CardDef.DiscardByOpponent and
// CardDef.DiscardByOpponentForHand.
var CatalogDiscardByOpponent func(key string) (own, forHand DiscardByOpponentKind)

// DiscardByOpponentOf is the wire value for a card in a hand: its own
// declaration, or, when it has none, the one a permanent its owner
// controls makes for every card in their hand. Empty for no claim.
// Caller must hold g.mu (read or write).
func (g *Game) DiscardByOpponentOf(c *Card) DiscardByOpponentKind {
	if c == nil || CatalogDiscardByOpponent == nil {
		return DiscardByOpponentNone
	}
	if own, _ := CatalogDiscardByOpponent(CatalogKey(*c)); own != DiscardByOpponentNone {
		return own
	}
	if g.Battlefield == nil {
		return DiscardByOpponentNone
	}
	for i := range g.Battlefield.Cards {
		p := &g.Battlefield.Cards[i]
		if p.Controller != c.Owner {
			continue
		}
		key := catalogAbilityKeyOf(p)
		if key == "" {
			continue
		}
		if _, forHand := CatalogDiscardByOpponent(key); forHand != DiscardByOpponentNone {
			return forHand
		}
	}
	return DiscardByOpponentNone
}
