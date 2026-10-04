package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "cf0492de-cec5-4455-839f-212246b7e9ea",
		Name:         "Gollum's Bite",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if err := (BoostUntilEOT{Target: t.ID, Power: -2, Toughness: -2, Label: "Gollum's Bite — -2/-2"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
		Activated: []ActivatedAbility{{
			Label:        "{3}{B}, Exile this card from your graveyard: The Ring tempts you",
			Cost:         Plus(ManaCost("{3}{B}"), ExileThis()),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Effect:       Do(TheRingTemptsYou{}),
		}},
	})
}
