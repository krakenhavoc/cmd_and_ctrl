package cardsa

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Adrestia — Legendary Artifact — Vehicle {3}, 4/3 (Assassin's Creed, #5948):
//
//	"Islandwalk
//	 Whenever Adrestia attacks, if an Assassin crewed it this turn, draw a
//	 card. Adrestia becomes an Assassin in addition to its other types until
//	 end of turn.
//	 Crew 1"
//
// One triggered ability with an intervening "if" (CR 603.4): it triggers only
// when an Assassin crewed this Vehicle this turn, and the check is made again
// as it resolves; both sentences are under it, so with no Assassin crew the
// Vehicle neither draws nor becomes an Assassin. The crew record names the
// creatures that paid for this object's crew cost (CR 702.122c), and each is
// read as it is now, or as it last existed if it has left the battlefield
// (CR 608.2h), so a crewer that died after crewing still counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "94779ca4-c202-4163-9c77-1036921a34bc",
		Name:            "Adrestia",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"islandwalk"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    effects.CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Adrestia"),
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack,
				func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
					return effects.Self(ev, source, ch, g) && adrestiaAssassinCrewed(g, game.ObjectRef{ID: source.InstanceID, Epoch: source.ObjectEpoch}, source.TextSubtype("Assassin"))
				},
				"Adrestia — draw a card; it becomes an Assassin until end of turn",
				adrestiaAttacks),
		},
	})
}

// adrestiaAssassinCrewed is "an Assassin crewed it this turn" for the Vehicle
// object `ref`.
func adrestiaAssassinCrewed(g *game.Game, ref game.ObjectRef, assassin string) bool {
	for _, r := range g.CrewRecordsForEffect(ref, false) {
		if c, ok := g.LookupCardForEffect(r.Creature.ID); ok && c.ObjectEpoch == r.Creature.Epoch && c.HasSubtype(assassin) {
			return true
		}
		if info, ok := g.LastKnownPermanentForEffect(r.Creature.ID); ok && info.Epoch == r.Creature.Epoch &&
			slices.Contains(info.Characteristic.Subtypes, assassin) {
			return true
		}
	}
	return false
}

func adrestiaAttacks(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok || !adrestiaAssassinCrewed(g, ref, ctx.TextSubtype("Assassin")) {
		return nil
	}
	if err := (effects.DrawCards{Player: ctx.Controller(), N: 1}).Apply(ctx); err != nil {
		return err
	}
	return effects.UntilEndOfTurn(ctx, ref.ID, nil, "Adrestia — an Assassin in addition to its other types", game.AddSubtypesMod(ctx.TextSubtype("Assassin")))
}
