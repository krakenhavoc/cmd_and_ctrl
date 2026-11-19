package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Joven and Chandler — Legendary Creature — Human Rogue {2}{R}{R}{R}, 3/3
// (Mystery Booster Commander Edition):
//
//	"Whenever an artifact is put into a graveyard from the battlefield, put
//	 a number of +1/+1 counters equal to that artifact's mana value on Joven
//	 and Chandler.
//	 {R}{R}{R}, {T}: Gain control of target artifact until end of turn.
//	 Untap it. It gains haste until end of turn."
//
// The trigger is Disciple of the Vault's: anyone's artifact going from the
// battlefield to a graveyard, a token too (mana value 0, so no counters).
// The mana value is the artifact's as it last existed on the battlefield
// (CR 608.2h, departedManaValue). The source must still be the
// same permanent to get the counters (CounterOnThis's own guard).
//
// The activation is Zealous Conscripts's theft, for an artifact.
//
// No simplification.
func init() {
	const label = "Joven and Chandler — put +1/+1 counters equal to that artifact's mana value on this creature"
	dies := On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
		return b23ArtifactPutIntoGraveyardFromBattlefield(ev, g)
	}, label, func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		if item.Trigger == nil {
			return nil
		}
		n := departedManaValue(g, item.Trigger.Event.CardID)
		if n <= 0 {
			return nil
		}
		return CounterOnThis{Kind: game.CounterPlusOne, N: n}.Apply(ctx)
	})
	Register(Spec{
		OracleID:     "11db8545-eca6-43f5-b9e8-f302acef53a5",
		Name:         "Joven and Chandler",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{dies},
		Activated: []ActivatedAbility{{
			Label:   "{R}{R}{R}, {T}: Gain control of target artifact until end of turn. Untap it. It gains haste until end of turn.",
			Cost:    Plus(ManaCost("{R}{R}{R}"), TapCost()),
			Targets: TargetPermanent("target artifact", Artifact()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (GainControl{
						Target:   t.ID,
						Duration: DurationUntilEndOfTurn(ctx),
						Label:    "Joven and Chandler — gain control until end of turn",
					}).Apply(ctx); err != nil {
						return err
					}
					if err := (UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{
						Target:   t.ID,
						Keywords: []string{"haste"},
						Label:    "Joven and Chandler — haste until end of turn",
					}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
