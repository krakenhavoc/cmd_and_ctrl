package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Borg Cube — {5}{W}{B} Artifact — Spacecraft, 6/6 (Star Trek, #4973):
//
//	"Affinity for artifact creatures
//	 Whenever this Spacecraft enters or attacks, assimilate a creature
//	 card exiled with it. If you can't, exile up to one target creature.
//	 (To assimilate a card, put it onto the battlefield under your
//	 control with a +1/+1 counter. It's a Borg artifact creature and
//	 loses all other creature types.)
//	 Station
//	 12+ | Artifact creatures you control have flying."
//
// Affinity is Mycosynth Golem's self cost modifier over artifact
// creatures. The station body is the ordinary Spacecraft's: an artifact
// creature 6/6 from twelve charge counters (CR 721.2b), and the 12+ line
// a layer-6 grant gated on the same count (CR 721.2a).
//
// The trigger is one ability with two conditions (Sun Titan's shape)
// and Captain Marvel's "up to one target creature", chosen as it is put
// on the stack (CR 603.3d) whether or not it will be needed. On
// resolution it reads the creature cards in exile linked to THIS object
// (CR 607.2a, Card.ExiledWith, ADR 0155 §8): the linked exile is the
// "if you can't" clause below, so a Cube that left and came back is a
// new object with nothing exiled with it (CR 400.7), and a Cube killed
// in response still assimilates through its last object. One such card
// is assimilated (the controller picks when, rarely, there are several)
// through the shared assimilate action (trekAssimilateFrom, batch 11's
// body). When there is none, or the put brings nothing onto the
// battlefield, the ability exiles its target instead, linked to the
// Cube, so the next trigger assimilates that creature if it is a
// creature card in exile.
//
// No simplification.
func init() {
	assimilate := WhenThisEntersOrAttacks(
		"Borg Cube — assimilate a creature card exiled with it; if you can't, exile up to one target creature",
		borgCubeAssimilateOrExile)
	assimilate.Targets = TargetCreature("up to one target creature").WithCount(0, 1)
	Register(Spec{
		OracleID:     "a59b854c-0e04-4581-b9d3-1516263d7ace",
		Name:         "Borg Cube",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			AffinityFor("Affinity for artifact creatures", And(Artifact(), Creature())),
		},
		Activated: []ActivatedAbility{Station()},
		Static:    append(SpacecraftAt(12, 6, 6), borgCubeFlyingGrant()),
		Triggered: []game.TriggeredAbility{assimilate},
	})
}

// borgCubeFlyingGrant is "12+ | Artifact creatures you control have
// flying": a layer-6 grant, live on the charge-counter count.
func borgCubeFlyingGrant() game.StaticAbility {
	grant := b16GrantKeywords(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.IsArtifact() && target.IsCreature() && target.Controller == source.Controller
	}, "flying")
	grant.ActiveWhen = AtChargeCounters(12)
	return grant
}

const borgCubeAssimilatedLabel = "Borg Cube — assimilated: a Borg artifact creature"

// borgCubeAssimilateOrExile is the trigger's resolution.
func borgCubeAssimilateOrExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	exiled := borgCubeCreatureCardsExiledWith(ctx)
	switch len(exiled) {
	case 0:
		return borgCubeExileTarget(ctx)
	case 1:
		return trekAssimilateFrom(ctx, game.ZoneExile, exiled[0], borgCubeAssimilatedLabel, borgCubeAssimilated)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Borg Cube — assimilate a creature card exiled with it",
		Cards:    exiled,
		Min:      1,
		Max:      1,
		ThenKey:  borgCubePick,
		LiveZone: game.ZoneExile,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// borgCubeCreatureCardsExiledWith is "a creature card exiled with it":
// the creature cards in exile linked to the ability's source object,
// read where they sit (printed types; no layer applies in exile).
func borgCubeCreatureCardsExiledWith(ctx *Context) []uuid.UUID {
	src, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	var out []uuid.UUID
	for _, c := range ctx.Game.ExiledWithForEffect(game.PermissionCardRef(src)) {
		if c.IsCreature() {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// borgCubePick is the creature card chosen among several (an on-disk key).
var borgCubePick = askedPick("choose-cards/borg-cube-assimilate", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return borgCubeExileTarget(ctx)
	}
	return trekAssimilateFrom(ctx, game.ZoneExile, r.Picked[0], borgCubeAssimilatedLabel, borgCubeAssimilated)
})

// borgCubeAssimilated is told what the assimilate put onto the
// battlefield; nothing means "if you can't" (an on-disk key).
var borgCubeAssimilated = CardTail("borg-cube/assimilated", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) > 0 {
		return nil
	}
	return borgCubeExileTarget(ctx)
})

// borgCubeExileTarget is "exile up to one target creature", linked to
// the Cube (CR 607.2a) so a later trigger can assimilate it. No target
// chosen, or none still legal, exiles nothing (CR 608.2b).
func borgCubeExileTarget(ctx *Context) error {
	id := firstLegalCardTarget(ctx)
	if id == uuid.Nil {
		return nil
	}
	return ExileTarget{Target: id, Linked: true}.Apply(ctx)
}
