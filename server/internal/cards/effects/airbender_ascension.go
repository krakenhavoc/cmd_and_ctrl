package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Airbender Ascension — Enchantment {1}{W} (Avatar: The Last Airbender,
// #4459):
//
//	"When this enchantment enters, airbend up to one target creature.
//	 Whenever a creature you control enters, put a quest counter on this
//	 enchantment.
//	 At the beginning of your end step, if this enchantment has four or
//	 more quest counters on it, exile up to one target creature you
//	 control, then return it to the battlefield under its owner's
//	 control."
//
// Three printed triggers, three rows:
//
//   - The entry airbends through the shared keyword action (Airbend, by
//     way of AirbendOtherTarget, whose "other" check can never bite here:
//     the source is an enchantment, never a creature target). The card's
//     owner may cast it from exile for {2} while it stays there; a token
//     ceases to exist (the ruling).
//   - "A creature you control enters" counts tokens and every creature of
//     a group entering together, one counter each; the counter goes on
//     the enchantment only while it is on the battlefield (CounterOnThis).
//   - The end-step flicker is an intervening "if" (CR 603.4): checked as
//     the step begins and again as the ability resolves, so a counter
//     removed in response stops it. The creature returns under its
//     OWNER's control as a new object (CR 400.7), its Auras and counters
//     gone (flickerFirstLegalTargetUnderOwner); a token does not return.
//
// No simplification.
func init() {
	fourQuest := CondBoard{
		Name: "this enchantment has four or more quest counters on it",
		Test: func(ctx *Context) bool { return airbenderAscensionCharged(ctx.Game, ctx.Source()) },
	}
	Register(Spec{
		OracleID:     "c4d2fdf9-637d-4e97-ae33-45f2d27bf8cd",
		Name:         "Airbender Ascension",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Airbender Ascension — airbend up to one target creature", AirbendOtherTarget),
				TargetCreature("up to one target creature").WithCount(0, 1)),
			On(game.EventETB, CreatureEnteredUnderYourControl,
				"Airbender Ascension — put a quest counter on it", Do(CounterOnThis{Kind: "quest", N: 1})),
			InterveningIf(fourQuest, Targeting(On(game.EventBeginEndStep,
				func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
					return ByYou(ev, source, ch, g) && airbenderAscensionCharged(g, source.InstanceID)
				},
				"Airbender Ascension — exile up to one target creature you control, then return it to the battlefield under its owner's control",
				flickerFirstLegalTargetUnderOwner),
				TargetCreature("up to one target creature you control", YouControl()).WithCount(0, 1))),
		},
	})
}

// airbenderAscensionCharged is "this enchantment has four or more quest
// counters on it": false once it has left the battlefield.
func airbenderAscensionCharged(g *game.Game, id uuid.UUID) bool {
	c, ok := g.LookupCardForEffect(id)
	return ok && onBattlefield(g, id) && c.Counters["quest"] >= 4
}
