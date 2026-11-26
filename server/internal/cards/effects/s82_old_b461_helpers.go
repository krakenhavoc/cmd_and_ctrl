package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b461_helpers.go — New Phyrexia, batch 461 (S82, #5575).

// ImprintExileTheTarget is the "Imprint — When this enters, exile target …"
// body of Exclusion Ritual and Invader Parasite: the first legal target is
// exiled with the exiled-with link to this object (CR 607.2a), which the
// card's other ability reads. A permanent that has left and come back in
// response is a new object, so nothing is linked to it (CR 400.7).
func ImprintExileTheTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if _, ok := ctx.SourceRef(); !ok {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return ExileTarget{Target: t.ID, Linked: true}.Apply(ctx)
		}
	}
	return nil
}

// PlayersCantCastSpellsNamedLikeTheExiledCard is "Players can't cast spells
// with the same name as the exiled card" (Exclusion Ritual): every player,
// the controller included. Read live off the cards exiled with the
// permanent (Card.ExiledWith), so it lifts the moment the permanent is gone.
func PlayersCantCastSpellsNamedLikeTheExiledCard() game.CastRestriction {
	return game.CastRestriction{
		Label: "Players can't cast spells with the same name as the exiled card.",
		Forbids: func(q game.CastQuery) bool {
			if q.Game == nil {
				return false
			}
			for _, c := range q.Game.ExiledWithForEffect(game.PermissionCardRef{ID: q.Source.InstanceID, Epoch: q.Source.ObjectEpoch}) {
				if c.Name != "" && (game.PermissionFilter{Named: c.Name}).Matches(q.Card) {
					return true
				}
			}
			return false
		},
	}
}

// OpponentLandNamedLikeTheExiledCardEntered is Invader Parasite's trigger
// condition: a land an opponent controls entered with the same name as a
// card exiled with this object (CR 607.2a). The name is compared as the land
// is now, case-folded (CR 201.2).
func OpponentLandNamedLikeTheExiledCardEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	land, ok := B34LandAnOpponentControlsEntered(ev, source, g)
	if !ok || land.Name == "" {
		return false
	}
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef{ID: source.InstanceID, Epoch: source.ObjectEpoch}) {
		if strings.EqualFold(c.Name, land.Name) {
			return true
		}
	}
	return false
}

// TheEnteringLandsController is the player a land-entered trigger's event
// names: the entering card's controller ("that player").
func TheEnteringLandsController(g *game.Game, item *game.StackItem) uuid.UUID {
	if item.Trigger == nil {
		return uuid.Nil
	}
	if c, ok := g.LookupCardForEffect(item.Trigger.Event.CardID); ok {
		return c.Controller
	}
	return uuid.Nil
}

// ChancellorOfTheForgeETB is "create X 1/1 red Phyrexian Goblin creature
// tokens with haste, where X is the number of creatures you control", counted
// as the ability resolves, the Chancellor included.
func ChancellorOfTheForgeETB(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	n := len(PermanentsControlledByMatching(g, ctx.Controller(), Creature()))
	if n <= 0 {
		return nil
	}
	return CreateToken{Template: TokenCard("1/1 red Phyrexian Goblin with haste"), N: n}.Apply(ctx)
}

// LifesFinaleSearch is the second half of Life's Finale: search target
// opponent's library for up to three creature cards, put them into their
// graveyard, then that player shuffles.
func LifesFinaleSearch(ctx *Context, victim uuid.UUID) error {
	return SearchLibrary{
		Player:       ctx.Controller(),
		LibraryOwner: victim,
		Predicate:    game.Card.IsCreature,
		Dest:         game.ZoneGraveyard,
		Limit:        3,
		Shuffle:      true,
		Reason:       "Life's Finale — search target opponent's library for up to three creature cards and put them into their graveyard",
	}.Apply(ctx)
}

// DestroyTheFirstLegalCardTarget is "destroy target …" as an effect body: the
// first target still legal as it resolves is destroyed (CR 608.2b).
func DestroyTheFirstLegalCardTarget(ctx *Context) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return DestroyTarget{Target: t.ID}.Apply(ctx)
		}
	}
	return nil
}
