package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Frontier Siege — Enchantment {3}{G}:
//
//	"As this enchantment enters, choose Khans or Dragons.
//	 • Khans — At the beginning of each of your main phases, add {G}{G}.
//	 • Dragons — Whenever a creature you control with flying enters,
//	   you may have it fight target creature you don't control."
//
// A #1572 anchor-word card: each bullet is gated on its word (ADR
// 0071).
//
//   - Khans is ONE printed ability with two trigger conditions, so it
//     is written as the two main-phase constructors sharing one gate
//     and one body; each main phase of yours triggers it once, an
//     extra one included. It is not a mana ability (CR 605.1b — it
//     triggers off a phase, not off a mana ability), so it uses the
//     stack and the {G}{G} lands in the pool as it resolves, early in
//     that main phase, and empties with it (CR 106.4).
//   - Dragons reads "with flying" on the entering creature, post-layer,
//     as it enters. The fighter is that creature: a creature that has
//     left the battlefield by resolution fights nothing (CR 701.14b),
//     and the target is re-checked as usual (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4cfaa5cf-cc3d-49a7-9544-38a8bb7e9ec1",
		Name:         "Frontier Siege",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseOptionAsEnters("Frontier Siege", "Khans", "Dragons"),
		Triggered: []game.TriggeredAbility{
			effects.WhenChosen("Khans", effects.AtYourPrecombatMain("Frontier Siege — add {G}{G} (first main phase)",
				effects.Do(effects.AddMana{Produced: "{G}{G}"}))),
			effects.WhenChosen("Khans", effects.AtYourPostcombatMain("Frontier Siege — add {G}{G} (second main phase)",
				effects.Do(effects.AddMana{Produced: "{G}{G}"}))),
			effects.WhenChosen("Dragons", effects.Optional(effects.Targeting(
				effects.On(game.EventETB, aFlyingCreatureEnteredUnderYourControl,
					"Frontier Siege — it fights target creature you don't control", frontierSiegeFight),
				effects.TargetCreature("target creature you don't control", effects.OpponentControls())),
				"Frontier Siege — have it fight target creature you don't control?")),
		},
	})
}

// aFlyingCreatureEnteredUnderYourControl is "whenever a creature you
// control with flying enters".
func aFlyingCreatureEnteredUnderYourControl(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	if !effects.CreatureEnteredUnderYourControl(ev, source, lki, g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && game.HasKeyword(&c, "flying")
}

// frontierSiegeFight is the Dragons body: the entering creature fights
// the chosen target.
func frontierSiegeFight(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	fighter := ctx.Trigger().Event.CardID
	for _, t := range ctx.LegalTargets() {
		return effects.B10Fight(ctx, fighter, t.ID)
	}
	return nil
}
