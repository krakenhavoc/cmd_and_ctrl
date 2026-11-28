package cardsbl

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blood Oath — Instant {3}{R} (Mercadian Masques, S82 OLD, #5619):
//
//	"Choose a card type. Target opponent reveals their hand. Blood Oath deals 3
//	 damage to that player for each card of the chosen type revealed this way.
//	 (Artifact, battle, creature, enchantment, instant, kindred, land,
//	 planeswalker, and sorcery are card types.)"
//
// The caster chooses the type first, as the spell resolves, from the nine card
// types the reminder lists. The target opponent then reveals their whole hand to
// the table (CR 701.20a), the cards of the chosen type are counted in it, and the
// damage is 3 for each, dealt as one instance. A target that is illegal on
// resolution fizzles the spell before the choice (CR 608.2b). A card with
// several types (an artifact creature) counts for either.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "877c34ca-6eff-42ce-b4d7-8a2071d0290a",
		Name:         "Blood Oath",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 3)),
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			victim := effects.TargetedPlayer(ctx)
			if victim == uuid.Nil {
				return nil
			}
			return effects.PickOption{
				Player:   ctx.Controller(),
				Question: "Blood Oath — choose a card type",
				Options:  bloodOathTypes,
				ThenKey:  bloodOathChosen,
				Carry:    []uuid.UUID{victim},
			}.Apply(ctx)
		},
	})
}

var bloodOathTypes = []game.ChoiceOption{
	{Label: "Artifact"}, {Label: "Battle"}, {Label: "Creature"}, {Label: "Enchantment"}, {Label: "Instant"},
	{Label: "Kindred"}, {Label: "Land"}, {Label: "Planeswalker"}, {Label: "Sorcery"},
}

// bloodOathChosen reveals the target's hand and deals the damage. An on-disk
// key: never renamed, never reused.
var bloodOathChosen = effects.OptionPickThen("option-pick/blood-oath-card-type", func(ctx *effects.Context, r game.OptionPicked) error {
	if len(r.Carry) == 0 || r.Index < 0 || r.Index >= len(bloodOathTypes) {
		return nil
	}
	victim := r.Carry[0]
	p := ctx.Game.PlayerByIDForEffect(victim)
	if p == nil || p.Hand == nil {
		return nil
	}
	ctx.Game.RevealHandForEffect(victim)
	cardType := strings.ToLower(bloodOathTypes[r.Index].Label)
	n := 0
	for _, c := range p.Hand.Cards {
		if c.HasCardType(cardType) {
			n++
		}
	}
	if n == 0 {
		return nil
	}
	return effects.DealDamage{Source: ctx.Source(), Target: victim, Amount: 3 * n}.Apply(ctx)
})
