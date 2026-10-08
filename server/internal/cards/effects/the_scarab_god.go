package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Scarab God — Legendary Creature — God {3}{U}{B}, 5/5 (EDHREC
// rank 340):
//
//	"At the beginning of your upkeep, each opponent loses X life and
//	 you scry X, where X is the number of Zombies you control.
//	 {2}{U}{B}: Exile target creature card from a graveyard. Create a
//	 token that's a copy of it, except it's a 4/4 black Zombie.
//	 When The Scarab God dies, return it to its owner's hand at the
//	 beginning of the next end step."
//
// Three abilities:
//
//   - The upkeep trigger reads X live off the battlefield
//     (scarabGodZombieCount) at the moment it fires; a zero count
//     drains nothing and scries nothing, which is a legal (if
//     pointless) upkeep.
//   - The activated ability is Hashaton, Scarab's Fist's token-copy
//     shape: exile the chosen graveyard creature, then
//     CreateTokenCopy with the same "except it's a 4/4 black Zombie"
//     exception (scarabGodZombieException, which REPLACES the copy's
//     creature types with Zombie rather than adding to them, per
//     CR 707.9a — the contrast with eternalize is Hashaton's own file's
//     point and applies here unchanged). "A graveyard" is any
//     player's, unrestricted.
//   - The dies trigger is The Locust God's exact body: a CR 603.7
//     delayed trigger scheduled for the next end step, returning the
//     card from wherever it sits in a graveyard back to its owner's
//     hand — a God reanimated or tucked into the command zone in the
//     meantime is a different object and is left alone, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c75e55f2-fd6c-4816-9d96-21eeb8369aff",
		Name:         "The Scarab God",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("The Scarab God — each opponent loses X life and you scry X", scarabGodUpkeep),
			WhenThisDies("The Scarab God — return it to hand at the next end step",
				func(g *game.Game, item *game.StackItem) error {
					return ScheduleDelayedTrigger{
						Label: "The Scarab God — return to its owner's hand",
						Cards: []uuid.UUID{item.SourceCardID},
						Body:  returnListedGraveyardToHandBody,
					}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{{
			Label: "{2}{U}{B}: Exile target creature card from a graveyard. Create a token " +
				"that's a copy of it, except it's a 4/4 black Zombie.",
			Cost:    ManaCost("{2}{U}{B}"),
			Targets: TargetCardInGraveyard("target creature card from a graveyard", Creature()),
			Effect:  scarabGodReanimateAsZombie,
		}},
	})
}

func scarabGodUpkeep(g *game.Game, item *game.StackItem) error {
	x := scarabGodZombieCount(g, item.Controller)
	if x <= 0 {
		return nil
	}
	if err := eachOpponentLosesLife(g, item, x); err != nil {
		return err
	}
	return Scry{Player: item.Controller, N: x}.Apply(NewContext(g, item))
}

// scarabGodZombieCount is "the number of Zombies you control" — a
// live battlefield read, changeling and layer-4 grants included
// (HasSubtype reads effective subtypes).
func scarabGodZombieCount(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsCreature() && c.HasSubtype("Zombie") {
			n++
		}
	}
	return n
}

func scarabGodReanimateAsZombie(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		id := t.ID
		controller := item.Controller
		return ExileTarget{Target: id, Then: func(ctx *Context, exiled bool) error {
			if !exiled {
				return nil
			}
			return CreateTokenCopy{
				Controller: controller,
				Copy:       id,
				N:          1,
				Except:     scarabGodZombieException,
			}.Apply(ctx)
		}}.Apply(ctx)
	}
	return nil
}

// scarabGodZombieException is "except it's a 4/4 black Zombie" —
// REPLACES the copy's creature types (CR 707.9a), same as Hashaton's
// exception but with no "tapped" clause.
func scarabGodZombieException(t *game.Card) {
	t.Power = 4
	t.Toughness = 4
	t.VariableToughness = false
	t.SetCopyExceptionColors("B")
	t.TypeLine = retypedTypeLine(t.TypeLine, "Zombie")
}
