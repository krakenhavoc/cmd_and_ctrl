package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gollum's Bite — Instant {B}:
//
//	"Target creature gets -2/-2 until end of turn.
//	 {3}{B}, Exile this card from your graveyard: The Ring tempts you.
//	 Activate only as a sorcery."
//
// The second ability works only from the graveyard (Zones), and
// exiling the card is part of its cost, as Buried Treasure's is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cf0492de-cec5-4455-839f-212246b7e9ea",
		Name:         "Gollum's Bite",
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, -2, -2)),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if err := (effects.BoostUntilEOT{Target: t.ID, Power: -2, Toughness: -2, Label: "Gollum's Bite — -2/-2"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{3}{B}, Exile this card from your graveyard: The Ring tempts you",
			Purpose:      game.Purpose{RingTempts: 1},
			Cost:         effects.Plus(effects.ManaCost("{3}{B}"), effects.ExileThis()),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Effect:       effects.Do(effects.TheRingTemptsYou{}),
		}},
	})
}
