package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_dmu_b12_helpers.go — shared pieces of Dominaria United batch 12
// (S82, #5487): The Raven Man's Bird and Phoenix Chick's return.

// printedRavenManBirdToken is the 1/1 black Bird creature token with
// flying and "This token can't block." that The Raven Man makes, as printed:
// the can't-block clause is a static it always has (the Fungus's shape), so
// it is a catalog template rather than a tokens_table row. Its slug is an
// on-disk identity: never renamed, never reused.
func printedRavenManBirdToken() tokenTemplate {
	return tokenTemplate{
		Slug: "bird-black-flying-cant-block",
		Card: game.Card{
			Name:      "Bird",
			TypeLine:  "Token Creature — Bird",
			Power:     1,
			Toughness: 1,
			Colors:    []string{"B"},
			Keywords:  []string{"flying"},
		},
		Static: []game.StaticAbility{RestrictSelf(game.CantBlock)},
		Text:   "This token can't block.",
	}
}

// RavenManBirdToken is that Bird's card value for CreateToken.
func RavenManBirdToken() game.Card { return tokenFromCatalog(printedRavenManBirdToken) }

// phoenixChickReturn puts `card`, still in its owner's graveyard, onto the
// battlefield under its controller's control tapped and attacking with a
// +1/+1 counter on it (CR 506.3c, CR 508.4): the controller picks what it
// attacks when more than one player, planeswalker or battle could be. The
// counter is an entry counter (CR 614.1c). It is putAttackingYourChoice's
// shape with the counter this one card prints.
func phoenixChickReturn(ctx *Context, card uuid.UUID) error {
	me := ctx.Controller()
	refs := ctx.Game.AttackTargetsForEffect(me)
	if len(refs) == 0 {
		return nil
	}
	if len(refs) == 1 {
		return phoenixChickPut(ctx.Game, me, card, refs[0].ID)
	}
	options := make([]game.ChoiceOption, len(refs))
	carry := []uuid.UUID{card}
	for i, r := range refs {
		options[i] = game.ChoiceOption{Label: "Attack " + attackTargetName(ctx.Game, r.ID)}
		if r.Kind != game.AttackTargetPlayer {
			options[i].Cards = []uuid.UUID{r.ID}
		}
		carry = append(carry, r.ID)
	}
	return PickOption{
		Question: "Phoenix Chick — what does it attack?",
		Options:  options,
		ThenKey:  phoenixChickAttackPicked,
		Carry:    carry,
	}.Apply(ctx)
}

func phoenixChickPut(g *game.Game, me, card, target uuid.UUID) error {
	if z := g.FindCardZoneForEffect(card); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return g.PutOntoBattlefieldTogetherKeyedForEffect(
		[]game.BatchEntry{{CardID: card, From: game.ZoneGraveyard}},
		game.ZoneEntryOptions{Controller: me, Tapped: true, Attacking: target,
			EntersWithCounters: map[string]int{game.CounterPlusOne: 1}}, game.KeyedTail{})
}

// phoenixChickAttackPicked is the keyed answer to "what does it attack?":
// the card and then the attack targets, in the order the options list them.
// An on-disk key: never renamed, never reused.
var phoenixChickAttackPicked = OptionPickThen("option-pick/phoenix-chick-attack", func(ctx *Context, r game.OptionPicked) error {
	if len(r.Carry) < 2 {
		return nil
	}
	card, targets := r.Carry[0], r.Carry[1:]
	index := r.Index
	if index < 0 || index >= len(targets) {
		index = 0
	}
	return phoenixChickPut(ctx.Game, ctx.Controller(), card, targets[index])
})
