package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Taskmaster, Mercenary Mimic — Legendary Creature — Human Mercenary
// Villain {2}{U}{B}, 3/5 (Marvel Super Heroes, #4455):
//
//	"Photographic Reflexes — At the beginning of your first main phase,
//	 until your next turn, Taskmaster becomes a copy of up to one target
//	 creature on the battlefield or creature card in a graveyard, except
//	 his name is Taskmaster, Mercenary Mimic and he's a legendary Human
//	 Mercenary Villain creature."
//
// Absorbing Man's trigger (BecomeCopy with CopyUntilYourNextTurn, CR 611.2b)
// aimed at Angel of Serenity's two zones, with Lazav's copy of a card in a
// graveyard. The "except" clause is an overwrite, not an "in addition":
// the copy's card types become exactly Creature (SetCardTypes drops the
// subtypes that went with the old types), then he is a legendary Human
// Mercenary Villain under his own name. Power and toughness are the copied
// creature's. "Up to one", so no target changes nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dcd97730-b793-4c0c-b158-625ac3656b4a",
		Name:         "Taskmaster, Mercenary Mimic",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				AtYourPrecombatMain("Taskmaster — becomes a copy of up to one target creature on the battlefield or creature card in a graveyard until your next turn", mshA8TaskmasterCopy),
				mshA8TaskmasterTargets()),
		},
	})
}

// mshA8TaskmasterTargets is "up to one target creature on the battlefield
// or creature card in a graveyard".
func mshA8TaskmasterTargets() *game.TargetSpec {
	spec := TargetCreature("up to one target creature on the battlefield or creature card in a graveyard").WithCount(0, 1)
	spec.Zones = []game.ZoneKind{game.ZoneBattlefield, game.ZoneGraveyard}
	return spec
}

// mshA8TaskmasterCopy is the trigger's resolution.
func mshA8TaskmasterCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	targets := ctx.LegalTargets()
	if len(targets) == 0 {
		return nil
	}
	return BecomeCopy{
		Targets:  []uuid.UUID{ctx.Source()},
		Of:       targets[0].ID,
		Duration: CopyUntilYourNextTurn,
		Except:   mshA8TaskmasterExcept,
		Label:    "Taskmaster — becomes a copy until your next turn",
	}.Apply(ctx)
}

// mshA8TaskmasterExcept is the "except" clause: his name, and a legendary
// Human Mercenary Villain creature.
func mshA8TaskmasterExcept(v *game.PrintedValues) {
	v.SetName("Taskmaster, Mercenary Mimic")
	v.SetCardTypes("Creature")
	v.AddSupertype("Legendary")
	v.AddSubtype("Human")
	v.AddSubtype("Mercenary")
	v.AddSubtype("Villain")
}
