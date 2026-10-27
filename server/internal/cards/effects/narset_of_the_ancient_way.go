package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Narset of the Ancient Way — Legendary Planeswalker — Narset
// {1}{U}{R}{W}, loyalty 4:
//
//	"+1: You gain 2 life. Add {U}, {R}, or {W}. Spend this mana only to
//	     cast a noncreature spell.
//	 −2: Draw a card, then you may discard a card. When you discard a
//	     nonland card this way, Narset deals damage equal to that card's
//	     mana value to target creature or planeswalker.
//	 −6: You get an emblem with "Whenever you cast a noncreature spell,
//	     this emblem deals 2 damage to any target.""
//
// +1 is a loyalty ability, so it uses the stack and is not a mana
// ability (CR 605.1a). The colour is chosen as it resolves (an option
// pick of the three), and the mana carries Chandra, Chill of
// Compliance's "only to cast a noncreature spell" restriction.
//
// −2 is Glacial Dragonhunt's loot: the draw lands, the "up to one"
// discard opens, and only a NONLAND discard creates the CR 603.12
// reflexive trigger, whose target is chosen as it goes on the stack.
// The amount is the discarded card's mana value, fixed when it was
// discarded, and Narset is the source of the damage (as it last existed,
// if she has left).
//
// −6's emblem triggers on each noncreature spell its owner casts and
// targets as it is put on the stack (CR 603.3d); the emblem is the
// source of its own damage.
//
// No simplification. Guffster's Party deck request (#4184).
func init() {
	Register(Spec{
		OracleID:        narsetAncientWayOracle,
		Name:            "Narset of the Ancient Way",
		Completeness:    CompletenessFull,
		StartingLoyalty: 4,
		Emblem: &EmblemSpec{
			Label: "Narset of the Ancient Way emblem",
			Text:  "Whenever you cast a noncreature spell, this emblem deals 2 damage to any target.",
			Triggered: []game.TriggeredAbility{
				Targeting(
					WheneverYouCast(Noncreature(), "Narset of the Ancient Way emblem — 2 damage to any target", narsetAncientWayEmblemDamage),
					TargetAny()),
			},
		},
		Activated: []ActivatedAbility{
			{
				Label: "+1: You gain 2 life. Add {U}, {R}, or {W}. Spend this mana only to cast a noncreature spell.",
				Cost:  LoyaltyCost(1),
				// The colour pick waits for a prompt the life gain asks (a CR 616 order; CR 608.2c, #4155).
				Effect: Do(
					GainLife{Amount: 2},
					PickOption{
						Question: "Narset of the Ancient Way — add {U}, {R}, or {W}",
						Options:  []game.ChoiceOption{{Label: "{U}"}, {Label: "{R}"}, {Label: "{W}"}},
						ThenKey:  narsetAncientWayMana,
					},
				),
			},
			{
				Label: "−2: Draw a card, then you may discard a card. When you discard a nonland card this way, Narset deals damage equal to that card's mana value to target creature or planeswalker.",
				Cost:  LoyaltyCost(-2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return g.DrawNThenForEffect(item.Controller, 1, game.DrawThen{
						Ref: narsetAncientWayDiscard, Player: item.Controller, Source: item.SourceCardID, N: 1,
					})
				},
			},
			{
				Label: "−6: You get an emblem with \"Whenever you cast a noncreature spell, this emblem deals 2 damage to any target.\"",
				Cost:  LoyaltyCost(-6),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateEmblem{}.Apply(NewContext(g, item))
				},
			},
		},
	})
}

const narsetAncientWayOracle = "b7ac6e89-941f-4ad9-a437-38ec38ec5263"

var (
	// The +1's mana: the colour picked, spendable only on a noncreature
	// spell. A pick, then a fixed-colour add, because a restriction
	// rides one fixed-colour slot and not a multi-option one. The key
	// is an on-disk identity.
	narsetAncientWayMana = OptionPickThen("option-pick/narset-of-the-ancient-way-mana",
		func(ctx *Context, r game.OptionPicked) error {
			if r.Option == nil {
				return nil
			}
			switch r.Option.Label {
			case "{U}", "{R}", "{W}":
			default:
				return nil
			}
			return AddMana{
				Produced:     r.Option.Label,
				Restrictions: []string{ManaRestrictCast, ManaRestrictNotType("Creature")},
			}.Apply(ctx)
		})

	// The −2's discard half: "then you may discard a card", run once the
	// draw has landed. A nonland discard creates the reflexive trigger
	// with the card's mana value. The key is an on-disk identity.
	narsetAncientWayDiscard = game.RegisterDrawThen("narset-of-the-ancient-way/may-discard", func(g *game.Game, d game.DrawThen) error {
		return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
			Player:   d.Player,
			Source:   d.Source,
			N:        1,
			UpTo:     true,
			Question: "Narset of the Ancient Way — you may discard a card",
		}, narsetAncientWayDiscarded, nil)
	})

	// The -2's reflexive trigger, raised once the discard has landed
	// (keyed, #4253). The key is an on-disk identity.
	narsetAncientWayDiscarded = game.RegisterDiscardThen("choose-cards/narset-of-the-ancient-way-discarded", func(g *game.Game, r game.DiscardsDone) error {
		parent := &game.StackItem{Controller: r.Player, SourceCardID: r.Source, Label: "Narset of the Ancient Way"}
		for _, id := range r.Discarded.Cards() {
			c, ok := g.LookupCardForEffect(id)
			if !ok || c.IsLand() {
				continue
			}
			mv, _ := g.ManaValueForEffect(c)
			t := WhenYouDo("Narset of the Ancient Way — damage equal to that card's mana value to target creature or planeswalker", narsetAncientWayDamageBody)
			t.Params = game.EffectParams{Amount: mv}
			return t.Apply(NewContext(g, parent))
		}
		return nil
	})

	// "When you discard a nonland card this way, Narset deals damage
	// equal to that card's mana value to target creature or planeswalker."
	// Breeches' damage body, the mana value in Params.Amount.
	narsetAncientWayDamageBody = game.ReflexiveBody("narset-of-the-ancient-way/damage", breechesBlastEffect,
		constTargets(func() *game.TargetSpec {
			return TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker()))
		}))
)

// narsetAncientWayEmblemDamage is the emblem's 2 damage to its target.
func narsetAncientWayEmblemDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	t, ok := firstLegalTarget(ctx)
	if !ok {
		return nil
	}
	return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 2}.Apply(ctx)
}
