package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sephiroth, One-Winged Angel — Legendary Creature — Angel Nightmare
// Avatar, 5/5, the back face of Sephiroth, Fabled SOLDIER (Edea
// steal-and-sac deck, #1565):
//
//	"Flying
//	 Super Nova — As this creature transforms into Sephiroth,
//	 One-Winged Angel, you get an emblem with "Whenever a creature
//	 dies, target opponent loses 1 life and you gain 1 life."
//	 Whenever Sephiroth attacks, you may sacrifice any number of other
//	 creatures. If you do, draw that many cards."
//
// Registered under "<oracle_id>#1" (game.CatalogKey), and so is its
// emblem: CreateEmblemForEffect files an emblem under the SOURCE's
// catalog key, and the source is on this face when the emblem is made.
//
// The attack trigger is the front face's sacrifice helper with no cap:
// a resolution-time pick of any number of your other creatures, all
// sacrificed together, one card drawn for each.
//
// The emblem's trigger watches every creature death, yours included,
// from the command zone (CR 114.3), and targets an opponent when it
// goes on the stack like any trigger.
//
// DECLARED SIMPLIFICATION (weaker than printed): Super Nova is made by
// the front face's own transform, the fourth-resolution clause. The
// engine's in-place transform has no "as this transforms" hook a face
// could declare, so if some other effect transforms Sephiroth, no
// emblem is made. In this deck nothing else transforms it.
func init() {
	Register(Spec{
		OracleID:     sephirothOracleID + "#1",
		Name:         "Sephiroth, One-Winged Angel",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Only Sephiroth's own transform gives you the Super Nova emblem. If another card's effect transforms it, you don't get one.",
		},
		PrintedKeywords: []string{"flying"},
		Emblem: &EmblemSpec{
			Label: "Sephiroth, One-Winged Angel emblem",
			Text:  "Whenever a creature dies, target opponent loses 1 life and you gain 1 life.",
			Triggered: []game.TriggeredAbility{
				Targeting(On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
					_, ok := diedCreature(ev, g)
					return ok
				}, "Sephiroth emblem — target opponent loses 1 life and you gain 1 life", drainTargetOpponentOne),
					TargetPlayer("target opponent", Opponent())),
			},
		},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Sephiroth, One-Winged Angel — sacrifice any number of other creatures to draw that many",
				func(g *game.Game, item *game.StackItem) error {
					return sephirothSacrificeOthersThenDraw(g, item, 0)
				}),
		},
	})
}
