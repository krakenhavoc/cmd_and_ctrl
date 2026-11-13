package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Barbed Foliage — Enchantment {2}{G}{G} (Mirage, #5633, #5784):
//
//	"Whenever a creature attacks you, it loses flanking until end of
//	 turn.
//	 Whenever a creature without flying attacks you, this enchantment
//	 deals 1 damage to it."
//
// "Attacks YOU" is the player, as Hissing Miasma reads it: an EventAttack
// names the attacked player or planeswalker in Target, so a creature
// attacking one of your planeswalkers or a battle you protect triggers
// neither ability. One trigger per attacking creature.
//
// The first is a layer-6 ability-removing effect with its own timestamp
// (CR 613.1f), ending at cleanup (CR 514.2). Flanking is an engine
// keyword trigger derived from the token (game/flanking.go), so a
// creature that lost it has no flanking trigger when it becomes blocked,
// however many instances it had, printed or granted. A flanking granted
// with a later timestamp puts it back (CR 613.7).
//
// The second reads "without flying" as the creature attacks (CR 603.2):
// one that gains flying in response is still dealt the damage. The
// enchantment deals the damage as an object, so one that has left the
// battlefield still deals it; a creature that has left the battlefield
// is dealt none.
//
// No simplification.
func init() {
	const (
		loses  = "Barbed Foliage — it loses flanking until end of turn"
		damage = "Barbed Foliage — 1 damage to the attacking creature without flying"
	)
	Register(Spec{
		OracleID:     "f9275c99-0616-44b3-a19e-98c683e87ad3",
		Name:         "Barbed Foliage",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, barbedFoliageAttacksYou, loses, barbedFoliageLosesFlanking),
			On(game.EventAttack, func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
				if !barbedFoliageAttacksYou(ev, source, ch, g) {
					return false
				}
				attacker, ok := g.LookupCardForEffect(ev.CardID)
				return ok && !game.HasKeyword(&attacker, "flying")
			}, damage, barbedFoliageDamage),
		},
	})
}

// barbedFoliageAttacksYou is "a creature attacks you": the attack names
// the enchantment's controller.
func barbedFoliageAttacksYou(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventAttack && ev.CardID != uuid.Nil && ev.Target != uuid.Nil && ev.Target == source.Controller
}

// barbedFoliageLosesFlanking is "it loses flanking until end of turn"
// for the attacker the trigger saw, while it is still that object (CR
// 400.7).
func barbedFoliageLosesFlanking(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if info, ok := ctx.TriggeringPermanent(); !ok || info.Left {
		return nil
	}
	return untilEndOfTurn(ctx, ctx.Trigger().Event.CardID, nil,
		"Barbed Foliage — loses flanking until end of turn", game.RemoveKeywordsMod(game.KeywordFlanking))
}

// barbedFoliageDamage is "this enchantment deals 1 damage to it".
func barbedFoliageDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if info, ok := ctx.TriggeringPermanent(); !ok || info.Left {
		return nil
	}
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	return DealDamage{SourceObject: &ref, Target: ctx.Trigger().Event.CardID, Amount: 1}.Apply(ctx)
}
