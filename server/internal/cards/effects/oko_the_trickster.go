package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Oko, the Trickster — Legendary Planeswalker — Oko {4}{G}{U}, loyalty
// 4:
//
//	"+1: Put two +1/+1 counters on up to one target creature you control.
//	 0: Until end of turn, Oko becomes a copy of target creature you
//	 control. Prevent all damage that would be dealt to him this turn.
//	 −7: Until end of turn, each creature you control has base power and
//	 toughness 10/10 and gains trample."
//
// ADR 0108 §7, Delivery PR 7 (#1904): the 0 is Mirage Mirror's "becomes
// a copy of target creature until end of turn" on Oko himself, then the
// not-one-use shield pinned to him for the rest of the turn. As a copy
// he is a creature and not a planeswalker (CR 707.2), keeping his
// loyalty counters; the shield is what keeps damage off him either way.
// The −7 fixes the creatures as it resolves (CR 611.2c).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "f4bf8bd1-71b0-4eb1-976c-54677238b2a6",
		Name:            "Oko, the Trickster",
		Completeness:    CompletenessFull,
		StartingLoyalty: 4,
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Put two +1/+1 counters on up to one target creature you control.",
				Cost:    LoyaltyCost(1),
				Targets: TargetCreature("up to one target creature you control", YouControl()).WithCount(0, 1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetCard {
							return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
						}
					}
					return nil
				},
			},
			{
				Label:   "0: Until end of turn, Oko becomes a copy of target creature you control. Prevent all damage that would be dealt to him this turn.",
				Cost:    LoyaltyCost(0),
				Targets: TargetCreature("target creature you control", YouControl()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if sourceIsNewObject(g, item) || !onBattlefield(g, item.SourceCardID) {
						return nil
					}
					if ts := ctx.LegalTargets(); len(ts) > 0 {
						if err := (BecomeCopy{Targets: []uuid.UUID{item.SourceCardID}, Of: ts[0].ID, Label: "Oko, the Trickster — becomes a copy"}).Apply(ctx); err != nil {
							return err
						}
					}
					return shieldThis(g, item, PreventDamageFromSource{})
				},
			},
			{
				Label: "−7: Until end of turn, each creature you control has base power and toughness 10/10 and gains trample.",
				Cost:  LoyaltyCost(-7),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return ScopedEffectFor{
						Match: And(Creature(), YouControl()),
						Mods: append(game.SetBasePTMods(10, 10),
							game.AddKeywordsMod("trample")),
						Duration: DurationUntilEndOfTurn(ctx),
						Label:    "Oko, the Trickster — base 10/10 and trample",
					}.Apply(ctx)
				},
			},
		},
	})
}
