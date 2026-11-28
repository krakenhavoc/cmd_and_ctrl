package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b771_helpers.go — shared shapes of S82 OLD batch 771 (Weatherlight:
// damage, #5629): Sawtooth Ogre's delayed damage, Spinning Darkness's alternative
// cost, Urborg Stalker's condition and the small damage triggers.

// S82OldB771DamageListedBody is "<source> deals 1 damage to that creature" at
// the beginning of the end of combat step (Sawtooth Ogre): the delayed trigger's
// source deals the damage to each listed card still on the battlefield. A source
// that has left deals it as it last existed (CR 608.2h).
var S82OldB771DamageListedBody = game.SimpleDelayedBody("s82-old-b771/damage-listed-cards", s82OldB771DamageListed)

func s82OldB771DamageListed(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range item.Targets {
		if t.Kind != game.TargetCard || t.ID == uuid.Nil || !OnBattlefield(g, t.ID) {
			continue
		}
		if err := (DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// S82OldB771SawtoothTrigger is "Whenever this creature blocks or becomes blocked
// by a creature, this creature deals 1 damage to that creature at end of combat":
// Corrosive Ooze's per-pair block trigger (CR 509.3b, 509.3d) scheduling a CR
// 603.7 delayed trigger at the beginning of the end of combat step.
func S82OldB771SawtoothTrigger() game.TriggeredAbility {
	const label = "Sawtooth Ogre — it deals 1 damage to that creature at end of combat"
	return On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return SkewerSlingerFoe(ev, source.InstanceID) != uuid.Nil
	}, label, func(g *game.Game, item *game.StackItem) error {
		if item.Trigger == nil {
			return nil
		}
		foe := SkewerSlingerFoe(item.Trigger.Event, item.SourceCardID)
		if foe == uuid.Nil {
			return nil
		}
		return ScheduleDelayedTrigger{
			At:    game.StepEndCombat,
			Label: label,
			Cards: []uuid.UUID{foe},
			Body:  S82OldB771DamageListedBody,
		}.Apply(NewContext(g, item))
	})
}

// S82OldB771ExileTopThreeBlack is Spinning Darkness's "You may exile the top
// three black cards of your graveyard rather than pay this spell's mana cost"
// (CR 118.9): a free alternative cost whose payment is the three topmost black
// cards of the caster's graveyard (CR 404.2), not the caster's to pick.
func S82OldB771ExileTopThreeBlack() game.AlternativeCost {
	const clause = "Exile the top three black cards of your graveyard"
	return game.AlternativeCost{
		Key:                "exile-top-three-black",
		Label:              clause,
		ExileFromGraveyard: CardsInYourGraveyard(3, clause, OfColor("B")),
		GraveyardTopOnly:   true,
		PayLabel:           "the top three black cards of your graveyard",
	}
}

// S82OldB771DamageTheCreatureThatDied is "this artifact deals 2 damage to that
// creature's controller" (Dingus Staff): the controller the creature had as it
// last existed on the battlefield (CR 603.10a).
func S82OldB771DamageTheDiedCreaturesController(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ev := item.Trigger.Event
	player := ev.LastKnownController
	if player == uuid.Nil {
		if info, ok := g.LastKnownPermanentForEffect(ev.CardID); ok {
			player = info.Controller
		}
	}
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if player == uuid.Nil || !ok {
		return nil
	}
	return DealDamage{SourceObject: &ref, Target: player, Amount: 2}.Apply(ctx)
}

// S82OldB771DamageTheEnteredCreature is "this enchantment deals 2 damage to it"
// (Aether Flash) for the creature whose entering set the trigger off, if it is
// still on the battlefield as the ability resolves.
func S82OldB771DamageTheEnteredCreature(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	id := item.Trigger.Event.CardID
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if id == uuid.Nil || !ok || !OnBattlefield(g, id) {
		return nil
	}
	return DealDamage{SourceObject: &ref, Target: id, Amount: 2}.Apply(ctx)
}

// S82OldB771ControlsANonblackNonlandPermanent is Urborg Stalker's intervening
// "if that player controls a nonblack, nonland permanent": a colorless permanent
// is nonblack.
func S82OldB771ControlsANonblackNonlandPermanent(g *game.Game, player uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && !c.IsLand() && !c.HasColorWord("B") {
			return true
		}
	}
	return false
}
