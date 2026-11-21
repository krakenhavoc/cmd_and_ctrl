package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Keen-Eyed Curator — Creature — Raccoon Scout {G}{G}, 3/3 (Bloomburrow):
//
//	"As long as there are four or more card types among cards exiled with this
//	 creature, it gets +4/+4 and has trample.
//	 {1}: Exile target card from a graveyard."
//
// The exile is linked (ExileTarget.Linked, ADR 0155 §8): the card lands with
// Card.ExiledWith naming this Curator object, so a card another effect exiled
// is not one of them, and a Curator that left and came back is a new object
// with none. The static counts the distinct card types printed on the linked
// cards (a card with several types, an artifact creature say, counts each), and
// both of its halves read the same count live, declaring DependsOnExile so
// the layers are re-read when a linked card leaves exile. A token exiled this
// way ceases to exist and counts for nothing.
//
// No simplification.
func init() {
	cond := func(target *game.Card, g *game.Game, source *game.Card) bool {
		return effects.SelfOnly(target, g, source) && keenEyedCuratorTypes(g, source) >= 4
	}
	pump := game.StaticAbility{
		Label:     "As long as there are four or more card types among cards exiled with this creature, it gets +4/+4.",
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7C_Modify,
		AppliesTo: cond,
		Apply:     effects.PumpBy(4, 4),

		DependsOnExile: true,
	}
	trample := effects.KeywordGrant(cond, "trample")
	trample.Label = "As long as there are four or more card types among cards exiled with this creature, it has trample."
	trample.DependsOnExile = true
	effects.Register(effects.Spec{
		OracleID:     "46ce3698-e4e6-44e9-8b99-2b12a703e91e",
		Name:         "Keen-Eyed Curator",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{pump, trample},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}: Exile target card from a graveyard.",
			Cost:    effects.ManaCost("{1}"),
			Targets: effects.TargetCardInGraveyard("target card from a graveyard"),
			Effect:  effects.AvB7ExileTargetLinked,
		}},
	})
}

// keenEyedCuratorTypes is the number of distinct card types among the cards
// exiled with `source`, read off their printed type lines.
func keenEyedCuratorTypes(g *game.Game, source *game.Card) int {
	if g == nil || source == nil {
		return 0
	}
	seen := map[string]bool{}
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef{ID: source.InstanceID, Epoch: source.ObjectEpoch}) {
		if c.InstanceID == source.InstanceID {
			continue
		}
		_, types, _ := game.ParseTypeLine(c.TypeLine)
		for _, t := range types {
			seen[t] = true
		}
	}
	return len(seen)
}
