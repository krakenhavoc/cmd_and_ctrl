package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dyadrine, Synthesis Amalgam — Legendary Artifact Creature — Construct
// {X}{G}{W}, 0/1 (Edge of Eternities):
//
//	"Trample
//	 Dyadrine enters with a number of +1/+1 counters on it equal to the
//	 amount of mana spent to cast it.
//	 Whenever you attack, you may remove a +1/+1 counter from each of two
//	 creatures you control. If you do, draw a card and create a 2/2
//	 colorless Robot artifact creature token."
//
// The entry counters are Kurbis' clause over the cast (ManaSpent); it
// reads what was really paid, so a cost reduction, the commander tax or a
// free cast are all counted as they happened and a copy gets none (the
// rulings). "Whenever you attack" is one trigger per attack declaration.
// The "you may" is Aegis Sculptor's "exactly two, or none" pick over two
// DIFFERENT creatures you control that carry a +1/+1 counter, re-checked
// as the answer lands; "if you do" is both counters coming off.
//
// No simplification.
func init() {
	trigger := OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return attackDeclaredByYou(ev, source.Controller)
	}, "Dyadrine — you may remove a +1/+1 counter from each of two creatures you control; if you do, draw a card and create a 2/2 Robot",
		dyadrineAsk))
	Register(Spec{
		OracleID:        "618a2272-62e0-4d0e-834d-7b81719eba12",
		Name:            "Dyadrine, Synthesis Amalgam",
		Completeness:    CompletenessFull,
		XMatters:        true,
		ReadsManaSpent:  game.ReadsOwnPayment,
		PrintedKeywords: []string{"trample"},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{{
			Kind:  game.CounterPlusOne,
			Count: func(cast game.CastCounts) int { return cast.ManaSpent },
		}},
		Triggered: []game.TriggeredAbility{trigger},
	})
}

// dyadrineHasCounter is a creature `you` control with a +1/+1 counter.
func dyadrineHasCounter(c game.Card, you uuid.UUID) bool {
	return c.Controller == you && c.IsCreature() && c.Counters[game.CounterPlusOne] > 0
}

func dyadrineAsk(g *game.Game, item *game.StackItem) error {
	var candidates []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if dyadrineHasCounter(c, item.Controller) {
			candidates = append(candidates, c.InstanceID)
		}
	}
	if len(candidates) < 2 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Dyadrine — remove a +1/+1 counter from each of two creatures you control to draw a card and create a 2/2 Robot (choose exactly two, or none)",
		Cards:    candidates,
		Min:      0,
		Max:      2,
		Zone:     game.ZoneBattlefield,
		LiveZone: game.ZoneBattlefield,
		ThenKey:  dyadrineRemove,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// dyadrineRemove is the two counters, then the card and the Robot. An
// on-disk key: never renamed, never reused.
var dyadrineRemove = askedPickRule("choose-cards/dyadrine-remove-two-counters",
	func(picked []game.Card) bool { return len(picked) == 2 },
	func(ctx *Context, r game.CardsChosen) error {
		if len(r.Picked) != 2 {
			return nil
		}
		g := ctx.Game
		for _, id := range r.Picked {
			c, ok := g.LookupCardForEffect(id)
			if !ok || !onBattlefield(g, id) || !dyadrineHasCounter(c, r.Chooser) {
				return nil // "if you do": a pick that went away pays nothing
			}
		}
		for _, id := range r.Picked {
			if err := g.AddCounterForEffect(id, game.CounterPlusOne, -1); err != nil {
				return err
			}
		}
		if err := (DrawCards{N: 1}).Apply(ctx); err != nil {
			return err
		}
		return CreateToken{Controller: ctx.Controller(), Template: TokenCard("2/2 colorless Robot artifact"), N: 1}.Apply(ctx)
	})
