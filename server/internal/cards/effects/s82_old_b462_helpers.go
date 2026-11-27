package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b462_helpers.go — New Phyrexia, batch 462 (S82, #5575).

// --- Postmortem Lunge --------------------------------------------------------

// PostmortemLungeReturn is "Return target creature card with mana value X
// from your graveyard to the battlefield. It gains haste. Exile it at the
// beginning of the next end step.": the keyed graveyard return, whose tail is
// told the permanent that entered (so a Clone's copy prompt or a replacement
// that pauses the entry is a restore point and the rest still happens).
func PostmortemLungeReturn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return g.ReturnFromGraveyardToBattlefieldKeyedForEffect(id, item.Controller, false, ctx.ItemTail(postmortemLungeEntered))
}

// postmortemLungeEntered gives the returned creature haste and schedules its
// exile at the next end step. An on-disk key.
var postmortemLungeEntered = EnteredTail("postmortem-lunge/haste-and-exile", func(ctx *Context, entered uuid.UUID, _ game.TailOutcome) error {
	if entered == uuid.Nil || !OnBattlefield(ctx.Game, entered) {
		return nil
	}
	if err := (GrantKeywordUntilEOT{Target: entered, Keywords: []string{"haste"}, Label: "Postmortem Lunge — haste"}).Apply(ctx); err != nil {
		return err
	}
	return ScheduleDelayedTrigger{
		Label: "Postmortem Lunge — exile the returned creature",
		Cards: []uuid.UUID{entered},
		Body:  ExileListedCardsBody,
	}.Apply(ctx)
})

// --- Auriok Survivors --------------------------------------------------------

// AuriokSurvivorsReturn is "return target Equipment card from your graveyard
// to the battlefield. If you do, you may attach it to this creature": the
// keyed return, then the question once the Equipment really is there.
func AuriokSurvivorsReturn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return g.ReturnFromGraveyardToBattlefieldKeyedForEffect(id, item.Controller, false, ctx.ItemTail(auriokSurvivorsEntered))
}

var auriokSurvivorsEntered = EnteredTail("auriok-survivors/offer-attach", func(ctx *Context, entered uuid.UUID, _ game.TailOutcome) error {
	if entered == uuid.Nil || !OnBattlefield(ctx.Game, entered) || !OnBattlefield(ctx.Game, ctx.Source()) || ctx.IsNewSourceObject(ctx.Source()) {
		return nil
	}
	return MayChoice{
		Question: "Auriok Survivors — attach the returned Equipment to it?",
		YesLabel: "Attach",
		NoLabel:  "Don't attach",
		Then:     auriokSurvivorsAttach,
		Carry:    []uuid.UUID{entered},
	}.Apply(ctx)
})

var auriokSurvivorsAttach = MayYes("confirm/s82-old-b462-auriok-survivors-attach", func(ctx *Context, r game.ConfirmAnswer) error {
	if len(r.Carry) == 0 || !OnBattlefield(ctx.Game, r.Carry[0]) || !OnBattlefield(ctx.Game, ctx.Source()) || ctx.IsNewSourceObject(ctx.Source()) {
		return nil
	}
	return ctx.Game.AttachForEffect(r.Carry[0], game.TargetRef{Kind: game.TargetCard, ID: ctx.Source()})
})

// --- Slag Fiend --------------------------------------------------------------

// ArtifactCardsInAllGraveyards is the count Slag Fiend's power and toughness
// are each equal to: the artifact cards in every player's graveyard.
func ArtifactCardsInAllGraveyards(g *game.Game) int {
	n := 0
	for _, p := range g.Seats {
		if p == nil || p.Graveyard == nil {
			continue
		}
		for _, c := range p.Graveyard.Cards {
			if c.IsArtifact() {
				n++
			}
		}
	}
	return n
}

// --- The Chancellors ---------------------------------------------------------

// CounterThatSpellUnlessItsCasterPays is "counter <the spell the triggering
// event cast> unless that player pays <cost>" (Chancellor of the Annex's two
// lines): the caster is asked the CR 118.12 question; a spell that has already
// left the stack raises none.
func CounterThatSpellUnlessItsCasterPays(ctx *Context, cost, label string) error {
	ev := ctx.Trigger().Event
	if ev.CardID == uuid.Nil || ctx.Game.StackItemForEffect(ev.CardID) == nil {
		return nil
	}
	return CounterUnlessPaid{StackID: ev.CardID, Cost: cost, Question: label}.Apply(ctx)
}

// --- Melira, Sylvok Outcast --------------------------------------------------

// CantGetPoisonCountersReplacement is "You can't get poison counters": a
// pure cancel of every poison placement onto the source's controller
// (CR 101.2), the permanent form of Melira, the Living Cure's rider.
func CantGetPoisonCountersReplacement(label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventCounterPlaced},
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			return ev.Kind == game.RepEventCounter && ev.CounterName == game.CounterPoison &&
				ev.CounterPlayer == src.Controller && ev.CounterDelta > 0
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.Cancel()
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		PureCancel: true,
		Label:      label,
	}
}

// CreaturesYouControlCantHaveMinusOneCounters is "Creatures you control can't
// have -1/-1 counters put on them" (Darksteel Angel, Melira, Sylvok Outcast):
// a pure cancel of every -1/-1 placement onto a creature the source's
// controller controls.
func CreaturesYouControlCantHaveMinusOneCounters(label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventCounterPlaced},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventCounter || ev.CounterName != game.CounterMinusOne || ev.CounterDelta <= 0 {
				return false
			}
			target, ok := g.LookupCardForEffect(ev.CounterTarget)
			return ok && target.IsCreature() && target.Controller == src.Controller
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.Cancel()
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		PureCancel: true,
		Label:      label,
	}
}
