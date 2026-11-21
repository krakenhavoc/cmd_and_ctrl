package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stay Hidden, Stay Silent — Enchantment — Aura {1}{U}:
//
//	"Enchant creature
//	 When this Aura enters, tap enchanted creature.
//	 Enchanted creature doesn't untap during its controller's untap
//	 step.
//	 {4}{U}{U}: Shuffle enchanted creature into its owner's library,
//	 then manifest dread. Activate only as a sorcery."
//
// Claustrophobia's tap and untap lock, plus the activated ability. The
// creature is shuffled into ITS OWNER's library (a stolen creature goes
// home), the owner's library is shuffled, and then the Aura's
// controller manifests dread from their own library. The shuffle and
// the manifest are the tuck's continuation, since the tuck can pause on
// a commander's command-zone question. If the Aura has left by the time
// the ability resolves, the creature it last enchanted is the one
// shuffled (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "266561cc-d014-48cc-ba5d-fa72f3a91736",
		Name:                  "Stay Hidden, Stay Silent",
		Completeness:          effects.CompletenessFull,
		Targets:               effects.EnchantCreature(),
		UntapStepRestrictions: []game.UntapStepRestriction{effects.EnchantedDoesntUntap()},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Stay Hidden, Stay Silent — tap enchanted creature", effects.TapEnchantedCreatureOnEntry),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{4}{U}{U}: Shuffle enchanted creature into its owner's library, then manifest dread. Activate only as a sorcery.",
			Cost:         effects.ManaCost("{4}{U}{U}"),
			SorcerySpeed: true,
			Effect:       stayHiddenShuffleAndManifest,
		}},
	})
}

func stayHiddenShuffleAndManifest(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	manifest := func(g *game.Game) error { return effects.ManifestDread{}.Apply(effects.NewContext(g, item)) }
	ref, ok := ctx.SourceRef()
	if !ok {
		return manifest(g)
	}
	aura, ok := g.PermanentForEffect(ref)
	if !ok || aura.AttachedTo.Kind != game.TargetCard {
		return manifest(g)
	}
	victim, ok := g.LookupCardForEffect(aura.AttachedTo.ID)
	if !ok {
		return manifest(g)
	}
	return g.TuckCardsToLibraryKeyedForEffect([]uuid.UUID{victim.InstanceID}, game.TuckOptions{},
		effects.WithCarry(ctx.ItemTail(stayHiddenShuffleThenManifest), victim.Owner))
}

// stayHiddenShuffleThenManifest is "its owner shuffles, then you manifest
// dread", after the tuck, keyed (#4253 slice 10a): the owner is Carry[0].
var stayHiddenShuffleThenManifest = effects.CardTail("then/stay-hidden-shuffle-manifest", func(ctx *effects.Context, o game.TailOutcome) error {
	if owner := effects.TailCarry0(o); owner != uuid.Nil {
		if err := ctx.Game.ShuffleLibraryForEffect(owner); err != nil {
			return err
		}
	}
	return effects.ManifestDread{}.Apply(ctx)
})
