package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "bb7638ce-3802-4198-89d2-6f5b647bd8fb",
		Name:         "Moon-Boy, Dino Rider",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Dinosaur spells you cast cost {1} less to cast.",
				YourSpell(), OtherSpellOfCreatureType("Dinosaur")),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attackDeclared(ev, source) && b05ControlsSubtype(g, source.Controller, source.TextSubtype("Dinosaur"))
			}, "Moon-Boy — gets +1/+1 until end of turn", thisGetsUntilEndOfTurn(1, 1, "Moon-Boy, Dino Rider")),
		},
	})
}
