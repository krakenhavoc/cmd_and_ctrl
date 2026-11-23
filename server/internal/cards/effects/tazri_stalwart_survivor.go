package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tazri, Stalwart Survivor — Legendary Creature — Human Warrior {2}{W},
// 3/3 (March of the Machine: The Aftermath, #5476):
//
//	"Each creature you control has '{T}: Add one mana of any of this
//	 creature's colors. Spend this mana only to activate an ability of a
//	 creature. Activate only if this creature has another activated
//	 ability.'
//	 {W}{U}{B}{R}{G}, {T}: Mill five cards. Put all creature cards with
//	 activated abilities that aren't mana abilities from among the milled
//	 cards into your hand."
//
// The first line is Cryptolith Rite's layer-6 grant of a mana bundle
// (ADR 0093) with three differences, all declared on the one granted row.
// The colours are read from the creature that has the ability as it is
// activated, through the layers (ProducedFunc, a pipe of its colours in
// WUBRG order): a colourless creature produces no mana, a one-colour
// creature produces that colour. The spend restriction is Shang-Chi's
// (the mana pays only for activating abilities of creature sources,
// CR 106.6). "Activate only if this creature has another activated
// ability" is the row's Condition (CR 602.5b): it counts the creature's
// activated abilities, non-mana and mana alike (a mana ability is an
// activated ability, CR 605.1a), and requires one besides this grant's
// own row, so a vanilla creature can't tap for mana while Tazri herself
// can through her mill ability.
//
// The mill is Cantankerous Keepers': the clause after it runs on the cards
// that really reached the graveyard, takes every creature card among them
// with an activated ability that isn't a mana ability (read from the card's
// own abilities: a CR 602 list, which excludes mana abilities and loyalty
// and triggered abilities only as the engine reads them), with no prompt.
//
// No simplification.
const tazriGrant = "tazri-stalwart-survivor/tap-for-colors"

func init() {
	Register(Spec{
		OracleID:     "13a3e61d-c9ff-447d-9016-09a183b7af1d",
		Name:         "Tazri, Stalwart Survivor",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: tazriGrant,
			Mana: []ManaAbility{{
				Cost:         ManaAbilityCost{Tap: true},
				Label:        "{T}: Add one mana of any of this creature's colors. Spend this mana only to activate an ability of a creature.",
				ProducedFunc: TazriProduced,
				Restrictions: []string{game.ManaRestrictAnyOf(
					[]string{game.ManaRestrictActivate, game.ManaRestrictType("Creature")},
				)},
				Condition: tazriHasAnotherActivatedAbility,
			}},
			Text: "{T}: Add one mana of any of this creature's colors. Spend this mana only to activate an ability of a creature. Activate only if this creature has another activated ability.",
		}},
		Static: []game.StaticAbility{
			Labelled(GrantAbilities(CreaturesYouControl, tazriGrant),
				"Each creature you control has \"{T}: Add one mana of any of this creature's colors. Spend this mana only to activate an ability of a creature. Activate only if this creature has another activated ability.\""),
		},
		Activated: []ActivatedAbility{{
			Label:   "{W}{U}{B}{R}{G}, {T}: Mill five cards. Put all creature cards with activated abilities that aren't mana abilities from among the milled cards into your hand.",
			Cost:    game.AbilityCost{Tap: true, Mana: "{W}{U}{B}{R}{G}"},
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return MillToZone{N: 5, ThenTail: ctx.ItemTail(tazriMilled)}.Apply(ctx)
			},
		}},
	})
}

// tazriProduced is "one mana of any of this creature's colors": a pipe of
// the colours the creature has now, or nothing when it is colourless.
func TazriProduced(g *game.Game, _, source uuid.UUID) string {
	c, ok := g.LookupCardForEffect(source)
	if !ok {
		return ""
	}
	have := map[string]bool{}
	for _, col := range c.EffectiveColors() {
		have[strings.ToUpper(col)] = true
	}
	var cols []string
	for _, col := range []string{"W", "U", "B", "R", "G"} {
		if have[col] {
			cols = append(cols, col)
		}
	}
	if len(cols) == 0 {
		return ""
	}
	return "{" + strings.Join(cols, "|") + "}"
}

// tazriHasAnotherActivatedAbility is the row's "Activate only if this
// creature has another activated ability": it has an activated ability
// (mana or not) besides the granted row itself.
func tazriHasAnotherActivatedAbility(g *game.Game, _, source uuid.UUID) bool {
	c, ok := g.LookupCardForEffect(source)
	if !ok {
		return false
	}
	return len(game.ActivatedAbilitiesForCard(c))+len(game.ManaAbilitiesOf(&c)) >= 2
}

// tazriMilled puts every milled creature card with a non-mana activated
// ability into its owner's hand.
var tazriMilled = CardTail("tazri-stalwart-survivor/after-mill", func(ctx *Context, o game.TailOutcome) error {
	return HobA11AllMilledToHand(ctx, o, func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.IsCreature() && len(game.ActivatedAbilitiesForCard(c)) > 0
	})
})
