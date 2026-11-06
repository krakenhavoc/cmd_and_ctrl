package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Koh, the Face Stealer — Legendary Creature — Shapeshifter Spirit
// {4}{B}{B}, 6/6:
//
//	"When Koh enters, exile up to one other target creature.
//	 Whenever another nontoken creature dies, you may exile it.
//	 Pay 1 life: Choose a creature card exiled with Koh.
//	 Koh has all activated and triggered abilities of the last chosen
//	 card."
//
// Both exiles are linked to this Koh object (ExileTarget.Linked, CR
// 607.2a), and the choice reads back only the cards THEY exiled
// (Game.ExiledWithRowForEffect over Koh's two rows): a card exiled with
// Koh by an ability Koh borrowed is linked to that ability, not to these.
// The dies trigger exiles the card only while it is still the object
// that died (the ruling: a card that left the graveyard in response is
// not exiled and can't be chosen).
//
// The choice is made as the ability resolves and stored on Koh as the
// object the card is (Card.ChosenCard); each choice replaces the last.
// The last line is ADR 0155's borrow with its 2026-11-05 amendment
// (HasActivatedAndTriggeredAbilitiesOf over LastChosenExiledCardLender):
// Koh has the chosen card's activated abilities, mana abilities included,
// and its triggered abilities, triggered keywords and triggered mana
// abilities included, until another card is chosen or the chosen card
// leaves exile. An ability already on the stack is unaffected by either
// (CR 113.7a). Koh leaving the battlefield is a new object with no
// exiled cards and no choice (CR 400.7).
//
// No simplification.
func init() {
	const clause = "Koh has all activated and triggered abilities of the last chosen card."
	borrow := HasActivatedAndTriggeredAbilitiesOf(game.AbilityBorrow{
		Label:   clause,
		Lenders: LastChosenExiledCardLender(),
	})
	borrow.DependsOnExile = true
	Register(Spec{
		OracleID:     "2bbb33e3-7d03-457c-92bb-b0c943615fa1",
		Name:         "Koh, the Face Stealer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters(kohEntersKey, exileLinkedTargetsEffect),
				Another(UpToOneTargetCreature("up to one other target creature")),
			),
			Optional(On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b17AnotherNontokenCreatureDied(ev, source, g)
			}, kohDiesKey, kohExileTheDead), "Koh, the Face Stealer — exile the creature card that died?"),
		},
		Activated: []ActivatedAbility{{
			Label:  "Pay 1 life: Choose a creature card exiled with Koh.",
			Cost:   PayLife(1),
			Effect: kohChooseExiledCreatureCard,
			// The choice itself changes no board: the rows it lends
			// declare their own answers.
			Purpose: game.Purpose{Answers: game.AnswerValue},
		}},
		Static: []game.StaticAbility{borrow},
	})
}

const (
	kohEntersKey = "Koh, the Face Stealer — exile up to one other target creature"
	kohDiesKey   = "Koh, the Face Stealer — exile the creature that died"
)

// kohExileTheDead is "you may exile it": the card that died, linked to
// Koh, while it is still that object in its graveyard (CR 400.7).
func kohExileTheDead(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	dead := item.Trigger.Event.CardID
	if !diedCardStillInGraveyard(ctx, dead) {
		return nil
	}
	return ExileTarget{Target: dead, Linked: true}.Apply(ctx)
}

// kohChooseExiledCreatureCard is "Choose a creature card exiled with
// Koh": among the creature cards Koh's own two exiles put there, still
// linked to this Koh object. With none, nothing is chosen and the last
// choice stands.
func kohChooseExiledCreatureCard(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	src, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	link := game.PermissionCardRef(src)
	var cands []*game.Card
	for _, key := range []string{kohEntersKey, kohDiesKey} {
		for _, c := range g.ExiledWithRowForEffect(link, game.AbilitySlotTriggered, key) {
			if c.IsCreature() {
				cands = append(cands, c)
			}
		}
	}
	g.QueueChooseExiledCardForEffect(ctx.Controller(), src.ID, "Koh, the Face Stealer — choose a creature card exiled with Koh", cands)
	return nil
}
