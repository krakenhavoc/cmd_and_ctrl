package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tezzeret, Betrayer of Flesh — Legendary Planeswalker — Tezzeret
// {2}{U}{U}, starting loyalty 4:
//
//	"The first activated ability of an artifact you activate each turn
//	 costs {2} less to activate.
//	 +1: Draw two cards. Then discard two cards unless you discard an
//	 artifact card.
//	 −2: Target artifact becomes an artifact creature. If it isn't a
//	 Vehicle, it has base power and toughness 4/4.
//	 −6: You get an emblem with "Whenever an artifact you control becomes
//	 tapped, draw a card.""
//
// The discount is an activation cost modifier on the first activation of
// any artifact's ability by the controller each turn (the record is
// Game.ArtifactActivationsThisTurn, mana abilities and an artifact that
// is sacrificed to pay for it included). The +1 is Teferi, Akosa of
// Zhalfir's shape with "an artifact card". The −2 states no duration, so
// the artifact stays a creature for good (CR 611.2a); a Vehicle keeps its
// printed size, anything else is 4/4.
func init() {
	Register(Spec{
		OracleID:        "5dedd5a9-bc5d-47bb-9ae1-7d683843d041",
		Name:            "Tezzeret, Betrayer of Flesh",
		Completeness:    CompletenessFull,
		StartingLoyalty: 4,
		CostModifiers: []game.CostModifier{
			ActivationCostsLess(2,
				"The first activated ability of an artifact you activate each turn costs {2} less to activate.",
				cmdr2S51ArtifactAbilityYouActivate(), cmdr2S51FirstArtifactActivationThisTurn()),
		},
		Emblem: &EmblemSpec{
			Label: "Tezzeret, Betrayer of Flesh emblem",
			Text:  "Whenever an artifact you control becomes tapped, draw a card.",
			Triggered: []game.TriggeredAbility{
				OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, cmdr2S51ArtifactBecameTapped,
					"Tezzeret, Betrayer of Flesh emblem — draw a card", Do(DrawCards{N: 1})),
			},
		},
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Draw two cards. Then discard two cards unless you discard an artifact card.",
				Cost:    LoyaltyCost(1),
				Purpose: game.Purpose{Draws: 2},
				Effect:  tezzeretBetrayerDrawThenDiscard,
			},
			{
				Label:   "−2: Target artifact becomes an artifact creature. If it isn't a Vehicle, it has base power and toughness 4/4.",
				Cost:    LoyaltyCost(-2),
				Targets: TargetPermanent("target artifact", Artifact()),
				Effect:  tezzeretBetrayerAnimate,
			},
			{
				Label:  "−6: You get an emblem with \"Whenever an artifact you control becomes tapped, draw a card.\"",
				Cost:   LoyaltyCost(-6),
				Effect: Do(CreateEmblem{}),
			},
		},
	})
}

func tezzeretBetrayerDrawThenDiscard(g *game.Game, item *game.StackItem) error {
	return g.DrawNThenForEffect(item.Controller, 2, game.DrawThen{
		Ref: tezzeretBetrayerDiscardThen, Player: item.Controller, Source: item.SourceCardID,
	})
}

// tezzeretBetrayerDiscardThen opens the conditional discard once both
// draws have resolved. Keyed: an on-disk identity.
var tezzeretBetrayerDiscardThen = game.RegisterDrawThen("tezzeret-betrayer-of-flesh-discard", func(g *game.Game, d game.DrawThen) error {
	p := g.PlayerByIDForEffect(d.Player)
	if p == nil || p.Hand == nil {
		return nil
	}
	// The full count is two, or the whole hand when it is smaller (CR
	// 701.9a). A hand of one card is asked the one discard, whatever it
	// is; a registered set rule is a function of the picks alone and
	// cannot read the count, so the two cases are two keys.
	need := min(2, p.Hand.Size())
	if need <= 0 {
		return nil
	}
	then := tezzeretBetrayerDiscardTwoOrArtifact
	if need == 1 {
		then = tezzeretBetrayerDiscardTheOne
	}
	return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
		Player:   d.Player,
		Source:   d.Source,
		N:        need,
		Min:      1,
		Question: "Tezzeret, Betrayer of Flesh — discard two cards, unless you discard an artifact card",
	}, then, nil)
})

var tezzeretBetrayerDiscardTwoOrArtifact = game.RegisterDiscardThenRule("choose-cards/tezzeret-betrayer-of-flesh-discard-two-or-artifact",
	func(picked []game.Card) bool {
		return len(picked) == 2 || (len(picked) == 1 && picked[0].IsArtifact())
	},
	func(*game.Game, game.DiscardsDone) error { return nil })

var tezzeretBetrayerDiscardTheOne = game.RegisterDiscardThen("choose-cards/tezzeret-betrayer-of-flesh-discard-the-one",
	func(*game.Game, game.DiscardsDone) error { return nil })

// tezzeretBetrayerAnimate is the −2. Whether it is a Vehicle is read as
// the ability resolves.
func tezzeretBetrayerAnimate(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var id uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			id = t.ID
			break
		}
	}
	if id == uuid.Nil {
		return nil
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	mods := []game.Mod{game.AddTypesMod("Creature")}
	if !c.HasSubtype("Vehicle") {
		mods = append(mods, game.SetBasePowerMod(4), game.SetBaseToughnessMod(4))
	}
	return ScopedEffectFor{
		Target:   id,
		Mods:     mods,
		Duration: g.PinnedTo(game.IndefiniteDuration(), id),
		Label:    "Tezzeret, Betrayer of Flesh — becomes an artifact creature",
	}.Apply(ctx)
}
