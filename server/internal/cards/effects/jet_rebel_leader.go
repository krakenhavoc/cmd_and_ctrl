package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jet, Rebel Leader — Legendary Creature — Human Rebel Ally {3}{W}, 3/4
// (Avatar: The Last Airbender Eternal, #5075):
//
//	"Whenever Jet attacks, look at the top five cards of your library.
//	 You may put a creature card with mana value 3 or less from among
//	 them onto the battlefield tapped and attacking. Put the rest on the
//	 bottom of your library in a random order."
//
// "Look at" (CR 701.20e): only Jet's controller learns the cards. The
// pick is an up-to-one choice over the creature cards of mana value 3 or
// less among them, a keyed prompt that carries the looked-at cards. Once
// it is answered the rest go to the bottom in a random order first, then
// the chosen card enters tapped and attacking through Raph & Mikey's
// shared put (putAttackingYourChoice): its controller chooses which
// player, planeswalker or battle it attacks, with no question when
// there is one choice. It was never declared as an attacker, so its own
// "whenever this attacks" does not trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b9ca8ead-c267-4170-a6c8-04498330c82b",
		Name:         "Jet, Rebel Leader",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Jet, Rebel Leader — look at the top five cards; you may put a creature card with mana value 3 or less onto the battlefield tapped and attacking",
				avB30dJetLook),
		},
	})
}

const avB30dJetName = "Jet, Rebel Leader"

// avB30dJetLook looks at the top five and asks for the pick.
func avB30dJetLook(g *game.Game, item *game.StackItem) error {
	you, source := item.Controller, item.SourceCardID
	looked := g.LookAtTopOfLibraryForEffect(you, 5)
	if len(looked) == 0 {
		return nil
	}
	var eligible []uuid.UUID
	for _, id := range looked {
		c, ok := g.LookupCardForEffect(id)
		if !ok || !c.IsCreature() {
			continue
		}
		if mv, readable := g.ManaValueForEffect(c); readable && mv <= 3 {
			eligible = append(eligible, id)
		}
	}
	if len(eligible) == 0 {
		return avB30dJetFinish(g, you, source, looked, uuid.Nil)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  you,
		Source:   source,
		Question: "Jet — put a creature card with mana value 3 or less onto the battlefield tapped and attacking?",
		Cards:    eligible,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneLibrary,
		ThenKey:  avB30dJetPicked,
		Carry:    looked,
	})
	return nil
}

// avB30dJetPicked is the pick answered: Carry is the looked-at cards. An
// on-disk key, never renamed.
var avB30dJetPicked = game.RegisterChooseCardsThen("avatar-b30/jet-rebel-leader-put-attacking", func(g *game.Game, r game.CardsChosen) error {
	if !r.Answered {
		return nil
	}
	var chosen uuid.UUID
	if len(r.Picked) > 0 {
		chosen = r.Picked[0]
	}
	return avB30dJetFinish(g, r.Chooser, r.Source, r.Carry, chosen)
})

// avB30dJetFinish bottoms the rest in a random order, then puts the
// chosen card, if any and still in the library, onto the battlefield
// tapped and attacking.
func avB30dJetFinish(g *game.Game, you, source uuid.UUID, looked []uuid.UUID, chosen uuid.UUID) error {
	var rest []uuid.UUID
	for _, id := range cardsStillInALibrary(g, looked) {
		if id != chosen {
			rest = append(rest, id)
		}
	}
	if err := g.PutOnBottomInRandomOrderForEffect(you, game.ZoneLibrary, rest); err != nil {
		return err
	}
	if chosen == uuid.Nil {
		return nil
	}
	return putAttackingYourChoice(keyedContext(g, you, source), chosen, game.ZoneLibrary, avB30dJetName)
}
