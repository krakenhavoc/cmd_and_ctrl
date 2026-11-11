package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_s37_helpers.go — batch 37 (Manifest and face-up payoffs).

const cmdr2S37WhisperwoodDiesManifest = "whisperwood-elemental/dies-manifest"

// cmdr2S37WhisperwoodGrant is Whisperwood Elemental's sacrifice ability:
// every face-up nontoken creature its controller controls, as it resolves,
// gains the dies-manifest trigger until end of turn.
func cmdr2S37WhisperwoodGrant(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != item.Controller || !c.IsCreature() || c.IsToken() || c.FaceDown {
			continue
		}
		if err := (GrantAbilitiesFor{
			Target: c.InstanceID,
			Keys:   []string{cmdr2S37WhisperwoodDiesManifest},
			Label:  "Whisperwood Elemental — it manifests the top card of your library when it dies",
		}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}

// cmdr2S37TurnedUpCreatureFights is Temur War Shaman's "you may have it
// fight target creature you don't control": the permanent that was turned
// face up (the trigger's event card) fights the announced target, when it
// is still a creature on the battlefield (CR 603.4's recheck) and the
// target is still legal (CR 701.14b).
func cmdr2S37TurnedUpCreatureFights(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ctx := NewContext(g, item)
	it, ok := battlefieldCardForEffect(g, item.Trigger.Event.CardID)
	if !ok || !it.IsCreature() {
		return nil
	}
	target, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return b10Fight(ctx, it.InstanceID, target)
}

// cmdr2S37GainLifePerCreature is Mastery of the Unseen's "you gain 1 life
// for each creature you control", counted as the ability resolves.
func cmdr2S37GainLifePerCreature(g *game.Game, item *game.StackItem) error {
	return GainLife{Amount: b14CreaturesControlled(g, item.Controller)}.Apply(NewContext(g, item))
}

// cmdr2S37ColorlessCreatureSpellCastByYou is Ugin's Mastery's "whenever you
// cast a colorless creature spell": your cast, of a spell that is a
// colorless creature as it sits on the stack.
func cmdr2S37ColorlessCreatureSpellCastByYou(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	return ok && spell.IsCreature() && spell.IsColorless()
}

// cmdr2S37AttackedWithTotalPower is "whenever you attack with creatures with
// total power N or greater": the creatures the controller declared as
// attackers in this declaration (the event's batch), each at its current
// power (CR 208.3).
func cmdr2S37AttackedWithTotalPower(n int) func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return attackDeclaredByYou(ev, source.Controller) &&
			packTacticsTotalPower(g, source.Controller, ev.Batch) >= n
	}
}

// cmdr2S37FaceDownCreaturesYouControl lists the face-down creatures `player`
// controls that an effect could turn face up.
func cmdr2S37FaceDownCreaturesYouControl(g *game.Game, player uuid.UUID) []uuid.UUID {
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.FaceDown && c.IsCreature() && game.CanTurnFaceUpForEffect(c) {
			ids = append(ids, c.InstanceID)
		}
	}
	return ids
}

// cmdr2S37OfferTurnFaceUp is Ugin's Mastery's "you may turn a face-down
// creature you control face up": the controller picks none or one of their
// face-down creatures as the ability resolves (nothing is targeted).
func cmdr2S37OfferTurnFaceUp(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ids := cmdr2S37FaceDownCreaturesYouControl(g, ctx.Controller())
	if len(ids) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Ugin's Mastery — you may turn a face-down creature you control face up",
		Cards:    ids,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneBattlefield,
		ThenKey:  cmdr2S37TurnFaceUpPick,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
}

// cmdr2S37TurnFaceUpPick turns the picked creature face up. An on-disk key.
var cmdr2S37TurnFaceUpPick = askedPick("choose-cards/cmdr2-s37-ugins-mastery-turn-face-up", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return TurnFaceUp{Targets: r.Picked[:1]}.Apply(ctx)
})

// cmdr2S37ExileFaceDownPermanentYouControl is "Exile a face-down permanent
// you control face up" as an activation cost (Primordial Mist): any
// face-down permanent of the activator's, not only a creature.
func cmdr2S37ExileFaceDownPermanentYouControl() game.AbilityCost {
	return game.AbilityCost{ExilePermanents: &game.ExilePermanentsCost{
		Count:    1,
		FaceDown: true,
		Label:    "a face-down permanent you control",
	}}
}

// cmdr2S37PlayExiledThisTurn is Primordial Mist's "You may play that card
// this turn": the card the cost exiled is, face up in exile, playable by the
// controller until end of turn (timing rules and costs still apply).
func cmdr2S37PlayExiledThisTurn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, id := range ctx.Exiled() {
		if err := g.ExileCardWithPermissionForEffect(id, game.CastPermission{
			Player: ctx.Controller(),
			Label:  "Primordial Mist — play it this turn",
		}); err != nil {
			return err
		}
	}
	return nil
}

// cmdr2S37FaceDownCreatureEntered is Trail of Mystery's "whenever a
// face-down creature you control enters".
func cmdr2S37FaceDownCreatureEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	return ok && c.FaceDown && c.IsCreature()
}

// cmdr2S37PlusTwoIfCreature is Trail of Mystery's turned-up rider: the
// permanent that was turned face up, if it is a creature, gets +2/+2 until
// end of turn (CR 603.4's recheck as it resolves).
func cmdr2S37PlusTwoIfCreature(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	c, ok := battlefieldCardForEffect(g, item.Trigger.Event.CardID)
	if !ok || !c.IsCreature() {
		return nil
	}
	return BoostUntilEOT{Target: c.InstanceID, Power: 2, Toughness: 2, Label: "Trail of Mystery — +2/+2"}.Apply(NewContext(g, item))
}

// cmdr2S37PanopticDoubler is Panoptic Projektor's "If turning a face-down
// permanent face up causes a triggered ability of a permanent you control to
// trigger, that ability triggers an additional time" (CR 603.2d).
func cmdr2S37PanopticDoubler() game.TriggerDoubler {
	return game.TriggerDoubler{
		Label:    "Panoptic Projektor",
		RowLabel: "Abilities that trigger when a face-down permanent is turned face up trigger an additional time.",
		Applies: func(_ *game.Game, q game.TriggerDoublingQuery) bool {
			return !q.FromSpell && q.SourceIsPermanent && q.Ability != nil &&
				q.Event.Kind == game.EventTurnedFaceUp &&
				q.SourceLKI.Controller == q.DoublerLKI.Controller
		},
	}
}
