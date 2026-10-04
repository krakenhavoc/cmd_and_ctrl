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
// for each mill to land (a commander can stop on the CR 903.9 prompt,
// and one that went to the command zone was not milled into a
// graveyard). Then a Treasure if any of those cards is a land — one
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
// One simplification, declared. CR 701.17c lets an effect that refers to
// a milled card find it in whatever public zone it went to, so a card a
// replacement exiled instead of milling into a graveyard (Rest in Peace)
// is still "those cards". The mill reports only the cards that reached a
// graveyard, so such a card is not covered by the permission. Strictly
// weaker than printed.
func init() {
	Register(Spec{
		OracleID:     "a800bec4-bacc-43f9-a773-dd795959935b",
		Name:         "Locke, Treasure Hunter",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"If a milled card is exiled instead of going to a graveyard, Mug doesn't let you cast it.",
		},
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
		return g.MillToZoneThenForEffect(seats[i], 1, game.ZoneGraveyard, nil, func(g *game.Game, landed []uuid.UUID) error {
			milled = append(milled, landed...)
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
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:     item.Controller,
		Zone:       game.ZoneGraveyard,
		CastOnly:   true,
		CastsLeft:  1,
		Source:     item.SourceCardID,
		SourceName: "Locke, Treasure Hunter",
		Label:      "Locke — cast a spell from among the milled cards",
	}, cards)
	return nil
}
