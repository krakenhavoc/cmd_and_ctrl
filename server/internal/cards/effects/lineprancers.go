package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "71433aef-5a06-449e-93b1-9565cb0afd93",
		Name:         "Lineprancers",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Lineprancers — you get {TK}{TK}, then you may put a power and toughness sticker on a creature you own",
				Do(GetTickets{N: 2}, PutASticker{On: StickerOnCreatureYouOwn, Kinds: []game.StickerKind{game.StickerPT}})),
		},
		Activated: []ActivatedAbility{{
			Label: "{3}{G}: Target creature you don't control blocks target creature you control with a power and toughness sticker on it other than Lineprancers this turn if able.",
			Cost:  ManaCost("{3}{G}"),
			Targets: Clauses(
				TargetCreature("target creature you don't control", OpponentControls()),
				Another(TargetCreature("target creature you control with a power and toughness sticker on it other than Lineprancers",
					YouControl(), WithSticker(game.StickerPT))),
			),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
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
				return BlocksAttackerUntilEOT{
					Blocker:  blocker.ID,
					Attacker: ref,
					Label:    "Lineprancers — blocks the stickered creature if able",
				}.Apply(ctx)
			},
		}},
	})
}
