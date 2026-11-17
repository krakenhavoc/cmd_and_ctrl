package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kaya, Intangible Slayer — Legendary Planeswalker — Kaya {3}{W}{W}{B}{B},
// starting loyalty 6 (Phyrexia: All Will Be One, #5479):
//
//	"Hexproof
//	 +2: Each opponent loses 3 life and you gain 3 life.
//	 0: You draw two cards. Then each opponent may scry 1.
//	 −3: Exile target creature or enchantment. If it wasn't an Aura, create
//	     a token that's a copy of it, except it's a 1/1 white Spirit
//	     creature with flying in addition to its other types."
//
// The 0 draws, then asks each opponent in turn order "scry 1?" (a yes
// scries before the next is asked). The −3 exiles the target and, unless it
// was an Aura as it last existed on the battlefield (CR 608.2h: read before
// it moves), makes a token copying its copiable values (CR 707.2) with the
// exception applied to the copy: 1/1, white, a creature, a Spirit and with
// flying, all in addition to what it already was (CR 707.9b). The copy is
// made whether or not the exile moved the card, as the sentences are
// independent.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9260ae3d-59bf-4e46-a13e-930754c3f0d2",
		Name:            "Kaya, Intangible Slayer",
		Completeness:    CompletenessFull,
		StartingLoyalty: 6,
		PrintedKeywords: []string{"hexproof"},
		Activated: []ActivatedAbility{
			{
				Label:   "+2: Each opponent loses 3 life and you gain 3 life.",
				Cost:    LoyaltyCost(2),
				Purpose: game.Purpose{LifeGain: 3},
				Effect:  Do(EachOpponentLosesLife{N: 3}, GainLife{Amount: 3}),
			},
			{
				Label:   "0: You draw two cards. Then each opponent may scry 1.",
				Cost:    LoyaltyCost(0),
				Purpose: game.Purpose{Draws: 2},
				Effect: Do(DrawCards{N: 2}, Instruction(func(ctx *Context) error {
					return kayaIntangibleSlayerAsk(ctx, ctx.Opponents())
				})),
			},
			{
				Label:   "−3: Exile target creature or enchantment. If it wasn't an Aura, create a token that's a copy of it, except it's a 1/1 white Spirit creature with flying in addition to its other types.",
				Cost:    LoyaltyCost(-3),
				Targets: TargetPermanent("target creature or enchantment", Or(Creature(), Enchantment())),
				Effect:  kayaIntangibleSlayerMinusThree,
			},
		},
	})
}

func kayaIntangibleSlayerMinusThree(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	targets := ctx.LegalTargets()
	if len(targets) == 0 || targets[0].Kind != game.TargetCard {
		return nil
	}
	id := targets[0].ID
	e := ExileTarget{Target: id}
	if c, ok := g.LookupCardForEffect(id); ok && !c.HasSubtype("Aura") {
		e.ThenTail = game.KeyedTail{
			Ref: kayaIntangibleSlayerCopy, Player: item.Controller, Source: item.SourceCardID,
			Carry: []uuid.UUID{id},
		}
	}
	return e.Apply(ctx)
}

// kayaIntangibleSlayerCopy is "create a token that's a copy of it, except
// it's a 1/1 white Spirit creature with flying in addition to its other
// types". An on-disk key: never renamed, never reused.
var kayaIntangibleSlayerCopy = CardTail("then/kaya-intangible-slayer-copy", func(ctx *Context, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	return CreateTokenCopy{
		Controller: o.Tail.Player,
		Copy:       o.Tail.Carry[0],
		N:          1,
		Except: func(t *game.Card) {
			t.Power, t.Toughness = 1, 1
			t.VariableToughness = false
			t.PrintedPTKnown = true
			t.SetCopyExceptionColors("W")
			t.TypeLine = socB19WithCreatureAndSubtypeInAddition(t.TypeLine, "Spirit")
			if !hasKeywordWord(t.Keywords, "flying") {
				t.Keywords = append(t.Keywords, "flying")
			}
		},
	}.Apply(ctx)
})

func hasKeywordWord(kws []string, kw string) bool {
	for _, k := range kws {
		if k == kw {
			return true
		}
	}
	return false
}

// kayaIntangibleSlayerAsk is "each opponent may scry 1", asked in turn
// order; a yes scries before the next is asked. On-disk keys: never renamed.
func kayaIntangibleSlayerAsk(ctx *Context, opponents []uuid.UUID) error {
	if len(opponents) == 0 {
		return nil
	}
	return MayChoice{
		Player:   opponents[0],
		Question: "Kaya, Intangible Slayer — scry 1?",
		Then:     kayaIntangibleSlayerAnswer,
		Carry:    opponents[1:],
	}.Apply(ctx)
}

var kayaIntangibleSlayerAnswer game.ConfirmThen

func init() {
	kayaIntangibleSlayerAnswer = MayAnswer("confirm/kaya-intangible-slayer-scry", func(ctx *Context, r game.ConfirmAnswer) error {
		if !r.Accepted {
			return kayaIntangibleSlayerAsk(ctx, r.Carry)
		}
		next := ctx.ItemTail(kayaIntangibleSlayerAskNext)
		next.Carry = r.Carry
		return Scry{Player: r.Chooser, N: 1, ThenTail: next}.Apply(ctx)
	})
}

var kayaIntangibleSlayerAskNext = CardTail("kaya-intangible-slayer/ask-next", func(ctx *Context, o game.TailOutcome) error {
	return kayaIntangibleSlayerAsk(ctx, o.Tail.Carry)
})
