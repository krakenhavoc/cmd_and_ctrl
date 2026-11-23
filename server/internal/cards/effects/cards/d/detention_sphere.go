package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Detention Sphere — Enchantment {1}{W}{U}:
//
//	"When this enchantment enters, you may exile target nonland
//	 permanent not named Detention Sphere and all other permanents with
//	 the same name as that permanent.
//	 When this enchantment leaves the battlefield, return the exiled
//	 cards to the battlefield under their owner's control."
//
// Fiend Hunter's two linked triggers (CR 607.2a): "the exiled cards" are
// the ones the entry trigger exiled, read off the event log by its key,
// and a Sphere that leaves before the entry trigger resolves exiles them
// for good (the printed ruling). The target and every other permanent
// with its name (any controller, CR 201.2) leave in one simultaneous
// exile, read as the trigger resolves; a face-down permanent has no name
// (CR 708.2a) and so matches nothing but itself. A token exiled this way
// ceases to exist. The "may" is answered as the trigger resolves
// (CR 603.5), after the target is chosen.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0a0b376d-7824-4ba5-8bc9-bc0be12800c5",
		Name:         "Detention Sphere",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches:        []game.EventKind{game.EventETB},
				AppliesTo:      effects.Self,
				OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Detention Sphere — exile target nonland permanent and all other permanents with the same name?"},
				Targets:        effects.TargetPermanent("target nonland permanent not named Detention Sphere", effects.Nonland(), DetentionSphereNotNamedSphere),
				Key:            detentionSphereExileLabel,
				Effect:         detentionSphereExile,
			},
			effects.On(game.EventLTB, effects.Self, "Detention Sphere — return the exiled cards to the battlefield under their owner's control",
				effects.B41ReturnCardsExiledWithToTheBattlefield(detentionSphereExileLabel)),
		},
	})
}

// detentionSphereExileLabel is the entry trigger's key, which is what the
// leave trigger reads the linked exile by.
const detentionSphereExileLabel = "Detention Sphere — exile target nonland permanent not named Detention Sphere and all other permanents with the same name"

// detentionSphereNotNamedSphere is "not named Detention Sphere".
func DetentionSphereNotNamedSphere(_ *game.Game, _ uuid.UUID, c game.Card) bool {
	return c.FaceDownIsPermanent() || !game.HasName(c, "Detention Sphere")
}

// detentionSphereExile exiles the chosen permanent and every other
// permanent that shares its name, together.
func detentionSphereExile(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	target, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	c, ok := g.LookupCardForEffect(target)
	if !ok {
		return nil
	}
	name, faceDown := c.Name, c.FaceDownIsPermanent()
	return effects.ExileAllMatching{
		Match: func(_ *game.Game, _ uuid.UUID, o game.Card) bool {
			if o.InstanceID == target {
				return true
			}
			return !faceDown && !o.FaceDownIsPermanent() && game.HasName(o, name)
		},
	}.Apply(ctx)
}
