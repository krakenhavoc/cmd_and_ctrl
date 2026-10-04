package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flayer of Loyalties — Creature — Eldrazi {10}, 10/10:
//
//	"When you cast this spell, gain control of target creature until
//	 end of turn. Untap that creature. Until end of turn, it has base
//	 power and toughness 10/10 and gains trample, annihilator 2, and
//	 haste.
//	 Annihilator 2
//	 Trample"
//
// The cast trigger targets as the Flayer is cast and resolves above
// it, so the creature is taken even if the Flayer is countered. The
// creature becomes yours until end of turn, untaps, has base power and
// toughness 10/10 (layer 7b, so its counters and pumps still apply on
// top), and gains trample, haste and annihilator 2 — the keyword the
// engine turns into an attack trigger (ADR 0113 §2). A target that is
// gone or no longer a creature by resolution is left alone (CR 608.2b).
//
// No simplification.
func init() {
	cast := WhenYouCastThisSpell("Flayer of Loyalties — gain control of target creature until end of turn",
		flayerOfLoyaltiesSteal)
	cast.Targets = TargetCreature("target creature")
	Register(Spec{
		OracleID:        "1e9053b5-5cca-486c-9dd8-b198a7b666bf",
		Name:            "Flayer of Loyalties",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"annihilator 2", "trample"},
		Triggered:       []game.TriggeredAbility{cast},
	})
}

// flayerOfLoyaltiesSteal takes the first still-legal target until end
// of turn, untaps it, sets its base power and toughness to 10/10 and
// gives it trample, annihilator 2 and haste.
func flayerOfLoyaltiesSteal(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, ref := range ctx.LegalTargets() {
		if ref.Kind != game.TargetCard {
			continue
		}
		if err := (ScopedEffectFor{
			Target:   ref.ID,
			Mods:     game.SetBasePTMods(10, 10),
			Duration: DurationUntilEndOfTurn(ctx),
			Label:    "Flayer of Loyalties — base power and toughness 10/10",
		}).Apply(ctx); err != nil {
			return err
		}
		return threatenAndGrant(ctx, ref.ID, "Flayer of Loyalties", "trample", "annihilator 2", "haste")
	}
	return nil
}
