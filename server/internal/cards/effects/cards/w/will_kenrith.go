package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Will Kenrith — Legendary Planeswalker — Will {4}{U}{U}, loyalty 4:
//
//	"+2: Until your next turn, up to two target creatures each have base
//	     power and toughness 0/3 and lose all abilities.
//	 −2: Target player draws two cards. Until your next turn, instant,
//	     sorcery, and planeswalker spells that player casts cost {2} less
//	     to cast.
//	 −8: Target player gets an emblem with "Whenever you cast an instant
//	     or sorcery spell, copy it. You may choose new targets for the
//	     copy."
//	 Partner with Rowan Kenrith
//	 Will Kenrith can be your commander."
//
// +2 is Warkite Marauder's pair of scoped effects (layer 6's ability loss
// and layer 7b's base 0/3), each on its own target, until Will's
// controller's next turn begins. Counters and anthems still apply on top
// of the new base.
//
// −2's reduction is the #6093 timed record
// (game.ModSpellsCostLessForPlayer) for the TARGET player, narrowed to
// instant, sorcery and planeswalker spells and lasting until Will's
// controller's next turn: {2} of generic mana comes off each (CR 601.2f),
// and nothing coloured.
//
// −8's emblem is Double Vision's copy without its first-each-turn gate,
// given to the target player (CR 114.2): "you" on it is that player, the
// copy is created rather than cast (CR 707.10), so it does not trigger the
// emblem again.
//
// "Partner with Rowan Kenrith" is the CR 702.124j entry search; "can be
// your commander" is deck construction, read from the oracle text by the
// deck checker.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "20dbfc9a-172e-4065-bfde-d40de3131787",
		Name:            "Will Kenrith",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Emblem: &effects.EmblemSpec{
			Label: "Will Kenrith emblem",
			Text:  "Whenever you cast an instant or sorcery spell, copy it. You may choose new targets for the copy.",
			Triggered: []game.TriggeredAbility{
				effects.On(game.EventCast, effects.YouCast(effects.Or(effects.Instant(), effects.Sorcery())),
					"Will Kenrith emblem — copy that spell; you may choose new targets for the copy",
					effects.CopyTheTriggeringSpellWithNewTargets),
			},
		},
		Triggered: []game.TriggeredAbility{
			effects.PartnerWith("Will Kenrith", "Rowan Kenrith"),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+2: Until your next turn, up to two target creatures each have base power and toughness 0/3 and lose all abilities.",
				Cost:    effects.LoyaltyCost(2),
				Targets: effects.TargetCreature("up to two target creatures").WithCount(0, 2),
				Effect:  willKenrithShrink,
			},
			{
				Label:   "−2: Target player draws two cards. Until your next turn, instant, sorcery, and planeswalker spells that player casts cost {2} less to cast.",
				Cost:    effects.LoyaltyCost(-2),
				Targets: effects.TargetPlayer("target player"),
				Purpose: effects.ForTargets(game.TargetPurpose{Slot: 0, Draws: 2}),
				Effect:  willKenrithDrawAndReduce,
			},
			{
				Label:   "−8: Target player gets an emblem with \"Whenever you cast an instant or sorcery spell, copy it. You may choose new targets for the copy.\"",
				Cost:    effects.LoyaltyCost(-8),
				Targets: effects.TargetPlayer("target player"),
				Effect:  effects.TargetPlayerGetsTheEmblem,
			},
		},
	})
}

// willKenrithShrink is +2: each still-legal target loses all abilities and
// has base power and toughness 0/3 until its controller's next turn.
func willKenrithShrink(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	d := effects.DurationUntilYourNextTurn(ctx, item.Controller)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (effects.ScopedEffectFor{
			Target:   t.ID,
			Mods:     append([]game.Mod{game.LoseAllAbilitiesMod()}, game.SetBasePTMods(0, 3)...),
			Duration: d,
			Label:    "Will Kenrith — base power and toughness 0/3 and loses all abilities until your next turn",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// willKenrithDrawAndReduce is −2.
func willKenrithDrawAndReduce(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	p, ok := effects.FirstLegalPlayerTarget(ctx)
	if !ok {
		return nil
	}
	if err := (effects.DrawCards{Player: p, N: 2}).Apply(ctx); err != nil {
		return err
	}
	g.RegisterScopedRuleEffectForEffect(item.SourceCardID, game.ScopeGame, item.Controller,
		[]game.Mod{game.SpellsCostLessForPlayerMod(p, 2, "instant", "sorcery", "planeswalker")},
		effects.DurationUntilYourNextTurn(ctx, item.Controller),
		"Will Kenrith — instant, sorcery, and planeswalker spells that player casts cost {2} less")
	return nil
}
