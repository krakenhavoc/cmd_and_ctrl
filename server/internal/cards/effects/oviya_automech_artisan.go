package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Oviya, Automech Artisan — Legendary Creature — Human Artificer {3}{G},
// 1/2 (Aetherdrift, #4464):
//
//	"Each creature that's attacking one of your opponents has trample.
//	 {G}, {T}: You may put a creature or Vehicle card from your hand onto
//	 the battlefield. If you put an artifact onto the battlefield this
//	 way, put two +1/+1 counters on it."
//
// The static is a layer 6 grant to EVERY creature attacking a player who
// is one of Oviya's controller's opponents, whoever controls it: in a
// multiplayer game an opponent's creature attacking another opponent has
// trample too, and a creature attacking you, a planeswalker or a battle
// does not (Martial Impetus's "attacking one of your opponents"). It is
// declared DependsOnAttackingStatus, so a declaration, a creature put
// onto the battlefield attacking, a removal from combat and the end of
// combat all rebuild the cached characteristics.
//
// The activated ability is Quicksilver Amulet's put, not a cast, so
// nothing can counter it and no cast trigger fires. "If you put an
// artifact onto the battlefield this way" reads the permanent that
// actually arrived, after the layers, in the put's keyed tail: a decline,
// an empty hand or an entry a replacement kept off the battlefield puts
// no counters. The two counters are put on after it enters, not "enters
// with", so they do not change how it entered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b01649b8-b0b5-4cf6-aea7-291f3407859e",
		Name:         "Oviya, Automech Artisan",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:                    game.Layer6Ability,
			Label:                    "Each creature that's attacking one of your opponents has trample.",
			DependsOnAttackingStatus: true,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.IsCreature() && socB8AttackingAnOpponentOf(g, source.Controller, *target)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				appendKeywordsTo(c, []string{"trample"})
			},
		}},
		Activated: []ActivatedAbility{{
			Label:   "{G}, {T}: You may put a creature or Vehicle card from your hand onto the battlefield. If you put an artifact onto the battlefield this way, put two +1/+1 counters on it.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{G}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return PutFromHandOntoBattlefield{
					Match:    Or(Creature(), Subtype("Vehicle")),
					Optional: true,
					Label:    "Oviya, Automech Artisan — you may put a creature or Vehicle card from your hand onto the battlefield",
					ThenTail: oviyaArtifactCounters,
				}.Apply(NewContext(g, item))
			},
		}},
	})
}

// oviyaArtifactCounters is "If you put an artifact onto the battlefield
// this way, put two +1/+1 counters on it", as a keyed tail. An on-disk
// key: never renamed, never reused.
var oviyaArtifactCounters = PutFromHandTail("oviya-automech-artisan/two-counters-on-an-artifact", func(ctx *Context, res PutFromHandResult) error {
	if res.Entered == uuid.Nil {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(res.Entered)
	if !ok || !c.IsArtifact() || !onBattlefield(ctx.Game, res.Entered) {
		return nil
	}
	return AddCounter{Target: res.Entered, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
})
