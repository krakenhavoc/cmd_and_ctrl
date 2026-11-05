package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Best of Both Worlds — Enchantment — Saga {4}{W}{U} (Star Trek
// Commander, #4974):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Assimilate target nonartifact creature an opponent controls.
//	 (Gain control of it and put a +1/+1 counter on it. It's a Borg
//	 artifact creature and loses all other creature types.)
//	 II — Artifact creatures you control get +2/+2 until end of turn.
//	 III — Exile the Borg assimilated with this Saga, then return it to
//	 the battlefield under its owner's control."
//
// Chapter I is assimilate on a permanent: the Saga's controller gains
// control of the target with no duration (CR 611.2a, layer 2), and in one
// record with no duration it becomes an artifact in addition to its other
// types and its creature types become Borg alone (layer 4, CR 205.1a:
// its other card types, supertypes and non-creature subtypes stay), then
// it gets a +1/+1 counter. A target that is no longer legal as the
// chapter resolves (it became an artifact, changed sides) is not
// assimilated (CR 608.2b). The creature is remembered on the Saga as the
// object it was (Card.LinkedPermanent, CR 607.2e), and only by the Saga
// object that assimilated it.
//
// Chapter II is a pump over the artifact creatures you control as it
// resolves (CR 611.2c), the assimilated Borg among them.
//
// Chapter III finds that Borg through the Saga as it is, or as it last
// existed if it has left in response (CR 608.2h), and flickers it to its
// owner: it returns as a new object (CR 400.7), so the control change,
// the type change and the counter are gone. A Borg that has left the
// battlefield in the meantime, even if it came back, is not the one
// assimilated with this Saga, and nothing happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a53d9c36-a033-450b-a095-7645876c9ef3",
		Name:         "The Best of Both Worlds",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTriggerTargeting(1, "The Best of Both Worlds — I: assimilate target nonartifact creature an opponent controls",
				TargetCreature("target nonartifact creature an opponent controls", Not(Artifact()), OpponentControls()),
				theBestOfBothWorldsAssimilate),
			ChapterTrigger(2, "The Best of Both Worlds — II: artifact creatures you control get +2/+2 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return BoostUntilEOT{Match: And(Artifact(), Creature(), YouControl()), Power: 2, Toughness: 2,
						Label: "The Best of Both Worlds — +2/+2"}.Apply(NewContext(g, item))
				}),
			ChapterTrigger(3, "The Best of Both Worlds — III: exile the Borg assimilated with this Saga, then return it under its owner's control",
				theBestOfBothWorldsReturn),
		},
	})
}

// theBestOfBothWorldsAssimilate is chapter I: assimilate the target and
// link it to the Saga.
func theBestOfBothWorldsAssimilate(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	borg := game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}
	const label = "The Best of Both Worlds — assimilated"
	if err := (GainControl{Target: id, Duration: game.IndefiniteDuration(), Label: label}).Apply(ctx); err != nil {
		return err
	}
	if err := (ScopedEffectFor{
		Target:   id,
		Mods:     []game.Mod{game.AddTypesMod("Artifact"), game.SetCreatureTypesMod(ctx.TextSubtypes("Borg")...)},
		Duration: game.IndefiniteDuration(),
		Label:    label + ": a Borg artifact creature",
	}).Apply(ctx); err != nil {
		return err
	}
	if sourceIsStillThisPermanent(g, item) {
		g.SetLinkedPermanentForEffect(item.SourceCardID, borg)
	}
	return AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
}

// theBestOfBothWorldsReturn is chapter III.
func theBestOfBothWorldsReturn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	borg := g.LinkedPermanentForEffect(ref)
	if borg == nil {
		return nil
	}
	return Flicker{Target: borg.InstanceID}.Apply(ctx)
}
