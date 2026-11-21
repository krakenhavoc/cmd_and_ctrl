package cardsa

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Anointed Peacekeeper — Creature — Human Cleric {2}{W}, 3/3 (Dominaria United):
//
//	"Vigilance
//	 As this creature enters, look at an opponent's hand, then choose any
//	 card name.
//	 Spells your opponents cast with the chosen name cost {2} more to cast.
//	 Activated abilities of sources with the chosen name cost {2} more to
//	 activate unless they're mana abilities."
//
// Arachne's two questions in a row, queued from the AsEnters hook: first the
// opponent whose hand the controller looks at (only they learn the cards, a
// "look at", not a reveal), then, from the continuation, the card name, stored
// on the permanent (ChosenName, CR 614.12). With no opponent left to look at,
// the name is still chosen. The name is any card name, lands included (the
// ability tax then reaches a named land's non-mana abilities).
//
// Two cost modifiers read that name, matched on any face (CR 201.2b): an
// increase on the spells opponents cast with it, and an activation increase on
// every non-mana activated ability of a source with it, whoever controls the
// source. Until the name is chosen nothing costs more.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a67f78be-8fbe-414a-9574-bb11ac50e1a5",
		Name:            "Anointed Peacekeeper",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		AsEnters:        anointedPeacekeeperAsEnters,
		CostModifiers: []game.CostModifier{
			effects.CostsMore(2, "Spells your opponents cast with the chosen name cost {2} more to cast.",
				effects.OpponentsSpell(), anointedPeacekeeperNamedSpell()),
			effects.ActivationCostsMore(2, "Activated abilities of sources with the chosen name cost {2} more to activate unless they're mana abilities.",
				anointedPeacekeeperNamedNonManaAbility()),
		},
	})
}

const anointedPeacekeeperName = "Anointed Peacekeeper"

// anointedPeacekeeperAsEnters asks for the opponent whose hand is looked at.
func anointedPeacekeeperAsEnters(card *game.Card, ctx *effects.Context) error {
	if ctx.Game.QueueChoosePlayerForEffect(game.ChoosePlayerPrompt{
		Chooser:  card.Controller,
		Among:    effects.Opponents.Seats(ctx, card.Controller),
		Source:   card.InstanceID,
		Question: anointedPeacekeeperName + " — look at an opponent's hand",
		ThenKey:  anointedPeacekeeperLooked,
		Carry:    []uuid.UUID{card.InstanceID},
	}) != uuid.Nil {
		return nil
	}
	ctx.Game.QueueCardNameChoiceForEffect(card.Controller, card.InstanceID, anointedPeacekeeperName+" — choose any card name")
	return nil
}

// anointedPeacekeeperLooked is the opponent's answer: the controller looks at
// that hand, then names a card. An on-disk key: never renamed, never reused.
var anointedPeacekeeperLooked = game.RegisterOptionPickThen("option-pick/anointed-peacekeeper-look", func(g *game.Game, r game.OptionPicked) error {
	if len(r.Carry) == 0 {
		return nil
	}
	if r.Option != nil {
		if p := g.PlayerByIDForEffect(r.Option.Player); p != nil && p.Hand != nil {
			for i := range p.Hand.Cards {
				p.Hand.Cards[i].AddKnower(r.Chooser)
			}
		}
	}
	g.QueueCardNameChoiceForEffect(r.Chooser, r.Carry[0], anointedPeacekeeperName+" — choose any card name")
	return nil
})

// anointedPeacekeeperNamedSpell is a spell with the chosen name.
func anointedPeacekeeperNamedSpell() effects.CostPredicate {
	return func(q game.CostQuery) bool {
		named := strings.TrimSpace(q.Source.ChosenName)
		return named != "" && q.Ability == nil && q.SpecialAction == nil && game.CardNameMatches(q.Card, named)
	}
}

// anointedPeacekeeperNamedNonManaAbility is an activated ability, not a mana
// ability, of a source with the chosen name.
func anointedPeacekeeperNamedNonManaAbility() effects.CostPredicate {
	return func(q game.CostQuery) bool {
		named := strings.TrimSpace(q.Source.ChosenName)
		return named != "" && q.Ability != nil && !q.Ability.Mana && game.CardNameMatches(q.Card, named)
	}
}
