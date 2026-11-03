package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Plargg and Nassari — Legendary Creature — Orc Efreet {3}{R}{R}, 5/4
// (Secrets of Strixhaven Commander, #4834):
//
//	"At the beginning of your upkeep, each player exiles cards from the top
//	 of their library until they exile a nonland card. An opponent chooses a
//	 nonland card exiled this way. You may cast up to two spells from among
//	 the other cards exiled this way without paying their mana costs."
//
// Etali, Primal Conqueror's walk: each player in turn order exiles from the
// top of their library until a nonland card (MillToZone with UntilNonland, so
// a commander offers its owner the command zone), one library after another,
// each exile waiting on a keyed tail. The lands exiled on the way stay exiled,
// and a player with no nonland card in their library exiles the whole library
// and adds nothing. The nonland cards are then put to an opponent: with one
// opponent they choose, with several the controller names which one chooses,
// as Coin of Fate does. The chosen card is not cast; one
// permission over the others, with two casts to spend (CastPermission.CastsLeft,
// Baron Helmut Zemo's three), lets the controller cast up to two of them for
// {0}. The casts are made during the resolution (CR 608.2g), each offered in
// turn through the resolution_cast prompt until two are made or the player
// declines; a spell with {X} in its cost is cast with X = 0 (CR 107.3b) and a
// mandatory additional cost is still paid. A card not cast stays in exile.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9e769a65-7aa5-484d-9493-badb485b5964",
		Name:         "Plargg and Nassari",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Plargg and Nassari — each player exiles until a nonland card; an opponent chooses one; you may cast up to two of the others free",
				func(g *game.Game, item *game.StackItem) error {
					return socB10PlarggExileFrom(NewContext(g, item), apnapLivingSeats(g), nil)
				}),
		},
	})
}

const socB10PlarggName = "Plargg and Nassari"

// socB10PlarggExileFrom exiles from the first player's library until a
// nonland card, then moves on to the next, and puts the cards to an opponent
// at the end. `hits` is the nonland cards exiled so far.
func socB10PlarggExileFrom(ctx *Context, players, hits []uuid.UUID) error {
	if len(players) == 0 {
		return socB10PlarggPutToOpponent(ctx, hits)
	}
	rest := players[1:]
	return MillToZone{
		Player:   players[0],
		To:       game.ZoneExile,
		Until:    UntilNonland,
		ThenTail: withN(withCarry(ctx.ItemTail(socB10PlarggMilled), append(append([]uuid.UUID(nil), rest...), hits...)...), len(rest)),
	}.Apply(ctx)
}

// socB10PlarggMilled is the clause after one player's exile: the nonland card
// is added to the hits and the next library is walked.
var socB10PlarggMilled game.TailRef

func init() {
	socB10PlarggMilled = CardTail("plargg-and-nassari/after-mill", func(ctx *Context, o game.TailOutcome) error {
		rest, hits := o.Tail.Carry[:o.Tail.N], o.Tail.Carry[o.Tail.N:]
		next := append([]uuid.UUID(nil), hits...)
		for _, id := range o.IDs {
			if c, ok := ctx.Game.LookupCardForEffect(id); ok && !c.IsLand() {
				next = append(next, id)
			}
		}
		return socB10PlarggExileFrom(ctx, rest, next)
	})
}

// socB10PlarggPutToOpponent has an opponent choose one of the exiled nonland
// cards. "An opponent chooses": the controller names which opponent when there
// are several. With one nonland card there is nothing left to cast, so no one
// is asked.
func socB10PlarggPutToOpponent(ctx *Context, hits []uuid.UUID) error {
	if len(hits) < 2 {
		return nil
	}
	opps := ctx.Opponents()
	switch len(opps) {
	case 0:
		return nil
	case 1:
		return socB10PlarggAsk(ctx.Game, ctx.Controller(), ctx.Source(), hits, opps[0])
	}
	return ChoosePlayer{
		Among:    Opponents,
		Question: socB10PlarggName + " — choose an opponent to choose one of the exiled cards",
		ThenKey:  socB10PlarggOpponentChosen,
		Carry:    hits,
	}.Apply(ctx)
}

// socB10PlarggOpponentChosen runs once the controller has named the opponent.
// An on-disk key: never renamed, never reused.
var socB10PlarggOpponentChosen = ChoosePlayerThen("option-pick/plargg-and-nassari-opponent", func(ctx *Context, chosen uuid.UUID, hits []uuid.UUID) error {
	if chosen == uuid.Nil {
		return nil
	}
	return socB10PlarggAsk(ctx.Game, ctx.Controller(), ctx.Source(), hits, chosen)
})

// socB10PlarggAsk puts the "choose a nonland card" prompt to the opponent. The
// controller is the prompt's FromPlayer and the exiled cards its Carry, so the
// prompt holds no closure.
func socB10PlarggAsk(g *game.Game, controller, source uuid.UUID, hits []uuid.UUID, chooser uuid.UUID) error {
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    chooser,
		FromPlayer: controller,
		Source:     source,
		Question:   socB10PlarggName + " — choose a nonland card exiled this way; its controller may cast up to two of the other cards free",
		Cards:      hits,
		Min:        1,
		Max:        1,
		ThenKey:    socB10PlarggPicked,
		LiveZone:   game.ZoneExile,
		Carry:      hits,
	})
	return nil
}

// socB10PlarggPicked is the keyed answer: one permission over the other cards
// still in exile, good for two casts, offered during the resolution. An
// on-disk key: never renamed, never reused.
var socB10PlarggPicked = game.RegisterChooseCardsThen("choose-cards/plargg-and-nassari-picked", func(g *game.Game, r game.CardsChosen) error {
	controller, source := r.From, r.Source
	picked := uuid.Nil
	if len(r.Picked) > 0 {
		picked = r.Picked[0]
	}
	var others []game.Card
	var ids []uuid.UUID
	for _, id := range r.Carry {
		if id == picked {
			continue
		}
		z := g.FindCardZoneForEffect(id)
		c, ok := g.LookupCardForEffect(id)
		if !ok || z == nil || z.Kind != game.ZoneExile {
			continue
		}
		others = append(others, c)
		ids = append(ids, id)
	}
	if len(others) == 0 {
		return nil
	}
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:      controller,
		Zone:        game.ZoneExile,
		Duration:    g.UntilEndOfTurnDuration(),
		Cost:        "{0}",
		Timing:      game.TimingFlash,
		CastOnly:    true,
		CastsLeft:   2,
		AnyNumber:   true,
		LapseOnPass: game.LapseStaysInExile,
		Source:      source,
		SourceName:  socB10PlarggName,
		Label:       socB10PlarggName + " — cast up to two of the other exiled cards without paying their mana costs",
	}, others)
	offerResolutionCastOver(g, controller, source, socB10PlarggName,
		game.MayCastKeywordFreeCast, declineLeaveInExile, ids...)
	return nil
})
