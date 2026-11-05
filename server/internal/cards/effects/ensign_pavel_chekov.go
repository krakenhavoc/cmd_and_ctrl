package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ensign Pavel Chekov — Legendary Creature — Human Pilot {3}{U}, 1/5 (Star
// Trek, #4973):
//
//	"Strange new worlds — When Ensign Chekov enters, each opponent exiles
//	 the top X cards of their library, where X is the number of
//	 differently named lands you control. For each opponent, choose a land
//	 card that player owns exiled this way. You may play the chosen cards
//	 for as long as they remain exiled."
//
// "Strange new worlds" is an ability word with no rules meaning. X is
// counted once, as the trigger resolves. Each opponent's exile is
// sequenced through a keyed tail (MillToZone to exile, Plargg and
// Nassari's walk), so a commander among the cards offers its owner the
// command zone and what landed is known before anything is chosen. Then,
// one opponent at a time in seat order, Chekov's controller chooses one
// of the land cards that opponent owns among the cards exiled this way
// (mandatory, as printed; an opponent with none exiled gets no prompt).
// Each chosen card gets a play permission (a land may be played with a
// land drop) that lasts while it stays in exile; the cards not chosen
// stay exiled with nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7d7e232f-e4f9-41db-a2d9-cfe88721b7a9",
		Name:         "Ensign Pavel Chekov",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Ensign Pavel Chekov — each opponent exiles the top X cards of their library; choose a land card from each to play",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return trekB6ChekovExileFrom(ctx, ctx.Opponents(), nil, socB23DifferentlyNamedLands(ctx))
				}),
		},
	})
}

const trekB6ChekovName = "Ensign Pavel Chekov"

// trekB6ChekovExileFrom exiles the top x cards of the first opponent's
// library and walks on to the next; `lands` is the land cards exiled so
// far. With every library done, the choosing starts.
func trekB6ChekovExileFrom(ctx *Context, opps, lands []uuid.UUID, x int) error {
	if len(opps) == 0 {
		return trekB6ChekovChooseNext(ctx.Game, ctx.Controller(), ctx.Source(), lands)
	}
	rest := opps[1:]
	tail := withCarry(ctx.ItemTail(trekB6ChekovExiled), append(append([]uuid.UUID(nil), rest...), lands...)...)
	tail.Nums = []int{len(rest), x}
	return MillToZone{Player: opps[0], N: x, To: game.ZoneExile, ThenTail: tail}.Apply(ctx)
}

// trekB6ChekovExiled is the clause after one opponent's exile: the land
// cards that landed are added and the next library is walked. An on-disk
// key: never renamed, never reused.
var trekB6ChekovExiled game.TailRef

func init() {
	trekB6ChekovExiled = CardTail("ensign-pavel-chekov/after-exile", trekB6ChekovAfterExile)
}

func trekB6ChekovAfterExile(ctx *Context, o game.TailOutcome) error {
	nRest, x := 0, 0
	if len(o.Tail.Nums) == 2 {
		nRest, x = o.Tail.Nums[0], o.Tail.Nums[1]
	}
	if nRest > len(o.Tail.Carry) {
		nRest = len(o.Tail.Carry)
	}
	rest := append([]uuid.UUID(nil), o.Tail.Carry[:nRest]...)
	lands := append([]uuid.UUID(nil), o.Tail.Carry[nRest:]...)
	for _, id := range o.IDs {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.IsLand() {
			lands = append(lands, id)
		}
	}
	return trekB6ChekovExileFrom(ctx, rest, lands, x)
}

// trekB6ChekovChooseNext asks Chekov's controller to choose one land card
// owned by the next opponent among `lands` still in exile, carrying the
// rest for the opponents after.
func trekB6ChekovChooseNext(g *game.Game, controller, source uuid.UUID, lands []uuid.UUID) error {
	var owner uuid.UUID
	var group, rest []uuid.UUID
	for _, id := range lands {
		z := g.FindCardZoneForEffect(id)
		c, ok := g.LookupCardForEffect(id)
		if !ok || z == nil || z.Kind != game.ZoneExile {
			continue
		}
		if owner == uuid.Nil {
			owner = c.Owner
		}
		if c.Owner == owner {
			group = append(group, id)
		} else {
			rest = append(rest, id)
		}
	}
	if len(group) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    controller,
		FromPlayer: controller,
		Source:     source,
		Question:   trekB6ChekovName + " — choose a land card that opponent owns exiled this way; you may play it for as long as it remains exiled",
		Cards:      group,
		Min:        1,
		Max:        1,
		ThenKey:    trekB6ChekovPicked,
		LiveZone:   game.ZoneExile,
		Carry:      rest,
	})
	return nil
}

// trekB6ChekovPicked grants the play permission over the chosen card and
// asks about the next opponent's lands. An on-disk key: never renamed,
// never reused. Assigned in init: its body queues the next prompt, which
// names it.
var trekB6ChekovPicked game.ChooseCardsThen

func init() {
	trekB6ChekovPicked = game.RegisterChooseCardsThen("choose-cards/ensign-pavel-chekov-picked", trekB6ChekovGrantAndContinue)
}

func trekB6ChekovGrantAndContinue(g *game.Game, r game.CardsChosen) error {
	for _, id := range r.Picked {
		z := g.FindCardZoneForEffect(id)
		c, ok := g.LookupCardForEffect(id)
		if !ok || z == nil || z.Kind != game.ZoneExile {
			continue
		}
		g.GrantCastPermissionToCardsForEffect(game.CastPermission{
			Player:     r.From,
			Zone:       game.ZoneExile,
			Duration:   game.WhileInZoneDuration(),
			Source:     r.Source,
			SourceName: trekB6ChekovName,
			Label:      trekB6ChekovName + " — you may play this card for as long as it remains exiled",
		}, []game.Card{c})
	}
	return trekB6ChekovChooseNext(g, r.From, r.Source, r.Carry)
}
