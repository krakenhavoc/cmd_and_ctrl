package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OldB795DealsOneToTargetAndToYou is "<this> deals 1 damage to any target and 1
// damage to you" (Shadow Guildmage): one source, one simultaneous damage event
// (CR 120.2). A target that is no longer legal makes the whole ability do
// nothing (CR 608.2b), the damage to you included.
func OldB795DealsOneToTargetAndToYou(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	legal := ctx.LegalTargets()
	if len(legal) == 0 {
		return nil
	}
	return g.DamageInstanceForEffect(func() error {
		for _, t := range legal {
			if err := (DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 1}).Apply(ctx); err != nil {
				return err
			}
		}
		return DealDamage{Source: item.SourceCardID, Target: ctx.Controller(), Amount: 1}.Apply(ctx)
	})
}

// OldB795SuperiorNumbers is Superior Numbers: damage to the target creature
// equal to the number of creatures you control in excess of the number the
// target opponent controls, both counted as the spell resolves. It needs both
// targets legal: with the opponent gone there is no number to subtract.
func OldB795SuperiorNumbers(_ *game.StackItem, ctx *Context) error {
	creature, ok1 := ctx.ClauseTarget(0)
	opponent, ok2 := ctx.ClauseTarget(1)
	if !ok1 || !ok2 || !ctx.IsTargetLegal(creature) || !ctx.IsTargetLegal(opponent) {
		return nil
	}
	n := len(CreaturesControlledByPlayer(ctx.Game, ctx.Controller())) -
		len(CreaturesControlledByPlayer(ctx.Game, opponent.ID))
	if n <= 0 {
		return nil
	}
	return DealDamage{Source: ctx.Source(), Target: creature.ID, Amount: n}.Apply(ctx)
}

// OldB795TapAndDealItsPower is Abyssal Hunter's "Tap target creature. This
// creature deals damage equal to its power to that creature": the Hunter's
// power as the ability resolves, or as it last existed if it has left
// (CR 608.2h).
func OldB795TapAndDealItsPower(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	t, ok := FirstLegalTarget(ctx)
	if !ok {
		return nil
	}
	if err := (TapTarget{Target: t.ID}).Apply(ctx); err != nil {
		return err
	}
	info, ok := ctx.SourcePermanent()
	if !ok || info.Power <= 0 {
		return nil
	}
	return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: info.Power}.Apply(ctx)
}

// oldB795DaggerParams is the data Acidic Dagger's two delayed triggers carry:
// the targeted creature as the object it was (Object), and the Dagger
// (Objects[0]).
func oldB795DaggerParams(creature, dagger game.ObjectRef) game.EffectParams {
	return game.EffectParams{Object: creature, Objects: []game.ObjectRef{dagger}}
}

// OldB795AcidicDaggerResolve is Acidic Dagger's ability: until end of turn,
// whenever the targeted creature deals combat damage to a non-Wall creature,
// destroy that creature; and when the targeted creature leaves the battlefield
// this turn, sacrifice the Dagger.
func OldB795AcidicDaggerResolve(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	t, ok := FirstLegalTarget(ctx)
	if !ok {
		return nil
	}
	creature := game.ObjectRef{ID: t.ID, Epoch: g.ObjectEpochForEffect(t.ID)}
	dagger, _ := ctx.SourceRef()
	params := oldB795DaggerParams(creature, dagger)
	if err := (DelayedOnEvent{
		Label:      "Acidic Dagger — destroy the non-Wall creature it dealt combat damage to",
		On:         []game.EventKind{game.EventDealDamage},
		Condition:  oldB795DealtCombatDamageToNonWall,
		CondParams: params,
		Body:       oldB795DestroyDamaged,
		Params:     params,
		Repeats:    true,
	}).Apply(ctx); err != nil {
		return err
	}
	return DelayedOnEvent{
		Label:      "Acidic Dagger — sacrifice this artifact",
		On:         []game.EventKind{game.EventLTB},
		Condition:  oldB795TargetedCreatureLeft,
		CondParams: params,
		Body:       oldB795SacrificeDagger,
		Params:     params,
		Repeats:    true,
	}.Apply(ctx)
}

// oldB795DealtCombatDamageToNonWall: the targeted creature (the same object,
// CR 400.7) dealt combat damage to a creature that is not a Wall.
var oldB795DealtCombatDamageToNonWall = game.DelayedCondition("damage/acidic-dagger-non-wall",
	func(ev game.Event, _ *game.DelayedTrigger, g *game.Game, p game.EffectParams) bool {
		if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 || ev.Source != p.Object.ID ||
			g.ObjectEpochForEffect(p.Object.ID) != p.Object.Epoch {
			return false
		}
		victim, ok := g.LookupCardForEffect(ev.Target)
		return ok && g.Battlefield.Contains(ev.Target) && victim.IsCreature() && !HasSubtype("Wall")(g, uuid.Nil, victim)
	})

// oldB795DestroyDamaged destroys the creature the event damaged.
var oldB795DestroyDamaged = game.DelayedBody("old-b795/acidic-dagger-destroy",
	func(g *game.Game, item *game.StackItem, _ game.EffectParams) error {
		ctx := NewContext(g, item)
		victim := ctx.Trigger().Event.Target
		if victim == uuid.Nil || !OnBattlefield(g, victim) {
			return nil
		}
		return DestroyTarget{Target: victim}.Apply(ctx)
	})

// oldB795TargetedCreatureLeft: the targeted creature left the battlefield.
var oldB795TargetedCreatureLeft = game.DelayedCondition("ltb/acidic-dagger-creature-left",
	func(ev game.Event, _ *game.DelayedTrigger, _ *game.Game, p game.EffectParams) bool {
		return ev.Kind == game.EventLTB && ev.CardID == p.Object.ID
	})

// oldB795SacrificeDagger sacrifices the Dagger if it is still the same object.
var oldB795SacrificeDagger = game.DelayedBody("old-b795/acidic-dagger-sacrifice",
	func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		if len(p.Objects) == 0 {
			return nil
		}
		d := p.Objects[0]
		if d.ID == uuid.Nil || !OnBattlefield(g, d.ID) || g.ObjectEpochForEffect(d.ID) != d.Epoch {
			return nil
		}
		return SacrificePermanent{Target: d.ID}.Apply(NewContext(g, item))
	})

// OldB795BeforeBlockersAreDeclared is "Activate only before blockers are
// declared" (Acidic Dagger): any step of the turn before the declare blockers
// step, which is the beginning of the turn through the declare attackers step.
func OldB795BeforeBlockersAreDeclared(g *game.Game) bool {
	for _, s := range game.TurnSequence() {
		if s == game.StepDeclareBlockers {
			return false
		}
		if s == g.Turn.Step {
			return true
		}
	}
	return false
}

// OldB795ChosenColorCreatureHurtYouOrAWhiteCreature is Mangara's Equity's
// trigger condition: a creature of the chosen colour (as the damage was dealt)
// dealt damage to the source's controller or to a white creature they control.
func OldB795ChosenColorCreatureHurtYouOrAWhiteCreature(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || ev.Amount <= 0 {
		return false
	}
	chosen := ChosenColorOf(g, source.InstanceID)
	if chosen == "" {
		return false
	}
	dealer, ok := g.LookupCardForEffect(ev.Source)
	if !ok || !g.Battlefield.Contains(ev.Source) || !dealer.IsCreature() || !OfColor(chosen)(g, source.Controller, dealer) {
		return false
	}
	if ev.Target == source.Controller {
		return true
	}
	victim, ok := g.LookupCardForEffect(ev.Target)
	return ok && g.Battlefield.Contains(ev.Target) && victim.IsCreature() && victim.Controller == source.Controller &&
		OfColor("W")(g, source.Controller, victim)
}

// OldB795PreventOneFromTheGrantingSpell is the ability Torrent of Lava grants
// to each creature: "Prevent the next 1 damage that would be dealt to this
// creature by Torrent of Lava this turn". The spell is the grantor
// (StackItem.GrantedBy); the creature is the ability's source.
func OldB795PreventOneFromTheGrantingSpell(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	spell := ctx.Grantor()
	if spell == uuid.Nil {
		return nil
	}
	return PreventDamageFromSource{
		From:    spell,
		Protect: ShieldThis,
		Amount:  1,
		Label:   "Torrent of Lava — prevent the next 1 damage to this creature",
	}.Apply(ctx)
}

// OldB795TaintedSpecter is Tainted Specter's ability: the target player is
// offered "put a card from your hand on top of your library"; if they do not,
// they discard a card, and if they do discard one the Specter deals 1 damage to
// each creature and each player.
func OldB795TaintedSpecter(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	who, ok := FirstLegalPlayerTarget(ctx)
	if !ok {
		return nil
	}
	return AnyPlayerPays{
		Pay:          HandToLibraryTopPayment(1),
		Only:         who,
		Question:     "Tainted Specter — put a card from your hand on top of your library, or discard a card?",
		AcceptLabel:  "Put a card on top",
		DeclineLabel: "Discard a card",
		Then:         oldB795SpecterDeclined,
		Carry:        []uuid.UUID{who},
	}.Apply(ctx)
}

var oldB795SpecterDeclined = AnyPlayerConsequences("any-player/old-b795-tainted-specter-discard", nil,
	func(ctx *Context, r game.AnyPlayerAnswer) error {
		if len(r.Carry) == 0 {
			return nil
		}
		return ctx.Game.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
			Player:   r.Carry[0],
			Source:   ctx.Source(),
			N:        1,
			Question: "Tainted Specter — discard a card",
			Asked:    game.NewAskedItem(ctx.Item),
		}, oldB795SpecterDiscarded, nil)
	}, nil)

var oldB795SpecterDiscarded = game.RegisterDiscardThen("choose-cards/old-b795-tainted-specter-damage", func(g *game.Game, r game.DiscardsDone) error {
	if !r.Discarded.Discarded(r.Player) {
		return nil
	}
	item := r.Item
	if item == nil {
		item = KeyedContext(g, uuid.Nil, r.Source).Item
	}
	return ThisDealsDamageToEachCreatureMatchingAndEachPlayer(Creature(), 1)(g, item)
})
