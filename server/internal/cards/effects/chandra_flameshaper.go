package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chandra, Flameshaper — Legendary Planeswalker — Chandra {5}{R}{R},
// loyalty 6 (Foundations, #4465):
//
//	"+2: Add {R}{R}{R}. Exile the top three cards of your library. Choose
//	 one. You may play that card this turn.
//	 +1: Create a token that's a copy of target creature you control,
//	 except it has haste and "At the beginning of the end step, sacrifice
//	 this token."
//	 −4: Chandra deals 8 damage divided as you choose among any number of
//	 target creatures and/or planeswalkers."
//
// The +2 is not a mana ability (the ruling): it uses the stack, adds the
// mana (the pool empties at the end of the step, so spend it on that card
// from exile in the same main phase), exiles the top three cards face up
// through the keyed exile, and then asks its controller to choose one of the
// cards that landed. Only that card gets the play permission, "play" so a
// land may be played with the turn's land drop, until end of turn; the
// other two stay exiled with no permission. Cards that did not land in exile
// (a commander that went to the command zone) are not offered.
//
// The +1 is Electroduplicate's copy (TokenCopyWithHasteAndGrant): haste and
// the quoted trigger are gained as part of the copying process (CR 707.9a),
// so they are copiable values of the token, and the token's own trigger
// sacrifices it at the beginning of every end step (so on the turn it was
// made). The target is re-read as it resolves (CR 608.2b).
//
// The −4 divides eight damage among any number of target creatures and/or
// planeswalkers, each chosen target getting at least 1 (CR 601.2d) as the
// activation is announced; a target that became illegal takes nothing and
// its share is not redistributed (the ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "818f908c-9af6-4edc-bb0e-193bc830e2e4",
		Name:            "Chandra, Flameshaper",
		Completeness:    CompletenessFull,
		StartingLoyalty: 6,
		Grants: []AbilityGrant{
			SacrificeThisTokenAtEndStepGrant(chandraFlameshaperGrant, "Chandra, Flameshaper — sacrifice this token"),
		},
		Activated: []ActivatedAbility{
			{
				Label:  "+2: Add {R}{R}{R}. Exile the top three cards of your library. Choose one. You may play that card this turn.",
				Cost:   LoyaltyCost(2),
				Effect: chandraFlameshaperPlus2,
			},
			{
				Label:   "+1: Create a token that's a copy of target creature you control, except it has haste and \"At the beginning of the end step, sacrifice this token.\"",
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1},
				Cost:    LoyaltyCost(1),
				Targets: TargetCreature("target creature you control", YouControl()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return TokenCopyWithHasteAndGrant(chandraFlameshaperGrant)(item, NewContext(g, item))
				},
			},
			{
				Label: "−4: Chandra deals 8 damage divided as you choose among any number of target creatures and/or planeswalkers.",
				Cost:  LoyaltyCost(-4),
				Targets: TargetPermanent("any number of target creatures and/or planeswalkers", Or(Creature(), Planeswalker())).
					WithCount(0, 0).Dividing(Divide(8)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return DealDividedDamage(NewContext(g, item))
				},
			},
		},
	})
}

// chandraFlameshaperGrant is the bundle the +1's token carries: an on-disk
// identity, never renamed, never reused.
const chandraFlameshaperGrant = "chandra-flameshaper/sacrifice-token"

// chandraFlameshaperPlus2 adds {R}{R}{R}, then exiles the top three cards
// and asks which one may be played.
func chandraFlameshaperPlus2(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (AddMana{Produced: "{R}{R}{R}"}).Apply(ctx); err != nil {
		return err
	}
	return g.ExileTopKeyedForEffect(item.Controller, item.Controller, 3, ctx.ItemTail(chandraFlameshaperExiled))
}

// chandraFlameshaperExiled is the exile's keyed rest: "Choose one." among
// the cards that landed in exile. An on-disk key: never renamed, never reused.
var chandraFlameshaperExiled = CardTail("chandra-flameshaper/choose-exiled", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Chandra, Flameshaper — choose one of the exiled cards; you may play it this turn",
		Cards:    o.IDs,
		Min:      1,
		Max:      1,
		LiveZone: game.ZoneExile,
		ThenKey:  chandraFlameshaperChosen,
	})
	return nil
})

// chandraFlameshaperChosen grants the play permission over the chosen card
// alone, locked to the object it is in exile now (CR 611.2c, 400.7). An
// on-disk key: never renamed, never reused.
var chandraFlameshaperChosen = game.RegisterChooseCardsThen("choose-cards/chandra-flameshaper-play", func(g *game.Game, r game.CardsChosen) error {
	if !r.Answered || len(r.Picked) == 0 {
		return nil
	}
	c, ok := g.LookupCardForEffect(r.Picked[0])
	if !ok {
		return nil
	}
	if z := g.FindCardZoneForEffect(c.InstanceID); z == nil || z.Kind != game.ZoneExile {
		return nil
	}
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:     r.Chooser,
		Zone:       game.ZoneExile,
		Source:     r.Source,
		SourceName: "Chandra, Flameshaper",
		Label:      "Play the card Chandra, Flameshaper let you choose",
	}, []game.Card{c})
	return nil
})
