package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Daughter of the Deep — Creature — Merfolk Noble {1}{U}, 2/1 (Marvel
// Super Heroes Commander, #4455):
//
//	"Whenever you draw your second card each turn, create a 1/1 blue
//	 Merfolk creature token.
//	 {U}, {T}: Target Merfolk can't be blocked this turn."
//
// The trigger is Knights of Dol Amroth's second-draw read
// (b41DrewYourSecondCardThisTurn). The ability is Merfolk Sovereign's
// unblockable grant, read at declaration (CR 509.1b) through the shared
// msc17CantBeBlockedTheTarget; Daughter can target herself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "63e0f03f-16da-4c09-ad03-8945f660942d",
		Name:         "Daughter of the Deep",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Daughter of the Deep — create a 1/1 blue Merfolk token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 blue Merfolk"), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}, {T}: Target Merfolk can't be blocked this turn.",
			Purpose: effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
			Cost:    effects.Plus(effects.ManaCost("{U}"), effects.TapCost()),
			Targets: effects.TargetCreature("target Merfolk", effects.HasSubtype("Merfolk")),
			Effect:  effects.Msc17CantBeBlockedTheTarget("Daughter of the Deep — can't be blocked"),
		}},
	})
}
