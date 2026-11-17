package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Angelic Guardian — Creature — Angel {4}{W}{W}, 5/5 (M19 Gift Pack,
// #5524):
//
//	"Flying
//	 Whenever one or more creatures you control attack, they gain
//	 indestructible until end of turn."
//
// Ancestor Dragon's "whenever you attack": one trigger per attack
// declaration of yours (OncePerBatch over attackDeclaredByYou, CR 603.2c),
// whether or not the Guardian is among the attackers. "They" are that
// declaration's attackers, read from the declaration's own batch as the
// trigger resolves, as Tyvar the Bellicose reads its Elves: a creature
// that only entered attacking was not declared (CR 508.4) and gets
// nothing, and a creature attacking under another player's declaration
// is not one of yours. One removed from combat before the trigger
// resolves still gains it, since the text names the creatures, not
// "attacking creatures". One that left the battlefield after it attacked
// is a new object if it comes back (CR 400.7) and gains nothing.
//
// No simplification.
const angelicGuardianLabel = "Angelic Guardian — the attacking creatures gain indestructible until end of turn"

func init() {
	Register(Spec{
		OracleID:        "2f959c67-2b0b-4ba8-a8aa-90c25a3806ec",
		Name:            "Angelic Guardian",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller)
			}, angelicGuardianLabel, angelicGuardianIndestructible)),
		},
	})
}

// angelicGuardianIndestructible gives indestructible until end of turn to
// each creature the trigger's controller declared as an attacker in the
// declaration that triggered it, while it is still that object on the
// battlefield.
func angelicGuardianIndestructible(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ctx := NewContext(g, item)
	batch := item.Trigger.Event.Batch
	var attackers []uuid.UUID
	declared := map[uuid.UUID]bool{}
	moved := map[uuid.UUID]bool{}
	for _, ev := range g.Events {
		switch {
		case ev.Kind == game.EventAttack && ev.Batch == batch && ev.Actor == item.Controller && ev.CardID != uuid.Nil:
			if !declared[ev.CardID] {
				declared[ev.CardID] = true
				attackers = append(attackers, ev.CardID)
			}
		case ev.Kind == game.EventZoneMove && declared[ev.CardID]:
			// It left the battlefield after it attacked: whatever is
			// there now is a new object (CR 400.7).
			moved[ev.CardID] = true
		}
	}
	for _, id := range attackers {
		if moved[id] || !onBattlefield(g, id) {
			continue
		}
		grant := GrantKeywordUntilEOT{Target: id, Keywords: []string{"indestructible"}, Label: "Angelic Guardian — indestructible"}
		if err := grant.Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
