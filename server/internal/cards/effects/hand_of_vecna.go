package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hand of Vecna — Legendary Artifact — Equipment {3} (Adventures in the
// Forgotten Realms, #5683):
//
//	"At the beginning of combat on your turn, equipped creature or a
//	 creature you control named Vecna gets +X/+X until end of turn,
//	 where X is the number of cards in your hand.
//	 Equip—Pay 1 life for each card in your hand.
//	 Equip {2}"
//
// The combat ability targets nothing. As it resolves its controller has
// one creature get the bonus: the creature the Hand is attached to,
// whoever controls it, or a creature they control named Vecna. With one
// such creature there is nothing to choose; with more, they choose (CR
// 608.2d), and X is the hand size as the ability resolves. On a Vecna
// made from an exiled Hand (The Book of Vile Darkness) the source is not
// an Equipment and equips nothing, so the creature is a Vecna its
// controller controls, normally itself.
//
// The first equip ability's cost is a life payment counted when it is
// activated (CR 601.2f through CR 602.2b): one life for each card in the
// activator's hand then, and nothing with an empty hand. A player can't
// pay more life than they have (CR 119.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3b7f0dd3-5b4c-4ac0-a43a-ddc62a44663e",
		Name:         "Hand of Vecna",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat("Hand of Vecna — equipped creature or a creature you control named Vecna gets +X/+X until end of turn", handOfVecnaPump),
		},
		Activated: []ActivatedAbility{
			EquipPayingAbility("Equip—Pay 1 life for each card in your hand", PayLifeCount(handOfVecnaLifePerCardInHand)),
			EquipAbility("{2}"),
		},
	})
}

// handOfVecnaLifePerCardInHand is "Pay 1 life for each card in your
// hand": the activator's hand size as the cost is determined.
var handOfVecnaLifePerCardInHand = game.LifeCount("one-per-card-in-your-hand",
	func(g *game.Game, activator, _ uuid.UUID) int {
		return b14HandSize(g, activator)
	})

// handOfVecnaPump is the combat ability's resolution: the one creature
// it can apply to gets the bonus, or its controller chooses among them.
func handOfVecnaPump(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	var options []game.ChoiceOption
	add := func(c game.Card, what string) {
		for _, id := range ids {
			if id == c.InstanceID {
				return
			}
		}
		ids = append(ids, c.InstanceID)
		options = append(options, game.ChoiceOption{Label: c.Name + what})
	}
	if host, ok := cmdr2B21EquippedCreature(g, item.SourceCardID); ok {
		add(host, " (equipped creature)")
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsCreature() && c.Effective().Name == "Vecna" {
			add(c, "")
		}
	}
	switch len(ids) {
	case 0:
		return nil
	case 1:
		return handOfVecnaBoost(ctx, ids[0])
	}
	return PickOption{
		Question: "Hand of Vecna — which creature gets +X/+X until end of turn?",
		Options:  options,
		ThenKey:  handOfVecnaChosen,
		Carry:    ids,
	}.Apply(ctx)
}

// handOfVecnaChosen applies the bonus to the creature the controller
// chose, keyed so a table waiting on the choice is a restore point.
var handOfVecnaChosen = OptionPickThen("option-pick/hand-of-vecna-creature",
	func(ctx *Context, r game.OptionPicked) error {
		if r.Index < 0 || r.Index >= len(r.Carry) {
			return nil
		}
		return handOfVecnaBoost(ctx, r.Carry[r.Index])
	})

// handOfVecnaBoost gives the creature +X/+X until end of turn, X the
// number of cards in the controller's hand now.
func handOfVecnaBoost(ctx *Context, creature uuid.UUID) error {
	c, ok := ctx.Game.LookupCardForEffect(creature)
	if !ok || !onBattlefield(ctx.Game, creature) || !c.IsCreature() {
		return nil
	}
	x := b14HandSize(ctx.Game, ctx.Controller())
	if x <= 0 {
		return nil
	}
	return BoostUntilEOT{Target: creature, Power: x, Toughness: x, Label: "Hand of Vecna"}.Apply(ctx)
}
