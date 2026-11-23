package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The War Games — Enchantment — Saga {2}{W}{W} (Doctor Who):
//
//	"I — Each player creates three tapped 1/1 white Warrior creature
//	 tokens. The tokens are goaded for as long as this Saga remains on the
//	 battlefield.
//	 II, III — Put a +1/+1 counter on each Warrior creature.
//	 IV — You may exile a nontoken creature you control. When you do, exile
//	 all Warriors."
//
// I makes the tokens in turn order from the active player, each as a group
// of three tapped ones, and goads each by the Saga's controller for as long
// as the Saga remains (ScopedEffectFor with a while-source-remains duration:
// the goad is a continuous effect, not the marker that ends at the goader's
// next turn, CR 701.15). II and III put a counter on every Warrior creature
// at the table. IV is an optional pick of a nontoken creature you control
// (exiled), then a reflexive trigger (CR 603.12) that exiles every Warrior.
//
// No simplification.
func init() {
	const name = "The War Games"
	effects.Register(effects.Spec{
		OracleID:     "d7cfe3a3-e8e6-446d-a855-46c845f657fa",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 3, CreatureTokens: 3},
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "each player creates three tapped 1/1 white Warrior creature tokens, goaded while this Saga remains"), theWarGamesTokens),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "put a +1/+1 counter on each Warrior creature"), theWarGamesCounters),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "put a +1/+1 counter on each Warrior creature"), theWarGamesCounters),
			effects.TriggerWithPurpose(
				effects.ChapterTrigger(4, effects.SagaChapterLabel(name, 4, "you may exile a nontoken creature you control; when you do, exile all Warriors"), theWarGamesFinale),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepExile, Partial: true}}),
		},
	})
}

func theWarGamesTokens(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	dur, ok := effects.DurationWhileSourceRemains(ctx, ctx.Source())
	var made []uuid.UUID
	for _, p := range effects.SeatsFromActive(g) {
		ids, err := g.CreateTokensForEffect(p, ctx.TextToken(effects.TokenCard("1/1 white Warrior")), 3, game.TokenEntryOptions{Tapped: true})
		if err != nil {
			return err
		}
		made = append(made, ids...)
	}
	if !ok {
		return nil
	}
	for _, id := range made {
		if err := (effects.ScopedEffectFor{
			Target:   id,
			Mods:     []game.Mod{game.GoadMod(ctx.Controller())},
			Duration: dur,
			Label:    "The War Games — goaded for as long as this Saga remains on the battlefield",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

func theWarGamesCounters(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var ids []uuid.UUID
	for _, c := range effects.MatchingBattlefield(ctx, effects.And(effects.Creature(), effects.HasSubtype("Warrior"))) {
		ids = append(ids, c.InstanceID)
	}
	return effects.B13PutCounterOnEach(ctx, ids)
}

func theWarGamesFinale(g *game.Game, item *game.StackItem) error {
	return effects.ChoosePermanents{
		Question: "The War Games — you may exile a nontoken creature you control",
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			var ids []uuid.UUID
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Controller == of && c.IsCreature() && !c.IsToken() {
					ids = append(ids, c.InstanceID)
				}
			}
			return ids, 0, 1
		},
		Keyed: theWarGamesPicked,
	}.Apply(effects.NewContext(g, item))
}

// theWarGamesPicked is the pick: the creature chosen, if any, is exiled, and
// that exile creates the reflexive trigger. An on-disk key.
var theWarGamesPicked = effects.RegisterPermanentsPick("prompt-run/the-war-games-exile", effects.PermanentsDef{
	Then: func(ctx *effects.Context, picked game.PromptedPicks) error {
		cards := picked.Cards()
		if len(cards) != 1 {
			return nil
		}
		return effects.ExileTarget{Target: cards[0], ThenTail: ctx.ItemTail(theWarGamesExiled)}.Apply(ctx)
	},
})

// theWarGamesExiled runs when the creature reached exile: "When you do,
// exile all Warriors." An on-disk key.
var theWarGamesExiled = effects.CardTail("the-war-games/exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return effects.WhenYouDo("The War Games — exile all Warriors", theWarGamesExileWarriorsBody).Apply(ctx)
})

// The reflexive body. An on-disk identity.
var theWarGamesExileWarriorsBody = game.SimpleDelayedBody("the-war-games/exile-all-warriors",
	func(g *game.Game, item *game.StackItem) error {
		return effects.ExileAllMatching{Match: effects.And(effects.Creature(), effects.HasSubtype("Warrior"))}.Apply(effects.NewContext(g, item))
	})
