package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Smoldering Egg // Ashmouth Dragon — Creature — Dragon Egg {1}{R}, 0/4 //
// Creature — Dragon 4/4 (Midnight Hunt, Innistrad: Double Feature):
//
//	Smoldering Egg
//	  "Defender
//	   Whenever you cast an instant or sorcery spell, put a number of ember
//	   counters on this creature equal to the amount of mana spent to cast
//	   that spell. Then if this creature has seven or more ember counters
//	   on it, remove them and transform this creature."
//	Ashmouth Dragon
//	  "Flying
//	   Whenever you cast an instant or sorcery spell, this creature deals 2
//	   damage to any target."
//
// The amount is the mana actually spent, read off the spell's payment record
// as it is cast and carried on the item (Aberrant Manawurm's shape). A spell
// cast for no mana puts no counters on but the seven-counter check still
// runs. The counters go through the counter window, so the check waits for
// them to land, and "remove them and transform" turns the same permanent
// over in place.
//
// No simplification.
func init() {
	const oracle = "7493d78c-26c4-402e-bec5-6449773d0344"
	const label = "Smoldering Egg — put ember counters equal to the mana spent on it; at seven, remove them and transform it"
	egg := WheneverYouCast(Or(Instant(), Sorcery()), label, smolderingEggEmber)
	egg.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Amount = g.StackItemPaidForEffect(ev.CardID).Spent().Total()
		return item
	}
	Register(Spec{
		OracleID:        oracle,
		Name:            "Smoldering Egg",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		ReadsManaSpent:  game.ReadsYourCasts,
		Triggered:       []game.TriggeredAbility{egg},
	})
	burn := WheneverYouCast(Or(Instant(), Sorcery()), "Ashmouth Dragon — it deals 2 damage to any target",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			for _, t := range ctx.LegalTargets() {
				return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 2}.Apply(ctx)
			}
			return nil
		})
	burn.Targets = TargetAny()
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Ashmouth Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{burn},
	})
}

// smolderingEggEmber is the cast trigger, in printed order.
func smolderingEggEmber(g *game.Game, item *game.StackItem) error {
	if !b09SourceStillOnBattlefield(g, item) {
		return nil
	}
	ctx := NewContext(g, item)
	if item.Params.Amount <= 0 {
		return smolderingEggCheckNow(ctx)
	}
	return g.AddCounterByKeyedForEffect(uuid.Nil, item.SourceCardID, smolderingEggEmberKind, item.Params.Amount,
		ctx.ItemTail(smolderingEggCheck))
}

// smolderingEggEmberKind is the printed counter kind.
const smolderingEggEmberKind = "ember"

// smolderingEggCheck is "then if it has seven or more ember counters on it,
// remove them and transform it", once the counters have landed. An on-disk
// key: never renamed, never reused.
var smolderingEggCheck = CardTail("smoldering-egg/check-ember", func(ctx *Context, _ game.TailOutcome) error {
	return smolderingEggCheckNow(ctx)
})

func smolderingEggCheckNow(ctx *Context) error {
	source := ctx.Item.SourceCardID
	n := b39CountersOn(ctx.Game, source, smolderingEggEmberKind)
	if n < 7 {
		return nil
	}
	return ctx.Game.AddCounterByKeyedForEffect(uuid.Nil, source, smolderingEggEmberKind, -n,
		ctx.ItemTail(smolderingEggTransform))
}

// smolderingEggTransform is the transform behind the ember counters'
// removal. An on-disk key: never renamed, never reused.
var smolderingEggTransform = CardTail("smoldering-egg/transform", func(ctx *Context, _ game.TailOutcome) error {
	return TransformThis{}.Apply(ctx)
})
