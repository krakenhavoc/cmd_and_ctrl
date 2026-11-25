package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const henryWuExploitGrant = "henry-wu-ingen-geneticist/exploit"

// Henry Wu, InGen Geneticist — Legendary Creature — Human Scientist
// {B}{G}{U}, 1/4:
//
//	"Henry Wu and other Human creatures you control have exploit. (When a
//	 creature with exploit enters, you may sacrifice a creature.)
//	 Whenever a creature you control exploits a non-Human creature, draw a
//	 card. If the exploited creature had power 3 or greater, create a
//	 Treasure token."
//
// Exploit is a triggered ability (CR 702.110a), so "have exploit" is a
// layer-6 grant of the shared Exploit trigger (ADR 0093): a Human entering
// under Henry Wu already has it as it enters (CR 603.6a), and each
// recipient's own exploit sacrifices for that recipient. Henry Wu is
// named on its own, so it keeps exploit even if it stops being a Human.
//
// The second ability watches game.EventExploited (CR 702.110b). Exploiting
// is a sacrifice, so it looks back (CR 603.10a): the exploited creature is
// judged non-Human as it last existed on the battlefield, and its power is
// read from that same last-known record as the ability resolves
// (CR 608.2h).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d94a792f-a48b-4181-b656-86fe452a4875",
		Name:         "Henry Wu, InGen Geneticist",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:       henryWuExploitGrant,
			Triggered: []game.TriggeredAbility{effects.Exploit("Henry Wu, InGen Geneticist", game.BodyRef{})},
			Text:      "Exploit (When this creature enters, you may sacrifice a creature.)",
		}},
		Static: []game.StaticAbility{henryWuGrantsExploit()},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventExploited, henryWuExploitedANonHuman,
				"Henry Wu — draw a card, and a Treasure if the exploited creature had power 3 or greater",
				henryWuDrawAndTreasure),
		},
	})
}

func henryWuGrantsExploit() game.StaticAbility {
	humans := effects.TribeFilter{Tribes: []string{"Human"}, YoursOnly: true}
	ab := effects.GrantAbilities(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID || humans.Matches(target, g, source)
	}, henryWuExploitGrant)
	ab.Label = "Henry Wu and other Human creatures you control have exploit."
	return ab
}

// henryWuExploitedANonHuman is "a creature you control exploits a
// non-Human creature": the exploiter's controller as it exploited, and the
// sacrificed creature as it last existed on the battlefield.
func henryWuExploitedANonHuman(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventExploited || ev.Actor != source.Controller {
		return false
	}
	if was := g.SimultaneousExitCardsForEffect([]uuid.UUID{ev.Target}); len(was) == 1 {
		return was[0].IsCreature() && !was[0].HasSubtype(source.TextSubtype("Human"))
	}
	if info, ok := g.LastKnownPermanentForEffect(ev.Target); ok {
		was := effects.CharacteristicCard(ev.Target, game.Card{}, info.Characteristic)
		return was.IsCreature() && !was.HasSubtype(source.TextSubtype("Human"))
	}
	return false
}

func henryWuDrawAndTreasure(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if err := g.DrawNForEffect(ctx.Controller(), 1); err != nil {
		return err
	}
	exploited := ctx.Trigger().Event.Target
	if info, ok := g.LastKnownPermanentForEffect(exploited); ok && info.Power >= 3 {
		return effects.TreasureTokens(g, ctx.Controller(), 1)
	}
	return nil
}
