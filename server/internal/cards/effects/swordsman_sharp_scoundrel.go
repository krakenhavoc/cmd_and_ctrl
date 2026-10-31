package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Swordsman, Sharp Scoundrel — Legendary Creature — Human Hero Villain
// {1}{B}, 2/2 (Marvel Super Heroes, #4455):
//
//	"Whenever another Villain you control enters, attach up to one target
//	 Equipment you control to target creature you control.
//	 Whenever an equipped creature you control attacks, it connives."
//
// The enter trigger is two target clauses in printed order, the Equipment
// ("up to one") and then the creature, each re-checked as the ability
// resolves (CR 608.2b), and the attach is Cloud's (AttachForEffect). "Villain"
// is a word of the card's text (mshB2AnotherVillainEntered reads it that way).
// The attack trigger judges "equipped" as the attack is declared and
// connives the attacker (ConniveEach, CR 701.50a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2f15abe7-54aa-40ed-a15d-5b827f1c3923",
		Name:         "Swordsman, Sharp Scoundrel",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return mshB2AnotherVillainEntered(ev, source, g)
				}, "Swordsman — attach up to one target Equipment you control to target creature you control",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						gear, ok := ctx.ClauseTarget(0)
						if !ok {
							return nil
						}
						creature, ok := ctx.ClauseTarget(1)
						if !ok {
							return nil
						}
						return g.AttachForEffect(gear.ID, game.TargetRef{Kind: game.TargetCard, ID: creature.ID})
					}),
				Clauses(
					TargetPermanent("up to one target Equipment you control", HasSubtype("Equipment"), YouControl()).WithCount(0, 1),
					TargetCreature("target creature you control", YouControl()),
				)),
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.CardID == uuid.Nil || ev.Actor != source.Controller {
					return false
				}
				attacker, ok := g.LookupCardForEffect(ev.CardID)
				return ok && attacker.Controller == source.Controller && isEquippedCreature(g, attacker)
			}, "Swordsman — the equipped creature connives",
				func(g *game.Game, item *game.StackItem) error {
					return ConniveEach{Controller: item.Controller, Creatures: []uuid.UUID{item.Trigger.Event.CardID}}.Apply(NewContext(g, item))
				}),
		},
	})
}
