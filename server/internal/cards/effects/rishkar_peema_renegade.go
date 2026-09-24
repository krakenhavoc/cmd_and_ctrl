package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rishkar, Peema Renegade — Legendary Creature — Elf Druid {2}{G}, 2/2:
//
//	"When Rishkar enters, put a +1/+1 counter on each of up to two
//	 target creatures.
//	 Each creature you control with a counter on it has "{T}: Add
//	 {G}.""
//
// The grant reads the recipient's counters of ANY kind, re-read on
// every layer pass (a counter change bumps the layer version), so a
// creature that loses its last counter loses the ability with it.
// ADR 0093's layer-6 grant; the ETB is an ordinary targeted trigger.
//
// No simplification.
const rishkarGrant = "rishkar-peema-renegade/tap-for-green"

func init() {
	Register(Spec{
		OracleID:     "761021ce-4559-464e-aa03-85c2fe78e267",
		Name:         "Rishkar, Peema Renegade",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Rishkar, Peema Renegade — put a +1/+1 counter on each of up to two target creatures",
					rishkarPutCounters),
				TargetCreature("up to two target creatures").WithCount(0, 2)),
		},
		Grants: []AbilityGrant{TapForManaGrant(rishkarGrant, "{G}", "Add {G}", "{T}: Add {G}.")},
		Static: []game.StaticAbility{GrantAbilities(creatureYouControlWithACounter, rishkarGrant)},
	})
}

// rishkarPutCounters is the ETB's resolution: one +1/+1 counter on each
// chosen creature still legal (CR 608.2b).
func rishkarPutCounters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			ids = append(ids, t.ID)
		}
	}
	for _, id := range ids {
		if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// creatureYouControlWithACounter is "each creature you control with a
// counter on it" — any kind, any number.
func creatureYouControlWithACounter(target *game.Card, _ *game.Game, source *game.Card) bool {
	if !target.IsCreature() || target.Controller != source.Controller {
		return false
	}
	for _, n := range target.Counters {
		if n > 0 {
			return true
		}
	}
	return false
}
