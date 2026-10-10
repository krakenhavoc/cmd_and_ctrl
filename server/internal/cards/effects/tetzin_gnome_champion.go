package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tetzin, Gnome Champion // The Golden-Gear Colossus — a transforming
// legendary artifact creature (#2709, ADR 0137):
//
//	Tetzin, Gnome Champion — Legendary Artifact Creature — Gnome
//	{U}{R}{W}, 2/2
//	  "Whenever Tetzin or another double-faced artifact you control
//	   enters, mill three cards. You may put an artifact card from among
//	   them into your hand.
//	   Craft with six artifacts {4}"
//	The Golden-Gear Colossus — Legendary Artifact Creature — Gnome, 6/6
//	  "Vigilance, trample
//	   Whenever The Golden-Gear Colossus enters or attacks, transform up
//	   to one other target double-faced artifact you control. Create two
//	   1/1 colorless Gnome artifact creature tokens."
//
// "Double-faced" is CR 712.1's: a transforming or modal double-faced
// card, a meld card, or a melded permanent
// (game.IsDoubleFacedForEffect), read off the entering permanent as it
// is now. Tetzin's trigger is Barrowgoyf's mill-then-take with an
// artifact filter. The craft is six artifacts from either zone.
//
// The Colossus may target any double-faced artifact you control but
// itself; only one that can transform does (the ruling: a modal
// double-faced artifact is a legal target that won't), through the
// in-place transform verb. The Gnomes come whether or not a target was
// chosen.
//
// No simplification.
const tetzinGnomeChampionOracleID = "e12b38e5-621b-4b2c-8c04-c15c41d4fc70"

func init() {
	Register(Spec{
		OracleID:     tetzinGnomeChampionOracleID,
		Name:         "Tetzin, Gnome Champion",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, tetzinDoubleFacedArtifactEntered,
				"Tetzin, Gnome Champion — mill three cards, you may put an artifact card from among them into your hand",
				tetzinMill),
		},
		Activated: []ActivatedAbility{
			Craft("Craft with six artifacts {4}", "{4}", CraftWithN(6, "artifact")),
		},
	})

	Register(Spec{
		OracleID:        tetzinGnomeChampionOracleID + "#1",
		Name:            "The Golden-Gear Colossus",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "trample"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEntersOrAttacks("The Golden-Gear Colossus — transform up to one other double-faced artifact, create two Gnomes", goldenGearColossus),
				Another(PermanentYouControl("up to one other target double-faced artifact you control", Artifact(), doubleFaced())).WithCount(0, 1),
			),
		},
	})
}

// doubleFaced matches a double-faced card or permanent (CR 712.1).
func doubleFaced() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return game.IsDoubleFacedForEffect(c) }
}

// tetzinDoubleFacedArtifactEntered is "Tetzin or another double-faced
// artifact you control enters".
func tetzinDoubleFacedArtifactEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.CardID == source.InstanceID {
		return true
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.Controller == source.Controller && c.IsArtifact() && game.IsDoubleFacedForEffect(c)
}

// tetzinMill is the trigger's resolution.
func tetzinMill(g *game.Game, item *game.StackItem) error {
	return MillToZone{
		N: 3,
		Then: func(ctx *Context, milled []uuid.UUID) error {
			return mayTakeOneFromAmongThem(ctx, milled, Artifact(),
				"Tetzin, Gnome Champion — you may put an artifact card from among them into your hand")
		},
	}.Apply(NewContext(g, item))
}

// goldenGearColossus transforms the chosen double-faced artifact, if
// one was chosen and is still legal, then makes two Gnomes.
func goldenGearColossus(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			if err := g.TransformPermanentForEffect(t.ID); err != nil {
				return err
			}
		}
	}
	return CreateToken{Controller: item.Controller, Template: TokenCard("1/1 colorless Gnome artifact"), N: 2}.Apply(ctx)
}
