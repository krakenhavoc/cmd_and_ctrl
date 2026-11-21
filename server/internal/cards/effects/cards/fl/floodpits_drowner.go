package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Floodpits Drowner — Creature — Merfolk {1}{U}, 2/1 (Duskmourn):
//
//	"Flash
//	 Vigilance
//	 When this creature enters, tap target creature an opponent controls
//	 and put a stun counter on it.
//	 {1}{U}, {T}: Shuffle this creature and target creature with a stun
//	 counter on it into their owners' libraries."
//
// The enters trigger is Alchemax Slayer-Bots' (tapAndStunTheTarget). The
// activation targets any creature with a stun counter, whoever controls it.
// Both go to their OWNERS' libraries as one simultaneous exit, then each
// owner shuffles (once, when both are the same player). Illegal target as it
// resolves: none of it happens and the Drowner stays (CR 608.2b), as the
// ruling says. A Drowner that has left or come back as a new object is not
// shuffled (CR 400.7).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "43c69d19-c06c-470a-bdba-a0335ff2d655",
		Name:            "Floodpits Drowner",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Floodpits Drowner — tap target creature an opponent controls and put a stun counter on it", effects.TapAndStunTheTarget),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}, {T}: Shuffle this creature and target creature with a stun counter on it into their owners' libraries.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{1}{U}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature with a stun counter on it", effects.HasAStunCounter()),
			Effect:  floodpitsDrownerShuffle,
		}},
	})
}

// floodpitsDrownerShuffle tucks the Drowner and the target into their
// owners' libraries, then each owner shuffles.
func floodpitsDrownerShuffle(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var victim uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			victim = t.ID
			break
		}
	}
	if victim == uuid.Nil {
		return nil
	}
	ids := []uuid.UUID{victim}
	if effects.SourceIsStillThisPermanent(g, item) && item.SourceCardID != victim {
		ids = append([]uuid.UUID{item.SourceCardID}, ids...)
	}
	var owners []uuid.UUID
	for _, id := range ids {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			continue
		}
		seen := false
		for _, o := range owners {
			seen = seen || o == c.Owner
		}
		if !seen {
			owners = append(owners, c.Owner)
		}
	}
	return g.TuckCardsToLibraryKeyedForEffect(ids, game.TuckOptions{},
		game.KeyedTail{Ref: effects.ThenShuffleEachCarried, Carry: owners})
}
