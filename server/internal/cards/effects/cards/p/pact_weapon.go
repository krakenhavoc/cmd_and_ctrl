package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pact Weapon — Artifact — Equipment {3}{B} (Commander Legends: Battle for
// Baldur's Gate, #5948):
//
//	"As long as this Equipment is attached to a creature, you don't lose the
//	 game for having 0 or less life.
//	 Whenever equipped creature attacks, draw a card and reveal it. The
//	 creature gets +X/+X until end of turn and you lose X life, where X is
//	 that card's mana value.
//	 Equip—Discard a card."
//
// The first ability is a game-end gate on life loss only (CR 704.5a), read at
// every check and only while the Equipment is attached to a creature, so
// unequipping it at 0 life loses the game at the next check. The attack
// trigger draws one card, reveals it, then pumps the creature it was equipped
// to and costs you life, both by that card's mana value (CR 202.3; a land is
// 0, an X in a mana cost is 0 off the stack). The drawn card is the one that
// arrived in your hand; when the draw was replaced or paused on a choice no
// card arrived and there is no X. Equip costs a discarded card rather than
// mana (CR 702.6a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bc2c02f6-2b6b-4016-b006-12ce0a1a2450",
		Name:         "Pact Weapon",
		Completeness: effects.CompletenessFull,
		GameEndGates: []game.GameEndGate{{
			Scope:    game.GateYou,
			CantLose: true,
			Causes:   []game.LossCause{game.LossLife},
			While: func(_ *game.Game, source game.Card) bool {
				return source.AttachedTo.Kind == game.TargetCard && source.AttachedTo.ID != uuid.Nil
			},
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttachedCreatureAttacked(ev, source)
			},
			Key:    "Pact Weapon — draw a card and reveal it; the creature gets +X/+X and you lose X life",
			Effect: pactWeaponAttack,
		}},
		Activated: []effects.ActivatedAbility{
			effects.EquipPayingAbility("Equip—Discard a card", effects.DiscardACard()),
		},
	})
}

func pactWeaponAttack(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	you := ctx.Controller()
	p := g.PlayerByIDForEffect(you)
	if p == nil || p.Hand == nil {
		return nil
	}
	before := map[uuid.UUID]bool{}
	for _, c := range p.Hand.Cards {
		before[c.InstanceID] = true
	}
	if err := (effects.DrawCards{Player: you, N: 1}).Apply(ctx); err != nil {
		return err
	}
	var drawn *game.Card
	for i := range p.Hand.Cards {
		if !before[p.Hand.Cards[i].InstanceID] {
			drawn = &p.Hand.Cards[i]
			break
		}
	}
	if drawn == nil {
		return nil
	}
	x := drawn.ManaValue()
	g.RevealForEffect(game.RevealSpec{Player: you, Source: ctx.Source(), Reason: "Pact Weapon — reveal the drawn card", Cards: []uuid.UUID{drawn.InstanceID}})
	if src, ok := g.LookupCardForEffect(ctx.Source()); ok && src.AttachedTo.Kind == game.TargetCard && src.AttachedTo.ID != uuid.Nil {
		if err := (effects.BoostUntilEOT{Target: src.AttachedTo.ID, Power: x, Toughness: x, Label: "Pact Weapon — +X/+X until end of turn"}).Apply(ctx); err != nil {
			return err
		}
	}
	return g.ChangePlayerLifeForEffect(ctx.Source(), you, -x)
}
