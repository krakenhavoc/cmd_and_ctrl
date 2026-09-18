package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Undead Butler — Creature — Zombie {1}{B}, 1/2 (EDHREC rank 4336):
//
//	"When this creature enters, mill three cards.
//	 When this creature dies, you may exile it. When you do, return
//	 target creature card from your graveyard to your hand."
//
// Two mana that fills your graveyard and then buys the best thing it
// found. The Butler is played in reanimator and aristocrats decks for
// the same reason: the mill is the setup and the death trigger is the
// payoff, and the deck was going to sacrifice it anyway.
//
// The second trigger is the interesting one and it is a CR 603.12
// REFLEXIVE trigger — "when you do" is a second triggered ability
// created by the first one while it resolves, not a rider on it. That
// shape decides three things:
//
//   - The "you may" belongs to the PARENT. You choose whether to exile
//     the Butler; once you have, the return is mandatory.
//   - The exile is a real cost of the rebuy, and the Butler is the
//     thing exiled. A deck that wanted the Butler back cannot have
//     both.
//   - The reflexive trigger's TARGET is chosen when IT goes on the
//     stack (CR 603.3d), which is after the Butler has already left
//     the graveyard — so the Butler can never return itself, with no
//     "another" clause needed to say so.
//
// The dies trigger fires from the graveyard, where the Butler now is,
// so exiling it is a plain exile of the source card. A Butler that
// something else moved out of the graveyard in response is not there
// to exile, and then the reflexive trigger never happens — which is
// why the reflexive is applied only after the exile is confirmed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "426fa8e8-b0f9-4b6c-b22b-04ec2073187a",
		Name:         "Undead Butler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Undead Butler — mill three cards", Do(MillCards{N: 3})),
			Optional(
				WhenThisDies("Undead Butler — exile it", b41UndeadButlerExileThenRebuy),
				"Undead Butler — exile it to return a creature card from your graveyard to your hand?"),
		},
	})
}

// b41UndeadButlerExileThenRebuy is the dies trigger's body: exile the
// Butler out of the graveyard, and only if that actually happened
// create the CR 603.12 reflexive trigger that returns a creature card.
//
// "Only if that actually happened" is the reflexive contract: nothing
// in the engine re-checks the condition, so the `if` is the card's. A
// Butler already moved out of the graveyard in response is not
// exiled, and the rebuy does not happen.
func b41UndeadButlerExileThenRebuy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (ExileTarget{Target: item.SourceCardID}).Apply(ctx); err != nil {
		return err
	}
	z := g.FindCardZoneForEffect(item.SourceCardID)
	if z == nil || z.Kind != game.ZoneExile {
		return nil
	}
	return ReflexiveTrigger{
		Label:   "Undead Butler — return target creature card from your graveyard to your hand",
		Targets: TargetCardInGraveyard("target creature card in your graveyard", YouOwn(), Creature()),
		Effect:  returnFirstLegalGraveyardTargetToHand,
	}.Apply(ctx)
}
