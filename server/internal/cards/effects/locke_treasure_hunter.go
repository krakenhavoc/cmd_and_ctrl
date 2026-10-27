package effects

import (
	"slices"

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
	// APNAP order (CR 101.4, #4268), not seat zero first.
	seats := apnapPlayers(g)
	return lockeMugStep(NewContext(g, item), seats, 0)
}

// lockeMugStep mills seats[i]'s card, or pays off once every seat has.
// Keyed (#4253 slice 10a): the seats ride the mill's tail as its Carry and
// the index as its N; the cards milled so far are the step record.
func lockeMugStep(ctx *Context, seats []uuid.UUID, i int) error {
	if i >= len(seats) {
		return lockeMugPayoff.Run(ctx)
	}
	// The "moved" form: a card Rest in Peace exiled instead of milling is
	// still one of "those cards" (CR 701.17c).
	return ctx.Game.MillToZoneKeyedMovedForEffect(seats[i], 1, game.ZoneGraveyard,
		withN(withCarry(ctx.ItemTail(lockeMugMilled), seats...), i))
}

// lockeMugMilled records what one seat's mill moved and mills the next.
// Set in init, because the next mill names it again.
var lockeMugMilled game.TailRef

func init() {
	lockeMugMilled = CardTail("then/locke-treasure-hunter-mug-step", func(ctx *Context, o game.TailOutcome) error {
		milled := slices.Clone(ctx.StepRecord(lockeMilledRecord).IDs)
		for _, id := range o.IDs {
			// Graveyard or exile only: a commander its owner moved to the
			// command zone is neither a land to count nor a card a stored
			// permission can open.
			if z := ctx.Game.FindCardZoneForEffect(id); z != nil && (z.Kind == game.ZoneGraveyard || z.Kind == game.ZoneExile) {
				milled = append(milled, id)
			}
		}
		ctx.RecordStep(lockeMilledRecord, milled)
		return lockeMugStep(ctx, o.Tail.Carry, o.Tail.N+1)
	})
}

// lockeMilledRecord is the cards milled so far, for the payoff (#4163).
const lockeMilledRecord = "locke-treasure-hunter/milled"

// lockeMugPayoff is the rest of Mug, over the cards that were milled: the
// Treasure for a land among them and then the cast permission, which waits
// for a prompt the Treasure's creation asks (a CR 616 order on tokens; CR
// 608.2c, #4155).
var lockeMugPayoff = RegisterThenSteps("then-steps/locke-treasure-hunter-payoff",
	Instruction(func(ctx *Context) error {
		for _, id := range ctx.StepRecord(lockeMilledRecord).IDs {
			if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.IsLand() {
				return CreateToken{Controller: ctx.Controller(), Template: TreasureToken(), N: 1}.Apply(ctx)
			}
		}
		return nil
	}),
	Instruction(func(ctx *Context) error {
		var cards []game.Card
		for _, id := range ctx.StepRecord(lockeMilledRecord).IDs {
			if c, ok := ctx.Game.LookupCardForEffect(id); ok {
				cards = append(cards, c)
			}
		}
		ctx.Game.GrantCastPermissionAcrossZonesForEffect(game.CastPermission{
			Player:     ctx.Controller(),
			CastOnly:   true,
			CastsLeft:  1,
			Source:     ctx.Item.SourceCardID,
			SourceName: "Locke, Treasure Hunter",
			Label:      "Locke — cast a spell from among the milled cards",
		}, cards)
		return nil
	}),
)
