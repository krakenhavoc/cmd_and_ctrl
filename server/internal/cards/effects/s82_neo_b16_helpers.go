package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_neo_b16_helpers.go — the shapes Kamigawa: Neon Dynasty batch 16
// (#5493) prints more than once: the Myojin's counter-from-hand entry,
// and the Go-Shintai end-step "you may pay {1}" opener.

// neoMyojinEntry is "<Myojin> enters with an indestructible counter on it
// if you cast it from your hand." (CR 614.1c, read off the resolving
// spell's cast record, so a Myojin cast from the command zone or put onto
// the battlefield without being cast has none). The counter is a keyword
// counter (CR 122.1b): it grants indestructible while it is there.
func neoMyojinEntry() []game.EntryCountersFromCast {
	return []game.EntryCountersFromCast{CountersIfCastFrom(game.ZoneHand, game.CounterIndestructible, 1)}
}

// neoRemoveIndestructible is the Myojins' cost, "Remove an indestructible
// counter from <this>".
func neoRemoveIndestructible() game.AbilityCost {
	return RemoveCountersFromThis(game.CounterIndestructible, 1)
}

// neoShrinesYouControl is "the number of Shrines you control" counted
// as the Go-Shintai's reflexive trigger or end-step trigger resolves.
func neoShrinesYouControl(g *game.Game, item *game.StackItem) int {
	return chkB1ShrinesYouControl(g, item)
}

// neoGoShintaiEndStep is the Go-Shintai opener, "At the beginning of your
// end step, you may pay {1}.": MayPay asks as the trigger resolves and
// runs `paid` (a registered PayUnlessBranches key) only when the mana was
// paid.
func neoGoShintaiEndStep(label, question string, paid game.PayUnlessElse) game.TriggeredAbility {
	return AtYourEndStep(label, func(g *game.Game, item *game.StackItem) error {
		return MayPay{
			Chooser:  item.Controller,
			Cost:     "{1}",
			Question: question,
			Then:     paid,
		}.Apply(NewContext(g, item))
	})
}

// neoPutCreatureCardsThisTurn is the card IDs of every creature card now
// in a graveyard that was put there from anywhere this turn (Myojin of
// Grim Betrayal). A token is not a card; a card is judged by the types it
// has in the graveyard.
func neoCreatureCardsPutIntoGraveyardsThisTurn(ctx *Context) []uuid.UUID {
	arrived := map[uuid.UUID]bool{}
	for _, ev := range ctx.Game.EventsThisTurn() {
		if ev.Kind == game.EventZoneMove && ev.NewZone == game.ZoneGraveyard && ev.OldZone != game.ZoneGraveyard && ev.CardID != uuid.Nil {
			arrived[ev.CardID] = true
		}
	}
	return allGraveyardsCardIDs(ctx, func(c game.Card) bool {
		return arrived[c.InstanceID] && !c.IsToken() && c.IsCreature()
	})
}

// The Go-Shintai's reflexive bodies ("When you do, …"), each counting
// the Shrines as the trigger resolves (CR 608.2h). Keys are on-disk
// identities and never renamed.
var (
	// Go-Shintai of Hidden Cruelty: destroy target creature with toughness
	// X or less. The clause's ceiling is the count when the trigger goes on
	// the stack (Params.Amount); the body re-reads it, so a target whose
	// toughness no longer fits X as it resolves is an illegal one.
	neoHiddenCrueltyBody = game.ReflexiveBody("neo-go-shintai/hidden-cruelty-destroy", simpleBody(func(g *game.Game, item *game.StackItem) error {
		if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
			return nil
		}
		t, ok := g.LookupCardForEffect(item.Targets[0].ID)
		if !ok || t.CurrentToughness() > neoShrinesYouControl(g, item) {
			return nil
		}
		return DestroyTarget{Target: t.InstanceID}.Apply(NewContext(g, item))
	}), func(_ uuid.UUID, p game.EffectParams) *game.TargetSpec {
		return TargetCreature("target creature with toughness X or less", ToughnessLE(p.Amount))
	})

	// Go-Shintai of Ancient Wars: X damage to target player or planeswalker.
	neoAncientWarsBody = game.ReflexiveBody("neo-go-shintai/ancient-wars-damage", simpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		t, ok := firstLegalTarget(ctx)
		if !ok {
			return nil
		}
		return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: neoShrinesYouControl(g, item)}.Apply(ctx)
	}), constTargets(targetPlayerOrPlaneswalker))

	// Go-Shintai of Boundless Vigor: a +1/+1 counter on target Shrine for
	// each Shrine you control.
	neoBoundlessVigorBody = game.ReflexiveBody("neo-go-shintai/boundless-vigor-counters", simpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		t, ok := firstLegalTarget(ctx)
		if !ok {
			return nil
		}
		return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: neoShrinesYouControl(g, item)}.Apply(ctx)
	}), constTargets(func() *game.TargetSpec {
		return TargetPermanent("target Shrine", OfSubtype("Shrine"))
	}))

	// Go-Shintai of Lost Wisdom: target player mills X cards.
	neoLostWisdomBody = game.ReflexiveBody("neo-go-shintai/lost-wisdom-mill", simpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		t, ok := firstLegalTarget(ctx)
		if !ok {
			return nil
		}
		return MillCards{Player: t.ID, N: neoShrinesYouControl(g, item)}.Apply(ctx)
	}), constTargets(func() *game.TargetSpec { return TargetPlayer("target player") }))
)

// neoGoShintaiReflexive is the paid branch of a Go-Shintai whose second
// sentence is "When you do, …": the reflexive trigger carries the Shrine
// count when it is created.
func neoGoShintaiReflexive(slug, label string, body game.BodyRef) game.PayUnlessElse {
	return PayUnlessBranches("pay-unless/neo-go-shintai-"+slug, func(ctx *Context, _ game.PayUnlessAnswer) error {
		return ReflexiveTrigger{
			Label:  "Go-Shintai — " + label,
			Body:   body,
			Params: game.EffectParams{Amount: neoShrinesYouControl(ctx.Game, ctx.Item)},
		}.Apply(ctx)
	}, nil)
}
