package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Krang & Shredder — Legendary Creature — Utrom Human Ninja
// {4}{U/B}{U/B}, 6/7 (Teenage Mutant Ninja Turtles, #4457):
//
//	"Whenever Krang & Shredder enter or attack, each opponent exiles
//	 cards from the top of their library until they exile a nonland card.
//	 Disappear — At the beginning of your end step, if a permanent left
//	 the battlefield under your control this turn, you may cast a card
//	 exiled with Krang & Shredder without paying its mana cost."
//
// Etali, Primal Conqueror's walk, one opponent after another in turn
// order, but each card is exiled LINKED to Krang & Shredder (CR 607.2a,
// ExileTarget.Linked) one at a time from the top, so the second ability
// can find them: the run for an opponent stops at the first nonland card
// that reaches exile, and an opponent whose library runs out exiles all
// of it. The lands exiled on the way are linked too, and stay exiled.
//
// Disappear is an ability word (CR 207.2c): it is revolt's condition
// (AtYourEndStepIfRevolt), checked as the trigger would go on the stack
// and again as it resolves (CR 603.4). It offers one free cast from among
// every card still in exile that this Krang & Shredder object exiled,
// from any of its triggers, during the resolution (CR 608.2g): one
// permission over the cards with one cast to spend, as Court of
// Locthwain's is, and a card not cast stays in exile. A Krang & Shredder
// that left and came back is a new object with no claim on the cards
// the old one exiled (CR 400.7). A spell with {X} in its cost is cast
// with X = 0 (CR 107.3b), and additional costs are still paid.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2385c8fb-9c38-4ff3-8f61-4e25a8c7d46b",
		Name:         "Krang & Shredder",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Krang & Shredder — each opponent exiles cards from the top of their library until they exile a nonland card",
				func(g *game.Game, item *game.StackItem) error {
					var opponents []uuid.UUID
					for _, id := range apnapLivingSeats(g) {
						if id != item.Controller {
							opponents = append(opponents, id)
						}
					}
					return krangAndShredderExileFrom(NewContext(g, item), opponents)
				}),
			AtYourEndStepIfRevolt("Krang & Shredder — you may cast a card exiled with Krang & Shredder without paying its mana cost",
				krangAndShredderFreeCast),
		},
	})
}

const krangAndShredderName = "Krang & Shredder"

// krangAndShredderExileFrom exiles the top card of the first player's
// library, linked to the source; the tail decides whether that player
// goes on. A player with an empty library is passed over.
func krangAndShredderExileFrom(ctx *Context, players []uuid.UUID) error {
	for len(players) > 0 {
		p := ctx.Game.PlayerByIDForEffect(players[0])
		if p == nil || p.Eliminated || p.Library == nil || p.Library.Size() == 0 {
			players = players[1:]
			continue
		}
		top, err := p.Library.Top()
		if err != nil {
			players = players[1:]
			continue
		}
		carry := append([]uuid.UUID{top.InstanceID}, players...)
		return ExileTarget{
			Target:   top.InstanceID,
			Linked:   true,
			ThenTail: withCarry(ctx.ItemTail(krangAndShredderExiled), carry...),
		}.Apply(ctx)
	}
	return nil
}

// krangAndShredderExiled is the clause after one card's exile: a nonland
// card that reached exile ends that player's run. A card that never left
// the library ends it too, so the walk cannot repeat itself. An on-disk
// key, registered in init because the walk asks again from inside it.
var krangAndShredderExiled game.TailRef

func init() {
	krangAndShredderExiled = CardTail("krang-and-shredder/after-exile", krangAndShredderAfterExile)
}

func krangAndShredderAfterExile(ctx *Context, o game.TailOutcome) error {
	if len(o.Tail.Carry) < 2 {
		return nil
	}
	card, players := o.Tail.Carry[0], o.Tail.Carry[1:]
	done := false
	if len(o.IDs) > 0 {
		c, ok := ctx.Game.LookupCardForEffect(card)
		done = ok && !c.IsLand()
	} else if z := ctx.Game.FindCardZoneForEffect(card); z != nil && z.Kind == game.ZoneLibrary {
		done = true
	}
	if done {
		players = players[1:]
	}
	return krangAndShredderExileFrom(ctx, players)
}

// krangAndShredderFreeCast is Disappear's resolution: offer one free cast
// from among the cards in exile linked to this object.
func krangAndShredderFreeCast(g *game.Game, item *game.StackItem) error {
	ref, ok := g.SourceObjectForEffect(item)
	if !ok {
		return nil
	}
	var cards []game.Card
	var ids []uuid.UUID
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		if c.IsLand() {
			continue
		}
		cards = append(cards, *c)
		ids = append(ids, c.InstanceID)
	}
	if len(cards) == 0 {
		return nil
	}
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:      item.Controller,
		Zone:        game.ZoneExile,
		Duration:    g.UntilEndOfTurnDuration(),
		Cost:        "{0}",
		Timing:      game.TimingFlash,
		CastOnly:    true,
		CastsLeft:   1,
		LapseOnPass: game.LapseStaysInExile,
		Source:      item.SourceCardID,
		SourceName:  krangAndShredderName,
		Label:       krangAndShredderName + " — cast a card exiled with Krang & Shredder without paying its mana cost",
	}, cards)
	offerResolutionCastOver(g, item.Controller, item.SourceCardID, krangAndShredderName,
		game.MayCastKeywordFreeCast, declineLeaveInExile, ids...)
	return nil
}
