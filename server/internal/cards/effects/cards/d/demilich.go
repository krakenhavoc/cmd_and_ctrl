package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Demilich — Creature — Skeleton Wizard {U}{U}{U}{U}, 4/3 (Adventures in the
// Forgotten Realms):
//
//	"This spell costs {U} less to cast for each instant and sorcery spell
//	 you've cast this turn.
//	 Whenever this creature attacks, exile up to one target instant or sorcery
//	 card from your graveyard. Copy it. You may cast the copy.
//	 You may cast this card from your graveyard by exiling four instant and/or
//	 sorcery cards from your graveyard in addition to paying its other costs."
//
// The reduction is a self cost modifier (Cavern-Hoard Dragon's) whose unit is
// {U} (Eluge's: CR 118.7b and 118.7c, so it eats blue first, then generic
// mana), counted off the per-turn tally of instant and sorcery spells the
// caster has cast (CastTally.InstantSorcery, bumped after pricing, so the
// spell being priced is not counted). The attack trigger is Blue Mage's Cane's
// body over your own graveyard: the card that really reached exile is copied
// (CR 707.12) and the copy is offered for its printed mana cost while the
// trigger resolves (CR 608.2g); a copy not cast ceases to exist and the
// original stays exiled. The graveyard cast is Retrace's model with an
// escape-style exile of four instant and/or sorcery cards as the added cost
// (CR 601.2h), the spell never being among them; the printed mana cost is
// still paid and the cost reduction applies to it.
//
// No simplification.
func init() {
	reduction := effects.CostsLessEach(demilichInstantSorcerySpellsCast,
		"This spell costs {U} less to cast for each instant and sorcery spell you've cast this turn.")
	unit, err := game.ParseCost("{U}")
	if err != nil {
		panic(err)
	}
	reduction.Unit = &unit
	effects.Register(effects.Spec{
		OracleID:          "eeb50bfc-8443-482b-bfc1-ea96b37809cd",
		Name:              "Demilich",
		Completeness:      effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{reduction},
		CastableZones:     []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{
			effects.S82AfrB16CastFromGraveyardByExiling("{U}{U}{U}{U}", 4, "instant and/or sorcery", effects.Or(effects.Instant(), effects.Sorcery())),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisAttacks("Demilich — exile up to one target instant or sorcery card from your graveyard, copy it, and you may cast the copy",
				demilichExileAndCopy),
				effects.TargetCardInGraveyard("up to one target instant or sorcery card from your graveyard",
					effects.YouOwn(), effects.Or(effects.Instant(), effects.Sorcery())).WithCount(0, 1)),
		},
	})
}

// demilichInstantSorcerySpellsCast is the number of instant and sorcery
// spells the caster has cast this turn.
func demilichInstantSorcerySpellsCast(q game.CostQuery) int {
	if q.Game == nil {
		return 0
	}
	return q.Game.SpellsCastThisTurn[q.Controller].InstantSorcery
}

// demilichExileAndCopy exiles the chosen card and offers a copy of what
// reached exile for its mana cost.
func demilichExileAndCopy(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			ids = append(ids, t.ID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return g.ExileCardsKeyedForEffect(ids, ctx.ItemTail(demilichCopy))
}

// demilichCopy copies the card that reached exile. An on-disk key: never
// renamed, never reused.
var demilichCopy = effects.CardTail("demilich/copy-exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	return effects.CopyAndCast{Cards: o.IDs}.Apply(effects.NewContext(ctx.Game, ctx.Item))
})
