package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b496_helpers.go — shared pieces of S82 OLD batch 496 (Zendikar
// landfall, pump, token makers, #5583).

// Old496ForestsYouControl counts the Forests `player` controls as the effect
// resolves, by land type, a word of the card's text (CR 612.2).
func Old496ForestsYouControl(ctx *Context, player uuid.UUID) int {
	forest := ctx.TextSubtype("Forest")
	n := 0
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == player && c.HasSubtype(forest) {
			n++
		}
	}
	return n
}

// Old496PumpTargetPerForest is Primal Bellow's body: "Target creature gets +1/+1
// until end of turn for each Forest you control." The Forests are counted as the
// spell resolves; a target that is no longer legal does nothing (CR 608.2b).
func Old496PumpTargetPerForest(item *game.StackItem, ctx *Context) error {
	t, ok := ctx.ClauseTarget(0)
	if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
		return nil
	}
	n := Old496ForestsYouControl(ctx, item.Controller)
	return BoostUntilEOT{Target: t.ID, Power: n, Toughness: n, Label: "Primal Bellow — +1/+1 for each Forest"}.Apply(ctx)
}

// Old496LarvaAttackPump is Timbermaw Larva's trigger body: "it gets +1/+1 until
// end of turn for each Forest you control", counted as the trigger resolves.
func Old496LarvaAttackPump(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if !OnBattlefield(g, item.SourceCardID) || SourceIsNewObject(g, item) {
		return nil
	}
	n := Old496ForestsYouControl(ctx, item.Controller)
	return BoostUntilEOT{Target: item.SourceCardID, Power: n, Toughness: n, Label: "Timbermaw Larva — +1/+1 for each Forest"}.Apply(ctx)
}

// Old496ZektarElemental is Zektar Shrine Expedition's body: a 7/1 red Elemental
// token with trample and haste, exiled at the beginning of the next end step
// (Rakdos Guildmage's token and delayed exile).
func Old496ZektarElemental(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	made, err := g.CreateTokensForEffect(ctx.Controller(), ctx.TextToken(TokenCard("7/1 red Elemental with trample and haste")), 1, game.TokenEntryOptions{})
	if err != nil || len(made) == 0 {
		return err
	}
	return ScheduleDelayedTrigger{
		At:         game.StepEnd,
		Controller: ctx.Controller(),
		Label:      "Zektar Shrine Expedition — exile the Elemental token",
		Cards:      made,
		Body:       ExileListedCardsBody,
	}.Apply(ctx)
}

// Old496EternityVesselBecomes is the landfall body of Eternity Vessel: "you may
// have your life total become the number of charge counters on this artifact"
// (the "may" is the trigger's). The total is set the way CR 119.5 says, by
// gaining or losing the difference, and only while the Vessel is still the
// permanent that triggered.
func Old496EternityVesselBecomes(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if !OnBattlefield(g, item.SourceCardID) || SourceIsNewObject(g, item) {
		return nil
	}
	vessel, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok {
		return nil
	}
	return B31LifeBecomes(ctx, item.Controller, vessel.Counters["charge"])
}

// Old496KalitasDestroy is Kalitas, Bloodchief of Ghet's ability: destroy target
// creature; if that creature dies this way, create a black Vampire token whose
// power and toughness are that creature's, as it last existed (read before it is
// destroyed). A creature that regenerates, is indestructible or is exiled
// instead of dying makes no token. The token is the ability's controller's.
func Old496KalitasDestroy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	t, ok := ctx.ClauseTarget(0)
	if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
		return nil
	}
	victim, ok := g.LookupCardForEffect(t.ID)
	if !ok {
		return nil
	}
	// The token is made from the destroy's continuation, told what was actually
	// destroyed (the clause is gated on that list, #911); the size rides as data.
	nums := []int{victim.CurrentPower(), victim.CurrentToughness(), int(B25LastEventSeq(g))}
	tail := ctx.Tail(old496KalitasToken)
	tail.Nums = nums
	return g.DestroyPermanentsKeyedForEffect([]uuid.UUID{t.ID}, WithCarry(tail, t.ID))
}

// old496KalitasToken is Kalitas's "if that creature dies this way": the creature
// (carried) was destroyed and went to a graveyard as that destruction happened
// (not exiled in its place), so a Vampire token with its power and toughness
// (Nums) is created for the ability's controller. An on-disk key: never
// renamed, never reused.
var old496KalitasToken = CardTail("destroy/old-b496-kalitas-vampire", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 || len(o.Tail.Carry) == 0 || len(o.Tail.Nums) < 3 {
		return nil
	}
	died := false
	for _, ev := range ctx.Game.Events {
		if int(ev.Seq) > o.Tail.Nums[2] && ev.Kind == game.EventLTB && ev.CardID == o.Tail.Carry[0] && ev.NewZone == game.ZoneGraveyard {
			died = true
			break
		}
	}
	if !died {
		return nil
	}
	tok := ctx.TextToken(TokenCard("0/0 black Vampire"))
	tok.Power, tok.Toughness = o.Tail.Nums[0], o.Tail.Nums[1]
	return ctx.Game.CreateTokenForEffect(ctx.Controller(), tok, 1)
})
