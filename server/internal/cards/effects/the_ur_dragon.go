package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Ur-Dragon — Legendary Creature — Dragon Avatar {4}{W}{U}{B}{R}{G},
// 10/10 (EDHREC rank 2641):
//
//	"Eminence — As long as The Ur-Dragon is in the command zone or on
//	 the battlefield, other Dragon spells you cast cost {1} less to cast.
//	 Flying
//	 Whenever one or more Dragons you control attack, draw that many
//	 cards, then you may put a permanent card from your hand onto the
//	 battlefield."
//
// The eminence line is ADR 0140's cost form, The Ur-Sphinx's shape
// with Dragon for Sphinx (#2802). The attack trigger is one per
// declaration (OncePerBatch), and "that many" is the Dragons declared
// in that batch, read off the log so a Dragon that died in response
// still counts (attackersOfSubtypeInTheSameBatch, shared with The
// Ur-Sphinx).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "87b22b09-4f6d-4bc5-9cfc-663e4c7c6981",
		Name:            "The Ur-Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		CostModifiers: []game.CostModifier{
			Eminence(CostsLess(1, "Eminence — other Dragon spells you cast cost {1} less to cast.",
				YourSpell(), OtherSpellOfCreatureType("Dragon"))),
		},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller) && attackerHasSubtype(g, ev, "Dragon")
			}, "The Ur-Dragon — draw that many cards, then you may put a permanent card from your hand onto the battlefield", urDragonDrawAndPut)),
		},
	})
}

// urDragonDrawAndPut is the attack trigger's body.
func urDragonDrawAndPut(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	n := attackersOfSubtypeInTheSameBatch(g, item.Trigger.Event, item.Controller, "Dragon")
	ctx := NewContext(g, item)
	if err := (DrawCards{Player: item.Controller, N: n}).Apply(ctx); err != nil {
		return err
	}
	return PutFromHandOntoBattlefield{
		Player:   item.Controller,
		Optional: true,
		Label:    "The Ur-Dragon — you may put a permanent card from your hand onto the battlefield",
	}.Apply(ctx)
}
