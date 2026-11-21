package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Master, Mesmerist — Legendary Creature — Time Lord Rogue {2}{U}{B}, 3/3
// (Doctor Who):
//
//	"{T}: Target creature an opponent controls with power less than or equal to
//	 The Master's power gains skulk until end of turn. Goad it. (A creature
//	 with skulk can't be blocked by creatures with greater power.)
//	 Whenever a creature with skulk deals combat damage to one of your
//	 opponents, put a +1/+1 counter on The Master and draw a card."
//
// The target compares with The Master's power as the ability goes on the stack
// and again as it resolves (CR 608.2b); the skulk and the goad (CR 701.15) go
// to the same creature. The trigger is any creature's, whoever controls it,
// with skulk read through the layers when the damage is dealt; it fires once
// per creature that dealt damage. The counter and the draw are one ability, so
// the draw happens even when The Master has left.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "306b6b9b-f488-4ea8-9f5e-9e8e7354df64",
		Name:         "The Master, Mesmerist",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Target creature an opponent controls with power less than or equal to The Master's power gains skulk until end of turn. Goad it.",
			Purpose: gapB85SkulkGrantPurpose(),
			Cost:    TapCost(),
			Targets: RelativeToSource(TargetCreature("target creature an opponent controls with power less than or equal to The Master's power", OpponentControls()), PowerNoGreater()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"skulk"}, Label: "The Master, Mesmerist — gains skulk until end of turn"}).Apply(ctx); err != nil {
					return err
				}
				return goadAndScheduleClear(ctx, []uuid.UUID{id})
			},
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !creatureDealtCombatDamageToAnOpponentOf(ev, source.Controller, g) {
					return false
				}
				dealer, ok := g.LookupCardForEffect(ev.Source)
				return ok && HasKeyword("skulk")(g, source.Controller, dealer)
			}, "The Master, Mesmerist — put a +1/+1 counter on The Master and draw a card",
				func(g *game.Game, item *game.StackItem) error {
					if b09SourceStillOnBattlefield(g, item) {
						if err := plusOneCountersOnThis(1)(g, item); err != nil {
							return err
						}
					}
					return DrawCards{N: 1}.Apply(NewContext(g, item))
				}),
		},
	})
}
