package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// CrownOfTheAgesOfferHostsB831 asks the controller which creature other than
// the Aura's current host to attach it to, among the ones it can enchant, and
// attaches it (Crown of the Ages; S82 OLD batch 831, #5636). Caller holds g.mu.
func CrownOfTheAgesOfferHostsB831(g *game.Game, item *game.StackItem, auraID uuid.UUID) error {
	aura, ok := g.LookupCardForEffect(auraID)
	if !ok || aura.AttachedTo.Kind != game.TargetCard {
		return nil
	}
	current := aura.AttachedTo.ID
	var hosts []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID != current && c.IsCreature() && g.AttachLegalForEffect(auraID, c.InstanceID) {
			hosts = append(hosts, c.InstanceID)
		}
	}
	if len(hosts) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Crown of the Ages — attach " + aura.Name + " to another creature",
		Cards:    hosts,
		Min:      1,
		Max:      1,
		ThenKey:  crownOfTheAgesAttachPick,
		LiveZone: game.ZoneBattlefield,
		Carry:    []uuid.UUID{auraID},
	})
	return nil
}

// crownOfTheAgesAttachPick attaches the carried Aura to the creature picked.
var crownOfTheAgesAttachPick = game.RegisterChooseCardsThen("choose-cards/old-b831-crown-of-the-ages-attach", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 || len(r.Carry) == 0 {
		return nil
	}
	return g.AttachForEffect(r.Carry[0], game.TargetRef{Kind: game.TargetCard, ID: r.Picked[0]})
})

// CrownOfTheAgesMove is the ability's effect: the first legal Aura target.
func CrownOfTheAgesMove(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return CrownOfTheAgesOfferHostsB831(g, item, t.ID)
		}
	}
	return nil
}

// ElkinBottleImpulseB831 is "Exile the top card of your library. Until the
// beginning of your next upkeep, you may play that card."
func ElkinBottleImpulseB831(ctx *Context) error {
	window := DurationUntilYourNextUpkeep(ctx, ctx.Controller())
	return ExileTopWithPermission{
		From:    ctx.Controller(),
		GrantTo: ctx.Controller(),
		N:       1,
		Window:  &window,
	}.Apply(ctx)
}

// MeriekeStealB831 is Merieke Ri Berit's tap ability: gain control of the
// target creature for as long as you control Merieke, and remember it as the
// permanent her linked trigger destroys (CR 607.2e).
func MeriekeStealB831(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target := FirstLegalBattlefieldTarget(ctx)
	if target == uuid.Nil || !SourceIsStillThisPermanent(g, item) {
		return nil
	}
	if err := (GainControl{
		Target: target, Controller: ctx.Controller(), Duration: ForAsLongAsYouControlThisWindow(g, item),
		Label: "Merieke Ri Berit — control while you control her",
	}).Apply(ctx); err != nil {
		return err
	}
	if c, ok := g.LookupCardForEffect(target); ok {
		g.SetLinkedPermanentForEffect(item.SourceCardID, game.ObjectRef{ID: c.InstanceID, Epoch: c.ObjectEpoch})
	}
	return nil
}

// MeriekeLeftOrUntappedB831 is the When of "When Merieke Ri Berit leaves the
// battlefield or becomes untapped".
func MeriekeLeftOrUntappedB831(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return (ev.Kind == game.EventLTB || ev.Kind == game.EventUntapCard) && ev.CardID == source.InstanceID
}

// MeriekeDestroyB831 destroys the creature Merieke took, while it is still that
// object; it can't be regenerated.
func MeriekeDestroyB831(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	stolen := g.LinkedPermanentForEffect(ref)
	if stolen == nil {
		return nil
	}
	return DestroyTarget{Target: stolen.InstanceID, CantBeRegenerated: true}.Apply(ctx)
}

// SpiderExileB831 is "Exile this creature and target creature": both go to
// exile as the ability resolves.
func SpiderExileB831(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if id, ok := B16FirstLegalTargetCard(ctx); ok {
		if err := (ExileTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return ExileTarget{Target: item.SourceCardID}.Apply(ctx)
}

// BounceThisAndTargetB831 is "Return this creature and target creature you control
// to their owner's hand" (Snow Hound, Wizard Mentor).
func BounceThisAndTargetB831(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if id, ok := B16FirstLegalTargetCard(ctx); ok && id != item.SourceCardID {
		if err := (BounceToHand{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return BounceToHand{Target: item.SourceCardID}.Apply(ctx)
}

// DuringCombatB831 is "Activate only during combat": any step of the combat
// phase (CR 506.1).
func DuringCombatB831() ActivationCondition {
	return func(g *game.Game, _, _ uuid.UUID) bool {
		switch g.Turn.Step {
		case game.StepBeginCombat, game.StepDeclareAttackers, game.StepDeclareBlockers,
			game.StepCombatDamage, game.StepEndCombat:
			return true
		}
		return false
	}
}

// ADefendingPlayerControlsASnowLandB831 is "only if defending player controls a
// snow land" for a source that is not itself attacking: some player attacked
// this combat (or the controller of a planeswalker or battle attacked) controls
// a snow land.
func ADefendingPlayerControlsASnowLandB831() ActivationCondition {
	return func(g *game.Game, _, _ uuid.UUID) bool {
		defenders := map[uuid.UUID]bool{}
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.AttackingTarget == uuid.Nil {
				continue
			}
			if g.PlayerByIDForEffect(c.AttackingTarget) != nil {
				defenders[c.AttackingTarget] = true
			} else if t, ok := g.LookupCardForEffect(c.AttackingTarget); ok {
				defenders[t.Controller] = true
			}
		}
		for _, c := range g.BattlefieldCardsForEffect() {
			if defenders[c.Controller] && c.IsLand() && c.HasSupertype("Snow") {
				return true
			}
		}
		return false
	}
}

// AugurySetupB831 is "Look at the top three cards of target player's library,
// then put them back in any order."
func AugurySetupB831(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		return LookAtLibraryThenPlace{
			Owner:     t.ID,
			N:         3,
			Placement: game.LibraryPlaceTop,
			Label:     "Elemental Augury — put them back in any order",
		}.Apply(ctx)
	}
	return nil
}

// SnowLandAnimateB831 is "Target snow land becomes a 2/2 creature until end of
// turn. It's still a land."
func SnowLandAnimateB831(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	land := FirstLegalBattlefieldTarget(ctx)
	if land == uuid.Nil {
		return nil
	}
	return BecomeCreatureUntilEOT{Target: land, Types: []string{"Creature"}, SetPower: 2, SetToughness: 2,
		Label: "Balduvian Conjurer — target snow land becomes a 2/2 creature until end of turn"}.Apply(ctx)
}

// VigilanceToTargetB831 is "Target creature gains vigilance until end of turn".
func VigilanceToTargetB831(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return GrantKeywordUntilEOT{Target: id, Keywords: []string{"vigilance"}, Label: "Arcum's Sleigh — vigilance until end of turn"}.Apply(ctx)
}

// PreventAllCombatDamageB831 is "Prevent all combat damage that would be dealt
// this turn."
func PreventAllCombatDamageB831(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return PreventAllCombatDamageThisTurn{Label: name + " — prevent all combat damage this turn"}.Apply(NewContext(g, item))
	}
}

// ElkinBottleActivation is Elkin Bottle's ability effect.
func ElkinBottleActivation(g *game.Game, item *game.StackItem) error {
	return ElkinBottleImpulseB831(NewContext(g, item))
}
