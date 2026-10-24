package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Goliath Daydreamer — Creature — Giant Wizard {2}{R}{R}, 4/4:
//
//	"Whenever you cast an instant or sorcery spell from your hand,
//	 exile that card with a dream counter on it instead of putting it
//	 into your graveyard as it resolves.
//	 Whenever this creature attacks, you may cast a spell from among
//	 cards you own in exile with dream counters on them without paying
//	 its mana cost."
//
// The first ability is a trigger on the cast (EventCast from the
// caster's hand, Narset Transcendent's "from your hand" gate) that
// marks the spell on the stack: game.MarkSpellExilesWithCounterForEffect
// makes "exile with a dream counter" one of the spell's resolution
// exits (ADR 0107 §3, amended 2026-10-13), so it replaces the move to
// the graveyard only when the spell RESOLVES. A countered, fizzled or
// copied spell is an ordinary one. The trigger resolves above the
// spell, so the mark is on before the spell resolves; if the spell has
// already left the stack the trigger does nothing.
//
// The second ability is Crabomination's grant: one permission over
// every card the controller owns in exile with a dream counter, priced
// {0} so X is 0 (CR 107.3b), spells only, for one spell (CastsLeft), at
// instant speed (CR 608.2g), closing on the controller's next priority
// pass with the cards left in exile. A card cast from exile leaves it
// with the counter gone (CR 400.7); one not cast keeps its counter for
// the next attack.
//
// No simplification.
func init() {
	mark := WheneverYouCast(Or(Instant(), Sorcery()),
		"Goliath Daydreamer — exile that card with a dream counter on it instead of putting it into your graveyard as it resolves",
		func(g *game.Game, item *game.StackItem) error {
			ev := NewContext(g, item).Trigger().Event
			if ev.OldZone != game.ZoneHand || ev.CardID == uuid.Nil {
				return nil
			}
			g.MarkSpellExilesWithCounterForEffect(ev.CardID, game.CounterDream)
			return nil
		})
	// The cast must come from the hand: gate the trigger itself, so
	// a flashback or exile cast never puts it on the stack.
	inner := mark.AppliesTo
	mark.AppliesTo = func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
		return ev.OldZone == game.ZoneHand && inner(ev, source, ch, g)
	}
	Register(Spec{
		OracleID:     "9207facd-fbf9-4faa-ac44-6c0bcac81551",
		Name:         "Goliath Daydreamer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			mark,
			WheneverThisAttacks("Goliath Daydreamer — you may cast a spell from among your exiled cards with dream counters", goliathDaydreamerAttack),
		},
	})
}

func goliathDaydreamerAttack(g *game.Game, item *game.StackItem) error {
	controller := item.Controller
	var cards []game.Card
	if g.Exile == nil {
		return nil
	}
	for _, c := range g.Exile.Cards {
		if c.Owner == controller && c.Counters[game.CounterDream] > 0 && !c.IsLand() {
			cards = append(cards, c)
		}
	}
	if len(cards) == 0 {
		return nil
	}
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:      controller,
		Zone:        game.ZoneExile,
		Duration:    g.UntilEndOfTurnDuration(),
		Cost:        "{0}",
		Timing:      game.TimingFlash,
		CastOnly:    true,
		CastsLeft:   1,
		LapseOnPass: game.LapseStaysInExile,
		Source:      item.SourceCardID,
		SourceName:  "Goliath Daydreamer",
		Label:       "Goliath Daydreamer — cast a spell with a dream counter without paying its mana cost",
	}, cards)
	// #4250: cast during the resolution (CR 608.2g); the one permission is
	// spent by the one cast.
	ids := make([]uuid.UUID, 0, len(cards))
	for _, c := range cards {
		ids = append(ids, c.InstanceID)
	}
	offerResolutionCastOver(g, controller, item.SourceCardID, "Goliath Daydreamer",
		game.MayCastKeywordFreeCast, declineLeaveInExile, ids...)
	return nil
}
