package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Elder Brain — Creature — Horror {5}{B}{B}, 6/6 (Commander Legends: Battle
// for Baldur's Gate, #5948):
//
//	"Menace
//	 Whenever this creature attacks a player, exile all cards from that
//	 player's hand, then they draw that many cards. You may play lands and
//	 cast spells from among the exiled cards for as long as they remain
//	 exiled. If you cast a spell this way, you may spend mana as though it
//	 were mana of any color to cast it."
//
// The trigger is "attacks a player", so an attack on a planeswalker or a battle
// is not one (CR 508.3a). The attacked player's whole hand is exiled face up in
// one event, linked to the Elder Brain (CR 607.2a; ExiledWithSource), and they
// draw as many cards as landed in exile, once the exile has settled. The
// permission is a standing one over the cards exiled with this permanent: it
// covers lands (a land drop is used, CR 305.2) and spells, whoever owns them,
// lasts as long as they stay exiled and ends if the Elder Brain leaves, and
// lets you spend mana as though it were of any colour (CR 609.4b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e537bd23-7368-459b-badf-b7b7c112c88a",
		Name:            "Elder Brain",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.CardID == source.InstanceID && effects.AttackedPlayer(g, ev) != uuid.Nil
			}, "Elder Brain — exile all cards from that player's hand, then they draw that many cards", elderBrainAttack),
		},
		CastPermissions: []game.CastPermission{{
			Zone:     game.ZoneExile,
			Filter:   game.PermissionFilter{ExiledWithSource: true},
			AnyColor: true,
			Label:    "Play lands and cast spells from among the cards exiled with this permanent; spend mana as though it were mana of any color (Elder Brain)",
		}},
	})
}

func elderBrainAttack(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	victim := effects.AttackedPlayer(g, item.Trigger.Event)
	p := g.PlayerByIDForEffect(victim)
	if victim == uuid.Nil || p == nil || p.Eliminated {
		return nil
	}
	ctx := effects.NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	var hand []uuid.UUID
	if p.Hand != nil {
		for _, c := range p.Hand.Cards {
			hand = append(hand, c.InstanceID)
		}
	}
	return g.ExileLinkedKeyedForEffect(ref, hand, game.KeyedTail{Ref: elderBrainDraw, Player: victim, Source: item.SourceCardID})
}

// elderBrainDraw is "then they draw that many cards", told what reached
// exile. An on-disk key: never renamed, never reused.
var elderBrainDraw = effects.CardTail("then/elder-brain-draw", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return effects.DrawCards{Player: o.Tail.Player, N: len(o.IDs)}.Apply(ctx)
})
