package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Altanak, the Thrice-Called — Legendary Creature — Insect Beast {5}{G}{G},
// 9/9 (Duskmourn: House of Horror, #4466):
//
//	"Trample
//	 Whenever Altanak becomes the target of a spell or ability an opponent
//	 controls, draw a card.
//	 {1}{G}, Discard this card: Return target land card from your graveyard
//	 to the battlefield tapped."
//
// Trample is a printed keyword. The trigger is Surrak, Elusive Hunter's watch
// on EventBecomesTarget, narrowed to Altanak itself and to an opponent's spell
// or ability: it fires as the targets are chosen, so it resolves before the
// spell or ability that did the targeting, even if that is countered or leaves
// the stack (the ruling). The engine emits one event per target slot (CR
// 115.3), so a spell or ability that targets Altanak twice would emit two;
// b24SelfTargetedByAnOpponentOnce lets only the first of them trigger, as the
// ruling says. The draw is not optional.
//
// The third ability is a hand-zone activated ability (Boseiju's channel
// shape): the discard is a cost paid as the ability is announced (CR 602.2b)
// and is not refunded if the ability is countered, and "target land card from
// your graveyard" is chosen then, so the discarded Altanak is not a legal
// pick (it is not a land). The land returns tapped under your control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6591f243-f94c-4a83-bb16-1fb838d9eb14",
		Name:            "Altanak, the Thrice-Called",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			On(game.EventBecomesTarget, b24SelfTargetedByAnOpponentOnce,
				"Altanak, the Thrice-Called — draw a card", Do(DrawCards{N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{G}, Discard this card: Return target land card from your graveyard to the battlefield tapped.",
			Cost:    game.AbilityCost{Mana: "{1}{G}", DiscardSelf: true},
			Zones:   []game.ZoneKind{game.ZoneHand},
			Targets: TargetCardInGraveyard("target land card from your graveyard", Land(), YouOwn()),
			Effect:  returnFirstLegalGraveyardTargetToBattlefieldTapped,
		}},
	})
}
