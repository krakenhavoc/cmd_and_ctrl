package effects

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b798_helpers.go — S82 OLD batch 798 (Mirage, #5633).

// SacrificeListedObjectsYouControlBody is the delayed body "sacrifice those
// permanents", over objects pinned by CR 400.7 epoch (Will Riker, First in
// the Field's body, shared by key): each one goes only while it is still the
// object it was, on the battlefield, under the trigger controller's control.
var SacrificeListedObjectsYouControlBody = trekB18SacrificeListedObjectsBody

var manaSymbolRE = regexp.MustCompile(`\{[^}]+\}`)

// OldB798ReducedManaCost is a mana cost with `by` generic mana taken off, for
// "its mana cost reduced by {2}" (Flash): only the generic part shrinks and
// never below none (CR 118.7a, 118.7e); {X} counts as 0 on the battlefield
// (CR 107.3i). It is "" when nothing is left to pay.
func OldB798ReducedManaCost(cost string, by int) string {
	generic := 0
	var rest []string
	for _, sym := range manaSymbolRE.FindAllString(cost, -1) {
		inner := strings.Trim(sym, "{}")
		if n, err := strconv.Atoi(inner); err == nil {
			generic += n
			continue
		}
		if inner == "X" {
			continue
		}
		rest = append(rest, sym)
	}
	generic -= by
	if generic < 0 {
		generic = 0
	}
	out := strings.Join(rest, "")
	if generic > 0 {
		out = "{" + strconv.Itoa(generic) + "}" + out
	}
	return out
}

// flashSacrificeUnlessPaid is Flash's rider on the creature it put onto the
// battlefield: "sacrifice it unless you pay its mana cost reduced by {2}".
// The cost is read off the permanent that arrived. Nothing arrived, or
// nothing to pay, asks nothing (a cost of {0} is paid at once).
var flashSacrificeUnlessPaid = PutFromHandTail("put-from-hand/s82-old-b798-flash-sacrifice-unless-paid", func(ctx *Context, res PutFromHandResult) error {
	if res.Entered == uuid.Nil || !OnBattlefield(ctx.Game, res.Entered) {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(res.Entered)
	if !ok {
		return nil
	}
	cost := OldB798ReducedManaCost(c.ManaCost, 2)
	if cost == "" {
		return nil
	}
	return PayUnless{
		Chooser:  ctx.Controller(),
		Cost:     cost,
		Question: "Flash — pay " + cost + " or sacrifice " + c.Name + "?",
		Else:     flashSacrificeDecline,
		Carry:    []uuid.UUID{res.Entered},
	}.Apply(ctx)
})

// flashSacrificeDecline sacrifices the creature Flash put onto the battlefield.
// An on-disk key: never renamed, never reused.
var flashSacrificeDecline = PayUnlessElse("pay-unless/s82-old-b798-flash-sacrifice", func(ctx *Context, r game.PayUnlessAnswer) error {
	for _, id := range r.Carry {
		if OnBattlefield(ctx.Game, id) {
			return SacrificePermanent{Target: id}.Apply(ctx)
		}
	}
	return nil
})

// OldB798Flash is Flash's whole resolution: you may put a creature card from
// your hand onto the battlefield, then (if you did) the pay-or-sacrifice rider.
func OldB798Flash(_ *game.StackItem, ctx *Context) error {
	return PutFromHandOntoBattlefield{
		Match:    Creature(),
		Optional: true,
		Label:    "Flash — you may put a creature card from your hand onto the battlefield",
		ThenTail: flashSacrificeUnlessPaid,
	}.Apply(ctx)
}

// OldB798OpponentCastACreatureSpellThisTurn is Lure of Prey's cast
// restriction: some opponent of the caster has cast a creature spell this turn
// (every spell they cast minus the noncreature ones). Read as the spell is
// announced, never at resolution.
func OldB798OpponentCastACreatureSpellThisTurn(g *game.Game, caster uuid.UUID, _ game.Card) bool {
	return EachOpponent(g, caster, func(opp uuid.UUID) bool {
		t := g.CastTallyFor(opp)
		return t.Total-t.Noncreature > 0
	})
}

// --- Sealed Fate -------------------------------------------------------------

// OldB798SealedFate is "look at the top X cards of target opponent's library.
// Exile one of those cards and put the rest back on top of that library in any
// order": the looker picks the card to exile (face up, CR 406.3), then orders
// the rest. A library of one card exiles it with no choice; an empty one does
// nothing.
func OldB798SealedFate(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	opp := TargetedPlayerOf(ctx)
	if opp == uuid.Nil || ctx.X() <= 0 {
		return nil
	}
	controller, source := item.Controller, item.SourceCardID
	looked := g.LookAtTopOfPlayersLibraryForEffect(controller, opp, ctx.X())
	if len(looked) == 0 {
		return nil
	}
	if len(looked) == 1 {
		return sealedFateFinish(g, controller, source, looked, looked)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    controller,
		FromPlayer: opp,
		Source:     source,
		Question:   "Sealed Fate — exile one of those cards",
		Cards:      looked,
		Min:        1,
		Max:        1,
		LiveZone:   game.ZoneLibrary,
		ThenKey:    sealedFatePicked,
		Carry:      looked,
	})
	return nil
}

// sealedFatePicked is the exiled card chosen; the carry is every card looked
// at. An on-disk key: never renamed, never reused.
var sealedFatePicked = game.RegisterChooseCardsThen("choose-cards/s82-old-b798-sealed-fate-exile", func(g *game.Game, r game.CardsChosen) error {
	return sealedFateFinish(g, r.Chooser, r.Source, r.Carry, r.Picked)
})

func sealedFateFinish(g *game.Game, controller, source uuid.UUID, looked, picked []uuid.UUID) error {
	then := game.KeyedTail{
		Ref: sealedFateOrderRest, Player: controller, Source: source,
		Carry: append(append([]uuid.UUID(nil), picked...), looked...), N: len(picked),
	}
	if len(picked) == 0 {
		return then.Run(g, nil)
	}
	return g.ExileCardsKeyedForEffect(picked, then)
}

// sealedFateOrderRest runs once the exile has settled: the rest of the cards
// looked at (the Carry after the first N, the picked ones) that are still in a
// library go back on top in an order the looker chooses. An on-disk key.
var sealedFateOrderRest = CardTail("then/s82-old-b798-sealed-fate-order-rest", func(ctx *Context, o game.TailOutcome) error {
	carry := o.Tail.Carry
	if o.Tail.N > len(carry) {
		return nil
	}
	took := make(map[uuid.UUID]bool, o.Tail.N)
	for _, id := range carry[:o.Tail.N] {
		took[id] = true
	}
	var rest []uuid.UUID
	for _, id := range CardsStillInALibrary(ctx.Game, carry[o.Tail.N:]) {
		if !took[id] {
			rest = append(rest, id)
		}
	}
	if len(rest) == 0 {
		return nil
	}
	return ctx.Game.PutInLibraryInChosenOrderThenForEffect(game.PutInLibrarySpec{
		Chooser:   o.Tail.Player,
		Source:    o.Tail.Source,
		Cards:     rest,
		From:      game.ZoneLibrary,
		Placement: game.LibraryPlaceTop,
		Reason:    LibraryOrderLabel(ctx.Game, o.Tail.Source, "put the rest back on top of that library in any order"),
	})
})

// --- Natural Balance ---------------------------------------------------------

func naturalBalanceLands(g *game.Game, seat uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == seat && c.IsLand() {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// OldB798NaturalBalance is Natural Balance: each player who controls six or
// more lands chooses five of them (each such player asked once, in turn order)
// and, when all have answered, the rest are sacrificed at once (CR 701.21a);
// then each player who controls four or fewer lands may search for up to
// 5 - (their lands) basic land cards, put onto the battlefield, and shuffles if
// they searched. A land count below five means a search of five minus it; a
// player with five lands, or one who declines, does nothing.
func OldB798NaturalBalance(g *game.Game, item *game.StackItem) error {
	var legs []game.ChooseCardsPrompt
	var seats []uuid.UUID
	for _, seat := range ApnapLivingSeats(g) {
		lands := naturalBalanceLands(g, seat)
		if len(lands) < 6 {
			continue
		}
		seats = append(seats, seat)
		legs = append(legs, game.ChooseCardsPrompt{
			Chooser:    seat,
			FromPlayer: seat,
			Source:     item.SourceCardID,
			Question:   "Natural Balance — choose five lands you control; the rest are sacrificed",
			Cards:      lands,
			Min:        5,
			Max:        5,
			Zone:       game.ZoneBattlefield,
		})
	}
	if len(legs) == 0 {
		return naturalBalanceSearch(NewContext(g, item))
	}
	return g.ChoosePermanentsInOrderKeyedForEffect(legs, naturalBalanceKeep, item.SourceCardID,
		game.RunData{Carry: seats, Asked: game.NewAskedItem(item)})
}

// naturalBalanceKeep is "sacrifices the rest", keyed: Carry is the seats that
// chose. An on-disk key.
var naturalBalanceKeep = game.RegisterPromptRunThen("prompt-run/s82-old-b798-natural-balance-keep", func(g *game.Game, r game.RunDone) error {
	kept := map[uuid.UUID]bool{}
	for _, id := range r.Picks().Cards() {
		kept[id] = true
	}
	var rest []uuid.UUID
	for _, seat := range r.Carry {
		for _, id := range naturalBalanceLands(g, seat) {
			if !kept[id] {
				rest = append(rest, id)
			}
		}
	}
	item := r.Item()
	if item == nil {
		return nil
	}
	if len(rest) == 0 {
		return naturalBalanceSearch(NewContext(g, item))
	}
	return g.SacrificeAllKeyedForEffect(r.Source, rest, game.KeyedTail{
		Ref: naturalBalanceSearchTail, Player: item.Controller, Source: r.Source, Asked: r.Asked,
	})
})

var naturalBalanceSearchTail = CardTail("then/s82-old-b798-natural-balance-search", func(ctx *Context, _ game.TailOutcome) error {
	return naturalBalanceSearch(ctx)
})

// naturalBalanceSearch is the second sentence: each player, in turn order,
// with four or fewer lands may search for up to five minus their lands.
func naturalBalanceSearch(ctx *Context) error {
	for _, seat := range ApnapLivingSeats(ctx.Game) {
		n := len(naturalBalanceLands(ctx.Game, seat))
		if n > 4 {
			continue
		}
		if err := (SearchLibrary{
			Player:    seat,
			Predicate: IsBasicLand,
			Dest:      game.ZoneBattlefield,
			Limit:     5 - n,
			Optional:  true,
			Shuffle:   true,
			Reason:    "Natural Balance — up to " + strconv.Itoa(5-n) + " basic land cards",
			Source:    ctx.Source(),
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
