package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Urborg Scavengers — Creature — Spirit {2}{B}, 2/2 (March of the Machine:
// The Aftermath, #5476):
//
//	"Whenever this creature enters or attacks, exile target card from a
//	 graveyard. Put a +1/+1 counter on this creature.
//	 This creature has flying as long as a card exiled with it has flying.
//	 The same is true for first strike, double strike, deathtouch, haste,
//	 hexproof, indestructible, lifelink, menace, reach, trample, and
//	 vigilance."
//
// One trigger with two conditions, one target (a card in any graveyard,
// required). The exile is LINKED to this object (CR 607.2a,
// ExileLinkedKeyedForEffect), so the card carries Card.ExiledWith, and the
// counter is the exile's continuation: put once the exile has finished,
// whether or not the card really reached exile (the counter is not
// conditional). A target that is illegal as the ability resolves does
// nothing at all (CR 608.2b). An object that left and came back is a new
// object that owns none of the cards (CR 400.7).
//
// The static is Soulflayer's layer-6 self-grant read from the linked
// cards: each is asked as it sits in exile, with its own printed
// characteristics (a keyword the card grants itself conditionally gives
// nothing). It is re-read when a linked card leaves exile
// (DependsOnExile). Any card counts, not only creature cards. A hexproof
// variant ("hexproof from white") is given as that variant.
//
// No simplification.
var urborgScavengersCounter = CardTail("urborg-scavengers/counter", func(ctx *Context, _ game.TailOutcome) error {
	return AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
})

func init() {
	Register(Spec{
		OracleID:     "d27f3338-4900-49eb-9d8b-994afd524abe",
		Name:         "Urborg Scavengers",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{urborgScavengersKeywords()},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEntersOrAttacks("Urborg Scavengers — exile target card from a graveyard, then put a +1/+1 counter on this creature",
					urborgScavengersExile),
				TargetCardInGraveyard("target card from a graveyard")),
		},
	})
}

// urborgScavengersExile exiles the still-legal target, linked to this
// object, then puts the counter.
func urborgScavengersExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			ids = append(ids, t.ID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	link, _ := ctx.SourceRef()
	return g.ExileLinkedKeyedForEffect(link, ids, ctx.ItemTail(urborgScavengersCounter))
}

// urborgScavengersKeywordList are the twelve keywords the static names.
var urborgScavengersKeywordList = []string{
	"flying", "first strike", "double strike", "deathtouch", "haste", "hexproof",
	"indestructible", "lifelink", "menace", "reach", "trample", "vigilance",
}

func urborgScavengersKeywords() game.StaticAbility {
	return game.StaticAbility{
		Layer:          game.Layer6Ability,
		AppliesTo:      selfOnly,
		DependsOnExile: true,
		Label:          "Urborg Scavengers — has each listed keyword a card exiled with it has",
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			linked := g.ExiledWithForEffect(game.PermissionCardRef{ID: source.InstanceID, Epoch: source.ObjectEpoch})
			for _, kw := range urborgScavengersKeywordList {
				for _, l := range linked {
					if game.HasKeyword(l, kw) {
						c.Abilities = game.AppendKeywordAbility(c.Abilities, kw)
						break
					}
				}
			}
			for _, l := range linked {
				for _, tok := range game.HexproofFromTokens(l) {
					c.Abilities = game.AppendKeywordAbility(c.Abilities, tok)
				}
			}
		},
	}
}
