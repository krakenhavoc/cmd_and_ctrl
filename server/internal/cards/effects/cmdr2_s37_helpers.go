package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_s37_helpers.go — batch 37 (Manifest and face-up payoffs).

const cmdr2S37WhisperwoodDiesManifest = "whisperwood-elemental/dies-manifest"

// cmdr2S37WhisperwoodGrant is Whisperwood Elemental's sacrifice ability:
// every face-up nontoken creature its controller controls, as it resolves,
// gains the dies-manifest trigger until end of turn.
func cmdr2S37WhisperwoodGrant(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != item.Controller || !c.IsCreature() || c.IsToken() || c.FaceDown {
			continue
		}
		if err := (GrantAbilitiesFor{
			Target: c.InstanceID,
			Keys:   []string{cmdr2S37WhisperwoodDiesManifest},
			Label:  "Whisperwood Elemental — it manifests the top card of your library when it dies",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// cmdr2S37CreaturesYouControl counts the creatures `player` controls, read
// through the layered view.
func cmdr2S37CreaturesYouControl(g *game.Game, player uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsCreature() {
			n++
		}
	}
	return n
}

// cmdr2S37TurnedUpCreatureFights is Temur War Shaman's "you may have it
// fight target creature you don't control": the permanent that was turned
// face up (the trigger's event card) fights the announced target, when it
// is still a creature on the battlefield (CR 603.4's recheck) and the
// target is still legal (CR 701.14b).
func cmdr2S37TurnedUpCreatureFights(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ctx := NewContext(g, item)
	it, ok := battlefieldCardForEffect(g, item.Trigger.Event.CardID)
	if !ok || !it.IsCreature() {
		return nil
	}
	target, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return b10Fight(ctx, it.InstanceID, target)
}

// cmdr2S37GainLifePerCreature is Mastery of the Unseen's "you gain 1 life
// for each creature you control", counted as the ability resolves.
func cmdr2S37GainLifePerCreature(g *game.Game, item *game.StackItem) error {
	return GainLife{Amount: cmdr2S37CreaturesYouControl(g, item.Controller)}.Apply(NewContext(g, item))
}
