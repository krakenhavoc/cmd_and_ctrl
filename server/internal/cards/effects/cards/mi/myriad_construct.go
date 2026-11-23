package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Myriad Construct — Artifact Creature — Construct {4}, 4/4 (Commander
// Masters):
//
//	"Kicker {3}
//	 If this creature was kicked, it enters with a +1/+1 counter on it for
//	 each nonbasic land your opponents control.
//	 When this creature becomes the target of a spell, sacrifice it and
//	 create a number of 1/1 colorless Construct artifact creature tokens
//	 equal to its power."
//
// The counters are an entry replacement (CR 614.1c) that reads whether the
// spell was kicked (EntryCastCountsForEffect, CR 702.33d) and counts the
// nonbasic lands (no Basic supertype) the caster's opponents control as it
// enters. The trigger is Dream Strix's "becomes the target of a spell"
// (SelfTargetedByASpell) sacrifice, then the tokens: the count is the
// Construct's power as it last was on the battlefield (CR 608.2h), so a
// counter-laden one pays out its counters, and nothing is made if it was
// already gone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "f54111c8-9c20-4a73-a7c4-24d216e7de07",
		Name:          "Myriad Construct",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{3}")},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventMove && ev.NewZone == game.ZoneBattlefield &&
					src != nil && ev.CardID == src.InstanceID
			},
			Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
				if g.EntryCastCountsForEffect(ev).Kicked <= 0 {
					return nil
				}
				n := 0
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller != src.Controller && c.IsLand() && !c.HasSupertype("basic") {
						n++
					}
				}
				if n > 0 {
					ev.AddCounterAtETB(game.CounterPlusOne, n)
				}
				return nil
			},
			Label: "Myriad Construct: if kicked, enters with a +1/+1 counter for each nonbasic land your opponents control",
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesTarget, effects.SelfTargetedByASpell,
				"Myriad Construct — sacrifice it and create a number of 1/1 colorless Construct artifact creature tokens equal to its power",
				myriadConstructSacrifice),
		},
	})
}

func myriadConstructSacrifice(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if !effects.OnBattlefield(g, item.SourceCardID) {
		return nil
	}
	if err := (effects.SacrificePermanent{Target: item.SourceCardID}).Apply(ctx); err != nil {
		return err
	}
	power := 0
	if info, ok := g.LastKnownPermanentForEffect(item.SourceCardID); ok {
		power = info.Power
	}
	if power <= 0 {
		return nil
	}
	return effects.CreateToken{Template: effects.TokenCard("1/1 colorless Construct artifact"), N: power}.Apply(ctx)
}
