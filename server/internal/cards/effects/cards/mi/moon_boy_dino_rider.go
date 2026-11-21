package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Moon-Boy, Dino Rider — Legendary Creature — Ape Human {1}{G}, 2/2
// (Marvel Super Heroes Commander, #4455):
//
//	"Dinosaur spells you cast cost {1} less to cast.
//	 Whenever Moon-Boy attacks while you control a Dinosaur, Moon-Boy gets
//	 +1/+1 until end of turn."
//
// The discount is Baron Strucker's board cost reduction on your own
// Dinosaur spells (a changeling counts). The attack trigger is Wargling's
// shape: the Dinosaur check is made once, as the attack is declared, and
// Moon-Boy gets the bonus if he is still on the battlefield when it
// resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bb7638ce-3802-4198-89d2-6f5b647bd8fb",
		Name:         "Moon-Boy, Dino Rider",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Dinosaur spells you cast cost {1} less to cast.",
				effects.YourSpell(), effects.OtherSpellOfCreatureType("Dinosaur")),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AttackDeclared(ev, source) && effects.B05ControlsSubtype(g, source.Controller, source.TextSubtype("Dinosaur"))
			}, "Moon-Boy — gets +1/+1 until end of turn", effects.ThisGetsUntilEndOfTurn(1, 1, "Moon-Boy, Dino Rider")),
		},
	})
}
