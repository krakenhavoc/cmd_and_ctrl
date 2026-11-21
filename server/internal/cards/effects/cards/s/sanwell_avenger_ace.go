package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sanwell, Avenger Ace — Legendary Creature — Human Pilot {1}{W}, 3/1 (The
// Brothers' War Commander, S82):
//
//	"As long as an artifact creature you control is attacking, prevent all
//	 damage that would be dealt to Sanwell.
//	 Whenever Sanwell becomes tapped, exile the top six cards of your
//	 library. You may cast a Vehicle or artifact creature spell from among
//	 them. Then put the rest on the bottom of your library in a random
//	 order."
//
// The first ability is a damage prevention (CR 615) that applies while any
// artifact creature you control is attacking, read as the damage would be
// dealt. The second is Annie Flash's tap trigger (the tap event and an attack
// that taps, ThisBecameTapped) over Boiling Rock Rioter's paid cast during
// the resolution (CR 608.2g): the six are exiled, each Vehicle card and
// artifact creature card among them gets one shared cast (spent by the first
// cast, mana cost paid), and once the offer is answered whatever is still in
// exile goes to the bottom in a random order. Lands and other cards stay in
// the "rest".
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4884ad6c-2290-44fc-ab9e-6e6b75b2d69a",
		Name:         "Sanwell, Avenger Ace",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.PreventAllDamageDealtTo(effects.PreventionStatic{
				To:    effects.ToThisCreature,
				While: sanwellAnArtifactCreatureIsAttacking,
				Label: "Sanwell, Avenger Ace — prevent all damage that would be dealt to it",
			}),
		},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, effects.ThisBecameTapped,
				"Sanwell, Avenger Ace — exile the top six cards of your library; you may cast a Vehicle or artifact creature spell from among them; put the rest on the bottom in a random order",
				sanwellExile),
		},
	})
}

// sanwellAnArtifactCreatureIsAttacking is the prevention's "as long as an
// artifact creature you control is attacking".
func sanwellAnArtifactCreatureIsAttacking(g *game.Game, src *game.Card) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == src.Controller && c.IsCreature() && c.IsArtifact() && c.AttackingTarget != uuid.Nil {
			return true
		}
	}
	return false
}

func sanwellExile(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	return effects.MillToZone{
		Player:   item.Controller,
		N:        6,
		To:       game.ZoneExile,
		ThenTail: ctx.ItemTail(sanwellExiled),
	}.Apply(ctx)
}

// sanwellExiled offers the cast over the Vehicle and artifact creature cards
// among the exiled six, then bottoms what is left.
var sanwellExiled = effects.CardTail("sanwell-avenger-ace/after-exile", func(ctx *effects.Context, o game.TailOutcome) error {
	g, controller, source := ctx.Game, ctx.Controller(), ctx.Source()
	var exiled []uuid.UUID
	var castable []game.Card
	var castableIDs []uuid.UUID
	for _, id := range o.IDs {
		z := g.FindCardZoneForEffect(id)
		c, ok := g.LookupCardForEffect(id)
		if !ok || z == nil || z.Kind != game.ZoneExile {
			continue
		}
		exiled = append(exiled, id)
		if c.HasSubtype("Vehicle") || (c.IsArtifact() && c.IsCreature()) {
			castable = append(castable, c)
			castableIDs = append(castableIDs, id)
		}
	}
	if len(exiled) == 0 {
		return nil
	}
	if len(castable) > 0 {
		g.GrantCastPermissionToCardsForEffect(game.CastPermission{
			Player:      controller,
			Zone:        game.ZoneExile,
			Timing:      game.TimingFlash,
			CastOnly:    true,
			CastsLeft:   1,
			Duration:    g.UntilEndOfTurnDuration(),
			LapseOnPass: game.LapseStaysInExile,
			Source:      source,
			SourceName:  "Sanwell, Avenger Ace",
			Label:       "Sanwell, Avenger Ace — cast a Vehicle or artifact creature spell from among the exiled cards",
		}, castable)
	}
	if g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      controller,
		Source:       source,
		Reason:       "Sanwell, Avenger Ace",
		Cards:        castableIDs,
		Keyword:      game.MayCastKeywordPaidCast,
		AcceptLabel:  "Cast it",
		DeclineLabel: "Put them on the bottom of your library",
		Then:         sanwellRestToBottom,
		Carry:        exiled,
	}) {
		return nil
	}
	return sanwellBottomRest(g, controller, exiled)
})

// sanwellRestToBottom is "Then put the rest on the bottom of your library in
// a random order", once the offer is answered: Carry is the exiled six. An
// on-disk key, never renamed.
var sanwellRestToBottom = game.RegisterChooseCardsThen("resolution-cast/sanwell-avenger-ace-rest-to-bottom",
	func(g *game.Game, r game.CardsChosen) error {
		return sanwellBottomRest(g, r.Chooser, r.Carry)
	})

// sanwellBottomRest puts every one of the cards still in exile on the bottom
// of its owner's library in a random order (the cast one is on the stack).
func sanwellBottomRest(g *game.Game, controller uuid.UUID, ids []uuid.UUID) error {
	var left []uuid.UUID
	for _, id := range ids {
		if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneExile {
			left = append(left, id)
		}
	}
	return g.PutOnBottomInRandomOrderForEffect(controller, game.ZoneExile, left)
}
