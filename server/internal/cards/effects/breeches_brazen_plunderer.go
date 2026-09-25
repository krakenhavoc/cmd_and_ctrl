package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Breeches, Brazen Plunderer — 3/3 Legendary Creature — Goblin
// Pirate for {3}{R}:
//
//	"Menace
//	 Whenever one or more Pirates you control deal damage to your
//	 opponents, exile the top card of each of those opponents'
//	 libraries. You may play those cards this turn, and you may
//	 spend mana as though it were mana of any color to cast those
//	 spells.
//	 Partner"
//
// Ragavan's effect widened to the whole team, and with the two
// clauses that Ragavan lacks: "play" rather than "cast", so a land
// off the top is castable — sorry, playable — and the mana
// relaxation that makes an off-colour steal actually usable in a
// two-colour deck.
//
// "One or more … deal damage to your opponents" — no "combat" in
// the printed text, so a Pirate's non-combat damage (an activated or
// triggered ability, not just an attack) counts too:
// damagedOpponentByAnyDamage, not damagedOpponent.
//
// It is also CR 603.2c's per-player collapse (#784): the engine
// still emits one damage event per source, but OncePerBatchPerPlayer
// fires this trigger at most once per opponent per damage batch, so
// two Pirates connecting with the SAME opponent in one batch now
// exile exactly one card from them, matching the printed card. Two
// Pirates hitting TWO DIFFERENT opponents in the same batch still
// correctly produce two separate triggers, one per opponent — which
// is what "exile the top card of EACH of those opponents' libraries"
// needs, since each trigger exiles from its own one victim.
//
// Partner is a deck-construction rule, not a game action, and is
// not modelled.
func init() {
	Register(Spec{
		OracleID:        "eb77f7dc-e9e4-44ef-8616-9f4e737e8ca5",
		Name:            "Breeches, Brazen Plunderer",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"Partner isn't supported, so Breeches can't be your commander."},
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			OncePerBatchPerPlayer(game.TriggeredAbility{
				Watches: []game.EventKind{game.EventDealDamage},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					victim := damagedOpponentByAnyDamage(ev, source.Controller, g)
					if victim == uuid.Nil {
						return false
					}
					dealer, ok := g.LookupCardForEffect(ev.Source)
					return ok && isPirate(dealer)
				},
				Key: "Breeches — exile their top card, playable this turn",
				Effect: func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					return ExileTopWithPermission{
						From:     item.Trigger.Event.Target,
						GrantTo:  item.Controller,
						N:        1,
						AnyColor: true,
					}.Apply(NewContext(g, item))
				},
			}),
		},
	})
}
