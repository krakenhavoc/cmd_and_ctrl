package effects

import (
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hunted by The Family — Sorcery {5}{U}{U}:
//
//	"Choose up to four target creatures you don't control. For each of
//	 them, that creature's controller faces a villainous choice — That
//	 creature becomes a 1/1 white Human creature and loses all
//	 abilities, or you create a token that's a copy of it."
//
// A villainous choice (CR 701.55, #3143) per target, faced by that
// creature's controller as the spell resolves. Several targets controlled
// by different players are asked one at a time in APNAP order (CR 701.55d,
// "performed for each … in APNAP order"): the targets are sorted by their
// controller's seat order and the choices are CHAINED, each answered body
// starting the next target's question when it is the last time its own
// choice is asked (Facing.Remaining, which is what keeps a Valeyard's
// additional times from starting the next target's question twice). A
// target that is no longer legal when its turn comes — it left, or came
// under the caster's control — is skipped (CR 608.2b), and the rest still
// go ahead.
//
// Option A is a continuous effect with NO stated duration (CR 611.2a), so
// it lasts for as long as the object does: layer 4 sets the card types to
// Creature and the creature type to Human (CR 205.1a: an Artifact Dalek
// stops being an artifact, and a supertype such as Legendary is kept),
// layer 5 makes it white, layer 6 takes away every ability including the
// ones it was given by another effect timestamped earlier (CR 613.6), and
// layer 7b sets its base power and toughness to 1/1, so counters and later
// pumps still apply. Option B copies the creature as it is on the
// battlefield; the controller of this spell creates the token.
//
// No simplifications.
//
// The choice is registered from init, not a var initialiser: its bodies
// start the next target's question, which reads the choice again.
var huntedChoice *VillainousChoice

func newHuntedChoice() *VillainousChoice {
	return NewVillainousChoice("hunted-by-the-family",
		"Hunted by The Family — the creature becomes a 1/1 Human with no abilities, or the caster copies it?",
		VillainousOption{
			Label: "That creature becomes a 1/1 white Human creature and loses all abilities",
			Harm:  2,
			Body: func(ctx *Context, f Facing) error {
				if len(f.Carry) == 0 {
					return nil
				}
				if err := huntedBecomeHuman(ctx, f.Carry[0]); err != nil {
					return err
				}
				return huntedNext(ctx, f)
			},
		},
		VillainousOption{
			Label: "The controller of Hunted by The Family creates a token that's a copy of that creature",
			Harm:  1,
			Body: func(ctx *Context, f Facing) error {
				if len(f.Carry) == 0 {
					return nil
				}
				if err := (CreateTokenCopy{Controller: ctx.Controller(), Copy: f.Carry[0], N: 1}).Apply(ctx); err != nil {
					return err
				}
				return huntedNext(ctx, f)
			},
		})
}

// huntedBecomeHuman is option A's indefinite effect on one creature.
func huntedBecomeHuman(ctx *Context, id uuid.UUID) error {
	if !onBattlefield(ctx.Game, id) {
		return nil
	}
	return ScopedEffectFor{
		Target: id,
		Mods: []game.Mod{
			game.SetTypesMod([]string{"Creature"}, []string{"Human"}),
			game.SetColorsMod("W"),
			game.LoseAllAbilitiesMod(),
			game.SetBasePowerMod(1),
			game.SetBaseToughnessMod(1),
		},
		Duration: ctx.Game.PinnedTo(game.IndefiniteDuration(), id),
		Label:    "Hunted by The Family — becomes a 1/1 white Human and loses all abilities",
	}.Apply(ctx)
}

// huntedNext starts the question for the next target once the last
// question about this one has been answered.
func huntedNext(ctx *Context, f Facing) error {
	if f.Remaining > 0 || len(f.Carry) < 2 {
		return nil
	}
	return huntedFace(ctx, f.Carry[1:])
}

// huntedFace asks the controller of the first creature in `ids` that is
// still a creature an opponent controls, carrying that creature and every
// one after it. Later calls come from an answered question, after the
// spell has left the stack, so the CR 608.2b re-check is the clause itself
// (a creature on the battlefield that the caster does not control) and not
// the stack item's target list; the first call has already narrowed `ids`
// by that list, which is what honours a hexproof gained in response.
func huntedFace(ctx *Context, ids []uuid.UUID) error {
	for i, id := range ids {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || !onBattlefield(ctx.Game, id) || !c.IsCreature() || c.Controller == ctx.Controller() {
			continue
		}
		return FaceVillainousChoice{
			Players: []uuid.UUID{c.Controller},
			Choice:  huntedChoice,
			Carry:   append([]uuid.UUID(nil), ids[i:]...),
		}.Apply(ctx)
	}
	return nil
}

func init() {
	huntedChoice = newHuntedChoice()
	Register(Spec{
		OracleID:     "7d178f1c-faf7-4efd-87a6-201a40c55ec8",
		Name:         "Hunted by The Family",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("up to four target creatures you don't control", OpponentControls()).WithCount(0, 4),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			var ids []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					ids = append(ids, t.ID)
				}
			}
			// APNAP order of the creatures' controllers; targets of
			// one controller keep the order they were chosen in.
			seat := map[uuid.UUID]int{}
			for i, p := range apnapPlayers(ctx.Game) {
				seat[p] = i
			}
			rank := func(id uuid.UUID) int {
				if c, ok := controllerOfTarget(ctx, id); ok {
					return seat[c]
				}
				return len(seat)
			}
			sort.SliceStable(ids, func(a, b int) bool { return rank(ids[a]) < rank(ids[b]) })
			return huntedFace(ctx, ids)
		},
	})
}
