package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Carrion Rats — Creature — Rat {B}, 2/1:
//
//	"Whenever this creature attacks or blocks, any player may exile a
//	 card from their graveyard. If a player does, this creature assigns
//	 no combat damage this turn."
//
// ADR 0187 (#3571): the offer is ADR 0150's any-player offer, in APNAP
// order from the active player with the Rats' controller included, paid
// by exiling from the payer's OWN graveyard (PayActionExileFromGraveyard,
// CR 118.3: a payer with an empty graveyard can't pay). A payment ends the
// offer, and "assigns no combat damage" is the rule gate of CR 510.1a, not
// a prevention shield, so it holds even where damage can't be prevented.
// The gate lands on the Rats only while they are still the permanent the
// trigger came from (CR 400.7): a Rats that left and came back is a new
// object and is unaffected.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1aa6b31b-cd30-4741-abc6-2ef72e445630",
		Name:         "Carrion Rats",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{carrionTrigger("Carrion Rats", 1)},
	})
}

// carrionTrigger is the shared "whenever this creature attacks or blocks,
// any player may exile <n> cards from their graveyard. If a player does,
// this creature assigns no combat damage this turn" of Carrion Rats and
// Carrion Wurm.
func carrionTrigger(name string, n int) game.TriggeredAbility {
	words := "a card"
	if n != 1 {
		words = numberWord(n) + " cards"
	}
	// "If a player does" is keyed (#4264), one key per card, so a table
	// waiting on any player's answer is a restore point. The Rats must
	// still be the permanent the trigger came from (CR 400.7): the asking
	// item carries its object.
	noDamage := AnyPlayerConsequences("any-player/"+strings.ReplaceAll(strings.ToLower(name), " ", "-")+"-no-combat-damage",
		func(ctx *Context, _ game.AnyPlayerAnswer) error {
			if !sourceIsStillThisPermanent(ctx.Game, ctx.Item) {
				return nil
			}
			ctx.Game.AssignsNoCombatDamageThisTurnForEffect(ctx.Source(), ctx.Item.SourceCardID, name+" — assigns no combat damage this turn")
			return nil
		}, nil, nil)
	return OnAny([]game.EventKind{game.EventAttack, game.EventBlock}, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return attackDeclared(ev, source) || selfBlocksOnce(ev, source)
	}, name+" — any player may exile "+words+" from their graveyard to stop its combat damage",
		func(g *game.Game, item *game.StackItem) error {
			return AnyPlayerPays{
				Pay:      ExileFromGraveyardPayment(n),
				Question: name + " — exile " + words + " from your graveyard so it assigns no combat damage this turn?",
				Then:     noDamage,
			}.Apply(NewContext(g, item))
		})
}
