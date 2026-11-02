package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Arachne, Psionic Weaver — {2}{W} Legendary Creature — Spider Human
// Hero 3/3 (Marvel's Spider-Man promos, #4460):
//
//	"Web-slinging {W} (You may cast this spell for {W} if you also
//	 return a tapped creature you control to its owner's hand.)
//	 As Arachne enters, look at an opponent's hand, then choose a card
//	 type other than creature.
//	 Spells of the chosen type cost {1} more to cast."
//
// Web-slinging is Amazing Spider-Girl's alternative cost at {W}.
//
// The as-enters clause (CR 614.12) is two questions in a row, queued
// from the AsEnters hook the way True-Name Nemesis's and Cloud Key's
// are: first the opponent whose hand Arachne's controller looks at
// (only that player learns the cards, Peek's "look at", not a reveal),
// then, from the continuation, the card type. The type is any card type
// a spell can have (CR 205.2a) except creature; the ones that never
// exist in a game of Commander (conspiracy, dungeon, phenomenon, plane,
// scheme, vanguard) are left off the list, since no spell of them can be
// cast. With no opponent left to look at, the type is still chosen.
//
// The increase is Cloud Key's predicate the other way: every player's
// spells of the chosen type, Arachne's controller's included. Until the
// type is chosen nothing costs more.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "28d500ea-720f-4332-891e-404fb03acb5c",
		Name:         "Arachne, Psionic Weaver",
		Completeness: CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{{
			Key:          "web_slinging",
			Label:        "Web-slinging {W}",
			ManaCost:     "{W}",
			ReturnToHand: PermanentYouControl("a tapped creature you control", Creature(), b751Tapped()),
			PayLabel:     "a tapped creature you control",
		}},
		AsEnters: arachneAsEnters,
		CostModifiers: []game.CostModifier{
			CostsMore(1, "Spells of the chosen type cost {1} more to cast.", arachneChosenTypeSpell()),
		},
	})
}

// arachneCardTypes are the card types Arachne offers, in prompt order.
var arachneCardTypes = []string{"instant", "sorcery", "artifact", "enchantment", "planeswalker", "battle", "kindred", "land"}

const arachneName = "Arachne, Psionic Weaver"

// arachneAsEnters asks for the opponent whose hand is looked at.
func arachneAsEnters(card *game.Card, ctx *Context) error {
	if ctx.Game.QueueChoosePlayerForEffect(game.ChoosePlayerPrompt{
		Chooser:  card.Controller,
		Among:    Opponents.seats(ctx, card.Controller),
		Source:   card.InstanceID,
		Question: arachneName + " — look at an opponent's hand",
		ThenKey:  arachneLooked,
		Carry:    []uuid.UUID{card.InstanceID},
	}) != uuid.Nil {
		return nil
	}
	arachneChooseType(ctx.Game, card.Controller, card.InstanceID)
	return nil
}

// arachneLooked is the opponent's answer: the controller looks at that
// hand, then chooses the card type. An on-disk key: never renamed, never
// reused.
var arachneLooked = game.RegisterOptionPickThen("option-pick/arachne-psionic-weaver-look", func(g *game.Game, r game.OptionPicked) error {
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
	arachneChooseType(g, r.Chooser, r.Carry[0])
	return nil
})

// arachneChooseType queues the stored card-type choice.
func arachneChooseType(g *game.Game, chooser, source uuid.UUID) {
	g.QueueChooseOptionAsEntersForEffect(chooser, source,
		arachneName+" — choose a card type other than creature", arachneCardTypes)
}

// arachneChosenTypeSpell matches a spell of the type Arachne's
// controller chose.
func arachneChosenTypeSpell() CostPredicate {
	return func(q game.CostQuery) bool {
		chosen := ChosenOptionOf(q.Game, q.Source.InstanceID)
		return chosen != "" && q.Card.HasCardType(chosen)
	}
}
