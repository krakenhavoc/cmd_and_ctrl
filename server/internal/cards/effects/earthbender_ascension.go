package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Earthbender Ascension — Enchantment {2}{G} (Avatar: The Last
// Airbender, #4459):
//
//	"When this enchantment enters, earthbend 2. Then search your library
//	 for a basic land card, put it onto the battlefield tapped, then
//	 shuffle.
//	 Landfall — Whenever a land you control enters, put a quest counter
//	 on this enchantment. When you do, if it has four or more quest
//	 counters on it, put a +1/+1 counter on target creature you control.
//	 It gains trample until end of turn."
//
// The entry earthbends its target land (the keyword action) and THEN
// searches: the search is Earthbend.Then, so it runs once the counters
// have landed. The land found is not on the battlefield when the target
// is chosen, so it can't be the land earthbent (the ruling). If the
// target land is illegal as the trigger resolves, the trigger doesn't
// resolve and there is no search (CR 608.2b, the ruling).
//
// Landfall puts the quest counter on the enchantment (CounterOnThis,
// nothing once it has left). Putting it is the "doing": the reflexive
// trigger (CR 603.12) is made only then, and only if the enchantment has
// four or more quest counters — an intervening "if" (CR 603.4), checked
// as the reflexive trigger is made and again as it resolves
// (airbenderAscensionCharged, false once it has left). Its target is
// chosen as it goes on the stack; the counter and the trample are for
// that creature as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f8966fd5-2926-4ba0-b7d6-cfea40cd4e05",
		Name:         "Earthbender Ascension",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisEnters("Earthbender Ascension — earthbend 2, then search for a basic land",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					land := FirstLegalBattlefieldTarget(ctx)
					if land == uuid.Nil {
						return nil
					}
					return Earthbend{Target: land, N: 2, Then: ctx.ItemTail(earthbenderAscensionSearch)}.Apply(ctx)
				}), EarthbendTargets()), game.Purpose{Lands: 1}),
			Landfall("Earthbender Ascension — put a quest counter on it",
				Do(CounterOnThis{Kind: "quest", N: 1}, Instruction(func(ctx *Context) error {
					if !airbenderAscensionCharged(ctx.Game, ctx.Source()) {
						return nil
					}
					return WhenYouDo("Earthbender Ascension — put a +1/+1 counter on target creature you control; it gains trample",
						earthbenderAscensionGrowBody).Apply(ctx)
				}))),
		},
	})
}

// earthbenderAscensionSearch is "Then search your library for a basic
// land card, put it onto the battlefield tapped, then shuffle", once the
// earthbend's counters have landed. An on-disk key.
var earthbenderAscensionSearch = CardTail("earthbender-ascension/search", func(ctx *Context, _ game.TailOutcome) error {
	return SearchLibrary{Player: ctx.Controller(), Predicate: IsBasicLand, Dest: game.ZoneBattlefield, Limit: 1,
		Shuffle: true, TappedOnEntry: true, Reason: "Earthbender Ascension — a basic land, tapped"}.Apply(ctx)
})

// earthbenderAscensionGrowBody is the reflexive half: the intervening
// "if" again (CR 603.4), then the counter and the trample on the target.
// An on-disk key.
var earthbenderAscensionGrowBody = game.ReflexiveBody("earthbender-ascension/grow",
	simpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		if !airbenderAscensionCharged(g, item.SourceCardID) {
			return nil
		}
		t, ok := ctx.ClauseTarget(0)
		if !ok {
			return nil
		}
		if err := (AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
			return err
		}
		return GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"trample"}, Label: "Earthbender Ascension — trample"}.Apply(ctx)
	}),
	constTargets(func() *game.TargetSpec {
		return TargetCreature("target creature you control", YouControl())
	}))
