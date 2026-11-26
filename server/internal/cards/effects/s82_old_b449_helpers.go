package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b449_helpers.go — Innistrad (S82 older sets, batch 449, #5572):
// the shared shapes of Scourge of Geier Reach and Village Cannibals.

// S82OldB449OpponentCreaturesPlusPT is "<name> gets +1/+1 for each creature
// your opponents control": a layer-7c modification counted live.
func S82OldB449OpponentCreaturesPlusPT(label string) game.StaticAbility {
	return game.StaticAbility{
		Label:     label,
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7C_Modify,
		AppliesTo: SelfOnly,
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			n := 0
			for _, p := range g.BattlefieldCardsForEffect() {
				if p.Controller != source.Controller && p.IsCreature() {
					n++
				}
			}
			c.Power += n
			c.Toughness += n
		},
	}
}

// S82OldB449AnotherSubtypeCreatureDied is "whenever another <subtype>
// creature dies": any controller's, the source excluded, the creature and
// its type read as it last existed (CR 603.10a).
func S82OldB449AnotherSubtypeCreatureDied(subtype string) When {
	return func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		dead, ok := DiedCreature(ev, g)
		return ok && ev.CardID != source.InstanceID && LeftAsSubtype(ev, dead, subtype)
	}
}

// S82OldB449HatchlingHatch is Ludevic's Test Subject's "put a hatchling
// counter on this creature. Then if there are five or more hatchling
// counters on it, remove all of them and transform it." The check reads the
// counters after the placement has settled; an ability whose source left the
// battlefield (or is a new object, CR 400.7) does nothing.
func S82OldB449HatchlingHatch(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id := item.SourceCardID
	if ctx.IsNewSourceObject(id) || !OnBattlefield(g, id) {
		return nil
	}
	if err := g.AddCounterForEffect(id, "hatchling", 1); err != nil {
		return err
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok || c.Counters["hatchling"] < 5 {
		return nil
	}
	if err := g.AddCounterForEffect(id, "hatchling", -c.Counters["hatchling"]); err != nil {
		return err
	}
	return TransformThis{}.Apply(ctx)
}

// s82OldB449ScholarLoot is Civilized Scholar's "draw a card, then discard a
// card" when the source is still the object that was activated: the discard
// carries the Scholar so its answer can turn it over.
var s82OldB449ScholarLoot = game.RegisterDrawThen("s82-old-b449-scholar-loot", func(g *game.Game, d game.DrawThen) error {
	return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
		Player: d.Player,
		Source: d.Source,
		N:      d.N,
	}, s82OldB449ScholarDiscarded, []uuid.UUID{d.Source})
})

// s82OldB449ScholarDiscarded is "If a creature card is discarded this way,
// untap this creature, then transform it." An on-disk key: never renamed,
// never reused.
var s82OldB449ScholarDiscarded = game.RegisterDiscardThen("choose-cards/s82-old-b449-scholar-discarded", func(g *game.Game, r game.DiscardsDone) error {
	if len(r.Carry) == 0 || !OnBattlefield(g, r.Carry[0]) {
		return nil
	}
	creature := false
	for _, id := range r.Discarded.Cards() {
		if c, ok := g.LookupCardForEffect(id); ok && c.HasCardType("creature") {
			creature = true
		}
	}
	if !creature {
		return nil
	}
	if err := g.UntapTargetForEffect(r.Carry[0]); err != nil {
		return err
	}
	if !g.CanTransformForEffect(r.Carry[0]) {
		return nil
	}
	return g.TransformPermanentForEffect(r.Carry[0])
})

// S82OldB449ScholarLootAndTurn is Civilized Scholar's activated ability.
func S82OldB449ScholarLootAndTurn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if ctx.IsNewSourceObject(item.SourceCardID) || !OnBattlefield(g, item.SourceCardID) {
		// The Scholar is gone or a new object: the loot still happens, but
		// it cannot untap or transform anything.
		return DrawThenDiscard(g, item.Controller, item.SourceCardID, 1, 1, "")
	}
	return g.DrawNThenForEffect(item.Controller, 1, game.DrawThen{
		Ref: s82OldB449ScholarLoot, Player: item.Controller, Source: item.SourceCardID, N: 1,
	})
}

// S82OldB449TapThenTransform is "tap this creature, then transform it".
func S82OldB449TapThenTransform(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (TapTarget{Target: item.SourceCardID}).Apply(ctx); err != nil {
		return err
	}
	return TransformThis{}.Apply(ctx)
}

// S82OldB449WitchSearch is Bitterheart Witch's search: a Curse card onto the
// battlefield attached to the trigger's target player, then shuffle.
func S82OldB449WitchSearch(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	t, ok := ctx.ClauseTarget(0)
	if !ok || t.Kind != game.TargetPlayer || !ctx.IsTargetLegal(t) {
		return nil
	}
	return SearchLibrary{
		Player:           item.Controller,
		Predicate:        func(c game.Card) bool { return c.HasSubtype("Curse") },
		Dest:             game.ZoneBattlefield,
		Limit:            1,
		Shuffle:          true,
		Reason:           "Bitterheart Witch — a Curse card",
		EntersAttachedTo: game.TargetRef{Kind: game.TargetPlayer, ID: t.ID},
	}.Apply(ctx)
}
