package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Minsc & Boo, Timeless Heroes — Legendary Planeswalker — Minsc {2}{R}{G},
// loyalty 3:
//
//	"When Minsc & Boo enters and at the beginning of your upkeep, you may
//	 create Boo, a legendary 1/1 red Hamster creature token with trample and
//	 haste.
//	 +1: Put three +1/+1 counters on up to one target creature with trample or
//	 haste.
//	 −2: Sacrifice a creature. When you do, Minsc & Boo deals X damage to any
//	 target, where X is that creature's power. If the sacrificed creature was a
//	 Hamster, draw X cards.
//	 Minsc & Boo, Timeless Heroes can be your commander."
//
// One optional trigger on two events (entering, and your upkeep). The −2 is
// Tip the Scales' shape: the sacrifice is a resolution choice, and "when you
// do" is a reflexive trigger raised once a creature has really left, so its
// any-target is picked then; X and "was a Hamster" are the sacrificed
// creature's last-known power and subtypes (CR 608.2h), fixed as the trigger
// is created. The draw belongs to the same trigger, so it needs a legal target.
// Commander eligibility is the importer's, from the printed line.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3a310554-3f69-4158-bb94-9d6965c7125a",
		Name:            "Minsc & Boo, Timeless Heroes",
		Completeness:    CompletenessFull,
		StartingLoyalty: 3,
		Purpose:         game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			Optional(OnAny([]game.EventKind{game.EventETB, game.EventBeginUpkeep}, minscAndBooCreatesBoo,
				"Minsc & Boo — you may create Boo, a legendary 1/1 red Hamster creature token with trample and haste",
				Do(CreateToken{Template: TokenCard(afrB15BooToken), N: 1})),
				"Create Boo, a legendary 1/1 red Hamster creature token with trample and haste?"),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Put three +1/+1 counters on up to one target creature with trample or haste.",
				Cost:    LoyaltyCost(1),
				Targets: TargetCreature("up to one target creature with trample or haste", Or(HasKeyword("trample"), HasKeyword("haste"))).WithCount(0, 1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return AddCounter{Target: id, Kind: game.CounterPlusOne, N: 3}.Apply(ctx)
				},
			},
			{
				Label: "−2: Sacrifice a creature. When you do, Minsc & Boo deals X damage to any target, where X is that " +
					"creature's power. If the sacrificed creature was a Hamster, draw X cards.",
				Cost: LoyaltyCost(-2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return mustSacrificeOne(g, item, "Minsc & Boo — sacrifice a creature", Creature(), minscAndBooSacrificed)
				},
			},
		},
	})
}

// minscAndBooCreatesBoo is the trigger's condition: this entered, or it is the
// controller's upkeep.
func minscAndBooCreatesBoo(ev game.Event, source *game.Card, c game.Characteristic, g *game.Game) bool {
	if ev.Kind == game.EventETB {
		return Self(ev, source, c, g)
	}
	return ByYou(ev, source, c, g)
}

// minscAndBooSacrificed raises the reflexive trigger once the creature has
// really gone, carrying its last-known power as X and whether it was a Hamster.
// An on-disk key: never renamed, never reused.
var minscAndBooSacrificed = RegisterSacrificedOne("prompt-run/minsc-and-boo-sacrifice", func(ctx *Context, sacrificed uuid.UUID) error {
	hamster := ""
	if ref, ok := ctx.Game.PermanentRefForEffect(sacrificed); ok {
		if info, ok := ctx.Game.PermanentForEffect(ref); ok {
			for _, st := range info.Characteristic.Subtypes {
				if st == "Hamster" {
					hamster = "Hamster"
				}
			}
		}
	}
	return ReflexiveTrigger{
		Label:  "Minsc & Boo — deals X damage to any target, X the sacrificed creature's power",
		Cards:  []uuid.UUID{sacrificed},
		Body:   minscAndBooDamageBody,
		Params: game.EffectParams{Amount: departedCreaturePower(ctx.Game, sacrificed), Name: hamster},
	}.Apply(ctx)
})

var minscAndBooDamageBody = game.ReflexiveBody("minsc-and-boo/damage", func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	ctx := NewContext(g, item)
	legal := ctx.LegalTargets()
	if len(legal) == 0 {
		return nil
	}
	if err := (DealDamage{Source: ctx.Source(), Target: legal[0].ID, Amount: p.Amount}).Apply(ctx); err != nil {
		return err
	}
	if p.Name == "Hamster" {
		return DrawCards{N: p.Amount}.Apply(ctx)
	}
	return nil
}, constTargets(func() *game.TargetSpec { return TargetAny() }))
