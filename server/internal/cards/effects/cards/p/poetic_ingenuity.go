package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Poetic Ingenuity — Enchantment {2}{R} (The Lost Caverns of Ixalan,
// #4470):
//
//	"Whenever one or more Dinosaurs you control attack, create that many
//	 Treasure tokens.
//	 Whenever you cast an artifact spell, create a 3/1 red Dinosaur
//	 creature token. This ability triggers only once each turn."
//
// The attack trigger is The Ur-Dragon's: one trigger per declaration
// (OncePerBatch, as the engine emits EventAttack per creature), and
// "that many" is the number of Dinosaurs declared in that batch, counted
// from the events so one that has left combat by resolution still
// counts (attackersOfSubtypeInTheSameBatch). The cast trigger is
// WheneverYouCast with Mirelurk Queen's once-each-turn limit,
// tallied as the ability is put on the stack. "Dinosaur" is a word of
// the card's text (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9780d975-acbe-40c6-b444-1e1e6b772788",
		Name:         "Poetic Ingenuity",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AttackDeclaredByYou(ev, source.Controller) && effects.AttackerHasSubtype(g, ev, source.TextSubtype("Dinosaur"))
			}, "Poetic Ingenuity — create that many Treasure tokens", poeticIngenuityTreasures)),
			effects.TriggersOnlyOnceEachTurn(effects.WheneverYouCast(effects.Artifact(), "Poetic Ingenuity — create a 3/1 red Dinosaur",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("3/1 red Dinosaur"), N: 1}))),
		},
	})
}

// poeticIngenuityTreasures is the attack trigger's resolution: one
// Treasure for each Dinosaur its controller declared in that batch.
func poeticIngenuityTreasures(g *game.Game, item *game.StackItem) error {
	n := effects.AttackersOfSubtypeInTheSameBatch(g, item.Trigger.Event, item.Controller, g.TextReaderForEffect(item).TextSubtype("Dinosaur"))
	if n <= 0 {
		return nil
	}
	return effects.CreateToken{Template: effects.TreasureToken(), N: n}.Apply(effects.NewContext(g, item))
}
