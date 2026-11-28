package cardsmi

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Musician — Creature — Human Wizard {2}{U}, 1/3 (Ice Age, S82 OLD batch 838,
// #5636):
//
//	"Cumulative upkeep {1}
//	 {T}: Put a music counter on target creature. If it doesn't have "At the
//	 beginning of your upkeep, destroy this creature unless you pay {1} for each
//	 music counter on it," it gains that ability."
//
// The music counter goes on first, then the ability is given if the creature
// does not already have it. "Has it" is read off the creature's triggered
// abilities as the layers leave them, so a creature given it by one Musician is
// not given it again by a second (CR 702.24b's instance counting is not
// involved: this is not a keyword). The gift is the ADR 0093 bundle pinned to
// the creature with no duration (CR 611.2a), as Balduvian Shaman's is: the
// creature's controller is asked at their own upkeep, the counters are the
// creature's own, and a decline destroys the creature (CR 701.8).
//
// No simplification.
const (
	musicianGrant = "musician/music-upkeep"
	musicianText  = "At the beginning of your upkeep, destroy this creature unless you pay {1} for each music counter on it"
)

func init() {
	effects.Register(effects.Spec{
		OracleID:     "0d6e9468-adb2-4feb-87cf-9113db8717b1",
		Name:         "Musician",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.CumulativeUpkeep("Musician — cumulative upkeep {1}", "{1}")},
		Grants: []effects.AbilityGrant{{
			Key:       musicianGrant,
			Triggered: []game.TriggeredAbility{effects.AtYourUpkeep(musicianText, musicianUpkeep)},
			Text:      musicianText + ".",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Put a music counter on target creature. If it doesn't have the music upkeep ability, it gains that ability.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Effect:  musicianPut,
		}},
	})
}

// musicianPut is the activated ability's effect: the counter, then the gift if
// the creature lacks the upkeep ability. A target that left does nothing
// (CR 608.2b).
func musicianPut(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard || !effects.OnBattlefield(g, t.ID) {
			continue
		}
		if err := (effects.AddCounter{Target: t.ID, Kind: "music", N: 1}).Apply(ctx); err != nil {
			return err
		}
		c, ok := g.LookupCardForEffect(t.ID)
		if !ok || musicianHasUpkeep(c) {
			return nil
		}
		return effects.GrantAbilitiesFor{
			Target:   t.ID,
			Keys:     []string{musicianGrant},
			Duration: g.PinnedTo(game.IndefiniteDuration(), t.ID),
			Label:    "Musician — gains the music upkeep ability",
		}.Apply(ctx)
	}
	return nil
}

// musicianHasUpkeep reports whether the creature already has the quoted ability.
func musicianHasUpkeep(c game.Card) bool {
	for _, t := range game.TriggersForCard(c) {
		if t.Key == musicianText {
			return true
		}
	}
	return false
}

// musicianUpkeep is the granted ability's effect: with N music counters on the
// creature, its controller pays {N} or it is destroyed. The creature is the
// trigger's source (ADR 0093 decision 4), so the counters and the destruction
// are its own.
func musicianUpkeep(g *game.Game, item *game.StackItem) error {
	n := effects.CounterCountOn(g, item.SourceCardID, "music")
	if n <= 0 || !effects.OnBattlefield(g, item.SourceCardID) {
		return nil
	}
	cost := "{" + strconv.Itoa(n) + "}"
	return effects.UpkeepPayUnless{
		Chooser:  item.Controller,
		Cost:     cost,
		Question: "Musician — pay " + cost + " or destroy this creature?",
		Else:     musicianDestroy,
	}.Apply(effects.NewContext(g, item))
}

// musicianDestroy is the decline: the creature is destroyed (it may be
// regenerated, CR 701.19). Keyed (#4264); an on-disk identity.
var musicianDestroy = effects.PayUnlessElse("pay-unless/musician-destroy", func(ctx *effects.Context, r game.PayUnlessAnswer) error {
	if r.Source == uuid.Nil {
		return nil
	}
	return effects.DestroyTarget{Target: r.Source}.Apply(ctx)
})
