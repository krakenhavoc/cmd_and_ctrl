package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Locke, Treasure Hunter — Legendary Creature — Human Rogue {1}{B}{R},
// 2/3 (#1729):
//
//	"Locke can't be blocked by creatures with greater power.
//	 Mug — Whenever Locke attacks, each player mills a card. If a land
//	 card was milled this way, create a Treasure token. Until end of
//	 turn, you may cast a spell from among those cards."
//
// The evasion is skulk's rule without the keyword (CR 702.118b's
// comparison, read off the attacker and the blocker as the block is
// declared), so it is a block rule of its own rather than a granted
// "skulk": the card does not have that keyword.
//
// Mug mills every player still in the game, one card each, and waits
// for each mill to settle (a commander can stop on the CR 903.9 prompt,
// and one that went to the command zone is no longer in a public zone to
// be cast from, so it is not counted). Then a Treasure if any of those cards is a land — one
// Treasure however many lands ("Additional land cards milled beyond the
// first won't cause you to create additional Treasures", ruling
// 2025-06-06) — and one permission over the cards, wherever they are:
// ScopeCards names the objects, so a card in an opponent's graveyard is
// castable from there (#1022). "A spell" is CastsLeft 1 (#1729): the
// first spell cast with it spends it. The rulings say the rest:
//
//   - "Losing control of Locke after the last ability has triggered
//     won't affect your ability to cast a spell from among the milled
//     cards." The permission is the player's, not Locke's.
//   - "If you cast a spell from among the milled cards using another
//     permission, Locke's effect doesn't apply." A card whose own text
//     opens its graveyard is cast by that text and spends nothing.
//   - "You must follow the normal timing permissions and restrictions
//     of the spell you cast" and "You must pay the costs to cast that
//     spell": the permission changes neither (TimingNormal, no price).
//
// Where a milled card went does not matter (CR 701.17c): a card a
// replacement exiled instead of milling into a graveyard (Rest in Peace)
// is still "those cards", and the permission covers it in exile. The mill
// reports every card it took out of the library
// (MillToZoneThenMovedForEffect), and the grant is one permission per
// zone those cards are in, sharing a single "a spell" (BudgetGroup).
func init() {
	Register(Spec{
		OracleID:     "a800bec4-bacc-43f9-a773-dd795959935b",
		Name:         "Locke, Treasure Hunter",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			CantBeBlockedByComparing(OnSelf(), GreaterPowerThanAttacker(), "creatures with greater power"),
		},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Locke, Treasure Hunter — Mug: each player mills a card", lockeMug),
		},
	})
}

// lockeMug mills each player in turn, collecting what landed, then
// finishes the ability.
func lockeMug(g *game.Game, item *game.StackItem) error {
	var seats []uuid.UUID
	for _, p := range g.Seats {
		if p != nil && !p.Eliminated {
			seats = append(seats, p.ID)
		}
	}
	var milled []uuid.UUID
	var step func(g *game.Game, i int) error
	step = func(g *game.Game, i int) error {
		if i == len(seats) {
			return lockeMugPayoff(g, item, milled)
		}
		// The "moved" form: a card Rest in Peace exiled instead of milling
		// is still one of "those cards" (CR 701.17c).
		return g.MillToZoneThenMovedForEffect(seats[i], 1, game.ZoneGraveyard, func(g *game.Game, moved []uuid.UUID) error {
			for _, id := range moved {
				// Graveyard or exile only: a commander its owner moved
				// to the command zone is neither a land to count nor a
				// card a stored permission can open.
				if z := g.FindCardZoneForEffect(id); z != nil && (z.Kind == game.ZoneGraveyard || z.Kind == game.ZoneExile) {
					milled = append(milled, id)
				}
			}
			return step(g, i+1)
		})
	}
	return step(g, 0)
}

// lockeMugPayoff is the rest of Mug, over the cards that were milled.
func lockeMugPayoff(g *game.Game, item *game.StackItem, milled []uuid.UUID) error {
	ctx := NewContext(g, item)
	var cards []game.Card
	land := false
	for _, id := range milled {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			continue
		}
		cards = append(cards, c)
		land = land || c.IsLand()
	}
	if land {
		if err := (CreateToken{Controller: item.Controller, Template: TreasureToken(), N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	g.GrantCastPermissionAcrossZonesForEffect(game.CastPermission{
		Player:     item.Controller,
		CastOnly:   true,
		CastsLeft:  1,
		Source:     item.SourceCardID,
		SourceName: "Locke, Treasure Hunter",
		Label:      "Locke — cast a spell from among the milled cards",
	}, cards)
	return nil
}
