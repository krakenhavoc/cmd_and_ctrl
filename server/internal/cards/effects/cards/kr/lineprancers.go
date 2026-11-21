package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lineprancers — Creature — Centaur Performer {1}{G}, 2/2:
//
//	"When Lineprancers enters, you get {TK}{TK}, then you may put a power
//	 and toughness sticker on a creature you own.
//	 {3}{G}: Target creature you don't control blocks target creature you
//	 control with a power and toughness sticker on it other than
//	 Lineprancers this turn if able."
//
// The sticker is the engine's keyed questions with only power and
// toughness stickers offered, on creatures (ADR 0208 §4). The block is
// Hunt Down's requirement (BlocksAttackerUntilEOT): two positional clauses,
// the first the blocker and the second the creature it must block if it
// attacks this turn (CR 509.1c), pinned to its current object.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "71433aef-5a06-449e-93b1-9565cb0afd93",
		Name:         "Lineprancers",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Lineprancers — you get {TK}{TK}, then you may put a power and toughness sticker on a creature you own",
				effects.Do(effects.GetTickets{N: 2}, effects.PutASticker{On: effects.StickerOnCreatureYouOwn, Kinds: []game.StickerKind{game.StickerPT}})),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "{3}{G}: Target creature you don't control blocks target creature you control with a power and toughness sticker on it other than Lineprancers this turn if able.",
			Cost:  effects.ManaCost("{3}{G}"),
			Targets: effects.Clauses(
				effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
				effects.Another(effects.TargetCreature("target creature you control with a power and toughness sticker on it other than Lineprancers",
					effects.YouControl(), effects.WithSticker(game.StickerPT))),
			),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				blocker, ok := ctx.ClauseTarget(0)
				if !ok || blocker.Kind != game.TargetCard {
					return nil
				}
				named, ok := ctx.ClauseTarget(1)
				if !ok || named.Kind != game.TargetCard {
					return nil
				}
				ref, ok := g.PermanentRefForEffect(named.ID)
				if !ok {
					return nil
				}
				return effects.BlocksAttackerUntilEOT{
					Blocker:  blocker.ID,
					Attacker: ref,
					Label:    "Lineprancers — blocks the stickered creature if able",
				}.Apply(ctx)
			},
		}},
	})
}
