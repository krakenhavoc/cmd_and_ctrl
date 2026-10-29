package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Desert Were-Worm — Creature — Dragon Wurm {4}{R}{R}, 0/5 (The Hobbit, HOB):
//
//	"This creature gets +2/+0 for each Mountain you control.
//	 Whenever you attack with creatures with total power 12 or greater for
//	 the first time each turn, untap all attacking creatures. After this
//	 phase, there is an additional combat phase."
//
// The pump is a layer 7c static on the Worm itself, re-read on every pass
// like Blanchwood Armor's Forest count: "Mountain" is a word of the Worm's
// text (CR 612.2), so a changed word counts the new land type.
//
// The trigger is Karlach's attack trigger, one per declaration
// (OncePerBatch over EventAttack), gated on the board. "Creatures with total
// power 12 or greater" is read off the creatures its controller is attacking
// with when the attack is declared, each at its power after the layers.
// "For the first time each turn" means the first time that attack met the
// total: an attack of 8 power in the first combat does not use it up, and
// the additional combat it makes can trigger it only if it has not already
// (b11TriggeredThisTurn, the engine's per-object trigger tally), so the
// combat it adds never chains a third. The body is Hellkite Charger's: untap
// every attacking creature, then add one combat phase after this one, with no
// main phase (the 2022 Karlach ruling).
//
// No simplification.
func init() {
	const label = "Desert Were-Worm — untap all attacking creatures; after this phase, an additional combat phase"
	Register(Spec{
		OracleID:     "81c2913f-42b7-4cfa-adc0-b3416fbe2ec6",
		Name:         "Desert Were-Worm",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "This creature gets +2/+0 for each Mountain you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				mountain := source.TextSubtype("Mountain")
				c.Power += 2 * permanentsYouControl(g, source, func(o *game.Card) bool { return o.HasSubtype(mountain) })
			},
		}},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller) &&
					desertWereWormAttackPower(g, source) >= 12 &&
					!b11TriggeredThisTurn(g, source.InstanceID, label)
			}, label, Do(UntapAttackingCreatures{}, ExtraCombatAfterThisPhase()))),
		},
	})
}

// desertWereWormAttackPower is the total power of the creatures the Worm's
// controller is attacking with right now.
func desertWereWormAttackPower(g *game.Game, source *game.Card) int {
	total := 0
	for _, id := range AttackingCreatures(g) {
		if c, ok := g.LookupCardForEffect(id); ok && c.Controller == source.Controller {
			total += c.CurrentPower()
		}
	}
	return total
}
