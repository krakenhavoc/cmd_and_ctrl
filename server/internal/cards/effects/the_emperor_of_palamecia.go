package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Emperor of Palamecia // The Lord Master of Hell (Final Fantasy,
// #4462) — a transforming legendary creature.
//
// Front face, Legendary Creature — Human Noble Wizard {U}{R}, 2/2:
//
//	"{T}: Add {U} or {R}. Spend this mana only to cast a noncreature spell.
//	 Whenever you cast a noncreature spell, if at least four mana was spent
//	 to cast it, put a +1/+1 counter on The Emperor of Palamecia. Then if it
//	 has three or more +1/+1 counters on it, transform it."
//
// Back face, The Lord Master of Hell — Legendary Creature — Demon Noble
// Wizard, 3/3 (the_lord_master_of_hell.go).
//
// The mana ability is a two-colour tap mana with a spending restriction
// (CR 106.12): it can pay for a noncreature spell only. The trigger is an
// intervening "if" (CR 603.4) read off the mana spent as the spell was
// cast, counted then and carried on the item, so it holds when the spell
// is countered or has left the stack (Tellah's shape). The counter and the
// transform are for the object that triggered (CR 400.7); the transform is
// the permanent's own, and "it has three or more counters" is read after
// the counter has been put on.
//
// No simplification.
func init() {
	const label = "The Emperor of Palamecia — put a +1/+1 counter on it; then if it has three or more +1/+1 counters, transform it"
	trig := On(game.EventCast, YouCast(Noncreature()), label, func(g *game.Game, item *game.StackItem) error {
		if item.Params.Amount < 4 || !sourceIsStillThisPermanent(g, item) {
			return nil
		}
		ctx := NewContext(g, item)
		if err := (AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
			return err
		}
		c, ok := g.LookupCardForEffect(item.SourceCardID)
		if !ok || c.Counters[game.CounterPlusOne] < 3 {
			return nil
		}
		return Transform{Target: item.SourceCardID}.Apply(ctx)
	})
	trig.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Amount = g.StackItemPaidForEffect(ev.CardID).Spent().Total()
		return item
	}
	trig = InterveningIf(CondBoard{Name: "at least four mana was spent to cast it", Test: func(ctx *Context) bool {
		return ctx.Item != nil && ctx.Item.Params.Amount >= 4
	}}, trig)
	mana := dualManaAbility("U", "R")
	mana.Restrictions = []string{ManaRestrictCast, ManaRestrictNotType("Creature")}
	mana.Label = "{T}: Add {U} or {R}. Spend this mana only to cast a noncreature spell."
	Register(Spec{
		OracleID:       theEmperorOfPalameciaOracleID,
		Name:           "The Emperor of Palamecia",
		ReadsManaSpent: game.ReadsYourCasts,
		Completeness:   CompletenessFull,
		ManaAbilities:  []ManaAbility{mana},
		Triggered:      []game.TriggeredAbility{trig},
	})
}

// theEmperorOfPalameciaOracleID is shared with the back face, which
// registers under it plus "#1" (game.CatalogKey).
const theEmperorOfPalameciaOracleID = "042c01ec-9f1f-432b-bea9-d64d822cd91e"
