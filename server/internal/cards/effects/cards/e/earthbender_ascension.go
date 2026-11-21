package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "f8966fd5-2926-4ba0-b7d6-cfea40cd4e05",
		Name:         "Earthbender Ascension",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Earthbender Ascension — earthbend 2, then search for a basic land",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					land := effects.FirstLegalBattlefieldTarget(ctx)
					if land == uuid.Nil {
						return nil
					}
					return effects.Earthbend{Target: land, N: 2, Then: ctx.ItemTail(earthbenderAscensionSearch)}.Apply(ctx)
				}), effects.EarthbendTargets()), game.Purpose{Lands: 1}),
			effects.Landfall("Earthbender Ascension — put a quest counter on it",
				effects.Do(effects.CounterOnThis{Kind: "quest", N: 1}, effects.Instruction(func(ctx *effects.Context) error {
					if !effects.AirbenderAscensionCharged(ctx.Game, ctx.Source()) {
						return nil
					}
					return effects.WhenYouDo("Earthbender Ascension — put a +1/+1 counter on target creature you control; it gains trample",
						earthbenderAscensionGrowBody).Apply(ctx)
				}))),
		},
	})
}

// earthbenderAscensionSearch is "Then search your library for a basic
// land card, put it onto the battlefield tapped, then shuffle", once the
// earthbend's counters have landed. An on-disk key.
var earthbenderAscensionSearch = effects.CardTail("earthbender-ascension/search", func(ctx *effects.Context, _ game.TailOutcome) error {
	return effects.SearchLibrary{Player: ctx.Controller(), Predicate: effects.IsBasicLand, Dest: game.ZoneBattlefield, Limit: 1,
		Shuffle: true, TappedOnEntry: true, Reason: "Earthbender Ascension — a basic land, tapped"}.Apply(ctx)
})

// earthbenderAscensionGrowBody is the reflexive half: the intervening
// "if" again (CR 603.4), then the counter and the trample on the target.
// An on-disk key.
var earthbenderAscensionGrowBody = game.ReflexiveBody("earthbender-ascension/grow",
	effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		if !effects.AirbenderAscensionCharged(g, item.SourceCardID) {
			return nil
		}
		t, ok := ctx.ClauseTarget(0)
		if !ok {
			return nil
		}
		if err := (effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
			return err
		}
		return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"trample"}, Label: "Earthbender Ascension — trample"}.Apply(ctx)
	}),
	effects.ConstTargets(func() *game.TargetSpec {
		return effects.TargetCreature("target creature you control", effects.YouControl())
	}))
