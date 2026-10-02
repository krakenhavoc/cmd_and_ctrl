package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Orcish Healer — Creature — Orc Cleric {R}{R}, 1/1:
//
//	"{R}{R}, {T}: Target creature can't be regenerated this turn.
//	 {B}{B}{R}, {T}: Regenerate target black or green creature.
//	 {R}{G}{G}, {T}: Regenerate target black or green creature."
//
// The first ability is ADR 0108 §2's turn-long mark (CR 701.19c): a
// shield the creature already has, or gets later this turn, is not
// applied. The other two are the same regeneration (CR 701.19a) for two
// different costs; colour is read as each resolves (CR 608.2b).
//
// No simplifications.
func init() {
	blackOrGreen := func() *game.TargetSpec {
		return TargetCreature("target black or green creature", Or(OfColor("B"), OfColor("G")))
	}
	Register(Spec{
		OracleID:     "9b7a0c9d-1045-4fc0-b5e3-18616106da08",
		Name:         "Orcish Healer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{R}{R}, {T}: Target creature can't be regenerated this turn.",
				Cost:    Plus(ManaCost("{R}{R}"), TapCost()),
				Targets: TargetCreature("target creature"),
				Effect:  targetCantBeRegeneratedThisTurn,
			},
			{
				Label:   "{B}{B}{R}, {T}: Regenerate target black or green creature.",
				Cost:    Plus(ManaCost("{B}{B}{R}"), TapCost()),
				Targets: blackOrGreen(),
				Effect:  regenerateTheTargetPermanent,
			},
			{
				Label:   "{R}{G}{G}, {T}: Regenerate target black or green creature.",
				Cost:    Plus(ManaCost("{R}{G}{G}"), TapCost()),
				Targets: blackOrGreen(),
				Effect:  regenerateTheTargetPermanent,
			},
		},
	})
}
