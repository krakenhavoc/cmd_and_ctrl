package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Teferi's Realm — World Enchantment {1}{U}{U}:
//
//	"At the beginning of each player's upkeep, that player chooses
//	 artifact, creature, land, or non-Aura enchantment. All nontoken
//	 permanents of that type phase out. (While they're phased out,
//	 they're treated as though they don't exist. Each one phases in
//	 before its controller untaps during their next untap step.)"
//
// Every player's upkeep; "that player" is the active player the upkeep
// event names, and they choose with a PickOption prompt as the trigger
// resolves. Every nontoken permanent of the chosen type, every
// player's, phases out at once with whatever is attached to it
// (CR 702.26g), and each phases back in during its controller's next
// untap step (CR 702.26a). The Realm phases itself out when it is the
// type chosen — it is an enchantment that is not an Aura.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9851d934-2e07-49c8-b08b-15f96d0f3f0c",
		Name:         "Teferi's Realm",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtEachUpkeep("Teferi's Realm — that player chooses a type; those nontoken permanents phase out", teferisRealmUpkeep),
		},
	})
}

// teferisRealmTypes are the four choices, in printed order, with the
// predicate each one phases out.
var teferisRealmTypes = []struct {
	label string
	match func(c game.Card) bool
}{
	{"Artifact", func(c game.Card) bool { return c.IsArtifact() }},
	{"Creature", func(c game.Card) bool { return c.IsCreature() }},
	{"Land", func(c game.Card) bool { return c.IsLand() }},
	{"Non-Aura enchantment", func(c game.Card) bool { return c.IsEnchantment() && !c.HasSubtype("Aura") }},
}

func teferisRealmUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player := ctx.Trigger().Event.Actor
	if player == uuid.Nil {
		return nil
	}
	options := make([]game.ChoiceOption, 0, len(teferisRealmTypes))
	for _, t := range teferisRealmTypes {
		options = append(options, game.ChoiceOption{Label: t.label})
	}
	return PickOption{
		Player:   player,
		Question: "Teferi's Realm — choose a type. All nontoken permanents of that type phase out.",
		Options:  options,
		Then: func(ctx *Context, index int) error {
			if index < 0 || index >= len(teferisRealmTypes) {
				return nil
			}
			match := teferisRealmTypes[index].match
			var ids []uuid.UUID
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if !c.IsToken() && match(c) {
					ids = append(ids, c.InstanceID)
				}
			}
			return PhaseOut{Targets: ids}.Apply(ctx)
		},
	}.Apply(ctx)
}
