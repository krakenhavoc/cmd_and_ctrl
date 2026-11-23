package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82BroB11AssemblyTrio reports whether `controller` controls a creature
// named `a` and a creature named `b`: the "If you control creatures named
// Power Plant Worker and Tower Worker" test of the three Assembly-Workers
// (The Brothers' War). The two names are two different creatures, so a
// card named for one of them (the asking Worker is the third) never
// satisfies both. Read live, as an ability resolves (Mine Worker, Power
// Plant Worker) or as mana is produced (Tower Worker), through the layered
// view: a creature that is not currently a creature does not count.
func S82BroB11AssemblyTrio(g *game.Game, controller uuid.UUID, a, b string) bool {
	return JmpB1YouControlACreatureNamed(g, controller, a) && JmpB1YouControlACreatureNamed(g, controller, b)
}

// s82BroB11BecomesArtifactCreature is "target noncreature artifact becomes
// a P/T artifact creature" (Alloy Animist; Levitating Statue's shape with
// a target): the permanent's card types become artifact and creature
// (CR 205.1a), so a type it has beyond those (a Land on an artifact land)
// goes, with the subtypes only that type allowed (CR 205.1a, 205.3d);
// the artifact subtypes it has stay. It is ONE continuous effect from
// the resolving ability (CR 611.2a) pinned to the object (CR 400.7) for
// `duration`: Creature in layer 4 and the base power and toughness in
// layer 7b.
func S82BroB11BecomesArtifactCreature(ctx *Context, id uuid.UUID, power, toughness int, duration game.Duration, label string) error {
	c, ok := ctx.Game.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	var drop []string
	for _, t := range []string{"Land", "Enchantment", "Kindred", "Battle", "Planeswalker"} {
		if c.HasCardType(strings.ToLower(t)) {
			drop = append(drop, t)
		}
	}
	mods := []game.Mod{}
	if len(drop) > 0 {
		mods = append(mods, game.RemoveTypesMod(drop...))
	}
	mods = append(mods, game.AddTypesMod("Creature"))
	return ScopedEffectFor{
		Target:   id,
		Mods:     append(mods, game.SetBasePTMods(power, toughness)...),
		Duration: duration,
		Label:    label,
	}.Apply(ctx)
}
