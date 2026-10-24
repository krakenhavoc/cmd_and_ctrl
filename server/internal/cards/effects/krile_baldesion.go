package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Krile Baldesion — Legendary Creature — Dwarf Wizard {W}{U}, 2/2:
//
//	"Lifelink
//	 Trace Aether — Whenever you cast a noncreature spell, you may return
//	 target creature card with mana value equal to that spell's mana
//	 value from your graveyard to your hand. Do this only once each turn."
//
// Lifelink is printed data. The target clause is built as the trigger goes
// on the stack (CR 603.3d), from the trigger's own CR 603.10 snapshot of
// the spell: a pure read of the trigger context, so restore re-derives it
// and the row needs no TargetsFromReadsBoard. That snapshot's mana value
// counts the X a spell was cast with (CR 202.3e, #3782), so a Fireball
// cast for X=3 offers the three-drops. A noncreature spell with no creature
// card of that value in the graveyard puts nothing on the stack.
//
// The "may" is made as the ability resolves (CR 603.5, MayChoice), after
// the target is chosen. "Do this only once each turn" limits the DOING, as
// G'raha Tia's and Pantlaza's do: a trigger whose "may" is declined, or
// whose target left in response, does nothing and leaves the turn's one use
// for a later spell. The use is recorded (Game.NoteDoneThisTurn) only when
// the card is returned, read when the ability triggers (so no trigger goes
// on the stack once the use is spent) and again as it resolves (so the
// second of two triggers stacked before either resolved does nothing).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "546a217a-ac24-4fc0-97a8-424cfe1aafc0",
		Name:            "Krile Baldesion",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered:       []game.TriggeredAbility{krileTraceAether()},
	})
}

const (
	krileLabel = "Krile Baldesion — Trace Aether: you may return a creature card with that mana value from your graveyard to your hand"
	// krileDoneLabel is the tally the once-each-turn limit reads.
	krileDoneLabel = "Krile Baldesion — Trace Aether (done)"
)

func krileTraceAether() game.TriggeredAbility {
	t := On(game.EventCast, AllOf(YouCast(Noncreature()), krileNotYetDone), krileLabel, krileReturn)
	t.TargetsFrom = krileClause
	return t
}

// krileNotYetDone keeps the trigger off the stack once the turn's one use
// has been spent.
func krileNotYetDone(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return source != nil && g.ResolvedThisTurn(source.InstanceID, krileDoneLabel) == 0
}

// krileClause is "target creature card with mana value equal to that
// spell's mana value from your graveyard", cut to the spell the trigger
// fired on. No snapshot at all returns a clause nothing satisfies (CR
// 603.3d removes the ability) rather than nil, which would mean "targets
// nothing".
func krileClause(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
	mv := -1
	if tc.Object != nil {
		mv = tc.Object.ManaValue
	}
	return TargetCardInGraveyard(
		"target creature card with that spell's mana value in your graveyard",
		YouOwn(), Creature(), ManaValueLE(mv), ManaValueGE(mv),
	)
}

// krileReturn asks, as the ability resolves, whether to return the chosen
// card, and records the use only when it is returned.
func krileReturn(g *game.Game, item *game.StackItem) error {
	if g.ResolvedThisTurn(item.SourceCardID, krileDoneLabel) > 0 {
		return nil
	}
	ctx := NewContext(g, item)
	ts := ctx.LegalTargets()
	if len(ts) == 0 || ts[0].Kind != game.TargetCard {
		return nil
	}
	id := ts[0].ID
	return MayChoice{
		Question: "Krile Baldesion — return that card to your hand?",
		Then:     krileAnswer,
		Carry:    []uuid.UUID{id},
	}.Apply(ctx)
}

// krileAnswer is the "yes": return the card (the carried ID) to the hand
// and record the use. Keyed (#4291).
var krileAnswer = MayYes("confirm/krile-baldesion-return", func(ctx *Context, r game.ConfirmAnswer) error {
	g, item, id := ctx.Game, ctx.Item, carryAt(r, 0)
	if g.ResolvedThisTurn(item.SourceCardID, krileDoneLabel) > 0 {
		return nil
	}
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
		return err
	}
	g.NoteDoneThisTurn(item.SourceCardID, krileDoneLabel)
	return nil
})
