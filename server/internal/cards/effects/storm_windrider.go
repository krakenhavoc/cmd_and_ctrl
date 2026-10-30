package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Storm, Windrider (#4601, Marvel Super Heroes) — Legendary Creature —
// Mutant Hero {1}{G}{W}{W}, 4/4:
//
//	"Flying
//	 Creatures with flying can't attack you or block creatures you control.
//	 Whenever you cast a spell that targets one or more creatures, those
//	 creatures gain flying until end of turn."
//
// The attack half is CreaturesMatchingCantAttackYou: the player alone, not
// their planeswalkers, judged on the attacker's flying at declaration. The
// block half is Chaosphere's blocker-side rule (CantBlockAttackers) with the
// attackers scoped to this controller's creatures. Both read flying as the
// layers leave it, so a creature that loses flying can attack and block as
// usual, and Storm's own flying does not exempt her from her own text.
//
// The cast trigger records the creatures the spell targets as it is cast,
// each as the object it was (CR 400.7, as Orvar, the All-Form does), because
// the spell may be gone by resolution. A creature that has left the
// battlefield or come back as a new object gains nothing (CR 608.2b). The
// grant is "until end of turn" on each; "those creatures" are any
// controller's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "06e9287b-56d4-443e-97ba-b5dcdc4fa989",
		Name:            "Storm, Windrider",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{CreaturesMatchingCantAttackYou(
			"Creatures with flying can't attack you.",
			game.PermanentQuery{Types: []string{"creature"}, Keyword: "flying"})},
		BlockRules: []game.BlockRule{
			CantBlockAttackers(OnMatching(HasKeyword("flying")), ControlledBySourceController(),
				"creatures with flying can't block creatures its controller controls"),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return len(stormTargetedCreatures(ev, source, g)) > 0
			},
			Key:   stormWindriderLabel,
			Build: stormRecordTargets,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, ref := range ctx.Payload() {
					if ref.Kind != game.TargetCard {
						continue
					}
					if info, ok := g.PermanentForEffect(ref.Object()); !ok || info.Left {
						continue
					}
					if err := (GrantKeywordUntilEOT{Target: ref.ID, Keywords: []string{"flying"}}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}

const stormWindriderLabel = "Storm, Windrider — those creatures gain flying until end of turn"

// stormTargetedCreatures lists the creatures on the battlefield the spell
// its controller just cast targets, in target order and without repeats.
func stormTargetedCreatures(ev game.Event, source *game.Card, g *game.Game) []uuid.UUID {
	if ev.Actor != source.Controller {
		return nil
	}
	item := g.StackItemForEffect(ev.CardID)
	if item == nil || item.Kind != game.StackItemSpell {
		return nil
	}
	var out []uuid.UUID
	seen := map[uuid.UUID]bool{}
	for _, t := range item.Targets {
		if t.Kind != game.TargetCard || seen[t.ID] {
			continue
		}
		if c, ok := g.LookupCardForEffect(t.ID); ok && c.IsCreature() && onBattlefield(g, t.ID) {
			seen[t.ID] = true
			out = append(out, t.ID)
		}
	}
	return out
}

// stormRecordTargets files those creatures on the item's payload as the
// objects they were when the spell was cast.
func stormRecordTargets(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
	ids := stormTargetedCreatures(ev, source, g)
	if len(ids) == 0 {
		return nil
	}
	item := game.NewTriggeredItem(source, stormWindriderLabel)
	for _, id := range ids {
		if ref, ok := g.PermanentRefForEffect(id); ok {
			item.Payload = append(item.Payload, game.PayloadObjectRef(ref))
		}
	}
	return item
}
