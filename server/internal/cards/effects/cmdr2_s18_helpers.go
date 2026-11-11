package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_s18_helpers.go — helpers for the Bloomburrow-legends batch (#5368).

// cmdr2S18TappedTreasuresForDamage is "create that many tapped Treasure
// tokens": one per point of the combat damage that triggered the ability
// (Prosperous Bandit). Its footprint is CreateTokensForEach's: a token step
// entering tapped changes only the new token's own state.
type cmdr2S18TappedTreasuresForDamage struct {
	// Count is how many: the combat damage that triggered the ability.
	Count func(ctx *Context) int
}

func (s cmdr2S18TappedTreasuresForDamage) Apply(ctx *Context) error {
	if s.Count == nil {
		return nil
	}
	n := s.Count(ctx)
	if n <= 0 {
		return nil
	}
	return CreateTokenAdvanced{Spec: Token(TreasureToken()).EntersTapped(), N: n}.Apply(ctx)
}

func (s cmdr2S18TappedTreasuresForDamage) footprints() ([]game.FootprintStep, bool) {
	if s.Count == nil {
		return nil, false
	}
	key, manual, keywords := game.FootprintTokenOf(TreasureToken())
	return []game.FootprintStep{
		condStep(game.FootprintCondBoard, 0),
		{Kind: game.FootprintCreateToken, N: 1, TokenKey: key, TokenManual: manual, TokenKeywords: keywords},
	}, true
}

// cmdr2S18OpponentCastsNoncreature is "an opponent casts a noncreature
// spell".
func cmdr2S18OpponentCastsNoncreature(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	return ByAnOpponent(ev, source, lki, g) && aPlayerCastsANoncreatureSpell(ev, source, lki, g)
}

// cmdr2S18KillerServiceOffer is Killer Service's end-step body: "you may pay
// {2} and sacrifice a token. If you do, create a 4/4 green Rhino Warrior".
// Both halves are costs, so it is offered only when the controller has a
// token to sacrifice; the sacrifice is then forced once the {2} is paid.
func cmdr2S18KillerServiceOffer(g *game.Game, item *game.StackItem) error {
	has := false
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsToken() {
			has = true
			break
		}
	}
	if !has {
		return nil
	}
	return MayPay{
		Chooser:  item.Controller,
		Cost:     "{2}",
		Question: "Killer Service — pay {2} and sacrifice a token to create a 4/4 Rhino Warrior?",
		Then:     cmdr2S18KillerServicePaid,
	}.Apply(NewContext(g, item))
}

var cmdr2S18KillerServicePaid = PayUnlessBranches("pay-unless/cmdr2-s18-killer-service",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return mustSacrificeOne(ctx.Game, ctx.Item, "Killer Service — sacrifice a token", IsTokenPredicate(), cmdr2S18KillerServiceRhino)
	}, nil)

var cmdr2S18KillerServiceRhino = RegisterSacrificedOne("prompt-run/cmdr2-s18-killer-service-rhino",
	func(ctx *Context, _ uuid.UUID) error {
		return CreateToken{Controller: ctx.Controller(), Template: TokenCard("4/4 green Rhino Warrior"), N: 1}.Apply(ctx)
	})

// cmdr2S18GreatestArtifactManaValue is the greatest mana value among the
// artifacts `player` controls, zero with none.
func cmdr2S18GreatestArtifactManaValue(g *game.Game, player uuid.UUID) int {
	best := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsArtifact() {
			if mv := c.ManaValue(); mv > best {
				best = mv
			}
		}
	}
	return best
}

// cmdr2S18BelloAppliesTo is Bello's set: "during your turn, each
// non-Equipment artifact and non-Aura enchantment you control with mana
// value 4 or greater".
func cmdr2S18BelloAppliesTo(target *game.Card, g *game.Game, source *game.Card) bool {
	if target.Controller != source.Controller || !IsYourTurn(g, source.Controller) || target.ManaValue() < 4 {
		return false
	}
	if target.IsArtifact() && !target.HasSubtype("Equipment") {
		return true
	}
	return target.IsEnchantment() && !target.IsAura()
}

// cmdr2S18BelloStatics are Bello's layered effects: the 4/4 Elemental
// creature type (layer 4) and base power and toughness (layer 7b), the
// indestructible and haste keywords (layer 6), and the granted combat
// damage draw (layer 6, a keyed grant).
func cmdr2S18BelloStatics(grantKey string) []game.StaticAbility {
	return []game.StaticAbility{
		{
			Label:     "Bello — each such artifact and enchantment is a creature in addition to its other types",
			Layer:     game.Layer4Type,
			AppliesTo: cmdr2S18BelloAppliesTo,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				if !eotHasType(c.Types, "Creature") {
					c.Types = append(append([]string(nil), c.Types...), "Creature")
				}
				if !eotHasType(c.Subtypes, "Elemental") {
					c.Subtypes = append(append([]string(nil), c.Subtypes...), "Elemental")
				}
			},
		},
		{
			Label:     "Bello — each such permanent has indestructible and haste",
			Layer:     game.Layer6Ability,
			AppliesTo: cmdr2S18BelloAppliesTo,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Abilities = game.AppendKeywordAbility(c.Abilities, "indestructible")
				c.Abilities = game.AppendKeywordAbility(c.Abilities, "haste")
			},
		},
		GrantAbilities(cmdr2S18BelloAppliesTo, grantKey),
		{
			Label:     "Bello — each such permanent is 4/4",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7B_Set,
			AppliesTo: cmdr2S18BelloAppliesTo,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power = 4
				c.Toughness = 4
			},
		},
	}
}

// cmdr2S18IsFood matches a Food permanent.
func cmdr2S18IsFood(_ *game.Game, _ uuid.UUID, c game.Card) bool {
	return hasSubtype(c, "Food")
}

// cmdr2S18ArtifactsYouControlOnYourTurn is Gourmand's Talent's set:
// "during your turn, artifacts you control".
func cmdr2S18ArtifactsYouControlOnYourTurn(target *game.Card, g *game.Game, source *game.Card) bool {
	return target.Controller == source.Controller && target.IsArtifact() && IsYourTurn(g, source.Controller)
}

// cmdr2S18FoodSacrificeAbility is "{2}, {T}, Sacrifice this artifact: You
// gain 3 life." as the Food rider a granted artifact carries.
func cmdr2S18FoodSacrificeAbility() ActivatedAbility {
	return ActivatedAbility{
		Label:   "{2}, {T}, Sacrifice this artifact: You gain 3 life.",
		Purpose: game.Purpose{Answers: game.AnswerValue},
		Cost:    Plus(ManaCost("{2}"), TapCost(), SacrificeThis()),
		Effect: func(g *game.Game, item *game.StackItem) error {
			return GainLife{Player: item.Controller, Amount: 3}.Apply(NewContext(g, item))
		},
	}
}
