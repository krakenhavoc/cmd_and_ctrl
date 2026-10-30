package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Photon, Mighty Marvel — Legendary Creature — Human Hero {3}{R}, 2/4
// (Marvel Super Heroes Commander, #4455):
//
//	"Flying
//	 Whenever Photon deals combat damage to a player, add that much mana of
//	 any one color. Until end of turn, you don't lose this mana as steps
//	 and phases end."
//
// Boommobile's colour-then-mana shape: the damage dealt rides the tail as
// its N, the colour is chosen as the trigger resolves, and the mana carries
// the keep mark (KeepManaUntilEndOfTurn, Savage Ventmaw) so it survives the
// step boundaries and is lost as the cleanup step begins.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c46d6f42-5e50-4c0d-8908-360337285f83",
		Name:            "Photon, Mighty Marvel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Photon, Mighty Marvel — add that much mana of any one color, kept until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil || item.Trigger.Event.Amount <= 0 {
						return nil
					}
					ChooseColorTail(game.ColorForMana, g, item.Controller, item.SourceCardID,
						"Photon, Mighty Marvel — choose a color of mana to add",
						withN(NewContext(g, item).Tail(photonColorChosen), item.Trigger.Event.Amount))
					return nil
				}),
		},
	})
}

// photonColorChosen adds that much mana of the colour once it is chosen (the
// tail's Word), kept through the steps. An on-disk key.
var photonColorChosen = game.RegisterTail("then/photon-mighty-marvel-color-chosen", func(g *game.Game, o game.TailOutcome) error {
	return g.AddManaWithOptionsForEffect(o.Tail.Player, o.Tail.Source,
		"{"+o.Tail.Word+strconv.Itoa(o.Tail.N)+"}",
		game.AddManaOptions{Riders: []game.ManaSpendRider{KeepManaUntilEndOfTurn()}})
})
