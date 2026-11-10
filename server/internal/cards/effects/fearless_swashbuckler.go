package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fearless Swashbuckler — Creature — Fish Pirate {1}{U}{R}, 3/3
// (Aetherdrift, #4464):
//
//	"Haste
//	 Vehicles you control have haste.
//	 Whenever you attack, if a Pirate and a Vehicle attacked this combat,
//	 draw three cards, then discard two cards."
//
// "Whenever you attack" is once per attack declaration of yours
// (OncePerBatch over attackDeclaredByYou), and the "if" is intervening
// (CR 603.4): the trigger waits for the settled declaration (AtBatchEnd with
// AttackersInclude, since the engine emits one attack event per creature) and
// is put on the stack only if a Pirate and a Vehicle are among the attackers,
// and InterveningIf checks it again as it resolves, reading the objects declared as attackers in this combat phase
// (game.AttackersThisCombatForEffect), so a creature that attacked and has
// since left still counts. One creature that is both a Pirate and a
// Vehicle satisfies both halves (the ruling). The draw and the discard
// are Careful Study's: the discard is chosen from the hand after the draw.
// The Vehicles' haste is a layer 6 grant over Vehicles you control.
//
// No simplification.
func init() {
	base := On(game.EventAttack,
		func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return attackDeclaredByYou(ev, source.Controller)
		},
		"Fearless Swashbuckler — draw three cards, then discard two cards",
		func(g *game.Game, item *game.StackItem) error {
			return drawThenDiscard(g, item.Controller, item.SourceCardID, 3, 2, "")
		})
	// The engine emits one attack event per attacker, so the "if" is asked
	// of the settled declaration (game.BatchEndCondition.AttackersInclude).
	base.AtBatchEnd = &game.BatchEndCondition{AttackersInclude: []string{"Pirate", "Vehicle"}}
	trigger := InterveningIf(CondPirateAndVehicleAttacked{}, OncePerBatch(base))
	Register(Spec{
		OracleID:        "6a6cdb3e-c850-4b3c-bf4b-c53dd4eaf56e",
		Name:            "Fearless Swashbuckler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Static: []game.StaticAbility{
			KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller && target.HasSubtype("Vehicle")
			}, "haste"),
		},
		Triggered: []game.TriggeredAbility{trigger},
	})
}
